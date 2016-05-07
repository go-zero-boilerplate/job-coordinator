package remote_comms_facade

type StartedDetails struct {
	Pid          int
	FeedbackChan <-chan string
	ErrorChan    <-chan error
}

func (s *StartedDetails) Wait() *Result {
	linesDone := false
	errorsDone := false
	result := &Result{}

outerFor:
	for {
		select {
		case feedbackLine, ok := <-s.FeedbackChan:
			if !ok {
				linesDone = true
			} else {
				result.AppendFeedback(feedbackLine)
			}
			if linesDone && errorsDone {
				break outerFor
			}
		case errLine, ok := <-s.ErrorChan:
			if !ok {
				errorsDone = true
			} else {
				result.AppendError(errLine)
			}
			if linesDone && errorsDone {
				break outerFor
			}
		}
	}

	return result
}
