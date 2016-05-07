package remote_file_system

func NewBaseDirFactory(baseDir string) Factory {
	return &baseDirFactory{baseDir: baseDir}
}

type baseDirFactory struct {
	baseDir string
}

func (b *baseDirFactory) New(remoteTempDir, jobId string) FileSystem {
	//Yes we are ignoring the 'remoteTempDir' but instead using the specified baseDir
	return NewBaseDirFileSystem(b.baseDir, jobId)
}
