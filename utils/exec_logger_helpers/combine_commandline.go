package exec_logger_helpers

import (
	"time"

	"github.com/go-zero-boilerplate/job-coordinator/utils/remote_file_system"
	"github.com/go-zero-boilerplate/osvisitors"
)

func CombineExecLoggerCommandline(remoteOsType osvisitors.OsType, remoteJobFS remote_file_system.FileSystem, commandLine []string, timeout *time.Duration, recordResourceUsage bool) (combinedArgs []string) {
	remoteExecLoggerFullPath := remoteJobFS.GetFullJobDir(GetExecLoggerBinFileName(remoteOsType))
	combinedArgs = []string{
		remoteExecLoggerFullPath,
		"-task",
		"exec",
	}
	if timeout != nil && *timeout > 0 {
		combinedArgs = append(combinedArgs, "-timeout-kill", timeout.String())
	}
	if recordResourceUsage {
		combinedArgs = append(combinedArgs, "-record-resource-usage")
	}
	combinedArgs = append(combinedArgs, commandLine...)
	return
}
