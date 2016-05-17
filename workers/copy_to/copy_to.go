package copy_to

import (
	"fmt"
	"path/filepath"

	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
	"github.com/go-zero-boilerplate/job-coordinator/utils/jobqueue"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job) error
	QueueJob(ctx *context.Context, maxGoRoutinesPerHost int, job Job, onResult OnResult)
}

func NewWorker(logger logger.Logger) Worker {
	return &copyTo{
		hq: &hostQueues{
			logger:         logger,
			hostQueues:     make(map[string]*jobqueue.Queue),
			resultHandlers: &queuedResultHandlers{handlers: make(map[*queuedJob]OnResult)},
		},
	}
}

type copyTo struct {
	hq *hostQueues
}

func (c *copyTo) getJobContext(ctx *context.Context, pendingJobFileSystem afero.Fs, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()
	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)

	//TODO: This `afero.FullBaseFsPath` is used in multiple spots, perhaps centralize?
	localJobExportDir := afero.FullBaseFsPath(pendingJobFileSystem.(*afero.BasePathFs), "")

	remoteFS := hostDetails.RemoteFileSystemFactory().New(job.Id())
	remoteJobPath := remoteFS.GetFullJobDir()
	logger := ctx.Logger.
		WithField("phase-id", job.Id()).
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
	defer jobCtx.logger.TraceDebug("Starting job").StopDebug(&err)

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

	if remoteAdditionalCacheSpecs := job.RemoteAdditionalCacheSpecs(); len(remoteAdditionalCacheSpecs) > 0 {
		jobCtx.logger.Debug("Job has additional remote caching")
		for _, spec := range remoteAdditionalCacheSpecs {
			remoteCacheDir := spec.RemoteCacheFS.GetFullJobDir()
			remoteJobSubDir := filepath.Join(jobCtx.remoteJobPath, spec.JobSubdir)
			if err = jobCtx.remoteComms.Copy(remoteCacheDir, remoteJobSubDir); err != nil {
				jobCtx.logger.WithError(err).WithField("cache-src", remoteCacheDir).WithField("cache-dest", remoteJobSubDir).Error("Cannot copy remote cache")
				return err
			}
		}
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

func (c *copyTo) QueueJob(ctx *context.Context, maxGoRoutinesPerHost int, job Job, onResult OnResult) {
	ctx.Logger.WithField("phase-id", job.Id()).Debug("TEMP: Queueing job")
	c.hq.QueueJob(c, ctx, job, onResult, maxGoRoutinesPerHost)
}
