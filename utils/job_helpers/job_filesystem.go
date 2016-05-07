package job_helpers

import "github.com/francoishill/afero"

func GetJobFileSystem(fs afero.Fs, jobId string) afero.Fs {
	return afero.NewBasePathFs(fs, jobId)
}
