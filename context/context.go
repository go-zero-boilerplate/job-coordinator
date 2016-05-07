package context

import (
	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

func New(logger logger.Logger, pendingLocalFileSystem, completedLocalFileSystem afero.Fs, localExecLoggerBinPath string, remoteCommsFactory remote_comms_facade.Factory) *Context {
	return &Context{
		Logger:                   logger,
		PendingLocalFileSystem:   pendingLocalFileSystem,
		CompletedLocalFileSystem: completedLocalFileSystem,
		LocalExecLoggerBinPath:   localExecLoggerBinPath,
		RemoteCommsFactory:       remoteCommsFactory,
	}
}

type Context struct {
	Logger                   logger.Logger
	PendingLocalFileSystem   afero.Fs
	CompletedLocalFileSystem afero.Fs
	LocalExecLoggerBinPath   string
	RemoteCommsFactory       remote_comms_facade.Factory
}

func (c *Context) CloneAndUseLogger(logger logger.Logger) *Context {
	return New(logger, c.PendingLocalFileSystem, c.CompletedLocalFileSystem, c.LocalExecLoggerBinPath, c.RemoteCommsFactory)
}
