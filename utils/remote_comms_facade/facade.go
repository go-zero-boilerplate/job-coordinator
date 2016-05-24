package remote_comms_facade

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-zero-boilerplate/osvisitors"

	"github.com/golang-devops/go-psexec/shared/dtos"
	"github.com/golang-devops/go-psexec/shared/tar_io"

	gpClient "github.com/golang-devops/go-psexec/client"

	"github.com/go-zero-boilerplate/job-coordinator/logger"
	"github.com/go-zero-boilerplate/job-coordinator/utils/host_details"
	"github.com/golang-devops/go-psexec/services/filepath_summary"
)

type Facade interface {
	Ping() error
	ConfirmVersionMatch(version string) error

	GetTempDir() (string, error)
	GetOsType() (osvisitors.OsType, error)

	// Start() (*StartedDetails, error)
	StartDetached(workingDir string, commandLine ...string) (*StartedDetails, error)
	// Run() (*Result, error)

	CheckPathsAreInSync(logger logger.Logger, localPath, remotePath string) (bool, error)

	Upload(localPath, remotePath string) error
	DownloadDir(remotePath, localPath string) error

	ReadFileContent(remotePath string) ([]byte, error)
	UploadFileContent(remotePath string, content []byte) error

	Stats(remotePath string) (*dtos.StatsDto, error)
	Copy(srcRemotePath, destRemotePath string) error
	Symlink(srcRemotePath, destRemoteSymlinkPath string) error
	Move(oldRemotePath, newRemotePath string) error

	Delete(remotePath string) error
}

type facade struct {
	logger                 logger.Logger
	gopsexecClient         *gpClient.Client
	hostDetails            host_details.HostDetails
	filepathSummaryService filepath_summary.Service
}

func (f *facade) newGoPsExecSession() (gpClient.Session, error) {
	//TODO: This port/address for the GoPsexec server on the remote machine should be customizable or obtained some central config server (like etcd or consul)
	return f.gopsexecClient.RequestNewSession(fmt.Sprintf("http://%s:62677", f.hostDetails.HostName()))
}

func (f *facade) startExecRequest(detached bool, exe, workingDir string, args ...string) (*StartedDetails, error) {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return nil, fmt.Errorf("Cannot obtain GoPsExec session, error: %s", err.Error())
	}

	remoteOsType, err := f.getOsTypeFromSession(session)
	if err != nil {
		return nil, fmt.Errorf("Cannot determine OsType, error: %s", err.Error())
	}

	createRespVisitor := &startExecVisitor{session: session, detached: detached, exe: exe, workingDir: workingDir, args: args}
	remoteOsType.Accept(createRespVisitor)
	if createRespVisitor.err != nil {
		return nil, fmt.Errorf("Error: %s. Cannot start command '%s' (args '%s')", createRespVisitor.err.Error(), exe, args)
	}

	feedbackChan, errorChan := createRespVisitor.resp.TextResponseChannel()
	return &StartedDetails{
		Pid:          createRespVisitor.resp.Pid,
		FeedbackChan: feedbackChan,
		ErrorChan:    errorChan,
	}, nil
}

func (f *facade) Ping() error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return fmt.Errorf("Cannot ping, error: %s", err.Error())
	}
	return session.Ping()
}

func (f *facade) ConfirmVersionMatch(version string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return fmt.Errorf("Cannot get session, error: %s", err.Error())
	}
	serverVersion, err := session.Version()
	if err != nil {
		return fmt.Errorf("Cannot get version, error: %s", err.Error())
	}

	if strings.TrimSpace(serverVersion) != strings.TrimSpace(version) {
		return fmt.Errorf("Server version is '%s' but was expected to be '%s'", serverVersion, version)
	}
	return nil
}

func (f *facade) GetTempDir() (string, error) {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return "", err
	}

	dto, err := session.GetTempDir()
	if err != nil {
		return "", err
	}
	return dto.TempDir, nil
}

func (f *facade) getOsTypeFromSession(session gpClient.Session) (osvisitors.OsType, error) {
	dto, err := session.GetOsTypeName()
	if err != nil {
		return nil, err
	}
	return osvisitors.ParseFromName(dto.Name)
}

