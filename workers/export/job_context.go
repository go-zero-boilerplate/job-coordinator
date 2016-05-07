package export

import (
	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

type jobContext struct {
	logger                 logger.Logger
	pendingJobFileSystem   afero.Fs
	localExecLoggerBinPath string
	remoteComms            remote_comms_facade.Facade
}
