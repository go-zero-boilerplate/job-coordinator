package remote_file_system

type CacheFileSystem interface {
	GetFullPath() string
}

func NewSimpleCacheFileSystem(fullPath string) CacheFileSystem {
	return &simpleCacheFileSystem{
		fullPath: fullPath,
	}
}

type simpleCacheFileSystem struct {
	fullPath string
}

func (s *simpleCacheFileSystem) GetFullPath() string {
	return s.fullPath
}
