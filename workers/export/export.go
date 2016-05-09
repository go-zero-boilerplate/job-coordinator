package export

import (
	"fmt"
	"os"

	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/utils/exec_logger_helpers"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job) error
}

func NewWorker() Worker {
	return &export{}
}

type export struct{}

func (e *export) getJobContext(ctx *context.Context, pendingJobFileSystem afero.Fs, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()

	logger := ctx.Logger.
		WithField("job", job.Id())

	jobCtx := &jobContext{
		logger:                 logger,
		pendingJobFileSystem:   pendingJobFileSystem,
		localExecLoggerBinPath: ctx.LocalExecLoggerBinPath,
		remoteComms:            ctx.RemoteCommsFactory.NewFacade(hostDetails),
	}
	return jobCtx, nil
}

func (e *export) runJob(jobCtx *jobContext, job Job) error {
	var err error
	defer jobCtx.logger.Trace("Starting job").Stop(&err)

	//Clean possible folder before exporting
	if err = jobCtx.pendingJobFileSystem.RemoveAll("."); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot remove local export dir")
		return err
	} else {
		jobCtx.logger.Info("Successfully ensured export dir gone")
	}

	additional := job.AdditionalCachedSpecs()
	if additional != nil {
		//TODO: This `afero.FullBaseFsPath` is used in multiple spots, perhaps centralize?
		localFullCacheDir := afero.FullBaseFsPath(additional.LocalFS.(*afero.BasePathFs), "")
		remoteFullCacheDir := additional.RemoteCacheFS.GetFullPath()
		err = jobCtx.remoteComms.Upload(localFullCacheDir, remoteFullCacheDir)
		if err != nil {
			jobCtx.logger.WithError(err).WithField("local-cache-dir", localFullCacheDir).WithField("remote-cache-dir", remoteFullCacheDir).Error("Upload cache failed")
			return err
		}
	}

	if err := job.ExportFiles(jobCtx.pendingJobFileSystem); err != nil {
		jobCtx.logger.WithError(err).Error("Export files failed")
		return err
	}

	execLoggerFile, err := os.Open(jobCtx.localExecLoggerBinPath)
	if err != nil {
		jobCtx.logger.WithField("local-exec-logger", jobCtx.localExecLoggerBinPath).WithError(err).Error("Cannot open file")
		return err
	}
	defer execLoggerFile.Close()

	osType, err := jobCtx.remoteComms.GetOsType()
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot get OsType")
		return err
	}

	execLoggerBinFileName := exec_logger_helpers.GetExecLoggerBinFileName(osType)
	if err = SaveFile(jobCtx.pendingJobFileSystem, execLoggerBinFileName, execLoggerFile); err != nil {
		jobCtx.logger.WithField("local-exec-logger", jobCtx.localExecLoggerBinPath).WithError(err).Error("Cannot copy file")
		return err
	}

	return nil
}

func (e *export) DoJob(ctx *context.Context, job Job) error {
	pendingJobFileSystem := job_helpers.GetJobFileSystem(ctx.PendingLocalFileSystem, job.Id())
	jobCtx, err := e.getJobContext(ctx, pendingJobFileSystem, job)
	if err != nil {
		return fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if err = e.runJob(jobCtx, job); err != nil {
		jobCtx.logger.WithError(err).Error("Could not run job")
		return fmt.Errorf("Could not run job, error: %s", err.Error())
	}

	return nil
}
