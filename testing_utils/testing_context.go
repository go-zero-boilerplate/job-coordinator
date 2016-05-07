package testing_utils

import (
	"fmt"
	"os"

	"github.com/francoishill/afero"

	gpClient "github.com/golang-devops/go-psexec/client"
	goPsexecShared "github.com/golang-devops/go-psexec/shared"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

func NewContext(virtualFileSystem bool) (*context.Context, error) {
	if virtualFileSystem {
		memFs := afero.NewMemMapFs()
		pendingLocalFileSystem := afero.NewBasePathFs(memFs, "pending")
		completedLocalFileSystem := afero.NewBasePathFs(memFs, "completed")
		return newContext(pendingLocalFileSystem, completedLocalFileSystem)
	} else {
		//TODO: Make the 'job-coordinator' folder part customizable
		pendingLocalFileSystem := afero.NewBasePathFs(afero.NewOsFs(), os.ExpandEnv(`$TEMP/job-coordinator/testing-pending`))
		completedLocalFileSystem := afero.NewBasePathFs(afero.NewOsFs(), os.ExpandEnv(`$TEMP/job-coordinator/testing-completed`))
		return newContext(pendingLocalFileSystem, completedLocalFileSystem)
	}
}

func newContext(pendingLocalFileSystem, completedLocalFileSystem afero.Fs) (*context.Context, error) {
	logger := NewLogger()

	pemPath := os.ExpandEnv(`$GOPATH/src/github.com/golang-devops/go-psexec/client/testdata/test_client.pem`)
	clientPemKey, err := goPsexecShared.ReadPemKey(pemPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read gopsexec client pem file '%s', error: %s", pemPath, err.Error())
	}
	goPsexecClient := gpClient.New(clientPemKey)

	//TODO: This will break on non-windows remote machines. Perhaps we must have a binary per OS (win32, win64, linux32, linux64, etc) and ensure that file exists locally before trying to copy it to the remote
	localExecLoggerBinPath := os.ExpandEnv(`$GOPATH/bin/exec-logger.exe`)

	remoteCommsFactory := remote_comms_facade.NewFactory(goPsexecClient)

	return context.New(
		logger,
		pendingLocalFileSystem,
		completedLocalFileSystem,
		localExecLoggerBinPath,
		remoteCommsFactory,
	), nil
}
