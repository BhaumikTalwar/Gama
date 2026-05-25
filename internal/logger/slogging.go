package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/buildinfo"
)

func NewSlogger(serviceName string, logConfig *config.LoggingConfig) *slog.Logger {
	level := getLogLevel(logConfig.Level)

	var writer io.Writer
	if logConfig.Output == "stdout" {
		writer = os.Stdout
	} else {
		writer = NewLumberjackLogger(logConfig)
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   logConfig.AddSource,
		ReplaceAttr: DefaultReplaceAttr(),
	}

	var handler slog.Handler
	if logConfig.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	} else {
		handler = slog.NewJSONHandler(writer, opts)
	}

	logger := slog.New(handler).With(
		"service", serviceName,
		"version", buildinfo.VersionStr(),
		"commit", buildinfo.CommitStr(),
	)

	return logger
}
