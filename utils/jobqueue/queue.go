package jobqueue

import (
	"sync"
)

//StartNewQueue will start a new job queue after initially adding the number of workers (worker==goroutine)
func StartNewQueue(initialWorkers int) *Queue {
	jobsChannel := make(chan Job)
	resultsChannel := make(chan JobResult)
	j := &Queue{jobsChannel: jobsChannel, resultsChannel: resultsChannel}
	return j.AddAndStartWorkers(initialWorkers)
}

//Queue is the manager containing all the queuing and closing of channels logic
type Queue struct {
	workersLock       sync.RWMutex
	queuedJobsLock    sync.RWMutex
	completedJobsLock sync.RWMutex

	jobsChannel    chan Job
	resultsChannel chan JobResult
	workers        []*worker

	totalQueuedCount    int
	totalCompletedCount int
	jobQueuingDone      bool
}

//AddAndStartWorkers will add the number of workers and also start them (each on their own goroutine)
func (j *Queue) AddAndStartWorkers(num int) *Queue {
	j.workersLock.Lock()
	defer j.workersLock.Unlock()

	for i := 0; i < num; i++ {
		worker := &worker{}
		j.workers = append(j.workers, worker)
		workerId := len(j.workers)
		go worker.start(workerId, j.jobsChannel, j.resultsChannel, j.onResultSent)
	}
	return j
}

//QueueJob will queue a job
func (j *Queue) QueueJob(job Job) {
	j.queuedJobsLock.Lock()
	defer j.queuedJobsLock.Unlock()

	j.totalQueuedCount++
	j.jobsChannel <- job
}

//ResultsChannel just returns a read-only result channel
func (j *Queue) ResultsChannel() <-chan JobResult {
	return j.resultsChannel
}

//JobQueuingDone is called from the "user" when all jobs are done
func (j *Queue) JobQueuingDone() {
	j.queuedJobsLock.Lock()
	j.completedJobsLock.Lock()
	defer j.queuedJobsLock.Unlock()
	defer j.completedJobsLock.Unlock()
	j.jobQueuingDone = true
	close(j.jobsChannel)
	j.checkAllResultsDone()
}

func (j *Queue) onResultSent(job Job) {
	j.completedJobsLock.Lock()
	defer j.completedJobsLock.Unlock()
	j.totalCompletedCount++
	j.checkAllResultsDone()
}

func (j *Queue) checkAllResultsDone() {
	if j.jobQueuingDone && j.totalQueuedCount <= j.totalCompletedCount {
		close(j.resultsChannel)
	}
}
