package remote_comms_facade

import (
	gpClient "github.com/golang-devops/go-psexec/client"

	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
)

type Factory interface {
	NewFacade(hostDetails host_details.HostDetails) Facade
}

func NewFactory(gopsexecClient *gpClient.Client) Factory {
	return &factory{
		gopsexecClient: gopsexecClient,
	}
}

type factory struct {
	gopsexecClient *gpClient.Client
}

func (f *factory) NewFacade(hostDetails host_details.HostDetails) Facade {
	return &facade{
		gopsexecClient: f.gopsexecClient,
		hostDetails:    hostDetails,
	}
}
