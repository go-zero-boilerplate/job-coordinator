package post_processing

import (
	"github.com/francoishill/afero"
	"github.com/go-zero-boilerplate/job-coordinator/logger"
)

type jobContext struct {
	logger                    logger.Logger
	completedJobFileSystem    afero.Fs
	exitedRelativePath        string
	logRelativePath           string
	resourceUsageRelativePath string
	localContextRelativePath  string
}
