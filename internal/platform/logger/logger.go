package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func New(configuredLevel string) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(configuredLevel))); err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", configuredLevel, err)
	}

	options := &slog.HandlerOptions{Level: level}
	return slog.New(slog.NewJSONHandler(os.Stdout, options)), nil
}
