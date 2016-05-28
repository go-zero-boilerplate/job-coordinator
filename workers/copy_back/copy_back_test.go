package copy_back

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/francoishill/afero"

	convey2 "github.com/go-zero-boilerplate/more_goconvey_assertions"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-zero-boilerplate/job-coordinator/testing_utils"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type testingJob struct {
	fileSystem  afero.Fs
	id          string
	hostDetails host_details.HostDetails
}

func (t *testingJob) Id() string                            { return t.id }
func (t *testingJob) HostDetails() host_details.HostDetails { return t.hostDetails }
func (t *testingJob) RemoteCleanupSpec() *CleanupSpec       { return &CleanupSpec{} }

func (t *testingJob) createTempRemoteFile(jobCtx *jobContext) error {
	fileContent := []byte("Content of copy-back job " + t.id)
	if err := os.MkdirAll(jobCtx.remoteJobPath, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(jobCtx.remoteJobPath, "tmp-copy-back-file-1.txt"), fileContent, 0655)
}

func TestExportWorker(t *testing.T) {
	Convey("Testing the copy-back worker\n", t, func() {
		ctx, err := testing_utils.NewContext(
			false, //Virtual FileSystem
		)
		So(err, ShouldBeNil)
		testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

		remoteFSFactory := testing_utils.NewTestingRemoteFSFactory()
		hostDetails := testing_utils.NewTestingHostDetails("localhost", remoteFSFactory)

		jobId := "testing-copyback-id-1"
		completedJobFileSystem := job_helpers.GetJobFileSystem(ctx.CompletedLocalFileSystem, jobId)
		oldCompletedJobFileSystem := job_helpers.GetJobFileSystem(ctx.CompletedLocalFileSystem, jobId+"_old")
		job := &testingJob{fileSystem: completedJobFileSystem, id: jobId, hostDetails: hostDetails}
		defer completedJobFileSystem.RemoveAll(".")
		defer oldCompletedJobFileSystem.RemoveAll(".")

		worker := &copyBack{}

		So(len(testingLogger.Handler.Lines), ShouldEqual, 0)

		jobCtx, err := worker.getJobContext(ctx, completedJobFileSystem, oldCompletedJobFileSystem, job)
		So(err, ShouldBeNil)
		defer os.RemoveAll(jobCtx.remoteJobPath)

		err = job.createTempRemoteFile(jobCtx)
		So(err, ShouldBeNil)

		So(jobCtx.remoteJobPath, convey2.AssertDirectoryExistance, true)

		//TODO: Should test scenario where event handlers get used inside worker
		var handlers Handlers
		err = worker.runJob(jobCtx, job, handlers)
		So(err, ShouldBeNil)

		So(jobCtx.remoteJobPath, convey2.AssertDirectoryExistance, false)

		expectation := testing_utils.LogExpectation{
			LineCount: 1,
			Lines: []testing_utils.ExpectedLine{
				testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Successfully deleted remote job dir"}},
			},
		}

		includeDebugLines := false
		err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
		So(err, ShouldBeNil)
	})
}
