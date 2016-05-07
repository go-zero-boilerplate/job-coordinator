package remote_file_system

type FileSystem interface {
	GetFullJobDir(relativeParts ...string) string
}
