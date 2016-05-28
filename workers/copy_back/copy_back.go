package copy_back

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job, handlers Handlers) error
}

func NewWorker() Worker {
	return &copyBack{}
}

type copyBack struct{}

func (c *copyBack) getJobContext(ctx *context.Context, completedJobFileSystem, oldCompletedJobFileSystem afero.Fs, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()
	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)

	//TODO: This `afero.FullBaseFsPath` is used in multiple spots, perhaps centralize?
	fullCompletedJobPath := afero.FullBaseFsPath(completedJobFileSystem.(*afero.BasePathFs), "")
	fullOldCompletedJobPath := afero.FullBaseFsPath(oldCompletedJobFileSystem.(*afero.BasePathFs), "")

	remoteFS := hostDetails.RemoteFileSystemFactory().New(job.Id())
	remoteJobPath := remoteFS.GetFullJobDir()
	logger := ctx.Logger.
		WithField("phase-id", job.Id()).
		WithField("local-dir", fullCompletedJobPath).
		WithField("local-old-dir", fullOldCompletedJobPath).
		WithField("host", hostDetails.HostName()).
		WithField("remote-dir", remoteJobPath)

	jobCtx := &jobContext{
		logger:                    logger,
		completedJobFileSystem:    completedJobFileSystem,
		oldCompletedJobFileSystem: oldCompletedJobFileSystem,
		fullCompletedJobPath:      fullCompletedJobPath,
		fullOldCompletedJobPath:   fullOldCompletedJobPath,
		remoteJobPath:             remoteJobPath,
		remoteComms:               remoteComms,
	}
	return jobCtx, nil
}

func (c *copyBack) moveCompletedToOld(jobCtx *jobContext) error {
	if err := jobCtx.oldCompletedJobFileSystem.RemoveAll("."); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot remove old dir")
		return err
	}

	if completedExists, err := afero.Exists(jobCtx.completedJobFileSystem, "."); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot check complete dir exists")
		return err
	} else if completedExists {
		if renameErr := os.Rename(jobCtx.fullCompletedJobPath, jobCtx.fullOldCompletedJobPath); renameErr != nil {
			jobCtx.logger.WithError(renameErr).Error("Cannot move job to old dir")
			return err
		}
	}

	return nil
}

func (c *copyBack) runJob(jobCtx *jobContext, job Job, handlers Handlers) error {
	var err error
	defer jobCtx.logger.TraceDebug("Starting job").StopDebug(&err)

	//Do this instead of ping
	err = jobCtx.remoteComms.ConfirmVersionMatch(job.HostDetails().ExpectedGopsexecVersion())
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot confirm GoPsexec version")
		return err
	}

	err = c.moveCompletedToOld(jobCtx)
	if err != nil {
		return err
	}

	//Cleanup before copying files back (remove binaries, etc)
	for _, relRemotePathToDelete := range job.RemoteCleanupSpec().RelativePathsToDelete {
		fullRemotePath := filepath.Join(jobCtx.remoteJobPath, relRemotePathToDelete)
		if err := jobCtx.remoteComms.Delete(fullRemotePath); err != nil {
			jobCtx.logger.WithError(err).WithField("remote-path", fullRemotePath).Error("Cannot delete remote path")
			handlers.FailedToCleanupRemotePathBeforeCopyBack(err, fullRemotePath)
			//Do not return error, we are able to continue
		}
	}

	err = jobCtx.remoteComms.DownloadDir(jobCtx.remoteJobPath, jobCtx.fullCompletedJobPath)
	if err != nil {
		jobCtx.logger.WithError(err).Error("Copy failed")
		return err
	}

	//Cleanup the remote dir after successfully copying back
	if err = jobCtx.remoteComms.Delete(jobCtx.remoteJobPath); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot remove remote job dir")
		handlers.FailedToRemoveRemoteJobDir(err, jobCtx.remoteJobPath)
		//Do not return error, we are able to continue
	} else {
		jobCtx.logger.Info("Successfully deleted remote job dir")
	}

	return nil
}

func (c *copyBack) DoJob(ctx *context.Context, job Job, handlers Handlers) error {
	completedJobFileSystem := job_helpers.GetJobFileSystem(ctx.CompletedLocalFileSystem, job.Id())
	oldCompletedJobFileSystem := job_helpers.GetJobFileSystem(ctx.CompletedLocalFileSystem, job.Id()+"_old")
	jobCtx, err := c.getJobContext(ctx, completedJobFileSystem, oldCompletedJobFileSystem, job)
	if err != nil {
		return fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if err := c.runJob(jobCtx, job, handlers); err != nil {
		return fmt.Errorf("Could not run job, error: %s", err.Error())
	}

	return nil
}
