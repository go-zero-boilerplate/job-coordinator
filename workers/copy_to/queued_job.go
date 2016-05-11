package copy_to

import (
	"github.com/go-zero-boilerplate/job-coordinator/context"
)

type queuedJob struct {
	copyToWorker *copyTo
	ctx          *context.Context
	job          Job
}

func (q *queuedJob) Do(workerId int) interface{} {
	return q.copyToWorker.DoJob(q.ctx, q.job)
}