func (f *facade) GetOsType() (osvisitors.OsType, error) {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return nil, err
	}

	return f.getOsTypeFromSession(session)
}

func (f *facade) StartDetached(workingDir string, commandLine ...string) (*StartedDetails, error) {
	/*close(startedDetails.ErrorChan)
	close(startedDetails.FeedbackChan)*/
	detached := true
	return f.startExecRequest(detached, commandLine[0], workingDir, commandLine[1:]...)
}

func (f *facade) timesAreEqual(t1, t2 time.Time) bool {
	if t1.Equal(t2) {
		return true
	}

	timestampFormat := "2006-01-02 15:04:05"
	timestamp1 := t1.Format(timestampFormat)
	timestamp2 := t2.Format(timestampFormat)
	if timestamp1 == timestamp2 {
		return true
	}

	return false
}

func (f *facade) fileSummariesAreInSync(logger logger.Logger, localSummary, remoteSummary *filepath_summary.FileSummary) bool {
	if !f.timesAreEqual(localSummary.ModTime, remoteSummary.ModTime) {
		logger.WithField("local-time", localSummary.ModTime).WithField("remote-time", remoteSummary.ModTime).Debug("File times out of sync")
		return false
	}
	if localSummary.Checksum.HexString() != remoteSummary.Checksum.HexString() {
		logger.WithField("local-checksum", localSummary.Checksum.HexString()).WithField("remote-checksum", remoteSummary.Checksum.HexString()).Debug("File checksums out of sync")
		return false
	}

	return true
}

func (f *facade) dirSummariesAreInSync(logger logger.Logger, localSummary, remoteSummary *filepath_summary.DirSummary) bool {
	if len(localSummary.FlattenedFileSummaries) > len(remoteSummary.FlattenedFileSummaries) {
		logger.
			WithField("local-count", len(localSummary.FlattenedFileSummaries)).
			WithField("remote-count", len(remoteSummary.FlattenedFileSummaries)).
			Debug("Remote has less files than local")
		return false
	}

	for _, localFileSummary := range localSummary.FlattenedFileSummaries {
		var foundRemoteSummary *filepath_summary.FileSummary
		for _, remoteFileSummary := range remoteSummary.FlattenedFileSummaries {
			if remoteFileSummary.HaveSamePath(localFileSummary) {
				foundRemoteSummary = remoteFileSummary
				break
			}
		}
		if foundRemoteSummary == nil {
			logger.WithField("local-file", localFileSummary.RelativePath).Debug("Dir out of sync - remote file missing")
			return false
		}

		subLogger := logger.WithField("local-file", localFileSummary.RelativePath).WithField("remote-file", foundRemoteSummary.RelativePath)
		if !f.fileSummariesAreInSync(subLogger, localFileSummary, foundRemoteSummary) {
			return false
		}
	}

	return true
}

func (f *facade) CheckPathsAreInSync(logger logger.Logger, localPath, remotePath string) (bool, error) {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return false, err
	}
	logger = logger.WithField("local-path", localPath).WithField("remote-path", remotePath)

	info, err := os.Stat(localPath)
	if err != nil {
		return false, fmt.Errorf("Unable to obtain stats of path '%s', error: %s", localPath, err.Error())
	}

	//TODO: We can perhaps make this two-phase checking so that we first check on `ModTime.Format("")` and `Size()`. The question is do we actually need the checksum?
	sessionFS := session.FileSystem()
	if info.IsDir() {
		localDirSummary, err := f.filepathSummaryService.GetDirSummary(localPath)
		if err != nil {
			return false, err
		}

		remoteDirSummary, err := sessionFS.DirSummary(remotePath)
		if err != nil {
			return false, err
		}

		if f.dirSummariesAreInSync(logger, localDirSummary, remoteDirSummary) {
			logger.Debug("Dirs are in sync")
			return true, nil
		}
	} else {
		baseDir := "" //TODO: refer to go-psexec todo `TODO: 2016-05-09 20:57` for why this is a hacky solution
		localFileSummary, err := f.filepathSummaryService.GetFileSummary(baseDir, localPath)
		if err != nil {
			return false, err
		}

		remoteFileSummary, err := sessionFS.FileSummary(remotePath)
		if err != nil {
			return false, err
		}

		if f.fileSummariesAreInSync(logger, localFileSummary, remoteFileSummary) {
			logger.Debug("Files are in sync")
			return true, nil
		}
	}
	return false, nil
}

