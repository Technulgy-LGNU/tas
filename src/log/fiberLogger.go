package log

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

// FiberCustomLogger accumulates logs in memory, writes them to disk at the end of the day,
// on buffer overflow, or on shutdown
type FiberCustomLogger struct {
	logBuffer []string
	mutex     sync.Mutex
}

// FiberLoggerMiddleware Fiber logging middleware to accumulate logs
func (l *FiberCustomLogger) FiberLoggerMiddleware() fiber.Handler {
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
			l.addToLogBuffer(logMessage)

			return err
		} else {
			return nil
		}
	}
}

// addToLogBuffer Adds log to buffer and flushes if it exceeds the max buffer size
func (l *FiberCustomLogger) addToLogBuffer(logMessage string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	logEntry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), logMessage)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Print(logEntry)
	l.logBuffer = append(l.logBuffer, logEntry)

	if len(l.logBuffer) >= maxLogBufferSize {
		l.writeLogToDisk()
	}
}

func (l *FiberCustomLogger) writeLogToDisk() {
	// Lock
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if logBuffer is empty
	if len(l.logBuffer) == 0 {
		return
	}

	// Write to disk
	fileName := fmt.Sprintf("./log/fiber_logs/%s.log", time.Now().Format("2006-1-02"))
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error opening logfile: %v\n", err)
		return
	}
	defer file.Close()

	for _, logEntry := range l.logBuffer {
		if _, err := file.WriteString(logEntry + "\n"); err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error writing log to disk: %v\n", err)
			return
		}
	}

	// Clear log buffer
	l.logBuffer = []string{}
}

// StartDailyFlush daily flush of all logs
func (l *FiberCustomLogger) StartDailyFlush() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			l.writeLogToDisk()
		}
	}()
}

// HandleShutdown Flush logs to disk on shutdown
func (l *FiberCustomLogger) HandleShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Flushing logs to disk before shutdown ...")
		l.writeLogToDisk()
		os.Exit(0)
	}()
}
