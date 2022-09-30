package log

import (
	"errors"
	"fmt"

	"github.com/alex-laties/gokitzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(name, lvl string, outputPaths []string) (*ZapLogger, error) {
	level, err := convertToZapLevel(lvl)
	if err != nil {
		return nil, fmt.Errorf("failed to convert level to zapcore level: %w", err)
	}

	if len(outputPaths) < 1 {
		return nil, errors.New("no output paths provided")
	}

	cfg := zap.Config{
		Level: zap.NewAtomicLevelAt(level),

		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},

		Encoding: "json",

		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},

		OutputPaths:      outputPaths,
		ErrorOutputPaths: outputPaths,
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap config: %w", err)
	}

	_, err = zap.RedirectStdLogAt(logger, level)
	if err != nil {
		return nil, fmt.Errorf("failed to redirect std logger: %w", err)
	}

	return &ZapLogger{
		logger: logger.WithOptions(zap.AddCallerSkip(1)).Named(name).Sugar(),
	}, nil
}

func (z *ZapLogger) Debug(msg string, keysAndValues ...any) {
	z.logger.Debugw(msg, keysAndValues...)
}

func (z *ZapLogger) Debugf(template string, args ...any) {
	z.logger.Debugf(template, args...)
}

func (z *ZapLogger) Info(msg string, keysAndValues ...any) {
	z.logger.Infow(msg, keysAndValues...)
}

func (z *ZapLogger) Infof(template string, args ...any) {
	z.logger.Infof(template, args...)
}

func (z *ZapLogger) Warn(msg string, keysAndValues ...any) {
	z.logger.Warnw(msg, keysAndValues...)
}

func (z *ZapLogger) Warnf(template string, args ...any) {
	z.logger.Warnf(template, args...)
}

func (z *ZapLogger) Error(msg string, keysAndValues ...any) {
	z.logger.Errorw(msg, keysAndValues...)
}

func (z *ZapLogger) Errorf(template string, args ...any) {
	z.logger.Errorf(template, args...)
}

func (z *ZapLogger) Panic(msg string, keysAndValues ...any) {
	z.logger.Panicw(msg, keysAndValues...)
}

func (z *ZapLogger) Panicf(template string, args ...any) {
	z.logger.Panicf(template, args...)
}

func (z *ZapLogger) Fatal(msg string, keysAndValues ...any) {
	z.logger.Fatalw(msg, keysAndValues...)
}

func (z *ZapLogger) Fatalf(template string, args ...any) {
	z.logger.Fatalf(template, args...)
}

func (z *ZapLogger) With(keysAndValues ...any) Logger {
	return &ZapLogger{
		logger: z.logger.With(keysAndValues...),
	}
}

func (z *ZapLogger) Simple() SimpleLogger {
	l := Must(gokitzap.FromZSLogger(z.logger.WithOptions(zap.AddCallerSkip(1))))
	l.SetMessageKey("msg")

	return l
}

func (z *ZapLogger) Close() error {
	if err := z.logger.Sync(); err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	return nil
}

// convertToZapLevel converts log level string to zapcore.Level.
func convertToZapLevel(lvl string) (zapcore.Level, error) {
	var level zapcore.Level
	if err := level.Set(lvl); err != nil {
		var zero zapcore.Level
		return zero, fmt.Errorf("failed to set level: %w", err)
	}

	return level, nil
}
