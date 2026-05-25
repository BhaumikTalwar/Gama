package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/buildinfo"
)

func NewZeroSlogger(serviceName string, logConfig *config.LoggingConfig) *slog.Logger {
	level := getLogLevel(logConfig.Level)

	var writer io.Writer
	if logConfig.Output == "stdout" {
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		writer = NewLumberjackLogger(logConfig)
	}

	zerologLogger := zerolog.New(writer)
	opts := slogzerolog.Option{
		Level:       level,
		Logger:      &zerologLogger,
		AddSource:   logConfig.AddSource,
		ReplaceAttr: DefaultReplaceAttr(),
	}

	logger := slog.New(opts.NewZerologHandler()).With(
		"service", serviceName,
		"version", buildinfo.VersionStr(),
		"commit", buildinfo.CommitStr(),
	)

	return logger
}
