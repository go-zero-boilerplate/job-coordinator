package post_processing

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/francoishill/afero"

	"github.com/golang-devops/exec-logger/exec_logger_dtos"
)

type Result struct {
	errors                    []string
	completedJobFileSystem    afero.Fs
	logRelativePath           string
	resourceUsageRelativePath string
	exitStatus                *exec_logger_dtos.ExitStatusDto
	localContext              *exec_logger_dtos.LocalContextDto
}

func (r *Result) appendError(err error) {
	r.errors = append(r.errors, err.Error())
}

func (r *Result) HasErrors() bool {
	return len(r.errors) > 0
}

func (r *Result) Errors() []string {
	return r.errors
}

func (r *Result) JobFileSystem() afero.Fs {
	return r.completedJobFileSystem
}

func (r *Result) OpenLogFile() (afero.File, error) {
	return r.completedJobFileSystem.Open(r.logRelativePath)
}

func (r *Result) OpenResourceUsagesFile() (afero.File, error) {
	return r.completedJobFileSystem.Open(r.resourceUsageRelativePath)
}

func (r *Result) ParseResourceUsagesFile() ([]*exec_logger_dtos.ResourceUsageDto, error) {
	fileContent, err := afero.ReadFile(r.completedJobFileSystem, r.resourceUsageRelativePath)
	if err != nil {
		return nil, fmt.Errorf("Cannot read resource-usage file ('%s'), error: %s", r.resourceUsageRelativePath, err.Error())
	}

	lines := strings.Split(string(fileContent), "\n")
	usages := []*exec_logger_dtos.ResourceUsageDto{}
	for _, l := range lines {
		trimmedLine := strings.TrimSpace(l)
		if trimmedLine == "" {
			continue
		}

		u := &exec_logger_dtos.ResourceUsageDto{}
		err = json.Unmarshal([]byte(l), u)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse resource usage line, error: %s. Line was: %s", err.Error(), trimmedLine)
		}
		usages = append(usages, u)
	}
	return usages, nil
}

func (r *Result) ExitStatus() *exec_logger_dtos.ExitStatusDto {
	return r.exitStatus
}

func (r *Result) LocalContext() *exec_logger_dtos.LocalContextDto {
	return r.localContext
}
