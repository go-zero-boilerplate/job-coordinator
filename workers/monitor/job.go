package monitor

import "github.com/go-zero-boilerplate/job-coordinator/utils/host_details"

type Job interface {
	Id() string
	HostDetails() host_details.HostDetails
	/*Pid() int*/
}
