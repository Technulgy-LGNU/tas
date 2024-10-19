package log

import (
	"context"
	"fmt"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

const (
	// maxLogBufferSize Maximum log entries before flushing to disk
	maxLogBufferSize = 1000
)

// GormCustomLogger accumulates logs in memory, writes them to disk at the end of the day,
// on buffer overflow, or on shutdown
type GormCustomLogger struct {
	logBuffer []string
	mutex     sync.Mutex
}

// LogMode sets the log mode (used to satisfy the GORM logger interface)
func (l *GormCustomLogger) LogMode(_ logger.LogLevel) logger.Interface {
	return l
}

// Setting the LogModes

func (l *GormCustomLogger) Info(_ context.Context, msg string, data ...interface{}) {
	toLog := fmt.Sprintf(msg, data...)
	l.addToLogBuffer("INFO", toLog)
}

func (l *GormCustomLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	toLog := fmt.Sprintf(msg, data...)
	l.addToLogBuffer("WARN", toLog)
}

func (l *GormCustomLogger) Error(_ context.Context, msg string, data ...interface{}) {
	toLog := fmt.Sprintf(msg, data...)
	l.addToLogBuffer("ERROR", toLog)
}

func (l *GormCustomLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	execTime := time.Since(begin)

	toLog := fmt.Sprintf("SQL: %s | Rows: %d | Time: %v | Error: %v", sql, rows, execTime, err)
	l.addToLogBuffer("TRACE", toLog)
}

// addToLogBuffer Adds log to buffer and flushes if it exceeds the max buffer size
func (l *GormCustomLogger) addToLogBuffer(level string, logMessage string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	logEntry := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format(time.RFC3339), level, logMessage)
	l.logBuffer = append(l.logBuffer, logEntry)
	log.Println(len(l.logBuffer))

	if len(l.logBuffer) >= maxLogBufferSize {
		l.mutex.Unlock()
		l.WriteLogToDisk()
		l.mutex.Lock()
	}
}

// WriteLogToDisk Writes log to disk, either creates a new file or appends to an existing one
func (l *GormCustomLogger) WriteLogToDisk() {
	// Lock
	l.mutex.Lock()
	defer l.mutex.Unlock()
	fmt.Println("Logging data to file")

	// Write to disk
	fileName := fmt.Sprintf("data/log/gorm_logs/%s.log", time.Now().Format("2006-1-02"))
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

// StartDailyFlush daily flush of all logs
func (l *GormCustomLogger) StartDailyFlush() {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			<-ticker.C
			l.WriteLogToDisk()
		}
	}()
}
