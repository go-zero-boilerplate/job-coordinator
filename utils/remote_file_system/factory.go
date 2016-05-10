package remote_file_system

type Factory interface {
	New(jobId string) FileSystem
}
