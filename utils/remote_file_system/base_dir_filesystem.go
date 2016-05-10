package remote_file_system

import "path/filepath"

func NewBaseDirFileSystem(baseDir, jobId string) FileSystem {
	baseJobDir := filepath.Join(baseDir, jobId)
	return &baseDirFileSystem{baseDir: baseJobDir}
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
