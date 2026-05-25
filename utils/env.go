package utils

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

func GetEnvInt(key string, dest *int) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			*dest = i
		} else {
			slog.Warn(fmt.Sprintf("Invalid %s", key), slog.String("value", v), slog.String("error", err.Error()))
		}
	}
}

func GetEnvBool(key string, dest *bool) {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			*dest = b
		} else {
			slog.Warn(fmt.Sprintf("Invalid %s", key), slog.String("value", v), slog.String("error", err.Error()))
		}
	}
}
