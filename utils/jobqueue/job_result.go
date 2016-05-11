package jobqueue

type JobResult struct {
	job    Job
	Result interface{}
}

func (j *JobResult) Job() Job {
	return j.job
}
