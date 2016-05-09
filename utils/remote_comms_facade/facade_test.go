package remote_comms_facade

import (
	"testing"

	// . "github.com/smartystreets/goconvey/convey"

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
	//TODO: Fix and expand this test
	/*Convey("Testing the ping command", t, func() {
		pemPath := os.ExpandEnv(`$GOPATH/src/github.com/golang-devops/go-psexec/client/testdata/test_client.pem`)
		clientPemKey, err := goPsexecShared.ReadPemKey(pemPath)
		So(err, ShouldBeNil)

		goPsexecClient := gpClient.New(clientPemKey)

		testingLogger := testing_utils.NewLogger()
		factory := NewFactory(testingLogger, goPsexecClient)
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
	})*/
}
