package copy_to

import (
	"fmt"

	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job) error
}

func NewWorker() Worker {
	return &copyTo{}
}

type copyTo struct{}

func (c *copyTo) getJobContext(ctx *context.Context, pendingJobFileSystem afero.Fs, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()
	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)
	remoteTempDir, err := remoteComms.GetTempDir()
	if err != nil {
		return nil, fmt.Errorf("Unable to get remote temp dir for host '%s', error: %s", hostDetails.HostName(), err.Error())
	}

	//TODO: This `afero.FullBaseFsPath` is used in multiple spots, perhaps centralize?
	localJobExportDir := afero.FullBaseFsPath(pendingJobFileSystem.(*afero.BasePathFs), "")

	remoteFS := hostDetails.RemoteFileSystemFactory().New(remoteTempDir, job.Id())
	remoteJobPath := remoteFS.GetFullJobDir()
	logger := ctx.Logger.
		WithField("job", job.Id()).
		WithField("local-dir", localJobExportDir).
		WithField("host", hostDetails.HostName()).
		WithField("remote-dir", remoteJobPath)

	jobCtx := &jobContext{
		logger:               logger,
		pendingJobFileSystem: pendingJobFileSystem,
		localExportDir:       localJobExportDir,
		remoteJobPath:        remoteJobPath,
		remoteComms:          remoteComms,
	}
	return jobCtx, nil
}

func (c *copyTo) runJob(jobCtx *jobContext, job Job) error {
	var err error
	defer jobCtx.logger.Trace("Starting job").Stop(&err)

	//Do this instead of ping
	err = jobCtx.remoteComms.ConfirmVersionMatch(job.HostDetails().ExpectedGopsexecVersion())
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot confirm GoPsexec version")
		return err
	}

	if err = jobCtx.remoteComms.Delete(jobCtx.remoteJobPath); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot remove remote job dir")
		return err
	} else {
		jobCtx.logger.Info("Successfully ensured remove job dir gone")
	}

	err = jobCtx.remoteComms.Upload(jobCtx.localExportDir, jobCtx.remoteJobPath)
	if err != nil {
		jobCtx.logger.WithError(err).Error("Copy failed")
		return err
	}

	if err = jobCtx.pendingJobFileSystem.RemoveAll("."); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot remove local export dir")
		//TODO: no need to return, we can still continue on this error, could potentially be cleaned by a "cleanup worker"?
	} else {
		jobCtx.logger.Info("Successfully deleted local export dir")
	}

	return nil
}

func (c *copyTo) DoJob(ctx *context.Context, job Job) error {
	pendingJobFileSystem := job_helpers.GetJobFileSystem(ctx.PendingLocalFileSystem, job.Id())
	jobCtx, err := c.getJobContext(ctx, pendingJobFileSystem, job)
	if err != nil {
		return fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if err := c.runJob(jobCtx, job); err != nil {
		return fmt.Errorf("Could not run job, error: %s", err.Error())
	}

	return nil
}
