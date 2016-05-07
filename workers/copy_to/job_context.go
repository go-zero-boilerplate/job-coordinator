package copy_to

import (
	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

type jobContext struct {
	logger               logger.Logger
	pendingJobFileSystem afero.Fs
	localExportDir       string
	remoteJobPath        string
	remoteComms          remote_comms_facade.Facade
}
