package utils

import (
	"os"
	"path/filepath"
)

// GetAssetsPath returns the absolute path to the assets directory
func GetAssetsPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "assets"
	}
	// On development (go run), os.Executable() points to a temp file
	// We handle this by checking if the assets folder exists relative to the exe
	exeDir := filepath.Dir(exe)
	assetsDir := filepath.Join(exeDir, "assets")

	if _, err := os.Stat(assetsDir); err == nil {
		return assetsDir
	}

	// Fallback to relative path for dev
	return "assets"
}
