package monitor

import "github.com/golang-devops/exec-logger/exec_logger_dtos"

type Status struct {
	IsAlive    bool
	ExitStatus *exec_logger_dtos.ExitStatusDto
}
