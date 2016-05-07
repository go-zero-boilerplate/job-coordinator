package logger

type LogTracer interface {
	Stop(errPtr *error)
}
