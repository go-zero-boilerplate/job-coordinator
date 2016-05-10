package remote_file_system

func NewBaseDirFactory(baseDir string) Factory {
	return &baseDirFactory{baseDir: baseDir}
}

type baseDirFactory struct {
	baseDir string
}

func (b *baseDirFactory) New(jobId string) FileSystem {
	return NewBaseDirFileSystem(b.baseDir, jobId)
}
