package log

type NoopLogger struct{}

func NewNoopLogger() NoopLogger { return NoopLogger{} }

func (n NoopLogger) Debug(_ string, _ ...any) {}

func (n NoopLogger) Debugf(_ string, _ ...any) {}

func (n NoopLogger) Info(_ string, _ ...any) {}

func (n NoopLogger) Infof(_ string, _ ...any) {}

func (n NoopLogger) Warn(_ string, _ ...any) {}

func (n NoopLogger) Warnf(_ string, _ ...any) {}

func (n NoopLogger) Error(_ string, _ ...any) {}

func (n NoopLogger) Errorf(_ string, _ ...any) {}

func (n NoopLogger) Panic(_ string, _ ...any) {}

func (n NoopLogger) Panicf(_ string, _ ...any) {}

func (n NoopLogger) Fatal(_ string, _ ...any) {}

func (n NoopLogger) Fatalf(_ string, _ ...any) {}

func (n NoopLogger) With(_ ...any) Logger { return NewNoopLogger() }

func (n NoopLogger) Simple() SimpleLogger { return NewNoopSimpleLogger() }

type NoopSimpleLogger struct{}

func NewNoopSimpleLogger() NoopSimpleLogger { return NoopSimpleLogger{} }

func (n NoopSimpleLogger) Log(_ ...interface{}) error { return nil }
