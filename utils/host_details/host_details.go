package host_details

import (
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type HostDetails interface {
	HostName() string
	ExpectedGopsexecVersion() string
	RemoteFileSystemFactory() remote_file_system.Factory
}

func NewSimple(hostName, expectedGopsexecVersion string, remoteFSFactory remote_file_system.Factory) HostDetails {
	return &simpleHostDetails{
		hostName:                hostName,
		expectedGopsexecVersion: expectedGopsexecVersion,
		remoteFSFactory:         remoteFSFactory,
	}
}

type simpleHostDetails struct {
	hostName                string
	expectedGopsexecVersion string
	remoteFSFactory         remote_file_system.Factory
}

func (s *simpleHostDetails) HostName() string {
	return s.hostName
}
func (s *simpleHostDetails) ExpectedGopsexecVersion() string {
	return s.expectedGopsexecVersion
}
func (s *simpleHostDetails) RemoteFileSystemFactory() remote_file_system.Factory {
	return s.remoteFSFactory
}
