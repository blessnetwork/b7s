package hclog

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

const (
	nameField = "module"
)

// Logger is a zerolog logger that fulfils the `go-hclog` Logger interface.
type Logger struct {
	name   string
	logger zerolog.Logger
}

// Ensure that we conform to the HC Logger interface.
var _ hclog.Logger = (*Logger)(nil)

// New creates a new Raft logger from a zerolog logger.
func New(logger zerolog.Logger) *Logger {

	l := Logger{
		logger: logger,
	}

	return &l
}

// IsTrace returns true if the logger should log Trace level messages.
func (l *Logger) IsTrace() bool {
	return l.shouldLog(zerolog.TraceLevel)
}

// IsDebug returns true if the logger should log Debug level messages.
func (l *Logger) IsDebug() bool {
	return l.shouldLog(zerolog.DebugLevel)
}

// IsInfo returns true if the logger should log Info level messages.
func (l *Logger) IsInfo() bool {
	return l.shouldLog(zerolog.InfoLevel)
}

// IsWarn returns true if the logger should log Warn level messages.
func (l *Logger) IsWarn() bool {
	return l.shouldLog(zerolog.WarnLevel)
}

// IsError returns true if the logger should log Error level messages.
func (l *Logger) IsError() bool {
	return l.shouldLog(zerolog.ErrorLevel)
}

// shouldLog return true if the logger level is higher than the specified level.
func (l *Logger) shouldLog(level zerolog.Level) bool {
	return l.logger.GetLevel() <= level
}

// SetLevel sets the log level for the logger.
func (l *Logger) SetLevel(level hclog.Level) {
	l.logger = l.logger.Level(hcToZerologLevel(level))
}

// GetLevel returns the current log level for the logger.
func (l *Logger) GetLevel() hclog.Level {

	switch l.logger.GetLevel() {

	case zerolog.Disabled:
		return hclog.Off
	case zerolog.TraceLevel:
		return hclog.Trace
	case zerolog.DebugLevel:
		return hclog.Debug
	case zerolog.InfoLevel:
		return hclog.Info
	case zerolog.WarnLevel:
		return hclog.Warn
	case zerolog.ErrorLevel:
		return hclog.Error

	default:
		// Return highest hclog level as a catch-all.
		return hclog.Error
	}
}

func hcToZerologLevel(level hclog.Level) zerolog.Level {

	switch level {

	case hclog.Off:
		return zerolog.Disabled

	case hclog.Trace:
		return zerolog.TraceLevel

	case hclog.Debug:
		return zerolog.DebugLevel

	case hclog.Info:
		return zerolog.InfoLevel

	case hclog.Warn:
		return zerolog.WarnLevel

	case hclog.Error:
		return zerolog.ErrorLevel

	default:
		// Should not happen but let's not err here.
		return zerolog.ErrorLevel
	}

}

// Trace logs a message with a trace level.
func (l *Logger) Trace(msg string, args ...interface{}) {
	l.log(zerolog.TraceLevel, msg, args...)
}

// Debug logs a message with a debug level.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(zerolog.DebugLevel, msg, args...)
}

// Info logs a message with an info level.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(zerolog.InfoLevel, msg, args...)
}

// Warn logs a message with a warn level.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(zerolog.WarnLevel, msg, args...)
}

// Error logs a message with an error level.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(zerolog.ErrorLevel, msg, args...)
}

// Log logs a message with the specified level.
func (l *Logger) Log(level hclog.Level, msg string, args ...interface{}) {
	lvl := hcToZerologLevel(level)
	l.log(lvl, msg, args...)
}

func (l *Logger) log(level zerolog.Level, msg string, args ...interface{}) {
	l.logger.WithLevel(level).Fields(args).Msg(msg)
}

// StandardLogger returns a value that conforms to the stdlib log.Logger interface
func (l *Logger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {

	stdlog := log.Default()
	stdlog.SetFlags(0)
	stdlog.SetOutput(l.logger)

	return stdlog
}

// StandardWriter return a value that conforms to io.Writer, which can be passed into log.SetOutput().
func (l *Logger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return l.logger
}

// ImpliedArgs return key/value pairs associated with the current logger. Not supported at the moment.
func (l *Logger) ImpliedArgs() []interface{} {
	return nil
}

// Named returns a named logger.
func (l *Logger) Named(name string) hclog.Logger {

	newLoggerName := name
	current := l.Name()
	if current != "" {
		// If the current logger already has a name, keep it as a prefix.
		newLoggerName = current + "." + name
	}

	logger := New(
		l.logger.With().Str(nameField, newLoggerName).Logger(),
	)
	logger.name = newLoggerName

	return logger
}

// Named returns a named logger.
func (l *Logger) ResetNamed(name string) hclog.Logger {

	logger := New(
		l.logger.With().Str(nameField, name).Logger(),
	)
	logger.name = name

	return logger
}

// Name returns the current name for the logger.
func (l *Logger) Name() string {
	return l.name
}

// With returns a logger with the specified fields.
func (l *Logger) With(args ...interface{}) hclog.Logger {

	logger := New(
		l.logger.With().Fields(args).Logger(),
	)

	return logger
}
