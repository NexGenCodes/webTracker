package health

import (
	"fmt"
	"os"
	"path/filepath"
	"webtracker-bot/internal/utils"
)

func VerifyEnvironment() error {
	assets := utils.GetAssetsPath()

	requiredFiles := []string{
		filepath.Join(assets, "fonts", "arial.ttf"),
		filepath.Join(assets, "fonts", "arial_bold.ttf"),
		filepath.Join(assets, "img", "logo.png"),
		filepath.Join(assets, "img", "approved_stamp.png"),
	}

	for _, f := range requiredFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("missing critical resource: %s", f)
		}
	}

	// Check for .env fallback
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Println("Warning: .env file not found, relying on system environment variables")
	}

	return nil
}
