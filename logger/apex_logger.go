package logger

import (
	"fmt"
	"strings"

	apex "github.com/apex/log"

	"github.com/go-zero-boilerplate/job-coordinator/utils/debug"
)

func NewApexLogger(level apex.Level, handler apex.Handler, apexEntry *apex.Entry) Logger {
	apex.SetLevel(level)
	apex.SetHandler(handler)

	return &apexLogger{
		level:          level,
		handler:        handler,
		apexEntry:      apexEntry,
		errStackTraces: true, //TODO: Is this fine by default?
	}
}

type apexLogger struct {
	level          apex.Level
	handler        apex.Handler
	apexEntry      *apex.Entry
	errStackTraces bool
}

func (l *apexLogger) Emergency(format string, params ...interface{}) {
	l.apexEntry.Fatalf(format, params...)
}
func (l *apexLogger) Alert(format string, params ...interface{}) {
	l.apexEntry.Errorf(format, params...)
}
func (l *apexLogger) Critical(format string, params ...interface{}) {
	l.apexEntry.Errorf(format, params...)
}
func (l *apexLogger) Error(format string, params ...interface{}) {
	l.apexEntry.Errorf(format, params...)
}
func (l *apexLogger) Warn(format string, params ...interface{}) {
	l.apexEntry.Warnf(format, params...)
}
func (l *apexLogger) Notice(format string, params ...interface{}) {
	l.apexEntry.Warnf(format, params...)
}
func (l *apexLogger) Info(format string, params ...interface{}) {
	l.apexEntry.Infof(format, params...)
}
func (l *apexLogger) Debug(format string, params ...interface{}) {
	l.apexEntry.Debugf(format, params...)
}
func (l *apexLogger) Trace(format string, params ...interface{}) LogTracer {
	return l.apexEntry.Trace(fmt.Sprintf(format, params...))
}

func (l *apexLogger) WithError(err error) Logger {
	newEntry := l.apexEntry.WithError(err)
	if l.errStackTraces {
		stack := strings.Replace(debug.GetFullStackTrace_Normal(false), "\n", "\\n", -1)
		newEntry = newEntry.WithField("stack", stack)
	}
	return NewApexLogger(l.level, l.handler, newEntry)
}

func (l *apexLogger) WithField(key string, value interface{}) Logger {
	return NewApexLogger(l.level, l.handler, l.apexEntry.WithField(key, value))
}

func (l *apexLogger) WithFields(fields map[string]interface{}) Logger {
	return NewApexLogger(l.level, l.handler, l.apexEntry.WithFields(apex.Fields(fields)))
}

func (l *apexLogger) DeferredRecoverStack(debugMessage string) {
	if r := recover(); r != nil {
		logger := l.WithField("recovery", r).WithField("debug", debugMessage)
		stack := strings.Replace(debug.GetFullStackTrace_Normal(false), "\n", "\\n", -1)
		logger = logger.WithField("stack", stack)
		logger.Alert("Unhandled panic recovered")
	}
}
