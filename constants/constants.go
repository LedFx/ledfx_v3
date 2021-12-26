package constants

import (
	"os"
	"path/filepath"
	"runtime"
)

var CONFIG_DIR = ".ledfx"
var VERSION = "v0.0.1"

func GetOsConfigDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return filepath.Join(os.Getenv("HOME"), CONFIG_DIR)
}
