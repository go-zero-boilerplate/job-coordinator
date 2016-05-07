package starter

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	convey2 "github.com/go-zero-boilerplate/more_goconvey_assertions"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils"
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils/mocks"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type testingJob struct {
	id          string
	ctx         *context.Context
	hostDetails host_details.HostDetails

	remoteDirMocker *mocks.RemoteDirMocker
}

func (t *testingJob) Id() string                            { return t.id }
func (t *testingJob) HostDetails() host_details.HostDetails { return t.hostDetails }
func (t *testingJob) Commandline(remoteFileSystem remote_file_system.FileSystem) ([]string, error) {
	dummyCliFlagInterval := 100 * time.Millisecond
	dummyCliFlagNumber := 5
	failureMode := false
	t.remoteDirMocker = mocks.NewRemoteDirMocker(t.ctx, dummyCliFlagInterval, dummyCliFlagNumber, failureMode)
	if err := t.remoteDirMocker.CreateScriptFile(remoteFileSystem, t.hostDetails); err != nil {
		return nil, err
	}
	return []string{t.remoteDirMocker.FullScriptFilePath}, nil
}
func (t *testingJob) Timeout() *time.Duration {
	//TODO: Add test for testing timeout works as expected
	return nil
}
func (t *testingJob) RecordResourceUsage() bool {
	//TODO: Add test for testing resource-usage words as expected
	return false
}

func waitForProcess(pid int) {
	//TODO: This process is run via GoPsexec server remotely, but the remote is localhost
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Wait()
	}
}

func TestCopyToWorker(t *testing.T) {
	Convey("Testing the copy-to worker", t, func() {
		ctx, err := testing_utils.NewContext(
			false, //Virtual FileSystem
		)
		So(err, ShouldBeNil)
		testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

		remoteFSFactory := testing_utils.NewTestingRemoteFSFactory()
		hostDetails := testing_utils.NewTestingHostDetails("localhost", remoteFSFactory)

		job := &testingJob{id: "testing-starter-id-1", hostDetails: hostDetails, ctx: ctx}

		worker := &starter{}

		testingLogger.Clear()
		So(len(testingLogger.Handler.Lines), ShouldEqual, 0)

		jobCtx, err := worker.getJobContext(ctx, job)
		So(err, ShouldBeNil)
		err = os.RemoveAll(jobCtx.remoteJobPath) //Remove now
		So(err, ShouldBeNil)
		defer os.RemoveAll(jobCtx.remoteJobPath) //Remove at end

		startedDetails, err := worker.runJob(jobCtx, job)
		So(err, ShouldBeNil)
		So(startedDetails, ShouldNotBeNil)
		So(startedDetails.Pid, ShouldNotEqual, 0)

		waitForProcess(startedDetails.Pid)

		So(job.remoteDirMocker.FullScriptFilePath, convey2.AssertFileExistance, true)
		So(job.remoteDirMocker.FullLogFilePath, convey2.AssertFileExistance, true)

		logFileContent, err := ioutil.ReadFile(job.remoteDirMocker.FullLogFilePath)
		So(err, ShouldBeNil)
		for _, expectedLine := range job.remoteDirMocker.ExpectedLogLines {
			So(string(logFileContent), ShouldContainSubstring, expectedLine)
		}

		expectation := testing_utils.LogExpectation{
			LineCount: 3,
			Lines: []testing_utils.ExpectedLine{
				testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Starting job"}},
				testing_utils.ExpectedLine{Index: 1, Info: true, RequiredSubstrings: []string{fmt.Sprintf("Started command with Pid %d", startedDetails.Pid)}},
				testing_utils.ExpectedLine{Index: 2, Info: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
			},
		}

		includeDebugLines := false
		err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
		So(err, ShouldBeNil)
	})
}
