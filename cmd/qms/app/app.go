package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/benbjohnson/clock"
	"github.com/grafana/dskit/modules"
	"github.com/grafana/dskit/services"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/Blinkuu/qms/internal/core/domain/cloud"
	"github.com/Blinkuu/qms/internal/core/services/alloc"
	"github.com/Blinkuu/qms/internal/core/services/memberlist"
	"github.com/Blinkuu/qms/internal/core/services/ping"
	"github.com/Blinkuu/qms/internal/core/services/rate"
	"github.com/Blinkuu/qms/internal/core/services/server"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/log"
)

const (
	Core         = "core"
	SingleBinary = "all"
)

type App struct {
	cfg                     Config
	clock                   clock.Clock
	logger                  log.Logger
	reg                     prometheus.Registerer
	tp                      trace.TracerProvider
	modulesManager          *modules.Manager
	servicesManager         *services.Manager
	serviceNamesAndServices map[string]services.Service
	server                  *server.Service
	ping                    *ping.Service
	memberlist              *memberlist.Service
	alloc                   *alloc.Service
	rate                    *rate.Service
}

func New(cfg Config, clock clock.Clock, logger log.Logger, reg prometheus.Registerer, tp trace.TracerProvider) (*App, error) {
	a := &App{
		cfg:            cfg,
		clock:          clock,
		logger:         logger,
		reg:            reg,
		tp:             tp,
		modulesManager: modules.NewManager(logger.Simple()),
	}

	a.modulesManager.RegisterModule(server.ServiceName, a.initServer, modules.UserInvisibleModule)
	a.modulesManager.RegisterModule(ping.ServiceName, a.initPing, modules.UserInvisibleModule)
	a.modulesManager.RegisterModule(memberlist.ServiceName, a.initMemberlist, modules.UserInvisibleModule)
	a.modulesManager.RegisterModule(alloc.ServiceName, a.initAlloc)
	a.modulesManager.RegisterModule(rate.ServiceName, a.initRate)
	a.modulesManager.RegisterModule(Core, nil)
	a.modulesManager.RegisterModule(SingleBinary, nil)

	deps := map[string][]string{
		server.ServiceName:     nil,
		ping.ServiceName:       {server.ServiceName},
		memberlist.ServiceName: {server.ServiceName},
		Core:                   {server.ServiceName, ping.ServiceName, memberlist.ServiceName},
		alloc.ServiceName:      {Core},
		rate.ServiceName:       {Core},
		SingleBinary:           {Core, alloc.ServiceName, rate.ServiceName},
	}

	for mod, targets := range deps {
		if err := a.modulesManager.AddDependency(mod, targets...); err != nil {
			return nil, fmt.Errorf("failed to add dependency: %w", err)
		}
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	if !a.modulesManager.IsUserVisibleModule(a.cfg.Target) {
		return fmt.Errorf("%s is not a runnable target", a.cfg.Target)
	}

	var err error
	a.serviceNamesAndServices, err = a.modulesManager.InitModuleServices(a.cfg.Target)
	if err != nil {
		return fmt.Errorf("failed to init module services: %w", err)
	}

	pingHandler := handlers.NewPingHTTPHandler(a.ping)
	a.server.HTTP.Handle("/ping", pingHandler.Ping())

	memberlistHandler := handlers.NewMemberlistHTTPHandler(a.memberlist)
	a.server.HTTP.Handle("/memberlist", memberlistHandler.Memberlist()).Methods(http.MethodGet)

	a.server.HTTP.Handle("/metrics", promhttp.Handler())

	{
		v1ApiRouter := a.server.HTTP.PathPrefix("/api/v1").Subrouter()

		rateHandler := handlers.NewRateHTTPHandler(a.rate)
		v1ApiRouter.HandleFunc("/allow", rateHandler.Allow()).Methods(http.MethodPost)

		allocHandler := handlers.NewAllocHTTPHandler(a.alloc)
		v1ApiRouter.HandleFunc("/alloc", allocHandler.Alloc()).Methods(http.MethodPost)
		v1ApiRouter.HandleFunc("/free", allocHandler.Free()).Methods(http.MethodPost)
	}

	svcs := make([]services.Service, 0, len(a.serviceNamesAndServices))
	for _, svc := range a.serviceNamesAndServices {
		svcs = append(svcs, svc)
	}

	a.servicesManager, err = services.NewManager(svcs...)
	if err != nil {
		return fmt.Errorf("failed to create services manager: %w", err)
	}

	healthy := func() { a.logger.Info("starting app") }
	stopped := func() { a.logger.Info("stopping app") }
	failed := func(service services.Service) {
		a.servicesManager.StopAsync()

		for name, svc := range a.serviceNamesAndServices {
			if svc == service {
				if service.FailureCase() == modules.ErrStopProcess {
					a.logger.Info("received stop signal via return error", "module", name, "err", service.FailureCase())
				} else {
					a.logger.Error("module failed", "module", name, "err", service.FailureCase())
				}

				return
			}
		}

		a.logger.Error("module failed", "module", "unknown", "err", service.FailureCase())
	}
	a.servicesManager.AddListener(services.NewManagerListener(healthy, stopped, failed))

	err = a.servicesManager.StartAsync(ctx)
	if err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	err = a.servicesManager.AwaitStopped(context.Background())
	if err != nil {
		return fmt.Errorf("failed to await stopped: %w", err)
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.servicesManager.StopAsync()
	err := a.servicesManager.AwaitStopped(ctx)
	if err != nil {
		return fmt.Errorf("failed to await stopped: %w", err)
	}

	return nil
}

func (a *App) initServer() (services.Service, error) {
	waitFor := func() []services.Service {
		namedSvcs := make([]services.Service, 0, len(a.serviceNamesAndServices))
		for name, svc := range a.serviceNamesAndServices {
			if name == server.ServiceName { // Do not wait for self
				continue
			}

			namedSvcs = append(namedSvcs, svc)
		}

		return namedSvcs
	}
	a.server = server.NewService(a.cfg.ServerConfig, a.clock, a.logger, a.reg, a.tp, waitFor)
	return a.server, nil
}

type loggingEventDelegate struct {
	logger log.Logger
}

func (l loggingEventDelegate) NotifyJoin(instance *cloud.Instance) {
	l.logger.Info("NotifyJoin()", "instance", instance)
}

func (l loggingEventDelegate) NotifyLeave(instance *cloud.Instance) {
	l.logger.Info("NotifyLeave()", "instance", instance)
}

func (l loggingEventDelegate) NotifyUpdate(instance *cloud.Instance) {
	l.logger.Info("NotifyUpdate()", "instance", instance)
}

func (a *App) initMemberlist() (services.Service, error) {
	var err error
	eventDelegate := loggingEventDelegate{logger: a.logger}
	a.memberlist, err = memberlist.NewService(
		a.cfg.MemberlistConfig,
		a.logger,
		eventDelegate,
		a.cfg.Target,
		a.cfg.ServerConfig.HTTPPort,
	)
	return a.memberlist, err
}

func (a *App) initPing() (services.Service, error) {
	a.ping = ping.NewService(a.logger)
	return a.ping, nil
}

func (a *App) initAlloc() (services.Service, error) {
	var err error
	a.alloc, err = alloc.NewService(a.cfg.AllocConfig, a.logger)
	return a.alloc, err
}

func (a *App) initRate() (services.Service, error) {
	var err error
	a.rate, err = rate.NewService(a.cfg.RateConfig, a.clock, a.logger)
	return a.rate, err
}
