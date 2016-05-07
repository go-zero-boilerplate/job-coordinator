package remote_comms_facade

import (
	gpClient "github.com/golang-devops/go-psexec/client"
)

type startExecVisitor struct {
	session gpClient.Session

	detached   bool
	exe        string
	workingDir string
	args       []string

	resp *gpClient.ExecResponse
	err  error
}

func (s *startExecVisitor) setFromExecBuilder(builder gpClient.SessionExecRequestBuilder) {
	if s.detached {
		builder = builder.Detached()
	}
	s.resp, s.err = builder.Exe(s.exe).Args(s.args...).WorkingDir(s.workingDir).BuildAndDoRequest()
}
func (s *startExecVisitor) VisitWindows() {
	s.setFromExecBuilder(s.session.ExecRequestBuilder().Winshell())
}
func (s *startExecVisitor) VisitLinux() {
	s.setFromExecBuilder(s.session.ExecRequestBuilder().Bash())
}
func (s *startExecVisitor) VisitDarwin() {
	s.VisitLinux()
}
