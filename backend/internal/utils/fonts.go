package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"webtracker-bot/internal/logger"

	"github.com/fogleman/gg"
)

var (
	fontDataMu sync.Mutex
)

func getFontsPath() string {
	return filepath.Join(GetAssetsPath(), "fonts")
}

const (
	FontCourierBold = "courier_bold.ttf"
	FontSignature   = "signature.ttf"
	FontBarcode     = "barcode.ttf"
	FontArial       = "arial.ttf"
	FontArialBold   = "arial_bold.ttf"
)

// EnsureFontsDownloader checks for fonts and downloads them if missing
func EnsureFontsDownloader() error {
	fontDataMu.Lock()
	defer fontDataMu.Unlock()

	if err := os.MkdirAll(getFontsPath(), 0755); err != nil {
		return fmt.Errorf("failed to create fonts dir: %w", err)
	}

	fonts := map[string]string{
		FontCourierBold: "https://github.com/google/fonts/raw/main/ofl/courierprime/CourierPrime-Bold.ttf",
		FontSignature:   "https://raw.githubusercontent.com/googlefonts/DancingScript/master/fonts/ttf/DancingScript-Bold.ttf",
		FontBarcode:     "https://github.com/google/fonts/raw/main/ofl/librebarcode39text/LibreBarcode39Text-Regular.ttf",
		FontArial:       "https://github.com/googlefonts/roboto/raw/main/src/hinted/Roboto-Regular.ttf",
		FontArialBold:   "https://github.com/googlefonts/roboto/raw/main/src/hinted/Roboto-Bold.ttf",
	}

	for filename, url := range fonts {
		path := filepath.Join(getFontsPath(), filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logger.Info().Str("font", filename).Msg("Downloading missing font...")
			if err := downloadFile(path, url); err != nil {
				return fmt.Errorf("failed to download font %s: %w", filename, err)
			}
		}
	}
	return nil
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// LoadFont helper to safely load a font face
func LoadFont(dc *gg.Context, fontName string, points float64) error {
	path := filepath.Join(getFontsPath(), fontName)
	return dc.LoadFontFace(path, points)
}
