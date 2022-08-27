package log

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// memorySink implements zap.Sink by writing all messages to a buffer.
type memorySink struct {
	*bytes.Buffer
}

func (s *memorySink) Close() error {
	return nil
}

func (s *memorySink) Sync() error {
	return nil
}

type failingMemorySink struct{}

func (f failingMemorySink) Write(_ []byte) (n int, err error) {
	return 0, errors.New("failing memory sink write error")
}

func (f failingMemorySink) Sync() error {
	return errors.New("failing memory sink sync error")
}

func (f failingMemorySink) Close() error {
	return errors.New("failing memory sink close error")
}

var (
	s   *memorySink
	fs  *failingMemorySink
	sMu sync.Mutex
)

func sink(t *testing.T) *memorySink {
	t.Helper()

	sMu.Lock()
	defer sMu.Unlock()

	if s != nil {
		return s
	}

	s = &memorySink{new(bytes.Buffer)}
	err := zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return s, nil
	})
	require.NoError(t, err)

	return s
}

func failingSink(t *testing.T) *failingMemorySink {
	t.Helper()

	sMu.Lock()
	defer sMu.Unlock()

	if fs != nil {
		return fs
	}

	fs = &failingMemorySink{}
	err := zap.RegisterSink("failingMemory", func(*url.URL) (zap.Sink, error) {
		return fs, nil
	})
	require.NoError(t, err)

	return fs
}

func TestNewZapLogger_ReturnsNoErrorAndNonNilLoggerWithCorrectArguments(t *testing.T) {
	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"stderr"}

	// When
	got, err := NewZapLogger(name, level, outputPaths)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestNewZapLogger_ReturnsErrorAndNilWithIncorrectLevelName(t *testing.T) {
	// Given
	name := "test"
	level := "incorrect"
	outputPaths := []string{"stderr"}

	// When
	got, err := NewZapLogger(name, level, outputPaths)

	// Then
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestNewZapLogger_ReturnsErrorAndNilWithEmptyOutputPaths(t *testing.T) {
	// Given
	name := "test"
	level := "debug"
	var outputPaths []string

	// When
	got, err := NewZapLogger(name, level, outputPaths)

	// Then
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestNewZapLogger_ReturnsErrorAndNilWithIncorrectOutputPaths(t *testing.T) {
	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"unregisteredSink://"}

	// When
	got, err := NewZapLogger(name, level, outputPaths)

	// Then
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestZapLogger_Debug_CorrectlyOutputsDebugLevelMessage(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	logger.Debug("test", "key", "value")
	got := sink.String()

	// Then
	assert.Contains(t, got, `"level":"debug"`)
	assert.Contains(t, got, `"msg":"test"`)
	assert.Contains(t, got, `"key":"value"`)
}

func TestZapLogger_Info_CorrectlyOutputsInfoLevelMessage(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	logger.Info("test", "key", "value")
	got := sink.String()

	// Then
	assert.Contains(t, got, `"level":"info"`)
	assert.Contains(t, got, `"msg":"test"`)
	assert.Contains(t, got, `"key":"value"`)

}

func TestZapLogger_Warn_CorrectlyOutputsWarnLevelMessage(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	logger.Warn("test", "key", "value")
	got := sink.String()

	// Then
	assert.Contains(t, got, `"level":"warn"`)
	assert.Contains(t, got, `"msg":"test"`)
	assert.Contains(t, got, `"key":"value"`)
}

func TestZapLogger_Error_CorrectlyOutputsErrorLevelMessage(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	logger.Error("test", "key", "value")
	got := sink.String()

	// Then
	assert.Contains(t, got, `"level":"error"`)
	assert.Contains(t, got, `"msg":"test"`)
	assert.Contains(t, got, `"key":"value"`)
}

func TestZapLogger_Panic_CorrectlyOutputsPanicLevelMessageAndPanics(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	panicFunc := func() {
		logger.Panic("test", "key", "value")
	}
	assert.Panics(t, panicFunc)
	got := sink.String()

	// Then
	assert.Contains(t, got, `"level":"panic"`)
	assert.Contains(t, got, `"msg":"test"`)
	assert.Contains(t, got, `"key":"value"`)

}

func TestZapLogger_Fatal_ExitsCommandWithoutSuccess(t *testing.T) {
	if os.Getenv("STOP_FATAL") == "1" {
		// Given
		name := "test"
		level := "debug"
		outputPaths := []string{"memory://"}
		logger, err := NewZapLogger(name, level, outputPaths)
		assert.NoError(t, err)

		// When
		logger.Fatal("test")

		return
	}

	// Then
	cmd := exec.Command(os.Args[0], "-test.run=TestZapLogger_Fatal")
	cmd.Env = append(os.Environ(), "STOP_FATAL=1")
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		assert.Fail(t, "error is not of type exit error")
	}

	assert.False(t, exitErr.Success())
}

func TestZapLogger_With_CorrectlyCreatesNewLoggerWithContext(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	logger.With("key", "value").Debug("")
	got := sink.String()

	// Then
	assert.Contains(t, got, `"key":"value"`)
}

func TestZapLogger_Simple_CorrectlyCreatesSimpleLogger(t *testing.T) {
	sink := sink(t)

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	err = logger.Simple().Log("key", "value")
	got := sink.String()

	// Then
	assert.NoError(t, err)
	assert.Contains(t, got, `"key":"value"`)
}

func TestZapLogger_Close_ReturnsNoErrorWithValidOutputPath(t *testing.T) {
	_ = sink(t) // Register if this test is run in a singular context

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"memory://"}
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	err = logger.Close()

	// Then
	assert.NoError(t, err)
}

