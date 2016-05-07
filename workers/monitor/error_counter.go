package monitor

import (
	"time"
)

type errorCounter struct {
	max     int
	current int
}

func (e *errorCounter) Inc() *errorCounter {
	e.current++
	return e
}

func (e *errorCounter) CapReached() bool {
	return e.current > e.max
}

func (e *errorCounter) Reset() *errorCounter {
	e.current = 0
	return e
}

func (e *errorCounter) GetSleepDuration() time.Duration {
	num := e.current
	if num <= 0 {
		num = 1
	}
	return time.Duration(2*num) * time.Second
}
