package remote_comms_facade

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	gpClient "github.com/golang-devops/go-psexec/client"
	goPsexecShared "github.com/golang-devops/go-psexec/shared"

	"github.com/go-zero-boilerplate/job-coordinator/testing_utils/testing_constants"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type testingHostDetails struct {
	h         string
	f         remote_file_system.Factory
	expGoVers string
}

func (t *testingHostDetails) HostName() string                                    { return t.h }
func (t *testingHostDetails) RemoteFileSystemFactory() remote_file_system.Factory { return t.f }
func (t *testingHostDetails) ExpectedGopsexecVersion() string                     { return t.expGoVers }

func TestFacadePing(t *testing.T) {
	Convey("Testing the ping command", t, func() {
		pemPath := os.ExpandEnv(`$GOPATH/src/github.com/golang-devops/go-psexec/client/testdata/test_client.pem`)
		clientPemKey, err := goPsexecShared.ReadPemKey(pemPath)
		So(err, ShouldBeNil)

		goPsexecClient := gpClient.New(clientPemKey)

		factory := NewFactory(goPsexecClient)
		So(factory, ShouldNotBeNil)

		hostDetails := &testingHostDetails{h: "localhost", f: nil, expGoVers: testing_constants.ExpectedGoPsexecVersion}
		facade := factory.NewFacade(hostDetails)
		So(facade, ShouldNotBeNil)

		err = facade.Ping()
		So(err, ShouldBeNil)

		err = facade.ConfirmVersionMatch(testing_constants.ExpectedGoPsexecVersion)
		So(err, ShouldBeNil)

		err = facade.ConfirmVersionMatch("0.0.0") //Just to ensure we get an error when not match.
		So(err, ShouldNotBeNil)
	})
}
