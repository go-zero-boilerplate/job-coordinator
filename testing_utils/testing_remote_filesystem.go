package testing_utils

import (
	"os"
	"path/filepath"

	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

func NewTestingRemoteFSFactory() remote_file_system.Factory {
	return &testingFactory{}
}

type testingFactory struct{}

func (t *testingFactory) New(jobId string) remote_file_system.FileSystem {
	baseDir := filepath.Join(os.ExpandEnv("$TEMP"), "job-coordinator", "testing-received")
	return remote_file_system.NewBaseDirFileSystem(baseDir, jobId)
}
