package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/francoishill/afero"
	"github.com/golang-devops/go-psexec/shared/io_throttler"
)

func SaveFile(fileSystem afero.Fs, pathInsideFileSystem string, reader io.Reader) error {
	err := fileSystem.MkdirAll(filepath.Dir(pathInsideFileSystem), 0755)
	if err != nil {
		return fmt.Errorf("Cannot create parent dir of %s, error: %s", pathInsideFileSystem, err.Error())
	}

	file, err := fileSystem.Create(pathInsideFileSystem)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io_throttler.CopyThrottled(io_throttler.DefaultIOThrottlingBandwidth, file, reader)
	return err
}

func ExportFile(sourceFs afero.Fs, relFilePath string, destFileSystem afero.Fs) error {
	sourceFile, err := sourceFs.Open(relFilePath)
	if err != nil {
		return fmt.Errorf("Cannot open source file '%s', error: %s", relFilePath, err.Error())
	}
	defer sourceFile.Close()

	if err := destFileSystem.MkdirAll("", 0755); err != nil {
		return fmt.Errorf("Unable to create parent dir of file '%s', error: %s", sourceFile, err.Error())
	}

	fileNameOnly := filepath.Base(relFilePath)
	destFile, err := destFileSystem.Create(fileNameOnly)
	if err != nil {
		return fmt.Errorf("Cannot open destination file '%s' (FileSystem %+v), error: %s", relFilePath, destFileSystem, err.Error())
	}
	defer destFile.Close()

	_, err = io_throttler.CopyThrottled(io_throttler.DefaultIOThrottlingBandwidth, destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("Unable to copy source file to dest file, error: %s", err.Error())
	}

	return nil
}

func ExportDir(sourceFs afero.Fs, exportFileSystem afero.Fs) error {
	//TODO: Is the 0755 permissions fine used below
	return afero.Walk(sourceFs, "", func(sourcePath string, info os.FileInfo, innerErr error) error {
		if innerErr != nil {
			return innerErr
		}

		relPath := sourcePath
		/*trimmedSourceDir := strings.Trim(sourceDir, "/\\ ")
		relPath := sourcePath[len(trimmedSourceDir):]
		if relPath == "" {
			return nil
		}
		relPath = relPath[1:]*/

		if info.IsDir() {
			return exportFileSystem.MkdirAll(relPath, 0755)
		}

		relDir := filepath.Dir(relPath)
		var subDirFileSystem afero.Fs
		if strings.TrimSpace(relDir) == "" || strings.TrimSpace(relDir) == "." {
			subDirFileSystem = exportFileSystem
		} else {
			subDirFileSystem = afero.NewBasePathFs(exportFileSystem, relDir)
		}
		return ExportFile(sourceFs, relPath, subDirFileSystem)
	})
}
