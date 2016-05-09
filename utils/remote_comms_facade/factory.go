package remote_comms_facade

import (
	gpClient "github.com/golang-devops/go-psexec/client"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/golang-devops/go-psexec/services/filepath_summary"
)

//Factory will create instances of Facade
type Factory interface {
	NewFacade(hostDetails host_details.HostDetails) Facade
}

//NewFactory creates a new Factory instance
func NewFactory(logger logger.Logger, gopsexecClient *gpClient.Client, filepathSummaryService filepath_summary.Service) Factory {
	return &factory{
		logger:                 logger,
		gopsexecClient:         gopsexecClient,
		filepathSummaryService: filepathSummaryService,
	}
}

type factory struct {
	logger                 logger.Logger
	gopsexecClient         *gpClient.Client
	filepathSummaryService filepath_summary.Service
}

func (f *factory) NewFacade(hostDetails host_details.HostDetails) Facade {
	return &facade{
		logger:                 f.logger,
		gopsexecClient:         f.gopsexecClient,
		hostDetails:            hostDetails,
		filepathSummaryService: f.filepathSummaryService,
	}
}
