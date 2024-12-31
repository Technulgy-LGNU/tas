package log

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

type FiberLogger struct {
	L *Logger
}

// FiberLoggerMiddleware Fiber logging middleware to accumulate logs
func (l *FiberLogger) FiberLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.OriginalURL() != "/monitor" {
			// Capture request
			start := time.Now()
			err := c.Next()
			latency := time.Since(start)
			if err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Printf("Error running next fiber function: %v\n", err)
				return err
			}
			logMessage := fmt.Sprintf("Request: %s %s | Status: %d | Latency: %v | Error: %v\n",
				c.Method(),
				c.OriginalURL(),
				c.Response().StatusCode(),
				latency,
				err,
			)
			l.L.AddToLogBuffer(logMessage, "INFO")

			return err
		} else {
			return nil
		}
	}
}
