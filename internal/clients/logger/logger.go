// Package logger constructs the application's structured logger.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Config controls logger construction. Writer defaults to [os.Stdout] when
// nil.
type Config struct {
	// Dev, when true, floors the effective level at slog.LevelDebug and
	// switches to a human-readable text handler with source locations.
	Dev bool

	// Level is the minimum log level to emit: "debug", "info", "warn", or
	// "error" (case-insensitive).
	Level string

	Writer io.Writer
}

// New builds a [slog.Logger] from cfg. It returns an error if cfg.Level does
// not name a known level.
func New(cfg Config) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("logger: unknown level %q: %w", cfg.Level, err)
	}

	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	if cfg.Dev && level > slog.LevelDebug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level, AddSource: cfg.Dev}

	var handler slog.Handler
	if cfg.Dev {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}

	return slog.New(handler), nil
}
