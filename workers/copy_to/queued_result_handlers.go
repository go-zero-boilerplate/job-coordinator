package copy_to

import "sync"

type queuedResultHandlers struct {
	lock     sync.Locker
	handlers map[*queuedJob]OnResult
}

func (q *queuedResultHandlers) AddHandler(j *queuedJob, onResult OnResult) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.handlers[j] = onResult
}
