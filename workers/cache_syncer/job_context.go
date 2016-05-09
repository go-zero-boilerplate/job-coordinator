package cache_syncer

import (
	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

type jobContext struct {
	logger      logger.Logger
	remoteComms remote_comms_facade.Facade
}
