package receipt

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand/v2"
	"path/filepath"
	"sync"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/shipment"

	"github.com/fogleman/gg"
)

const (
	Width      = 1800
	Height     = 1400
	Quality    = 100
	ColorBgHex = "#f4f2eb"
)

const TemplateVersion = "v27"

var ctxPool = sync.Pool{
	New: func() interface{} {
		return gg.NewContext(Width, Height)
	},
}

var (
	ColorBurgundy = color.RGBA{139, 0, 0, 255}
	ColorDarkText = color.RGBA{45, 45, 45, 235}
	ColorLabel    = color.RGBA{70, 70, 70, 220}
	ColorSlate    = color.RGBA{60, 60, 60, 255}
	ColorInkBlue  = color.RGBA{0, 0, 139, 230}
	ColorStampRed = color.RGBA{139, 0, 0, 160}
)

var (
	logoCache    image.Image
	stampCache   image.Image
	cacheMu      sync.RWMutex
	useOptimized bool
)

func InitReceiptRenderer(opt bool) error {
	useOptimized = opt
	return EnsureFontsDownloader()
}

func RenderReceipt(s shipment.Shipment, companyName string, lang i18n.Language) ([]byte, error) {
	if useOptimized {
		return RenderReceiptOptimized(s, companyName, lang)
	}
	return RenderReceiptLegacy(s, companyName, lang)
}

