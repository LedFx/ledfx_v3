package util

import "os"

func FileExists(location string) bool {
	_, err := os.Stat(location)
	return err == nil
}
