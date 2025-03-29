package log

import (
	"context"
	"fmt"
	"gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	L *Logger
}

func (l *GormLogger) LogMode(_ logger.LogLevel) logger.Interface {
	return l
}

// Setting the LogModes

func (l *GormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	toLog := fmt.Sprintf(msg, data...)
	l.L.LogEvent("INFO", toLog)
}

func (l *GormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	toLog := fmt.Sprintf(msg, data...)
	l.L.LogEvent("WARN", toLog)
}

func (l *GormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	toLog := fmt.Sprintf(msg, data...)
	l.L.LogEvent("ERROR", toLog)
}

func (l *GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	execTime := time.Since(begin)

	toLog := fmt.Sprintf("SQL: %s | Rows: %d | Time: %v | Error: %v", sql, rows, execTime, err)
	l.L.LogEvent("DEBUG", toLog)
}
