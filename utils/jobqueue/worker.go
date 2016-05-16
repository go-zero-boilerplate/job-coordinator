package jobqueue

type worker struct{}

func (w *worker) start(workerID int, jobs <-chan Job, results chan<- JobResult, onResultSent func(job Job)) {
	for job := range jobs {
		err := job.Do(workerID)
		results <- JobResult{job: job, Error: err}
		onResultSent(job)
	}
}
