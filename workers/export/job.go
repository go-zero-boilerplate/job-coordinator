package export

import (
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"

	"github.com/francoishill/afero"
)

type Job interface {
	Id() string
	HostDetails() host_details.HostDetails
	ExportFiles(fileSystem afero.Fs) error
}
