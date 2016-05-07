package testing_utils

import (
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils/testing_constants"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"

	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
)

func NewTestingHostDetails(hostName string, remoteFileSystemFactory remote_file_system.Factory) host_details.HostDetails {
	expectedGopsexecVersion := testing_constants.ExpectedGoPsexecVersion
	return &testingHostDetails{
		hostName:                hostName,
		expectedGopsexecVersion: expectedGopsexecVersion,
		remoteFileSystemFactory: remoteFileSystemFactory,
	}
}

type testingHostDetails struct {
	hostName                string
	expectedGopsexecVersion string
	remoteFileSystemFactory remote_file_system.Factory
}

func (t *testingHostDetails) HostName() string {
	return t.hostName
}

func (t *testingHostDetails) ExpectedGopsexecVersion() string {
	return t.expectedGopsexecVersion
}

func (t *testingHostDetails) RemoteFileSystemFactory() remote_file_system.Factory {
	return t.remoteFileSystemFactory
}
