package monitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-zero-boilerplate/osvisitors"

	convey2 "github.com/go-zero-boilerplate/more_goconvey_assertions"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils"
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils/mocks"
	"github.com/go-zero-boilerplate/job-coordinator/utils/exec_logger_helpers"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type testingJob struct {
	id          string
	ctx         *context.Context
	hostDetails host_details.HostDetails

	remoteOsType    osvisitors.OsType
	remoteJobFS     remote_file_system.FileSystem
	remoteDirMocker *mocks.RemoteDirMocker
	pid             int
}

func (t *testingJob) Id() string                            { return t.id }
func (t *testingJob) HostDetails() host_details.HostDetails { return t.hostDetails }

/*func (t *testingJob) Pid() int                              { return t.pid }*/

func waitForProcess(pid int) {
	//TODO: This process is run via GoPsexec server remotely, but the remote is localhost
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Wait()
	}
}

type startedProcessDetails struct {
	err                  error
	job                  *testingJob
	remoteJobDirToDelete string
	pid                  int
}

func startProcess(ctx *context.Context, dummyJobId string, dummyCliFlagInterval time.Duration, dummyCliFlagNumber int, failureMode bool) (details *startedProcessDetails) {
	details = &startedProcessDetails{}

	remoteFSFactory := testing_utils.NewTestingRemoteFSFactory()
	hostDetails := testing_utils.NewTestingHostDetails("localhost", remoteFSFactory)

	job := &testingJob{id: dummyJobId, hostDetails: hostDetails, ctx: ctx}
	details.job = job

	remoteComms := ctx.RemoteCommsFactory.NewFacade(hostDetails)
	remoteTempDir, err := remoteComms.GetTempDir()
	if err != nil {
		details.err = err
		return
	}

	remoteOsType, err := remoteComms.GetOsType()
	if err != nil {
		details.err = err
		return
	}

	job.remoteOsType = remoteOsType
	job.remoteJobFS = hostDetails.RemoteFileSystemFactory().New(remoteTempDir, job.Id())
	details.remoteJobDirToDelete = job.remoteJobFS.GetFullJobDir()

	//Remove dir before starting
	err = os.RemoveAll(details.remoteJobDirToDelete)
	if err != nil {
		details.err = err
		return
	}

	job.remoteDirMocker = mocks.NewRemoteDirMocker(ctx, dummyCliFlagInterval, dummyCliFlagNumber, failureMode)
	err = job.remoteDirMocker.CreateScriptFile(job.remoteJobFS, hostDetails)
	if err != nil {
		details.err = err
		return
	}

	commandLine := []string{job.remoteDirMocker.FullScriptFilePath}
	var timeout *time.Duration
	var recordResourceUsage bool
	allExecLoggerCommandLine := exec_logger_helpers.CombineExecLoggerCommandline(remoteOsType, job.remoteJobFS, commandLine, timeout, recordResourceUsage)
	workingDir := job.remoteJobFS.GetFullJobDir()
	startedDetails, err := remoteComms.StartDetached(workingDir, allExecLoggerCommandLine...)
	if err != nil {
		details.err = err
		return
	}
	if startedDetails.Pid == 0 {
		details.err = fmt.Errorf("Process ID should not be 0")
		return
	}

	details.pid = startedDetails.Pid
	job.pid = startedDetails.Pid
	return
}