func loadCompanyLogo() image.Image {
	cacheMu.RLock()
	if logoCache != nil {
		defer cacheMu.RUnlock()
		return logoCache
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()
	if logoCache != nil {
		return logoCache
	}

	logoPath := filepath.Join(GetAssetsPath(), "img", "logo.png")
	img, err := gg.LoadImage(logoPath)
	if err != nil {
		return nil
	}
	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()
	m := dx
	if dy > dx {
		m = dy
	}
	scale := 350.0 / float64(m)
	newWidth, newHeight := int(float64(dx)*scale), int(float64(dy)*scale)
	dc := gg.NewContext(newWidth, newHeight)
	dc.Scale(scale, scale)
	dc.DrawImage(img, 0, 0)
	logoCache = dc.Image()
	return logoCache
}

func loadApprovedStamp() image.Image {
	cacheMu.RLock()
	if stampCache != nil {
		defer cacheMu.RUnlock()
		return stampCache
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()
	if stampCache != nil {
		return stampCache
	}

	stampPath := filepath.Join(GetAssetsPath(), "img", "approved_stamp.png")
	img, err := gg.LoadImage(stampPath)
	if err != nil {
		return nil
	}
	bounds := img.Bounds()
	scale := 260.0 / float64(bounds.Dx())
	newWidth, newHeight := int(float64(bounds.Dx())*scale), int(float64(bounds.Dy())*scale)
	dc := gg.NewContext(newWidth, newHeight)
	dc.Scale(scale, scale)
	dc.DrawImage(img, 0, 0)
	stampCache = dc.Image()
	return stampCache
}

func charWrap(dc *gg.Context, text string, width float64) []string {
	var lines []string
	var current string
	for _, r := range text {
		w, _ := dc.MeasureString(current + string(r))
		if w > width && current != "" {
			lines = append(lines, current)
			current = string(r)
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func drawNoisePro(dc *gg.Context) {
	dc.Push()
	dc.SetRGBA(0, 0, 0, 0.02)
	for i := 0; i < 15000; i++ { // Reduced count for performance
		dc.DrawPoint(rand.Float64()*Width, rand.Float64()*Height, 1)
	}
	dc.Fill() // Batch fill instead of individual strokes
	dc.Pop()
}

func drawGuillochePatterns(dc *gg.Context) {
	dc.Push()
	dc.SetRGBA255(139, 0, 0, 8)
	dc.SetLineWidth(0.4)
	for y := 450.0; y < 1050.0; y += 90 { // Reduced frequency for performance
		dc.MoveTo(40, y)
		for x := 40.0; x < float64(Width)-40; x += 30 { // Increased step for performance
			dy := math.Sin(x/60) * 12
			dc.LineTo(x, y+dy)
		}
		dc.Stroke()
	}
	dc.Pop()
}

func drawFoldLines(dc *gg.Context) {
	dc.Push()
	dc.SetRGBA255(0, 0, 0, 15)
	dc.SetLineWidth(1)
	y1 := float64(Height) / 3.0
	dc.DrawLine(0, y1, float64(Width), y1)
	dc.Stroke()
	y2 := 2.0 * float64(Height) / 3.0
	dc.DrawLine(0, y2, float64(Width), y2)
	dc.Stroke()
	dc.Pop()
}

func drawSecurityFoilPro(dc *gg.Context, x, y float64) {
	dc.Push()
	dc.Translate(x, y)
	dc.Rotate(gg.Radians(15))

	dc.SetRGBA255(212, 175, 55, 220)
	dc.DrawRegularPolygon(12, 0, 0, 46, 0)
	dc.Fill()

	dc.SetRGBA255(130, 100, 30, 255)
	dc.SetLineWidth(1.5)
	dc.DrawRegularPolygon(12, 0, 0, 46, 0)
	dc.Stroke()

	dc.SetRGBA255(255, 255, 255, 140)
	dc.SetLineWidth(1)
	for i := 0; i < 8; i++ {
		dc.Rotate(gg.Radians(45))
		dc.DrawLine(-35, 0, 35, 0)
		dc.Stroke()
	}

	if err := LoadFont(dc, FontArialBold, 11); err == nil {
		dc.SetColor(color.Black)
		dc.DrawStringAnchored("SAFETY", 0, -3, 0.5, 0.5)
		dc.DrawStringAnchored("GUARANTEED", 0, 12, 0.5, 0.5)
	}

	dc.SetRGBA255(0, 0, 0, 40)
	dc.DrawCircle(0, 4, 34)
	dc.Stroke()

	dc.Pop()
}

func drawLinearBarcodePro(dc *gg.Context, x, y, w, h float64) {
	dc.SetColor(ColorDarkText)
	for i := 0.0; i < w; i += 6 {
		bw := 1.0 + float64(rand.IntN(4))
		dc.DrawRectangle(x-w/2+i, y-h/2, bw, h)
		dc.Fill()
	}
}

func drawQRCodePro(dc *gg.Context, x, y, size float64) {
	dc.Push()
	dc.SetColor(color.White)
	dc.DrawRectangle(x, y, size, size)
	dc.Fill()

	dc.SetColor(color.Black)
	dc.SetLineWidth(1)
	dc.DrawRectangle(x, y, size, size)
	dc.Stroke()

	dotSize := size / 10
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if (i < 3 && j < 3) || (i > 6 && j < 3) || (i < 3 && j > 6) {
				if i != 1 || j != 1 {
					dc.DrawRectangle(x+float64(i)*dotSize, y+float64(j)*dotSize, dotSize, dotSize)
					dc.Fill()
				}
				continue
			}
			if rand.Float64() > 0.5 {
				dc.DrawRectangle(x+float64(i)*dotSize, y+float64(j)*dotSize, dotSize, dotSize)
				dc.Fill()
			}
		}
	}
	dc.Pop()
}

func drawWarningV9(dc *gg.Context, x, y, w, h float64) {
	dc.SetHexColor("#f4f2eb")
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()
	dc.SetColor(color.Black)
	dc.SetLineWidth(2)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()
	if err := LoadFont(dc, FontArialBold, 22); err == nil {
		dc.SetColor(ColorBurgundy)
		dc.DrawStringAnchored("! CONFIDENTIAL !", x+w/2, y+50, 0.5, 0.5)
	}
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(color.Black)
		dc.DrawStringAnchored("• UNAUTHORIZED OPENING IS A FEDERAL OFFENSE", x+w/2, y+95, 0.5, 0.5)
		dc.DrawStringAnchored("• DIPLOMATIC SECURE TRANSIT", x+w/2, y+130, 0.5, 0.5)
		dc.DrawStringAnchored("• ANTI-TAMPER SEAL PROTECTED", x+w/2, y+165, 0.5, 0.5)
	}
}
