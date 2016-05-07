package mocks

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/francoishill/afero"
	"github.com/golang-devops/exec-logger/exec_logger_constants"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils/script_creators"
	"github.com/go-zero-boilerplate/job-coordinator/utils/exec_logger_helpers"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
	"github.com/go-zero-boilerplate/osvisitors"
	"github.com/go-zero-boilerplate/path_utils"
)

func NewRemoteDirMocker(ctx *context.Context, dummyCliFlagInterval time.Duration, dummyCliFlagNumber int, failureMode bool) *RemoteDirMocker {
	return &RemoteDirMocker{
		ctx:                  ctx,
		dummyCliFlagInterval: dummyCliFlagInterval,
		dummyCliFlagNumber:   dummyCliFlagNumber,
		failureMode:          failureMode,
	}
}

type RemoteDirMocker struct {
	ctx                  *context.Context
	dummyCliFlagInterval time.Duration
	dummyCliFlagNumber   int
	failureMode          bool

	FullScriptFilePath    string
	ExpectedScriptContent string
	FullLogFilePath       string
	ExpectedLogLines      []string
}

func (r *RemoteDirMocker) CreateScriptFile(remoteFS remote_file_system.FileSystem, hostDetails host_details.HostDetails) error {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "job-coordinator-script-")
	if err != nil {
		return fmt.Errorf("Unable to create temp script file, error: %s", err.Error())
	}
	tmpFile.Close()

	filePath := tmpFile.Name()
	defer os.Remove(filePath)

	remoteComms := r.ctx.RemoteCommsFactory.NewFacade(hostDetails)
	remoteOsType, err := remoteComms.GetOsType()
	if err != nil {
		return fmt.Errorf("Cannot get remote OsType, error: %s", err.Error())
	}

	visitor := script_creators.NewCreateStarterScriptFileVisitor(r.dummyCliFlagInterval, r.dummyCliFlagNumber, r.failureMode)
	remoteOsType.Accept(visitor)

	if exists, err := path_utils.FileExists(visitor.DummyCliBinPath); err != nil {
		return fmt.Errorf("Unable to determine if dummycli bin '%s' exists, error: %s. %s", visitor.DummyCliBinPath, err.Error(), script_creators.INSTALL_DUMMYCLI_HELP)
	} else if !exists {
		return fmt.Errorf("Dummycli bin '%s' does not exist. %s", visitor.DummyCliBinPath, script_creators.INSTALL_DUMMYCLI_HELP)
	}

	err = ioutil.WriteFile(filePath, []byte(visitor.Content), 0655)
	if err != nil {
		return fmt.Errorf("Unable to write script lines to file, error: %s", err.Error())
	}

	remoteJobScriptPath := remoteFS.GetFullJobDir(visitor.FileNameOnly)

	err = remoteComms.Upload(filePath, remoteJobScriptPath)
	if err != nil {
		return fmt.Errorf("Unable to upload temp script to remote '%s', error: %s", hostDetails.HostName(), err.Error())
	}

	r.FullScriptFilePath = remoteJobScriptPath
	r.ExpectedScriptContent = visitor.Content
	r.FullLogFilePath = remoteFS.GetFullJobDir(exec_logger_constants.LOG_FILE_NAME)
	r.ExpectedLogLines = visitor.ExpectedLogLines

	return r.copyExecLoggerBinFile(remoteOsType, remoteFS)
}

func (r *RemoteDirMocker) copyExecLoggerBinFile(remoteOsType osvisitors.OsType, remoteFS remote_file_system.FileSystem) error {
	execLoggerFile, err := os.Open(r.ctx.LocalExecLoggerBinPath)
	if err != nil {
		return err
	}
	defer execLoggerFile.Close()

	execLoggerBinFileName := exec_logger_helpers.GetExecLoggerBinFileName(remoteOsType)
	remoteExecLoggerExePath := remoteFS.GetFullJobDir(execLoggerBinFileName)
	osFs := afero.NewOsFs()
	return afero.CopyFile(osFs, r.ctx.LocalExecLoggerBinPath, osFs, remoteExecLoggerExePath)
}
