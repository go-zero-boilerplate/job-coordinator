package remote_file_system

import "path/filepath"

func NewDefaultFileSystem(remoteTempDir, receiveSubDir, jobId string) FileSystem {
	basePath := filepath.Join(remoteTempDir, "job-coordinator", receiveSubDir, jobId)
	return &defaultFileSystem{basePath: basePath}
}

type defaultFileSystem struct {
	basePath string
}

func (d *defaultFileSystem) GetFullJobDir(relativeParts ...string) string {
	elems := []string{
		d.basePath,
	}
	elems = append(elems, relativeParts...)
	return filepath.Join(elems...)
}