func TestMonitorWorker(t *testing.T) {
	for i := 1; i <= 3; i++ {
		Convey(fmt.Sprintf("Testing the copy-to worker (iteration %d)", i), t, func() {
			waitAlreadyCalled := false

			ctx, err := testing_utils.NewContext(
				false, //Virtual FileSystem
			)
			So(err, ShouldBeNil)
			testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

			startTime := time.Now()

			failureMode := i == 2
			longRunningAbortMode := i == 3
			dummyJobId := fmt.Sprintf("testing-monitor-id-%d", i)
			dummyCliFlagInterval := 100 * time.Millisecond
			dummyCliFlagNumber := 50

			if longRunningAbortMode {
				dummyCliFlagInterval = 500 * time.Millisecond
				dummyCliFlagNumber = 100
			}

			procDetails := startProcess(ctx, dummyJobId, dummyCliFlagInterval, dummyCliFlagNumber, failureMode)
			if strings.TrimSpace(procDetails.remoteJobDirToDelete) != "" {
				defer os.RemoveAll(procDetails.remoteJobDirToDelete)
			}
			So(procDetails.err, ShouldBeNil) //Yes this comes after the remoteJobDirToDelete removal because that path could be set already but the error occurred later
			defer func() {
				if !waitAlreadyCalled {
					waitForProcess(procDetails.pid)
				}
			}()

			testingLogger.Clear()
			So(len(testingLogger.Handler.Lines), ShouldEqual, 0)

			worker := &monitor{}
			jobCtx, err := worker.GetJobContext(ctx, procDetails.job)
			So(err, ShouldBeNil)
			jobCtx.initialSleepDelay = 1 * time.Second

			So(procDetails.job.remoteDirMocker.FullLogFilePath, ShouldEqual, jobCtx.remoteLogFile)

			var waitGrpPossibleAbortMsg sync.WaitGroup
			var notifyAbortErr error
			if longRunningAbortMode {
				waitGrpPossibleAbortMsg.Add(1)
				go func() {
					defer waitGrpPossibleAbortMsg.Done()
					time.Sleep(1 * time.Second) //Give it time to start running
					notifyAbortErr = worker.notifyShutdown(jobCtx, "Abort from monitor_test.go")
				}()
			}

			err = worker.runJobAndWaitWhileAlive(jobCtx, procDetails.job)
			So(err, ShouldBeNil)

			waitGrpPossibleAbortMsg.Wait()
			So(notifyAbortErr, ShouldBeNil)

			//TODO: Would a better way for waiting for the process be to use the startedDetails.Wait() of StartDetached above?
			waitForProcess(procDetails.pid)
			waitAlreadyCalled = true

			So(procDetails.job.remoteDirMocker.FullScriptFilePath, convey2.AssertFileExistance, true)
			So(procDetails.job.remoteDirMocker.FullLogFilePath, convey2.AssertFileExistance, true)

			logFileContent, err := ioutil.ReadFile(procDetails.job.remoteDirMocker.FullLogFilePath)
			So(err, ShouldBeNil)

			if failureMode {
				So(string(logFileContent), ShouldContainSubstring, "Command exited with code 2")

				expectation := testing_utils.LogExpectation{
					LineCount: 3,
					Lines: []testing_utils.ExpectedLine{
						testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Starting job"}},
						testing_utils.ExpectedLine{Index: 1, Error: true, RequiredSubstrings: []string{"Command exited with error: exit status 2. ExitCode was 2"}},
						testing_utils.ExpectedLine{Index: 2, Info: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
					},
				}

				includeDebugLines := false
				err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
				So(err, ShouldBeNil)
			} else if longRunningAbortMode {
				So(string(logFileContent), ShouldContainSubstring, "Command exited with code 1")

				So(time.Now().Sub(startTime).Seconds(), ShouldBeGreaterThan, 1)
				So(time.Now().Sub(startTime).Seconds(), ShouldBeLessThan, 10)

				expectation := testing_utils.LogExpectation{
					LineCount: 3,
					Lines: []testing_utils.ExpectedLine{
						testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Starting job"}},
						testing_utils.ExpectedLine{Index: 1, Error: true, RequiredSubstrings: []string{"Command exited with error: exit status 1. ExitCode was 1"}},
						testing_utils.ExpectedLine{Index: 2, Info: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
					},
				}

				includeDebugLines := false
				err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
				So(err, ShouldBeNil)
			} else {
				for _, expectedLine := range procDetails.job.remoteDirMocker.ExpectedLogLines {
					So(string(logFileContent), ShouldContainSubstring, expectedLine)
				}

				expectation := testing_utils.LogExpectation{
					LineCount: 2,
					Lines: []testing_utils.ExpectedLine{
						testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Starting job"}},
						testing_utils.ExpectedLine{Index: 1, Info: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
					},
				}

				includeDebugLines := false
				err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
				So(err, ShouldBeNil)
			}
		})
	}
}
