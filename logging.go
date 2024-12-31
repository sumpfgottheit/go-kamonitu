package main

import (
	"context"
	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func GetCurrentFunctionName() string {
	// Skip GetCurrentFunctionName
	return getFrame(4).Function
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

// Custom handler that adds the caller function name
type FuncNameHandler struct {
	handler slog.Handler
}

func (h *FuncNameHandler) Handle(ctx context.Context, record slog.Record) error {
	// Extract the caller function name
	functionName := GetCurrentFunctionName()
	if functionName == "unknown" {
		functionName = "unknown_function"
	}
	record.AddAttrs(slog.String("function", functionName))
	return h.handler.Handle(ctx, record)
}

func (h *FuncNameHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.handler.Enabled(context.Background(), level)
}

func (h *FuncNameHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FuncNameHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *FuncNameHandler) WithGroup(name string) slog.Handler {
	return &FuncNameHandler{handler: h.handler.WithGroup(name)}
}

// setupLogging initializes and configures the logging system based on the provided log level and directory.
// If logDir is empty, logs are written to stderr; otherwise, a rolling log file is created in the specified directory.
// loglevel sets the minimum severity of logs to display, and logs include a timestamp in RFC3339 format.
// Calling kamonitu with -d, loglevel is automatically set to DEBUG and output die STDERR
func setupLogging(loglevel slog.Level, logDir string) {
	var handler slog.Handler
	if logDir == "" {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      loglevel,
			TimeFormat: time.RFC3339,
		})
	} else {
		handler = tint.NewHandler(&lumberjack.Logger{
			Filename:   logDir + "kamonitu.log",
			MaxSize:    10, // megabytes
			MaxBackups: 5,
			MaxAge:     30,   // days
			Compress:   true, // disabled by default
		}, &tint.Options{
			AddSource:  false,
			Level:      loglevel,
			TimeFormat: time.RFC3339,
			NoColor:    true,
		})
	}

	slog.SetDefault(slog.New(&FuncNameHandler{handler: handler}))
	slog.Info("Initialized Logging.", "Level", loglevel.String())

}
