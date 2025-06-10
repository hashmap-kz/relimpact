package loggr

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

type LevelLogger struct {
	level   LogLevel
	appCode string
	l       *log.Logger
}

var (
	Logger = newDefaultLogger() // initialized safely
	once   sync.Once
)

func newDefaultLogger() *LevelLogger {
	return &LevelLogger{
		level:   LevelInfo,
		appCode: fmt.Sprintf("%d", os.Getpid()),
		l:       log.New(os.Stderr, "", 0),
	}
}

// Init safely sets a global logger (only once)
func Init(level LogLevel, appCode string) {
	once.Do(func() {
		Logger = &LevelLogger{
			level:   level,
			appCode: appCode,
			l:       log.New(os.Stderr, "", 0),
		}
	})
}

func (l *LevelLogger) log(level LogLevel, label, msg string) {
	if level < l.level {
		return
	}
	now := time.Now().Format("2006-01-02 15:04:05.000 -07")
	output := fmt.Sprintf("%s -- [%s] -- %-5s -- %s", now, l.appCode, label, msg)
	l.l.Println(output)
}

// Trace
func (l *LevelLogger) Trace(msg string) {
	l.log(LevelTrace, "TRACE", msg)
}

func (l *LevelLogger) Tracef(format string, args ...any) {
	l.Trace(fmt.Sprintf(format, args...))
}

// Debug
func (l *LevelLogger) Debug(msg string) {
	l.log(LevelDebug, "DEBUG", msg)
}

func (l *LevelLogger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Info
func (l *LevelLogger) Info(msg string) {
	l.log(LevelInfo, "INFO", msg)
}

func (l *LevelLogger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warn
func (l *LevelLogger) Warn(msg string) {
	l.log(LevelWarn, "WARN", msg)
}

func (l *LevelLogger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Error
func (l *LevelLogger) Error(msg string) {
	l.log(LevelError, "ERROR", msg)
}

func (l *LevelLogger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatal
func (l *LevelLogger) Fatal(msg string) {
	l.log(LevelError, "FATAL", msg)
	os.Exit(1)
}

func (l *LevelLogger) Fatalf(format string, args ...any) {
	l.Fatal(fmt.Sprintf(format, args...))
}

// wrappers

func Trace(msg string)                  { Logger.Trace(msg) }
func Tracef(format string, args ...any) { Logger.Tracef(format, args...) }

func Debug(msg string)                  { Logger.Debug(msg) }
func Debugf(format string, args ...any) { Logger.Debugf(format, args...) }

func Info(msg string)                  { Logger.Info(msg) }
func Infof(format string, args ...any) { Logger.Infof(format, args...) }

func Warn(msg string)                  { Logger.Warn(msg) }
func Warnf(format string, args ...any) { Logger.Warnf(format, args...) }

func Error(msg string)                  { Logger.Error(msg) }
func Errorf(format string, args ...any) { Logger.Errorf(format, args...) }

func Fatal(msg string)                  { Logger.Fatal(msg) }
func Fatalf(format string, args ...any) { Logger.Fatalf(format, args...) }
