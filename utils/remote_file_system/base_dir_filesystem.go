package remote_file_system

import "path/filepath"

func NewBaseDirFileSystem(remoteTempDir, jobId string) FileSystem {
	baseDir := filepath.Join(remoteTempDir, jobId)
	return &baseDirFileSystem{baseDir: baseDir}
}

type baseDirFileSystem struct {
	baseDir string
}

func (d *baseDirFileSystem) GetFullJobDir(relativeParts ...string) string {
	elems := []string{
		d.baseDir,
	}
	elems = append(elems, relativeParts...)
	return filepath.Join(elems...)
}
