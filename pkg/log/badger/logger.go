package badger

import (
	"github.com/Blinkuu/qms/pkg/log"
)

type Logger struct {
	logger log.Logger
}

func NewLogger(logger log.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (b Logger) Debugf(template string, args ...any) {
	b.logger.Debugf(template, args)
}

func (b Logger) Infof(template string, args ...any) {
	b.logger.Infof(template, args)
}

func (b Logger) Warningf(template string, args ...any) {
	b.logger.Warnf(template, args)
}

func (b Logger) Errorf(template string, args ...any) {
	b.logger.Errorf(template, args)
}
