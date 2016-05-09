package cache_syncer

import "github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
import "github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
import "github.com/francoishill/afero"

type Job interface {
	HostDetails() host_details.HostDetails
	LocalSourceCacheFS() afero.Fs
	RemoteDestCacheFS() remote_file_system.CacheFileSystem
}
