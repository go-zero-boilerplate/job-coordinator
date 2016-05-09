package cache_syncer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/francoishill/afero"
	"github.com/go-zero-boilerplate/job-coordinator/testing_utils"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
)

type testingJob struct {
	hostDetails        host_details.HostDetails
	localSourceCacheFS afero.Fs
	remoteDestCacheFS  remote_file_system.CacheFileSystem
}

func (t *testingJob) HostDetails() host_details.HostDetails { return t.hostDetails }
func (t *testingJob) LocalSourceCacheFS() afero.Fs          { return t.localSourceCacheFS }
func (t *testingJob) RemoteDestCacheFS() remote_file_system.CacheFileSystem {
	return t.remoteDestCacheFS
}

func TestCacheSyncerWorker(t *testing.T) {
	Convey("Testing the cache-syncer worker", t, func() {
		ctx, err := testing_utils.NewContext(
			false, //
		)
		So(err, ShouldBeNil)
		So(ctx, ShouldNotBeNil)

		// testingLogger := ctx.Logger.(*testing_utils.TestingLogger)

		//TODO: Implement these tests
	})
}
