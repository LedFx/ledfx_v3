package constants

import (
	"os"
	"path/filepath"
	"runtime"
)

const CONFIG_DIR = ".ledfx"
const VERSION = "v2.1.1"

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
