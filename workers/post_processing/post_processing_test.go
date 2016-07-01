package post_processing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/francoishill/afero"
	"github.com/golang-devops/exec-logger/exec_logger_constants"
	"github.com/golang-devops/exec-logger/exec_logger_dtos"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-zero-boilerplate/job-coordinator/testing_utils"
	"github.com/go-zero-boilerplate/job-coordinator/utils/job_helpers"
)

type testingJob struct {
	fileSystem      afero.Fs
	id              string
	logContent      string
	exitedDto       *exec_logger_dtos.ExitStatusDto
	localContextDto *exec_logger_dtos.LocalContextDto
}

func (t *testingJob) Id() string { return t.id }

func (t *testingJob) createTempLogFile() error {
	fileName := exec_logger_constants.LOG_FILE_NAME
	parentDir := filepath.Dir(fileName)

	fileContent := []byte(t.logContent)
	if err := t.fileSystem.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	return afero.WriteFile(t.fileSystem, fileName, fileContent, 0655)
}

func (t *testingJob) createExitAndLocalContextFileParentDirs() error {
	parentDir := filepath.Dir(exec_logger_constants.EXITED_FILE_NAME)
	if err := t.fileSystem.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("Cannot create exit file parent dir '%s', error: %s", parentDir, err.Error())
	}
	parentDir = filepath.Dir(exec_logger_constants.LOCAL_CONTEXT_FILE_NAME)
	if err := t.fileSystem.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("Cannot create local-context file parent dir '%s', error: %s", parentDir, err.Error())
	}
	return nil
}

func (t *testingJob) createTempExitedFile() error {
	fileContent, err := json.Marshal(t.exitedDto)
	if err != nil {
		return fmt.Errorf("Unable to marshal the mock ExitedDto. Error: %s. Attempted object to marshal was: %+v", err.Error(), t.exitedDto)
	}

	fileName := exec_logger_constants.EXITED_FILE_NAME
	parentDir := filepath.Dir(fileName)
	if err := t.fileSystem.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	return afero.WriteFile(t.fileSystem, fileName, fileContent, 0655)
}

func (t *testingJob) createTempLocalContextFile() error {
	fileContent, err := json.Marshal(t.localContextDto)
	if err != nil {
		return fmt.Errorf("Unable to marshal the mock LocalContextDto. Error: %s. Attempted object to marshal was: %+v", err.Error(), t.exitedDto)
	}

	fileName := exec_logger_constants.LOCAL_CONTEXT_FILE_NAME
	parentDir := filepath.Dir(fileName)
	if err := t.fileSystem.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	return afero.WriteFile(t.fileSystem, fileName, fileContent, 0655)
}

func TestExportWorker(t *testing.T) {
	Convey("Testing the post-processing worker\n", t, func() {
		ctx, err := testing_utils.NewContext(
			false, //Virtual FileSystem
		)
		So(err, ShouldBeNil)
		testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

		exitTime := time.Now()

		jobId := "testing-postprocessing-id-1"
		completedJobFileSystem := job_helpers.GetJobFileSystem(ctx.CompletedLocalFileSystem, jobId)
		job := &testingJob{
			fileSystem: completedJobFileSystem,
			id:         jobId,
			logContent: "Log content of job " + jobId,
			exitedDto: &exec_logger_dtos.ExitStatusDto{
				ExitCode: 11,
				Error:    "Dummy error in exit file",
				Duration: (33 * time.Second).String(),
				ExitTime: exitTime,
			},
			localContextDto: &exec_logger_dtos.LocalContextDto{
				HostName: "localhost",
				UserName: "dummy-user",
			},
		}
		defer completedJobFileSystem.RemoveAll(".")

		err = job.createTempLogFile()
		So(err, ShouldBeNil)
		err = job.createExitAndLocalContextFileParentDirs()
		So(err, ShouldBeNil)

		worker := &postProcessing{}

		So(len(testingLogger.Handler.Lines), ShouldEqual, 0)

		jobCtx, err := worker.getJobContext(ctx, completedJobFileSystem, job)
		So(err, ShouldBeNil)

		//First the failure case where exited file does not exist
		result := worker.runJob(jobCtx, job)
		So(result.HasErrors(), ShouldBeTrue)
		expectedErrCnt := 2
		if len(result.Errors()) != expectedErrCnt {
			So(fmt.Errorf("Results error count should be %d (was %d). Errors was: %+v", expectedErrCnt, len(result.Errors()), result.errors), ShouldBeNil)
		}
		exitedJSONFileName := "exited.json"
		So(result.Errors()[0], ShouldContainSubstring, "The system cannot find the file specified")
		So(result.Errors()[0], ShouldContainSubstring, exitedJSONFileName)
		localContextFileName := "local-context.json"
		So(result.Errors()[1], ShouldContainSubstring, "The system cannot find the file specified")
		So(result.Errors()[1], ShouldContainSubstring, localContextFileName)

		expectation := testing_utils.LogExpectation{
			LineCount: 3,
			Lines: []testing_utils.ExpectedLine{
				testing_utils.ExpectedLine{Index: 0, Error: true, RequiredSubstrings: []string{"Cannot read exit file"}},
				testing_utils.ExpectedLine{Index: 1, Error: true, RequiredSubstrings: []string{"Cannot read local-context file"}},
				testing_utils.ExpectedLine{Index: 2, Error: true, RequiredSubstrings: []string{"Starting job"}}, //This is the deferred Trace-Stop log
			},
		}

		includeDebugLines := false
		err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
		So(err, ShouldBeNil)

		//Create the exited fail and then it should succeed
		testingLogger.Clear()
		err = job.createTempExitedFile()
		So(err, ShouldBeNil)
		err = job.createTempLocalContextFile()
		So(err, ShouldBeNil)

		result = worker.runJob(jobCtx, job)
		So(result.HasErrors(), ShouldBeFalse)
		So(result.completedJobFileSystem, ShouldNotBeNil)
		So(result.logRelativePath, ShouldNotBeEmpty)

		exitStatus := result.ExitStatus()
		So(exitStatus, ShouldNotBeNil)
		So(exitStatus.HasError(), ShouldBeTrue)
		So(exitStatus.ExitCode, ShouldEqual, 11)
		So(exitStatus.Error, ShouldEqual, "Dummy error in exit file")
		So(exitStatus.Duration, ShouldEqual, "33s")
		if !exitStatus.ExitTime.Equal(exitTime) {
			So(fmt.Errorf("Actual exit time '%s' does not equal expected '%s'", exitStatus.ExitTime.String(), exitTime.String()), ShouldBeNil)
		}

		So(job.logContent, ShouldNotBeEmpty)
		logFile, err := result.OpenLogFile()
		So(err, ShouldBeNil)
		defer logFile.Close()

		logContent, err := ioutil.ReadAll(logFile)
		So(err, ShouldBeNil)
		So(logContent, ShouldNotBeNil)
		So(string(logContent), ShouldContainSubstring, job.logContent)

		expectation = testing_utils.LogExpectation{
			LineCount: 0,
			Lines:     []testing_utils.ExpectedLine{},
		}

		includeDebugLines = false
		err = expectation.MeetsExpectation(testingLogger, includeDebugLines)
		So(err, ShouldBeNil)
	})
}
