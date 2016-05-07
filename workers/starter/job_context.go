package starter

import (
	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type jobContext struct {
	logger         logger.Logger
	localExportDir string
	remoteJobFS    remote_file_system.FileSystem
	remoteJobPath  string
	remoteComms    remote_comms_facade.Facade
}
