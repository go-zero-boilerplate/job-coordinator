package logger

type LogTracer interface {
	Stop(errPtr *error)
}

type LogDebugTracer interface {
	StopDebug(errPtr *error)
}
