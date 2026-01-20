package utils

import (
	"os"
	"path/filepath"
)

// GetAssetsPath returns the absolute path to the assets directory
func GetAssetsPath() string {
	// Try to find project root by looking for go.mod
	dir, _ := os.Getwd()
	for dir != "" && dir != "." && dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Found root where go.mod is
			assetsDir := filepath.Join(dir, "assets")
			if _, err := os.Stat(assetsDir); err == nil {
				return assetsDir
			}
		}
		dir = filepath.Dir(dir)
	}

	// Fallback to executable relative path
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		assetsDir := filepath.Join(exeDir, "assets")
		if _, err := os.Stat(assetsDir); err == nil {
			return assetsDir
		}
	}

	return "assets"
}
