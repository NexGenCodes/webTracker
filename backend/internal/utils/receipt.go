package utils

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/models"

	"github.com/chromedp/chromedp"
)

var (
	receiptRenderer *ReceiptRenderer
	rendererOnce    sync.Once
	rendererMu      sync.RWMutex
	isHealthy       int32 = 1
)

type ReceiptRenderer struct {
	allocCtx    context.Context
	cancel      context.CancelFunc
	tmpl        *template.Template
	renderMu    sync.Mutex
	renderCount int
}

func InitReceiptRenderer() error {
	var initErr error
	rendererOnce.Do(func() {
		receiptRenderer, initErr = newReceiptRenderer()
	})
	return initErr
}

func newReceiptRenderer() (*ReceiptRenderer, error) {
	// Parse template
	tmplPath := filepath.Join("internal", "utils", "receipt_template.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse receipt template: %w", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("single-process", true), // Critical for low-memory environments
		chromedp.Flag("headless", true),
		chromedp.WindowSize(1200, 1600),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	testCtx, testCancel := chromedp.NewContext(allocCtx)
	defer testCancel()

	testTimeout, testTimeoutCancel := context.WithTimeout(testCtx, 15*time.Second) // Increased init timeout
	defer testTimeoutCancel()

	if err := chromedp.Run(testTimeout); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to launch Chrome: %w", err)
	}

	logger.Info().Msg("Chrome instance initialized successfully")

	return &ReceiptRenderer{
		allocCtx: allocCtx,
		cancel:   cancel,
		tmpl:     tmpl,
	}, nil
}

// RenderReceipt generates a high-resolution receipt image from shipment data
func RenderReceipt(shipment models.Shipment, companyName string) ([]byte, error) {
	if receiptRenderer == nil {
		return nil, fmt.Errorf("receipt renderer not initialized")
	}
	return receiptRenderer.render(shipment, companyName)
}

func (r *ReceiptRenderer) render(shipment models.Shipment, companyName string) ([]byte, error) {
	r.renderMu.Lock()
	defer r.renderMu.Unlock()

	r.renderCount++
	shouldRecycle := r.renderCount%10 == 0

	ctx, cancel := chromedp.NewContext(r.allocCtx)
	defer cancel()

	var htmlBuf bytes.Buffer
	data := struct {
		models.Shipment
		CompanyName string
	}{
		Shipment:    shipment,
		CompanyName: companyName,
	}

	if err := r.tmpl.Execute(&htmlBuf, data); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	var pngBuf []byte
	timeout, timeoutCancel := context.WithTimeout(ctx, 90*time.Second) // Increased to 90s
	defer timeoutCancel()

	// Use data:text/html to avoid about:blank overhead
	htmlContent := htmlBuf.String()

	err := chromedp.Run(timeout,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Navigate directly to content
			return chromedp.Navigate("data:text/html;charset=utf-8," + url.PathEscape(htmlContent)).Do(ctx)
		}),
		chromedp.Sleep(500*time.Millisecond),                                 // Wait for fonts/rendering
		chromedp.Screenshot(".awb-container", &pngBuf, chromedp.NodeVisible), // Auto-crop to receipt
	)

	if err != nil {
		logger.Warn().Err(err).Msg("Screenshot capture failed")

		// Check if Chrome died
		if isChromeDeadError(err) {
			logger.Error().Msg("Chrome process died, attempting auto-recovery")
			atomic.StoreInt32(&isHealthy, 0)

			// Attempt restart
			rendererMu.Lock()
			defer rendererMu.Unlock()

			r.cancel() // Kill old Chrome
			newRenderer, restartErr := newReceiptRenderer()
			if restartErr == nil {
				// Update fields individually to avoid copying the mutex
				r.allocCtx = newRenderer.allocCtx
				r.cancel = newRenderer.cancel
				r.tmpl = newRenderer.tmpl
				// renderMu stays in place (never copy mutexes)
				// renderCount is preserved
				atomic.StoreInt32(&isHealthy, 1)
				logger.Info().Msg("Chrome auto-recovery successful")
			} else {
				logger.Error().Err(restartErr).Msg("Chrome auto-recovery failed")
			}
		}

		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	jpegBuf, err := compressPNGToJPEG(pngBuf)
	if err != nil {
		logger.Warn().Err(err).Msg("JPEG compression failed, using PNG")
		return pngBuf, nil
	}

	logger.Info().
		Int("png_size_kb", len(pngBuf)/1024).
		Int("jpeg_size_kb", len(jpegBuf)/1024).
		Int("savings_pct", (len(pngBuf)-len(jpegBuf))*100/len(pngBuf)).
		Bool("recycled", shouldRecycle).
		Msg("Receipt rendered successfully")

	return jpegBuf, nil
}

func compressPNGToJPEG(pngBytes []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	quality := 85
	if envQuality := os.Getenv("RECEIPT_JPEG_QUALITY"); envQuality != "" {
		fmt.Sscanf(envQuality, "%d", &quality)
	}

	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, rgba, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	return jpegBuf.Bytes(), nil
}

func ShutdownRenderer() error {
	if receiptRenderer != nil {
		receiptRenderer.cancel()
		logger.Info().Msg("Receipt renderer shut down")
	}
	return nil
}

func IsRendererHealthy() bool {
	return atomic.LoadInt32(&isHealthy) == 1
}

func isChromeDeadError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "target closed") ||
		strings.Contains(errStr, "connection closed") ||
		strings.Contains(errStr, "session closed")
}
