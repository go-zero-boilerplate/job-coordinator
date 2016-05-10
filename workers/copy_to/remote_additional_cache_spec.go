package copy_to

import (
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type RemoteAdditionalCacheSpec struct {
	JobSubdir     string
	RemoteCacheFS remote_file_system.FileSystem
}
