package export

import (
	"github.com/francoishill/afero"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

//AdditionalCacheSpecs holds info regarding additional cache, like the case where
//there are large binaries we want copied with the job but these binaries do not change very often
type AdditionalCacheSpecs struct {
	LocalFS       afero.Fs
	RemoteCacheFS remote_file_system.CacheFileSystem
}
