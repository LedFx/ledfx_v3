package util

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

	tempDir := "temp"
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
