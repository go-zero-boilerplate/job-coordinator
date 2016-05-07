package remote_file_system

type Factory interface {
	New(remoteTempDir, jobId string) FileSystem
}
