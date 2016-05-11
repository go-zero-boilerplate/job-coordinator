package copy_to

import (
	"sync"

	"github.com/go-zero-boilerplate/job-coordinator/context"
	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/go-zero-boilerplate/job-coordinator/utils/jobqueue"
)

type hostQueues struct {
	lock           sync.RWMutex
	logger         logger.Logger
	hostQueues     map[string]*jobqueue.Queue
	resultHandlers *queuedResultHandlers
}

func (h *hostQueues) QueueJob(c *copyTo, ctx *context.Context, job Job, onResult OnResult, maxGoRoutinesPerHost int) {
	queuedJob := &queuedJob{
		copyToWorker: c,
		ctx:          ctx,
		job:          job,
	}
	h.resultHandlers.AddHandler(queuedJob, onResult)

	q := h.getQueueForHost(job.HostDetails(), maxGoRoutinesPerHost)
	q.QueueJob(queuedJob)
}

func (h *hostQueues) getQueueForHost(hostDetails host_details.HostDetails, maxGoRoutinesPerHost int) *jobqueue.Queue {
	h.lock.Lock()
	defer h.lock.Unlock()

	hostName := hostDetails.HostName()
	if q, ok := h.hostQueues[hostName]; ok {
		return q
	}

	q := jobqueue.StartNewQueue(maxGoRoutinesPerHost)
	h.hostQueues[hostName] = q
	go h.startQueueResultProcessing(hostDetails, q)
	return q
}

func (h *hostQueues) removeQueueForHost(hostDetails host_details.HostDetails) {
	h.lock.Lock()
	defer h.lock.Unlock()
	hostName := hostDetails.HostName()
	if _, ok := h.hostQueues[hostName]; ok {
		delete(h.hostQueues, hostName)
	}
}

func (h *hostQueues) startQueueResultProcessing(hostDetails host_details.HostDetails, queue *jobqueue.Queue) {
	defer h.logger.DeferredRecoverStack("Error processing results")

	for result := range queue.ResultsChannel() {
		h.logger.Debug("TEMP: Job done, result %+v (type = %T)", result, result)
		h.logger.WithField("result", result).Debug("TEMP: Got result")
		if err := h.resultHandlers.HandleResult(result); err != nil {
			h.logger.WithError(err).Error("Failed to handle OnResult")
		}
	}
	h.removeQueueForHost(hostDetails)
}
