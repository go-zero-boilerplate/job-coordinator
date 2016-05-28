package copy_to

import (
	"github.com/go-zero-boilerplate/job-coordinator/context"
)

type queuedJob struct {
	copyToWorker *copyTo
	ctx          *context.Context
	job          Job
	handlers     Handlers
}

//Do will execute the job
func (q *queuedJob) Do(workerID int) error {
	return q.copyToWorker.DoJob(q.ctx, q.job, q.handlers)
}
