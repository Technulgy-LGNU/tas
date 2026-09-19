package observability

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Configure writes structured logs to stderr, where Docker collects them.
func Configure(raw string) error {
	level := slog.LevelInfo
	if raw != "" {
		if err := level.UnmarshalText([]byte(strings.ToUpper(raw))); err != nil {
			return fmt.Errorf("logging.level must be debug, info, warn or error")
		}
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
	return nil
}
