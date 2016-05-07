package copy_back

import (
	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

type jobContext struct {
	logger                    logger.Logger
	completedJobFileSystem    afero.Fs
	oldCompletedJobFileSystem afero.Fs
	fullCompletedJobPath      string
	fullOldCompletedJobPath   string
	remoteJobPath             string
	remoteComms               remote_comms_facade.Facade
}
