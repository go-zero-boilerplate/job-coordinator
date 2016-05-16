package jobqueue

type JobResult struct {
	job   Job
	Error error
}

func (j *JobResult) Job() Job {
	return j.job
}
