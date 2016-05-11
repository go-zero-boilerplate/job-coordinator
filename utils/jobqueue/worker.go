package jobqueue

type worker struct{}

func (w *worker) start(workerId int, jobs <-chan Job, results chan<- JobResult, onResultSent func(job Job)) {
	for job := range jobs {
		result := job.Do(workerId)
		results <- JobResult{job: job, Result: result}
		onResultSent(job)
	}
}
