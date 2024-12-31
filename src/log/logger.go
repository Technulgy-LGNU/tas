package log

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Define max log buffer size
var maxLogBufferSize = 500

type Logger struct {
	logBuffer []string
	mutex     sync.Mutex
	LogLevel  string
}

// SetLogLevel sets the log level
func (l *Logger) SetLogLevel(level string) {
	switch level {
	case "DEBUG":
		l.LogLevel = "DEBUG"
	case "INFO":
		l.LogLevel = "INFO"
	case "WARN":
		l.LogLevel = "WARN"
	case "ERROR":
		l.LogLevel = "ERROR"
	default:
		l.LogLevel = "DEBUG"
	}
}

// LogEvent logs an event with a level
func (l *Logger) LogEvent(event string, level string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	switch level {
	case "DEBUG":
		if l.LogLevel == "DEBUG" {
			log.Printf("%s: %s\n", level, event)
			l.AddToLogBuffer(event, level)
		}
	case "INFO":
		if l.LogLevel == "DEBUG" || l.LogLevel == "INFO" {
			log.Printf("%s: %s\n", level, event)
			l.AddToLogBuffer(event, level)
		}
	case "WARN":
		if l.LogLevel == "DEBUG" || l.LogLevel == "INFO" || l.LogLevel == "WARN" {
			log.Printf("%s: %s\n", level, event)
			l.AddToLogBuffer(event, level)
		}
	case "ERROR":
		log.Printf("%s: %s\n", level, event)
		l.AddToLogBuffer(event, level)
	case "FATAL":
		log.Println("Error, shutting down ...")
		l.AddToLogBuffer(event, level)
		l.WriteLogToDisk()
		log.Fatalf("%s: %s\n", level, event)
	}
}

func (l *Logger) AddToLogBuffer(event string, level string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	logEntry := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format(time.RFC3339), level, event)
	l.logBuffer = append(l.logBuffer, logEntry)

	if len(l.logBuffer) >= maxLogBufferSize {
		l.WriteLogToDisk()
	}
}

func (l *Logger) WriteLogToDisk() {
	// Lock
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Write to disk
	if os.Args[1] == "prod" {
		fileName := fmt.Sprintf("/var/lib/tas/logs/%s.log", time.Now().Format(time.RFC3339))
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error opening logfile: %v\n", err)
			return
		}
		defer file.Close()

		for _, logEntry := range l.logBuffer {
			if _, err := file.WriteString(logEntry); err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Printf("Error writing log to disk: %v\n", err)
				return
			}
		}

		// Clear log buffer
		l.logBuffer = []string{}
	} else if os.Args[1] == "dev" {
		fileName := fmt.Sprintf("data/logs/%s.log", time.Now().Format("2006-1-02"))
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error opening logfile: %v\n", err)
			return
		}
		defer file.Close()

		for _, logEntry := range l.logBuffer {
			if _, err := file.WriteString(logEntry); err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Printf("Error writing log to disk: %v\n", err)
				return
			}
		}

		// Clear log buffer
		l.logBuffer = []string{}
	}
}

// StartDailyFlush daily flush of all logs
func (l *Logger) StartDailyFlush() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			l.WriteLogToDisk()
		}
	}()
}