func (f *facade) Upload(localPath, remotePath string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("Unable to obtain stats of path '%s', error: %s", localPath, err.Error())
	}

	var tarProvider tar_io.TarProvider
	if info.IsDir() {
		tarProvider = tar_io.Factories.TarProvider.Dir(localPath, "")
	} else {
		tarProvider = tar_io.Factories.TarProvider.File(localPath)
	}
	return session.FileSystem().UploadTar(tarProvider, remotePath, info.IsDir())
}

func (f *facade) DownloadDir(remotePath, localPath string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	tarReceiver := tar_io.Factories.TarReceiver.Dir(localPath)
	return session.FileSystem().DownloadTar(remotePath, nil, tarReceiver)
}

func (f *facade) ReadFileContent(remotePath string) ([]byte, error) {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = session.FileSystem().DownloadTar(remotePath, nil, tar_io.Factories.TarReceiver.Writer(buf))
	if err != nil {
		return nil, fmt.Errorf("Unable to read content of remote file '%s', error: %s", remotePath, err.Error())
	}
	return buf.Bytes(), nil
}

func (f *facade) UploadFileContent(remotePath string, content []byte) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	tempFile, err := ioutil.TempFile(os.TempDir(), "remote-facade-upload-")
	if err != nil {
		return fmt.Errorf("Cannot create temp file for uploading, error: %s", err.Error())
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("Unable to close tempfile before attempting upload, error: %s", err.Error())
	}

	if err = ioutil.WriteFile(tempFile.Name(), content, 0655); err != nil {
		return fmt.Errorf("Unable to write to temp file '%s', error: %s", tempFile.Name(), err.Error())
	}

	tarProvider := tar_io.Factories.TarProvider.File(tempFile.Name())
	isDir := false
	err = session.FileSystem().UploadTar(tarProvider, remotePath, isDir)
	if err != nil {
		return fmt.Errorf("Unable to upload local file '%s' to remote file '%s', error: %s", tempFile.Name(), remotePath, err.Error())
	}
	return nil
}

func (f *facade) Stats(remotePath string) (*dtos.StatsDto, error) {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return nil, err
	}

	statsDTO, err := session.FileSystem().Stats(remotePath)
	if err != nil {
		return nil, fmt.Errorf("Cannot determine stats for remote path '%s', error: %s", remotePath, err.Error())
	}

	return statsDTO, nil
}

func (f *facade) Copy(srcRemotePath, destRemotePath string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	if err := session.FileSystem().Copy(srcRemotePath, destRemotePath); err != nil {
		return fmt.Errorf("Unable to copy remote path from '%s' to '%s', error: %s", srcRemotePath, destRemotePath, err.Error())
	}
	return nil
}

func (f *facade) Symlink(srcRemotePath, destRemoteSymlinkPath string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	if err := session.FileSystem().Symlink(srcRemotePath, destRemoteSymlinkPath); err != nil {
		return fmt.Errorf("Unable to symlink remote path from '%s' to '%s', error: %s", srcRemotePath, destRemoteSymlinkPath, err.Error())
	}
	return nil
}

func (f *facade) Move(oldRemotePath, newRemotePath string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	if err := session.FileSystem().Move(oldRemotePath, newRemotePath); err != nil {
		return fmt.Errorf("Unable to move remote path from '%s' to '%s', error: %s", oldRemotePath, newRemotePath, err.Error())
	}
	return nil
}

func (f *facade) Delete(remotePath string) error {
	session, err := f.newGoPsExecSession()
	if err != nil {
		return err
	}

	if err := session.FileSystem().Delete(remotePath); err != nil {
		return fmt.Errorf("Unable to delete remote path '%s', error: %s", remotePath, err.Error())
	}
	return nil
}
