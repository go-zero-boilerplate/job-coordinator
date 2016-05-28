package copy_to

import (
	"github.com/francoishill/afero"
)

//Handlers is an interface for common events we might want to handle by the consumer of job-coordinator
type Handlers interface {
	FailedToRemoveLocalExportDir(err error, pendingFS afero.Fs)
}
