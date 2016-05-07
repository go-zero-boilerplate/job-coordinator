package logger

import (
	"github.com/go-zero-boilerplate/loggers"
)

type Logger interface {
	loggers.LoggerRFC5424
	Trace(format string, params ...interface{}) LogTracer

	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger

	DeferredRecoverStack(debugMessage string)
}
