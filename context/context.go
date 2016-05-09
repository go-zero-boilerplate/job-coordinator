package context

import (
	"github.com/francoishill/afero"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_comms_facade"
)

//New creates a new instance of Context
func New(
	logger logger.Logger,
	pendingLocalFileSystem,
	completedLocalFileSystem afero.Fs,
	localExecLoggerBinPath string,
	remoteCommsFactory remote_comms_facade.Factory,
	services *Services) *Context {

	return &Context{
		Logger:                   logger,
		PendingLocalFileSystem:   pendingLocalFileSystem,
		CompletedLocalFileSystem: completedLocalFileSystem,
		LocalExecLoggerBinPath:   localExecLoggerBinPath,
		RemoteCommsFactory:       remoteCommsFactory,
		Services:                 services,
	}
}

//Context holds globally used context like Logger and Services
type Context struct {
	Logger                   logger.Logger
	PendingLocalFileSystem   afero.Fs
	CompletedLocalFileSystem afero.Fs
	LocalExecLoggerBinPath   string
	RemoteCommsFactory       remote_comms_facade.Factory
	Services                 *Services
}

//CloneAndUseLogger will keep all the fields the same of the context but use a different logger
func (c *Context) CloneAndUseLogger(logger logger.Logger) *Context {
	return New(logger, c.PendingLocalFileSystem, c.CompletedLocalFileSystem, c.LocalExecLoggerBinPath, c.RemoteCommsFactory, c.Services)
}
