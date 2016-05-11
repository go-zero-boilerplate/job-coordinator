package jobqueue

//Job is the job that gets "executed"
type Job interface {
	//Do is called by the worker to run and return the result
	Do(workerId int) interface{}
}
