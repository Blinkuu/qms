package log

type SimpleLogger interface {
	Log(keysAndValues ...interface{}) error
}

type Logger interface {
	// Debug logs a message on a Debug level with optional key-value pairs.
	Debug(msg string, keysAndValues ...any)

	// Debugf uses fmt.Sprintf to log a templated message.
	Debugf(template string, args ...any)

	// Info logs a message on an Info level with optional key-value pairs.
	Info(msg string, keysAndValues ...any)

	// Infof uses fmt.Sprintf to log a templated message.
	Infof(template string, args ...any)

	// Warn logs a message on a Warn level with optional key-value pairs.
	Warn(msg string, keysAndValues ...any)

	// Warnf uses fmt.Sprintf to log a templated message.
	Warnf(template string, args ...any)

	// Error logs a message on an Error level with optional key-value pairs.
	Error(msg string, keysAndValues ...any)

	// Errorf uses fmt.Sprintf to log a templated message.
	Errorf(template string, args ...any)

	// Panic logs a message on a Panic level with optional key-value pairs and
	// then panics.
	Panic(msg string, keysAndValues ...any)

	// Panicf uses fmt.Sprintf to log a templated message, then panics.
	Panicf(template string, args ...any)

	// Fatal logs a message on a Fatal level with optional key-value pairs and
	// then calls os.Exit(1).
	Fatal(msg string, keysAndValues ...any)

	// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
	Fatalf(template string, args ...any)

	// With adds a variadic number of fields to the logging context.
	With(keysAndValues ...any) Logger

	// Simple creates a new SimpleLogger.
	Simple() SimpleLogger
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}
