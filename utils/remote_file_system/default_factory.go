package remote_file_system

func NewDefaultFactory() Factory {
	return &defaultFactory{}
}

type defaultFactory struct{}

func (d *defaultFactory) New(remoteTempDir, jobId string) FileSystem {
	return NewDefaultFileSystem(remoteTempDir, "received", jobId)
}
