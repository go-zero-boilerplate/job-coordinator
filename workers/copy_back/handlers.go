package copy_back

//Handlers is an interface for common events we might want to handle by the consumer of job-coordinator
type Handlers interface {
	FailedToCleanupRemotePathBeforeCopyBack(err error, fullRemotePath string)
	FailedToRemoveRemoteJobDir(err error, remoteJobPath string)
}
