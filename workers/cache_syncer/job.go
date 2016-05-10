package cache_syncer

import "github.com/go-zero-boilerplate/job-coordinator/utils/host_details"

import "github.com/francoishill/afero"

type Job interface {
	Id() string
	HostDetails() host_details.HostDetails
	LocalJobCacheFS() afero.Fs
}
