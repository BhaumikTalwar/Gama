package logger

import (
	"github.com/BhaumikTalwar/Gama/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLumberjackLogger(logConfig *config.LoggingConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   logConfig.FilePath,
		MaxSize:    logConfig.MaxSizeMB,
		MaxBackups: logConfig.MaxBackups,
		MaxAge:     logConfig.MaxAgeDays,
		Compress:   logConfig.Compress,
	}
}