func TestZapLogger_Close_ReturnsErrorWithInvalidOutputPath(t *testing.T) {
	_ = failingSink(t) // Register if this test is run in a singular context

	// Given
	name := "test"
	level := "debug"
	outputPaths := []string{"failingMemory://"} // "stderr" is invalid in test context because it is an inappropriate ioctl for device
	logger, err := NewZapLogger(name, level, outputPaths)
	assert.NoError(t, err)

	// When
	err = logger.Close()

	// Then
	assert.Error(t, err)
}

func Test_convertToZapLevel(t *testing.T) {
	type args struct {
		lvl string
	}
	tests := []struct {
		name    string
		args    args
		want    zapcore.Level
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ReturnsNoErrorAndDebugLevelForLowercaseDebugString",
			args: args{
				lvl: "debug",
			},
			want:    zapcore.DebugLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndDebugLevelForUppercaseDebugString",
			args: args{
				lvl: "DEBUG",
			},
			want:    zapcore.DebugLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndDebugLevelForMixedcaseDebugString",
			args: args{
				lvl: "DeBuG",
			},
			want:    zapcore.DebugLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndInfoLevelForLowercaseInfoString",
			args: args{
				lvl: "info",
			},
			want:    zapcore.InfoLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndInfoLevelForUppercaseInfoString",
			args: args{
				lvl: "INFO",
			},
			want:    zapcore.InfoLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndInfoLevelForMixedcaseInfoString",
			args: args{
				lvl: "InFo",
			},
			want:    zapcore.InfoLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndWarnLevelForLowercaseWarnString",
			args: args{
				lvl: "warn",
			},
			want:    zapcore.WarnLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndWarnLevelForUppercaseWarnString",
			args: args{
				lvl: "WARN",
			},
			want:    zapcore.WarnLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndWarnLevelForMixedcaseWarnString",
			args: args{
				lvl: "WaRn",
			},
			want:    zapcore.WarnLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndErrorLevelForLowercaseErrorString",
			args: args{
				lvl: "error",
			},
			want:    zapcore.ErrorLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndErrorLevelForUppercaseErrorString",
			args: args{
				lvl: "ERROR",
			},
			want:    zapcore.ErrorLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndErrorLevelForMixedcaseErrorString",
			args: args{
				lvl: "ErRoR",
			},
			want:    zapcore.ErrorLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndPanicLevelForLowercasePanicString",
			args: args{
				lvl: "panic",
			},
			want:    zapcore.PanicLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndPanicLevelForUppercasePanicString",
			args: args{
				lvl: "PANIC",
			},
			want:    zapcore.PanicLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndPanicLevelForMixedcasePanicString",
			args: args{
				lvl: "PaNiC",
			},
			want:    zapcore.PanicLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndFatalLevelForLowercaseFatalString",
			args: args{
				lvl: "fatal",
			},
			want:    zapcore.FatalLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndFatalLevelForUppercaseFatalString",
			args: args{
				lvl: "FATAL",
			},
			want:    zapcore.FatalLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndFatalLevelForMixedcaseFatalString",
			args: args{
				lvl: "FaTaL",
			},
			want:    zapcore.FatalLevel,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsErrorAndZeroValueForUnknownLevel",
			args: args{
				lvl: "unknown",
			},
			want:    0,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToZapLevel(tt.args.lvl)
			if !tt.wantErr(t, err, fmt.Sprintf("convertToZapLevel(%v)", tt.args.lvl)) {
				return
			}

			assert.Equalf(t, tt.want, got, "convertToZapLevel(%v)", tt.args.lvl)
		})
	}
}
