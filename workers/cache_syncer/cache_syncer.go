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

	logger := ctx.Logger.
		WithField("host", hostDetails.HostName())

	jobCtx := &jobContext{
		logger:      logger,
		remoteComms: remoteComms,
	}
	return jobCtx, nil
}

func (c *copyTo) runJob(jobCtx *jobContext, job Job) error {
	var err error
	defer jobCtx.logger.Trace("Starting job").Stop(&err)

	//TODO: This `afero.FullBaseFsPath` is used in multiple spots, perhaps centralize?
	localFullCacheDir := afero.FullBaseFsPath(job.LocalSourceCacheFS().(*afero.BasePathFs), "")
	remoteFinalCacheDir := job.RemoteDestCacheFS().GetFullPath()
	remotePendingCacheDir := remoteFinalCacheDir + "_pending"
	remoteOldCacheDir := remoteFinalCacheDir + "_old"
	logger := jobCtx.logger.
		WithField("local-dir", localFullCacheDir).
		WithField("remote-pending-dir", remotePendingCacheDir).
		WithField("remote-old-dir", remoteOldCacheDir).
		WithField("remote-final-dir", remoteFinalCacheDir)

	if areInSync, err := jobCtx.remoteComms.CheckPathsAreInSync(localFullCacheDir, remoteFinalCacheDir); err != nil {
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

	logger.Debug("Moving remote FINAL to OLD dir")
	if err = jobCtx.remoteComms.Move(remoteFinalCacheDir, remoteOldCacheDir); err != nil {
		logger.WithError(err).Error("Cannot rename pending to final")
		return err
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