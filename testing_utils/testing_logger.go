package testing_utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/debug"
)

func NewLogger() logger.Logger {
	return &TestingLogger{
		Handler: &testingLogHandler{},
		Fields:  []map[string]interface{}{},
	}
}

type testingLogLine struct {
	Fields    []map[string]interface{}
	Text      string
	Timestamp time.Time
	IsError   bool
	IsWarning bool
	IsInfo    bool
	IsDebug   bool
}

func (t *testingLogLine) String() string {
	fieldStrings := []string{}
	for _, fs := range t.Fields {
		for key, value := range fs {
			fieldStrings = append(fieldStrings, fmt.Sprintf("%s=%v", key, value))
		}
	}
	return fmt.Sprintf("[%s, E=%t, W=%t, I=%t, D=%t] %s |-| %s", t.Timestamp.String(), t.IsError, t.IsWarning, t.IsInfo, t.IsDebug, t.Text, strings.Join(fieldStrings, ", "))
}

type testingLogHandler struct {
	lock  sync.RWMutex
	Lines []*testingLogLine
}

func (t *testingLogHandler) appendLine(line *testingLogLine) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Lines = append(t.Lines, line)
}

func (t *testingLogHandler) clear() {
	t.Lines = nil
}

func (t *testingLogHandler) fullStrings(includeDebugLines bool) (lines []string) {
	for _, l := range t.Lines {
		if !includeDebugLines && l.IsDebug {
			continue
		}
		lines = append(lines, l.String())
	}
	return
}

type TestingLogger struct {
	locker  sync.RWMutex
	Handler *testingLogHandler
	Fields  []map[string]interface{}
}

func (t *TestingLogger) FullStrings(includeDebugLines bool) (lines []string) {
	return t.Handler.fullStrings(includeDebugLines)
}

func (t *TestingLogger) Clear() {
	t.Handler.clear()
}

func (t *TestingLogger) appendLine(line testingLogLine) {
	line.Fields = t.Fields
	t.Handler.appendLine(&line)
}

func (t *TestingLogger) Emergency(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsError: true})
}
func (t *TestingLogger) Alert(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsError: true})
}
func (t *TestingLogger) Critical(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsError: true})
}
func (t *TestingLogger) Error(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsError: true})
}
func (t *TestingLogger) Warn(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsWarning: true})
}
func (t *TestingLogger) Notice(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsInfo: true})
}
func (t *TestingLogger) Info(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsInfo: true})
}
func (t *TestingLogger) Debug(format string, params ...interface{}) {
	t.appendLine(testingLogLine{Text: fmt.Sprintf(format, params...), Timestamp: time.Now(), IsDebug: true})
}

type logTracer struct {
	startTime time.Time
	logger    *TestingLogger
	msg       string
}

func (l *logTracer) Stop(errPtr *error) {
	if *errPtr == nil {
		l.logger.WithField("duration", time.Since(l.startTime)).Info(l.msg)
	} else {
		l.logger.WithField("duration", time.Since(l.startTime)).WithError(*errPtr).Error(l.msg)
	}
}

func (t *TestingLogger) Trace(format string, params ...interface{}) logger.LogTracer {
	msg := fmt.Sprintf(format, params...)
	t.Info(format, params...)
	return &logTracer{startTime: time.Now(), logger: t, msg: msg}
}

func (t *TestingLogger) WithField(key string, value interface{}) logger.Logger {
	fields := map[string]interface{}{}
	fields[key] = value
	return t.WithFields(fields)
}
func (t *TestingLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return &TestingLogger{
		Handler: t.Handler,
		Fields:  append(t.Fields, fields),
	}
}
func (t *TestingLogger) WithError(err error) logger.Logger {
	return t.WithField("error", err)
}

func (t *TestingLogger) DeferredRecoverStack(debugMessage string) {
	if r := recover(); r != nil {
		logger := t.WithField("recovery", r).WithField("debug", debugMessage)
		stack := strings.Replace(debug.GetFullStackTrace_Normal(false), "\n", "\\n", -1)
		logger = logger.WithField("stack", stack)
		logger.Alert("Unhandled panic recovered")
	}
}
