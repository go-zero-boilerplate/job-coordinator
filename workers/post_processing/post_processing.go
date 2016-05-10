package post_processing

import (
	"encoding/json"
	"fmt"

	"github.com/francoishill/afero"
	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job) (*Result, error)
}

func NewWorker() Worker {
	return &postProcessing{}
}

type postProcessing struct{}

func (p *postProcessing) getJobContext(ctx *context.Context, completedJobFileSystem afero.Fs, job Job) (*jobContext, error) {
	logger := ctx.Logger.
		WithField("phase-id", job.Id())

	jobCtx := &jobContext{
		logger:                    logger,
		completedJobFileSystem:    completedJobFileSystem,
		exitedRelativePath:        exec_logger_constants.EXITED_FILE_NAME,
		logRelativePath:           exec_logger_constants.LOG_FILE_NAME,
		resourceUsageRelativePath: exec_logger_constants.RECORD_RESOURCE_USAGE_FILE_NAME,
		localContextRelativePath:  exec_logger_constants.LOCAL_CONTEXT_FILE_NAME,
	}
	return jobCtx, nil
}

func (p *postProcessing) runJob(jobCtx *jobContext, job Job) (*Result, error) {
	var err error
	defer jobCtx.logger.Trace("Starting job").Stop(&err)

	exitedDtoContent, err := afero.ReadFile(jobCtx.completedJobFileSystem, jobCtx.exitedRelativePath)
	if err != nil {
		jobCtx.logger.WithError(err).WithField("exit-file", jobCtx.exitedRelativePath).Error("Cannot read exit file")
		return nil, err
	}

	exitedDto := &exec_logger_dtos.ExitStatusDto{}
	if unmarshalError := json.Unmarshal(exitedDtoContent, exitedDto); unmarshalError != nil {
		jobCtx.logger.WithError(unmarshalError).WithField("exit-file", jobCtx.exitedRelativePath).Error("Cannot unmarshal exit file json")
		return nil, unmarshalError
	}

	localContextContent, err := afero.ReadFile(jobCtx.completedJobFileSystem, jobCtx.localContextRelativePath)
	if err != nil {
		jobCtx.logger.WithError(err).WithField("context-file", jobCtx.localContextRelativePath).Error("Cannot read local-context file")
		return nil, err
	}

	localContext := &exec_logger_dtos.LocalContextDto{}
	if unmarshalError := json.Unmarshal(localContextContent, localContext); unmarshalError != nil {
		jobCtx.logger.WithError(unmarshalError).WithField("context-file", jobCtx.localContextRelativePath).Error("Cannot unmarshal local-context file json")
		return nil, unmarshalError
	}

	result := &Result{
		completedJobFileSystem:    jobCtx.completedJobFileSystem,
		logRelativePath:           jobCtx.logRelativePath,
		resourceUsageRelativePath: jobCtx.resourceUsageRelativePath,
		exitStatus:                exitedDto,
		localContext:              localContext,
	}

	return result, nil
}

func (p *postProcessing) DoJob(ctx *context.Context, job Job) (*Result, error) {
	completedJobFileSystem := job_helpers.GetJobFileSystem(ctx.CompletedLocalFileSystem, job.Id())
	jobCtx, err := p.getJobContext(ctx, completedJobFileSystem, job)
	if err != nil {
		return nil, fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if result, err := p.runJob(jobCtx, job); err != nil {
		jobCtx.logger.WithError(err).Error("Could not run job")
		return nil, fmt.Errorf("Could not run job, error: %s", err.Error())
	} else {
		return result, nil
	}
}
