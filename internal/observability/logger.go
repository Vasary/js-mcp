package observability

import (
	"log/slog"
	"os"
)

func NewLogger(levelName string) *slog.Logger {
	level := new(slog.LevelVar)
	level.Set(parseLevel(levelName))

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func parseLevel(levelName string) slog.Level {
	switch levelName {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
