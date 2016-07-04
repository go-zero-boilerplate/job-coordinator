package monitor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"

	"github.com/go-zero-boilerplate/job-coordinator/context"
)

type Worker interface {
	GetJobContext(ctx *context.Context, job Job) (*jobContext, error)
	CheckCurrentStatus(jobCtx *jobContext) (*Status, error)
	RunJobAndWaitWhileAlive(ctx *context.Context, job Job) error
}

func NewWorker() Worker {
	return &monitor{}
}

type monitor struct{}

func (m *monitor) GetJobContext(ctx *context.Context, job Job) (*jobContext, error) {
	hostDetails := job.HostDetails()
	logger := ctx.Logger.
		WithField("phase-id", job.Id()).
		WithField("host", hostDetails.HostName())
		//WithField("pid", job.Pid())

	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)

	remoteJobFS := hostDetails.RemoteFileSystemFactory().New(job.Id())

	remoteLogFile := remoteJobFS.GetFullJobDir(exec_logger_constants.LOG_FILE_NAME)
	aliveFilePath := remoteJobFS.GetFullJobDir(exec_logger_constants.ALIVE_FILE_NAME)
	exitedFilePath := remoteJobFS.GetFullJobDir(exec_logger_constants.EXITED_FILE_NAME)
	requestAbortFilePath := remoteJobFS.GetFullJobDir(exec_logger_constants.MUST_ABORT_FILE_NAME)

	jobCtx := &jobContext{
		logger:               logger,
		remoteComms:          remoteComms,
		remoteLogFile:        remoteLogFile,
		aliveFilePath:        aliveFilePath,
		exitedFilePath:       exitedFilePath,
		requestAbortFilePath: requestAbortFilePath,
		initialSleepDelay:    1 * time.Second,
	}
	return jobCtx, nil
}

func (m *monitor) readExitStatusFile(jobCtx *jobContext) (*exec_logger_dtos.ExitStatusDto, error) {
	exitedFileContent, err := jobCtx.remoteComms.ReadFileContent(jobCtx.exitedFilePath)
	if err != nil {
		return nil, fmt.Errorf("Cannot read exit file '%s', error: %s", jobCtx.exitedFilePath, err.Error())
	}

	exitedDto := &exec_logger_dtos.ExitStatusDto{}
	if unmarshalError := json.Unmarshal(exitedFileContent, exitedDto); unmarshalError != nil {
		return nil, unmarshalError
	} else {
		return exitedDto, nil
	}
}

func (m *monitor) readAliveTime(jobCtx *jobContext) (*aliveTime, error) {
	aliveFileContent, err := jobCtx.remoteComms.ReadFileContent(jobCtx.aliveFilePath)
	if err != nil {
		return nil, fmt.Errorf("Cannot read alive file '%s', error: %s", jobCtx.aliveFilePath, err.Error())
	}

	timeStr := strings.TrimSpace(string(aliveFileContent))
	if aliveTimeStamp, err := time.Parse(exec_logger_constants.ALIVE_TIME_FORMAT, timeStr); err != nil {
		return nil, fmt.Errorf("Cannot parse alive time '%s', error: %s", timeStr, err.Error())
	} else {
		return &aliveTime{time: aliveTimeStamp}, nil
	}
}

func (m *monitor) notifyAbort(jobCtx *jobContext, message string) error {
	abortLines := []string{
		fmt.Sprintf("TIME: %s", time.Now().UTC().Format(exec_logger_constants.ALIVE_TIME_FORMAT)),
		fmt.Sprintf("MESSAGE: %s", message),
	}
	err := jobCtx.remoteComms.UploadFileContent(jobCtx.requestAbortFilePath, []byte(strings.Join(abortLines, "\n")))
	if err != nil {
		return fmt.Errorf("Cannot upload abort file to remote path '%s', error: %s", jobCtx.requestAbortFilePath, err.Error())
	}
	return nil
}

