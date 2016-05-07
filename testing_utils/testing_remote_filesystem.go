package testing_utils

import (
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

func NewTestingRemoteFSFactory() remote_file_system.Factory {
	return &testingFactory{}
}

type testingFactory struct{}

func (t *testingFactory) New(remoteTempDir, jobId string) remote_file_system.FileSystem {
	return remote_file_system.NewDefaultFileSystem(remoteTempDir, "testing-received", jobId)
}
