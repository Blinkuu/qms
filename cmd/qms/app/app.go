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
	"github.com/thanos-io/thanos/pkg/discovery/dns"
	"go.opentelemetry.io/otel/trace"

	"github.com/Blinkuu/qms/internal/core/domain"
	"github.com/Blinkuu/qms/internal/core/services/alloc"
	"github.com/Blinkuu/qms/internal/core/services/memberlist"
	"github.com/Blinkuu/qms/internal/core/services/ping"
	"github.com/Blinkuu/qms/internal/core/services/proxy"
	"github.com/Blinkuu/qms/internal/core/services/rate"
	"github.com/Blinkuu/qms/internal/core/services/server"
	allocstorage "github.com/Blinkuu/qms/internal/core/storage/alloc"
	"github.com/Blinkuu/qms/internal/handlers"
	"github.com/Blinkuu/qms/pkg/cloud"
	"github.com/Blinkuu/qms/pkg/cloud/native"
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
	discoverer              cloud.Discoverer
	modulesManager          *modules.Manager
	servicesManager         *services.Manager
	serviceNamesAndServices map[string]services.Service
	server                  *server.Service
	ping                    *ping.Service
	memberlist              *memberlist.Service
	proxy                   *proxy.Service
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
		discoverer:     native.NewDiscoverer(logger, dns.NewProvider(logger.Simple(), reg, dns.GolangResolverType)),
		modulesManager: modules.NewManager(logger.Simple()),
	}

	a.modulesManager.RegisterModule(server.ServiceName, a.initServer, modules.UserInvisibleModule)
	a.modulesManager.RegisterModule(ping.ServiceName, a.initPing, modules.UserInvisibleModule)
	a.modulesManager.RegisterModule(memberlist.ServiceName, a.initMemberlist, modules.UserInvisibleModule)
	a.modulesManager.RegisterModule(proxy.ServiceName, a.initProxy)
	a.modulesManager.RegisterModule(alloc.ServiceName, a.initAlloc)
	a.modulesManager.RegisterModule(rate.ServiceName, a.initRate)
	a.modulesManager.RegisterModule(Core, nil)
	a.modulesManager.RegisterModule(SingleBinary, nil)

	deps := map[string][]string{
		server.ServiceName:     nil,
		ping.ServiceName:       {server.ServiceName},
		memberlist.ServiceName: {server.ServiceName},
		Core:                   {server.ServiceName, ping.ServiceName, memberlist.ServiceName},
		proxy.ServiceName:      {Core},
		alloc.ServiceName:      {Core},
		rate.ServiceName:       {Core},
		SingleBinary:           {Core, proxy.ServiceName, alloc.ServiceName, rate.ServiceName},
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

	if err := a.servicesManager.StartAsync(ctx); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	if err := a.servicesManager.AwaitHealthy(ctx); err != nil {
		return fmt.Errorf("failed to await healthy: %w", err)
	}

	pingHandler := handlers.NewPingHTTPHandler(a.ping)
	a.server.HTTP.Handle("/ping", pingHandler.Ping())

	memberlistHandler := handlers.NewMemberlistHTTPHandler(a.memberlist)
	a.server.HTTP.Handle("/memberlist", memberlistHandler.Memberlist()).Methods(http.MethodGet)

	a.server.HTTP.Handle("/metrics", promhttp.Handler())

	{
		v1ApiRouter := a.server.HTTP.PathPrefix("/api/v1").Subrouter()

		rateProxyHandler := handlers.NewRateHTTPHandler(a.proxy)
		allocProxyHandler := handlers.NewAllocHTTPHandler(a.proxy)
		v1ApiRouter.Handle("/allow", rateProxyHandler.Allow()).Methods(http.MethodPost)
		v1ApiRouter.Handle("/view", allocProxyHandler.View()).Methods(http.MethodPost)
		v1ApiRouter.Handle("/alloc", allocProxyHandler.Alloc()).Methods(http.MethodPost)
		v1ApiRouter.Handle("/free", allocProxyHandler.Free()).Methods(http.MethodPost)

		{
			v1InternalApiRouter := v1ApiRouter.PathPrefix("/internal").Subrouter()

			rateHandler := handlers.NewRateHTTPHandler(a.rate)
			v1InternalApiRouter.Handle("/allow", rateHandler.Allow()).Methods(http.MethodPost)

			allocHandler := handlers.NewAllocHTTPHandler(a.alloc)
			v1InternalApiRouter.Handle("/view", allocHandler.View()).Methods(http.MethodPost)
			v1InternalApiRouter.Handle("/alloc", allocHandler.Alloc()).Methods(http.MethodPost)
			v1InternalApiRouter.Handle("/free", allocHandler.Free()).Methods(http.MethodPost)

			if a.cfg.AllocConfig.Storage.Backend == allocstorage.Raft {
				raftHandler := handlers.NewRaftHTTPHandler(a.alloc)
				v1InternalApiRouter.Handle("/raft/join", raftHandler.Join()).Methods(http.MethodPost)
				v1InternalApiRouter.Handle("/raft/exit", raftHandler.Exit()).Methods(http.MethodPost)
			}
		}
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

func (l loggingEventDelegate) NotifyJoin(instance domain.Instance) {
	l.logger.Info("NotifyJoin()", "instance", instance)
}

func (l loggingEventDelegate) NotifyLeave(instance domain.Instance) {
	l.logger.Info("NotifyLeave()", "instance", instance)
}

func (l loggingEventDelegate) NotifyUpdate(instance domain.Instance) {
	l.logger.Info("NotifyUpdate()", "instance", instance)
}

func (a *App) initMemberlist() (services.Service, error) {
	var err error
	eventDelegate := loggingEventDelegate{logger: a.logger}
	a.memberlist, err = memberlist.NewService(
		a.cfg.MemberlistConfig,
		a.logger.With("service", memberlist.ServiceName),
		a.discoverer,
		eventDelegate,
		a.cfg.Target,
		a.cfg.ServerConfig.HTTPPort,
	)
	return a.memberlist, err
}

func (a *App) initPing() (services.Service, error) {
	a.ping = ping.NewService(
		a.logger.With("service", ping.ServiceName),
	)
	return a.ping, nil
}

func (a *App) initProxy() (services.Service, error) {
	memberlistClient := memberlist.NewClient(
		a.logger.With("service", proxy.ServiceName, "component", memberlist.ClientName),
	)
	rateClient := rate.NewClient(
		a.logger.With("service", proxy.ServiceName, "component", rate.ClientName),
	)
	allocClient := alloc.NewClient(
		a.logger.With("service", proxy.ServiceName, "component", alloc.ClientName),
	)
	var err error
	a.proxy, err = proxy.NewService(
		a.cfg.ProxyConfig,
		a.logger.With("service", proxy.ServiceName),
		a.discoverer,
		memberlistClient,
		rateClient,
		allocClient,
	)
	return a.proxy, err
}

func (a *App) initAlloc() (services.Service, error) {
	var err error
	a.alloc, err = alloc.NewService(
		a.cfg.AllocConfig,
		a.logger.With("service", alloc.ServiceName),
		a.memberlist,
	)
	return a.alloc, err
}

func (a *App) initRate() (services.Service, error) {
	var err error
	a.rate, err = rate.NewService(
		a.cfg.RateConfig,
		a.clock,
		a.logger.With("service", rate.ServiceName),
	)
	return a.rate, err
}
