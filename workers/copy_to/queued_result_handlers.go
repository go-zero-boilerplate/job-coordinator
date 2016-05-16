package copy_to

import (
	"fmt"
	"sync"

	"github.com/go-zero-boilerplate/job-coordinator/utils/jobqueue"
)

type OnResult interface {
	OnSuccess(job Job)
	OnError(err error)
}

type queuedResultHandlers struct {
	lock     sync.RWMutex
	handlers map[*queuedJob]OnResult
}

func (q *queuedResultHandlers) AddHandler(j *queuedJob, onResult OnResult) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.handlers[j] = onResult
}

func (q *queuedResultHandlers) HandleResult(result jobqueue.JobResult) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	resultJob := result.Job()
	for qj, onResult := range q.handlers {
		if resultJob == qj {
			if result.Error == nil {
				onResult.OnSuccess(qj.job)
			} else {
				onResult.OnError(result.Error)
			}
			return nil
		}
	}

	return fmt.Errorf("No handler found for result-job %+v (type = %T)", resultJob, resultJob)
}
