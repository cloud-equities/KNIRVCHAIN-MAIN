package KNIRVCHAIN

import (
	"os"
	"runtime"
)

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}
