package log

import (
	"fmt"
	"os"
)

type Logger struct{}

// LogEvent logs an event with a level
func (l *Logger) LogEvent(event string, level string) {
	if level == "FATAL" {
		fmt.Printf("FATAL Event: %s\n", event)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", level, event)
}
