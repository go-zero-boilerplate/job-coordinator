package remote_comms_facade

import "strings"

type Result struct {
	FeedbackLines []string
	Errors        []error
}

func (r *Result) AppendFeedback(feedback string) {
	r.FeedbackLines = append(r.FeedbackLines, feedback)
}

func (r *Result) AppendError(err error) {
	r.Errors = append(r.Errors, err)
}

func (r *Result) ErrorStrings() (errStrs []string) {
	for _, e := range r.Errors {
		errStrs = append(errStrs, e.Error())
	}
	return
}

func (r *Result) CombinedErrorLines() string {
	return strings.Join(r.ErrorStrings(), "\\n")
}

func (r *Result) CombinedFeedbackLines() string {
	return strings.Join(r.FeedbackLines, "\\n")
}
