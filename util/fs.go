package util

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExecDir returns the absolute path of given relative path to the current executable dir.
// If relPath is "", it just returns the absolute path of current executable dir.
func ExecDir() (string, error) {
	// os.Executable requires Go 1.18+
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	ex, err = filepath.EvalSymlinks(ex)
	if err != nil {
		return "", err
	}
	ex, err = filepath.Abs(ex)

	// little check for developers. This should detect if we're running in a dev environment
	if !strings.HasPrefix(ex, "/tmp/go-build") {
		return filepath.Dir(ex), err
	}

	// for devs, extract to cwd
	ex, err = os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(ex)
}

func FileExists(location string) bool {
	_, err := os.Stat(location)
	return err == nil
}

func Unzip(path, dest string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer archive.Close()

	ex, err := ExecDir()
	if err != nil {
		return err
	}
	tempDir := filepath.Join(ex, "temp")
	defer os.RemoveAll(tempDir)

	for _, f := range archive.File {
		filePath := filepath.Join(tempDir, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(tempDir)+string(os.PathSeparator)) {
			return errors.New("invalid path")
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, fileInArchive)
		dstFile.Close()
		fileInArchive.Close()
		if err != nil {
			return err
		}
	}
	return os.Rename(filepath.Join(tempDir, "ledfx_frontend_v2"), dest)

}