func (m *monitor) CheckCurrentStatus(jobCtx *jobContext) (*Status, error) {
	aliveErrCounter := &errorCounter{max: 5}
	durationAfterWhichAssumeDead := 5 * time.Minute

RETRY_ON_ALIVE_ERROR:
	if exitStatus, err := m.readExitStatusFile(jobCtx); err == nil {
		jobCtx.logger.Debug("Exit file found (exit code %d)", exitStatus.ExitCode)
		// Got the exit status file, assume the process is finished and no need to check alive file
		return &Status{ExitStatus: exitStatus}, nil
	}

	aliveTime, err := m.readAliveTime(jobCtx)
	if err != nil {
		aliveErrCounter.Inc()

		if aliveErrCounter.CapReached() {
			jobCtx.logger.WithError(err).Warn("Consecutive alive error %d/%d", aliveErrCounter.current, aliveErrCounter.max)
		} else {
			jobCtx.logger.WithError(err).Info("Consecutive alive error %d/%d", aliveErrCounter.current, aliveErrCounter.max)
		}

		if aliveErrCounter.CapReached() {
			notifyMessage := fmt.Sprintf("Failed too many times checking if alive (%d/%d)", aliveErrCounter.current, aliveErrCounter.max)
			if notifyErr := m.notifyAbort(jobCtx, notifyMessage); notifyErr != nil {
				jobCtx.logger.WithError(notifyErr).Warn("Failure to 'notify abort'")
				//do not return, we can continue still if failed to notify remote machine to abort
			}

			return nil, fmt.Errorf("Failed too many times checking if alive (%d/%d)", aliveErrCounter.current, aliveErrCounter.max)
		}

		time.Sleep(aliveErrCounter.GetSleepDuration())
		goto RETRY_ON_ALIVE_ERROR
	}

	durationAgo := time.Now().UTC().Sub(aliveTime.time)
	if durationAgo > durationAfterWhichAssumeDead {
		jobCtx.logger.WithField("duration", durationAgo.String()).Warn("Assume process dead (tolerance %s)", durationAfterWhichAssumeDead.String())
		if exitStatus, readExitStatusErr := m.readExitStatusFile(jobCtx); readExitStatusErr == nil {
			// If we successfully read the exit status, "ignore" the alive file error
			return &Status{ExitStatus: exitStatus}, nil
		}

		notifyMessage := fmt.Sprintf("Assumed process died (tolerance was %s)", durationAfterWhichAssumeDead.String())
		if notifyErr := m.notifyAbort(jobCtx, notifyMessage); notifyErr != nil {
			jobCtx.logger.WithError(notifyErr).Warn("Failure to 'notify abort'")
			//do not return, we can continue still if failed to notify remote machine to abort
		}

		return nil, fmt.Errorf("Assuming process is dead (tolerance %s)", durationAfterWhichAssumeDead.String())
	}

	return &Status{IsAlive: true}, nil
}

func (m *monitor) waitWhileAlive(jobCtx *jobContext) (*exec_logger_dtos.ExitStatusDto, error) {
	sleepDurationBetweenNormalChecks := 2 * time.Second //Note the sleep duration is increased if we sleep withing "alive error checks" - see `aliveErrCounter.GetSleepDuration()` below

	for {
		currentStatus, err := m.CheckCurrentStatus(jobCtx)
		if err != nil {
			return nil, err
		}

		if !currentStatus.IsAlive {
			return currentStatus.ExitStatus, nil
		}

		jobCtx.logger.Debug("No issues. Process seems to be alive, sleeping for %s", sleepDurationBetweenNormalChecks.String())
		time.Sleep(sleepDurationBetweenNormalChecks)
	}
}

func (m *monitor) runJobAndWaitWhileAlive(jobCtx *jobContext, job Job) error {
	var err error
	defer jobCtx.logger.TraceDebug("Starting job").StopDebug(&err)

	time.Sleep(jobCtx.initialSleepDelay)

	if exitStatus, err := m.waitWhileAlive(jobCtx); err != nil {
		jobCtx.logger.WithError(err).Error("Cannot wait while alive")
		return err
	} else if exitStatus.HasError() {
		jobCtx.logger.Error("Command exited with error: %s. ExitCode was %d", exitStatus.Error, exitStatus.ExitCode)
		return err
	} else if exitStatus.ExitCode != 0 {
		jobCtx.logger.Error("Command exited with ExitCode %d", exitStatus.ExitCode)
		return fmt.Errorf("Command exited with ExitCode %d", exitStatus.ExitCode)
	}

	return nil
}

func (m *monitor) RunJobAndWaitWhileAlive(ctx *context.Context, job Job) error {
	jobCtx, err := m.GetJobContext(ctx, job)
	if err != nil {
		return fmt.Errorf("Cannot get job context, error: %s", err.Error())
	}

	if err = m.runJobAndWaitWhileAlive(jobCtx, job); err != nil {
		return fmt.Errorf("Could not run job, error: %s", err.Error())
	}

	return nil
}
