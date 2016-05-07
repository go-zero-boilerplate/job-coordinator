package exec_logger_helpers

import (
	"github.com/go-zero-boilerplate/osvisitors"
)

func GetExecLoggerBinFileName(osType osvisitors.OsType) string {
	visitor := &execLoggerBinFileNameVisitor{}
	osType.Accept(visitor)
	return visitor.answer
}

type execLoggerBinFileNameVisitor struct{ answer string }

func (e *execLoggerBinFileNameVisitor) VisitWindows() { e.answer = "exec-logger.exe" }
func (e *execLoggerBinFileNameVisitor) VisitLinux()   { e.answer = "exec-logger" }
func (e *execLoggerBinFileNameVisitor) VisitDarwin()  { e.answer = "exec-logger" }
