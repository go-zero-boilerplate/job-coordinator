package cache_syncer

import (
	"fmt"

	"github.com/francoishill/afero"
	"github.com/go-zero-boilerplate/job-coordinator/context"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job) error
}

func NewWorker() Worker {
	return &copyTo{}
}

type copyTo struct{}

func (c *copyTo) getJobContext(ctx *context.Context, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()
	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)

	remoteFS := hostDetails.RemoteFileSystemFactory().New(job.Id())
	remoteJobPath := remoteFS.GetFullJobDir()
	logger := ctx.Logger.
		WithField("phase-id", job.Id()).
		WithField("host", hostDetails.HostName()).
		WithField("remote-dir", remoteJobPath)

	jobCtx := &jobContext{
		logger:        logger,
		remoteJobPath: remoteJobPath,
		remoteComms:   remoteComms,
	}
	return jobCtx, nil
}

func (c *copyTo) runJob(jobCtx *jobContext, job Job) error {
	var err error
	defer jobCtx.logger.TraceDebug("Starting job").StopDebug(&err)

	//TODO: This `afero.FullBaseFsPath` is used in multiple spots, perhaps centralize?
	localFullCacheDir := afero.FullBaseFsPath(job.LocalJobCacheFS().(*afero.BasePathFs), "")
	remoteFinalCacheDir := jobCtx.remoteJobPath
	remotePendingCacheDir := remoteFinalCacheDir + "_pending"
	remoteOldCacheDir := remoteFinalCacheDir + "_old"
	logger := jobCtx.logger.
		WithField("local-dir", localFullCacheDir).
		WithField("remote-pending-dir", remotePendingCacheDir).
		WithField("remote-old-dir", remoteOldCacheDir).
		WithField("remote-final-dir", remoteFinalCacheDir)

	logger.Debug("Checking if paths are in sync")
	if areInSync, err := jobCtx.remoteComms.CheckPathsAreInSync(logger, localFullCacheDir, remoteFinalCacheDir); err != nil {
		logger.WithError(err).Error("Cannot ensure paths are in sync")
		//dont return here, it just means we will copy the files again even though they are in sync
	} else if areInSync {
		logger.Debug("Remote files are already in sync")
		return nil
	}

	logger.Debug("Deleting PENDING cache remote dir")
	if err = jobCtx.remoteComms.Delete(remotePendingCacheDir); err != nil {
		logger.WithError(err).Error("Cannot delete pending")
		return err
	}

	logger.Debug("Uploading to PENDING cache remote dir")
	if err = jobCtx.remoteComms.Upload(localFullCacheDir, remotePendingCacheDir); err != nil {
		logger.WithError(err).Error("Upload to pending failed")
		return err
	}

	logger.Debug("Deleting OLD cache remote dir")
	if err = jobCtx.remoteComms.Delete(remoteOldCacheDir); err != nil {
		logger.WithError(err).Error("Cannot delete old")
		return err
	}

	if finalDirInfo, err := jobCtx.remoteComms.Stats(remoteFinalCacheDir); err != nil {
		logger.WithError(err).Error("Cannot get stats for final dir (to be renamed to old)")
	} else if finalDirInfo.Exists {
		logger.Debug("Moving remote FINAL to OLD dir")
		if err = jobCtx.remoteComms.Move(remoteFinalCacheDir, remoteOldCacheDir); err != nil {
			logger.WithError(err).Error("Cannot rename pending to final")
			return err
		}
	} else {
		logger.Debug("Skipping move of remote FINAL to OLD dir (dir missing)")
	}

	logger.Debug("Moving remote PENDING to FINAL dir")
	if err = jobCtx.remoteComms.Move(remotePendingCacheDir, remoteFinalCacheDir); err != nil {
		logger.WithError(err).Error("Cannot rename pending to final")
		return err
	}

	return nil
}

func (c *copyTo) DoJob(ctx *context.Context, job Job) error {
	jobCtx, err := c.getJobContext(ctx, job)
	if err != nil {
		return fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if err := c.runJob(jobCtx, job); err != nil {
		return fmt.Errorf("Could not run job, error: %s", err.Error())
	}

	return nil
}
