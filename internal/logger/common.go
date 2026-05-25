package logger

import (
	"log/slog"
	"strings"
)

var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"token":         {},
	"api_key":       {},
	"apikey":        {},
	"auth":          {},
	"authorization": {},
}

func DefaultReplaceAttr() func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		key := strings.ToLower(a.Key)

		if a.Key == slog.TimeKey {
			return slog.Int64("timestamp", a.Value.Time().UTC().Unix())
		}

		if _, ok := sensitiveKeys[key]; ok {
			return slog.String(a.Key, "[REDACTED]")
		}

		if a.Value.Kind() == slog.KindString && a.Value.String() == "" {
			return slog.Attr{}
		}

		return a
	}
}

func getLogLevel(lvl string) slog.Level {
	switch strings.TrimSpace(strings.ToUpper(lvl)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	case "INFO":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
