package utils

import (
	"os"
	"path/filepath"
)

// IsStringInSlice check if a string is in a given slice
func IsStringInSlice(s string, sl []string) bool {
	for i := range sl {
		if sl[i] == s {
			return true
		}
	}
	return false
}

func CalculateDirSizeInMb(path string) (float64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	const bytesInMegabyte = 1000 * 1000
	return float64(size) / bytesInMegabyte, nil // Return size in Megabyte
}
