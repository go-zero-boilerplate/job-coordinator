package jobqueue

//Job is the job that gets "executed"
type Job interface {
	//Do is called by the worker to run and return the error (result should be set on the job object itself)
	Do(workerID int) error
}
