package starter

import (
	"fmt"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/utils/exec_logger_helpers"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

type Worker interface {
	DoJob(ctx *context.Context, job Job) error
}

func NewWorker() Worker {
	return &starter{}
}

type starter struct{}

func (s *starter) getJobContext(ctx *context.Context, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()
	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)

	remoteJobFS := hostDetails.RemoteFileSystemFactory().New(job.Id())
	remoteJobPath := remoteJobFS.GetFullJobDir()
	logger := ctx.Logger.
		WithField("job", job.Id()).
		WithField("host", hostDetails.HostName()).
		WithField("remote-dir", remoteJobPath)

	jobCtx := &jobContext{
		logger:        logger,
		remoteJobFS:   remoteJobFS,
		remoteJobPath: remoteJobPath,
		remoteComms:   remoteComms,
	}
	return jobCtx, nil
}

func (s *starter) runJob(jobCtx *jobContext, job Job) (*remote_comms_facade.StartedDetails, error) {
	var err error
	defer jobCtx.logger.Trace("Starting job").Stop(&err)

	//Do this instead of ping
	err = jobCtx.remoteComms.ConfirmVersionMatch(job.HostDetails().ExpectedGopsexecVersion())
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot confirm GoPsexec version")
		return nil, err
	}

	workingDir := jobCtx.remoteJobPath
	commandLine, err := job.Commandline(jobCtx.remoteJobFS)
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot get commandline")
		return nil, err
	}

	osType, err := jobCtx.remoteComms.GetOsType()
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot get OsType")
		return nil, err
	}

	timeout := job.Timeout()
	recordResourceUsage := job.RecordResourceUsage()

	allExecLoggerCommandLine := exec_logger_helpers.CombineExecLoggerCommandline(osType, jobCtx.remoteJobFS, commandLine, timeout, recordResourceUsage)
	startedDetails, err := jobCtx.remoteComms.StartDetached(workingDir, allExecLoggerCommandLine...)
	if err != nil {
		jobCtx.logger.WithError(err).Error("Cannot start remote command")
		return nil, err
	}

	jobCtx.logger.Info("Started command with Pid %d", startedDetails.Pid)

	return startedDetails, nil
}

func (s *starter) DoJob(ctx *context.Context, job Job) error {
	jobCtx, err := s.getJobContext(ctx, job)
	if err != nil {
		return fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if _, err := s.runJob(jobCtx, job); err != nil {
		return fmt.Errorf("Could not run job, error: %s", err.Error())
	}

	return nil
}
