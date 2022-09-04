package log

import (
	gokitlog "github.com/go-kit/log"
)

type SimpleLogger interface {
	gokitlog.Logger
}

type Logger interface {
	// Debug logs a message on a Debug level with optional key-value pairs.
	Debug(msg string, keysAndValues ...any)

	// Info logs a message on a Info level with optional key-value pairs.
	Info(msg string, keysAndValues ...any)

	// Warn logs a message on a Warn level with optional key-value pairs.
	Warn(msg string, keysAndValues ...any)

	// Error logs a message on a Error level with optional key-value pairs.
	Error(msg string, keysAndValues ...any)

	// Panic logs a message on a Panic level with optional key-value pairs and
	// then panics.
	Panic(msg string, keysAndValues ...any)

	// Fatal logs a message on a Fatal level with optional key-value pairs and
	// then calls os.Exit(1).
	Fatal(msg string, keysAndValues ...any)

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
