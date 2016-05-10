package copy_to

import (
	"os"
	"testing"

	"github.com/francoishill/afero"

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

func (t *testingJob) createTempFile() error {
	fileContent := []byte("Content of copy-to job " + t.id)
	if err := t.fileSystem.MkdirAll("", 0755); err != nil {
		return err
	}
	return afero.WriteFile(t.fileSystem, "tmp-cop-to-file-1.txt", fileContent, 0655)
}

func (t *testingJob) RemoteAdditionalCacheSpecs() []*RemoteAdditionalCacheSpec {
	//TODO: Implement tests
	return nil
}

func TestExportWorker(t *testing.T) {
	Convey("Testing the copy-to worker\n", t, func() {
		ctx, err := testing_utils.NewContext(
			false, //Virtual FileSystem
		)
		So(err, ShouldBeNil)
		testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

		remoteFSFactory := testing_utils.NewTestingRemoteFSFactory()
		hostDetails := testing_utils.NewTestingHostDetails("localhost", remoteFSFactory)

		jobId := "testing-copyto-id-1"
		pendingJobFileSystem := job_helpers.GetJobFileSystem(ctx.PendingLocalFileSystem, jobId)
		job := &testingJob{fileSystem: pendingJobFileSystem, id: jobId, hostDetails: hostDetails}
		//TODO: How to ensure that this filesystem is not the Root dir of the machine or similar? This RemoveAll happens in a few spots
		defer pendingJobFileSystem.RemoveAll(".")

		err = job.createTempFile()
		So(err, ShouldBeNil)

		worker := &copyTo{}

		So(len(testingLogger.Handler.Lines), ShouldEqual, 0)

		jobCtx, err := worker.getJobContext(ctx, pendingJobFileSystem, job)
		So(err, ShouldBeNil)
		defer os.RemoveAll(jobCtx.remoteJobPath)

		err = testing_utils.CheckFileSystemPathExistance(pendingJobFileSystem, ".", true)
		So(err, ShouldBeNil)

		err = worker.runJob(jobCtx, job)
		So(err, ShouldBeNil)

		err = testing_utils.CheckFileSystemPathExistance(pendingJobFileSystem, ".", false)
		So(err, ShouldBeNil)

		expectation := testing_utils.LogExpectation{
			LineCount: 4,
			Lines: []testing_utils.ExpectedLine{
				testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Starting job"}},
				testing_utils.ExpectedLine{Index: 1, Info: true, RequiredSubstrings: []string{"Successfully ensured remove job dir gone"}},
				testing_utils.ExpectedLine{Index: 2, Info: true, RequiredSubstrings: []string{"Successfully deleted local export dir"}},
				testing_utils.ExpectedLine{Index: 3, Info: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
			},
		}

		includeDebugLines := false
		err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
		So(err, ShouldBeNil)
	})
}
