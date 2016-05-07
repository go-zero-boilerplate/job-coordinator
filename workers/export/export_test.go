package export

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/francoishill/afero"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-zero-boilerplate/job-coordinator/testing_utils"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type testingJob struct {
	id          string
	hostDetails host_details.HostDetails
	tempJobFs   afero.Fs
}

func (t *testingJob) Id() string                            { return t.id }
func (t *testingJob) HostDetails() host_details.HostDetails { return t.hostDetails }

func (t *testingJob) ExportFiles(fileSystem afero.Fs) error {
	err := fileSystem.MkdirAll(".", 0755)
	if err != nil {
		return fmt.Errorf("Unable to create temp export filesystem dir, error: %s", err.Error())
	}

	files := []string{"myfile-1.txt", "myfile-2.py"}
	for _, f := range files {
		if err := t.exportTempFile(fileSystem, f); err != nil {
			return fmt.Errorf("Cannot export temp file '%s', error: %s", f, err.Error())
		}
	}
	return nil
}

func (t *testingJob) exportTempFile(fileSystem afero.Fs, name string) error {
	return afero.WriteFile(fileSystem, name, []byte("This is a dummy file with name "+name), 0655)
}

func TestExportWorker(t *testing.T) {
	Convey("Testing the export worker", t, func() {
		ctx, err := testing_utils.NewContext(
			//true, //Virtual FileSystem
			false, //
		)
		So(err, ShouldBeNil)
		testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

		tempJobDir, err := ioutil.TempDir(os.TempDir(), "job-coord-export-test-")
		So(err, ShouldBeNil)
		defer os.RemoveAll(tempJobDir)

		remoteFSFactory := testing_utils.NewTestingRemoteFSFactory()
		hostDetails := testing_utils.NewTestingHostDetails("localhost", remoteFSFactory)

		tempJobFs := afero.NewBasePathFs(afero.NewOsFs(), tempJobDir)
		jobId := "testing-export-id-1"
		job := &testingJob{tempJobFs: tempJobFs, id: jobId, hostDetails: hostDetails}

		worker := &export{}

		pendingJobFileSystem := job_helpers.GetJobFileSystem(ctx.PendingLocalFileSystem, jobId)
		jobCtx, err := worker.getJobContext(ctx, pendingJobFileSystem, job)
		So(err, ShouldBeNil)
		defer pendingJobFileSystem.RemoveAll(".")

		So(len(testingLogger.Handler.Lines), ShouldEqual, 0)
		err = worker.runJob(jobCtx, job)
		So(err, ShouldBeNil)

		expectation := testing_utils.LogExpectation{
			LineCount: 3,
			Lines: []testing_utils.ExpectedLine{
				testing_utils.ExpectedLine{Index: 0, Info: true, RequiredSubstrings: []string{"Starting job"}},
				testing_utils.ExpectedLine{Index: 1, Info: true, RequiredSubstrings: []string{"Successfully ensured export dir gone"}},
				testing_utils.ExpectedLine{Index: 2, Info: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
			},
		}
		includeDebugLines := false
		err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
		So(err, ShouldBeNil)

		fileInfo, err := pendingJobFileSystem.Stat("myfile-1.txt")
		So(err, ShouldBeNil)
		So(fileInfo.Mode().IsRegular(), ShouldBeTrue)

		_, err = pendingJobFileSystem.Stat("myfile-does-not-exist.noexist")
		So(err, ShouldNotBeNil)
		// So(err.Error(), ShouldContainSubstring, afero.ErrFileNotFound.Error())

		err = os.RemoveAll(tempJobDir)
		So(err, ShouldBeNil)
	})
}
