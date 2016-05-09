package context

import (
	"github.com/golang-devops/go-psexec/services/encoding/checksums"
	"github.com/golang-devops/go-psexec/services/filepath_summary"
)

//Services holds the different services used globally
type Services struct {
	Checksums         checksums.Service
	FilePathSummaries filepath_summary.Service
}

//NewServices creates a new instance of Services
func NewServices(checksums checksums.Service, filePathSummaries filepath_summary.Service) *Services {
	return &Services{
		Checksums:         checksums,
		FilePathSummaries: filePathSummaries,
	}
}
