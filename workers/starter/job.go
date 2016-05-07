package starter

import (
	"time"

	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"

	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
)

type Job interface {
	Id() string
	HostDetails() host_details.HostDetails
	Commandline(remoteFileSystem remote_file_system.FileSystem) ([]string, error)
	Timeout() *time.Duration
	RecordResourceUsage() bool
}
