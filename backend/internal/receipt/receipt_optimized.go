package receipt

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/logger"
	"webtracker-bot/internal/shipment"

	"image/draw"

	"github.com/fogleman/gg"
)

var (
	templateCache = make(map[i18n.Language]*image.RGBA)
	tmplMu        sync.RWMutex

	bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func RenderReceiptOptimized(s shipment.Shipment, companyName string, lang i18n.Language) ([]byte, error) {
	tmpl := loadStaticTemplate(companyName, lang)
	if tmpl == nil {
		return nil, fmt.Errorf("static template not available")
	}

	dc := ctxPool.Get().(*gg.Context)
	defer ctxPool.Put(dc)

	// BRUTAL OPTIMIZATION 1: Bit-Blitting (raw memory copy)
	// This replaces dc.Clear() and dc.DrawImage(tmpl, 0, 0)
	// It's the fastest possible way to initialize the base layer.
	target := dc.Image().(*image.RGBA)
	copy(target.Pix, tmpl.Pix)

	drawDynamicFields(dc, s, companyName, lang)

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	err := jpeg.Encode(buf, dc.Image(), &jpeg.Options{Quality: 82})
	return buf.Bytes(), err
}

func loadStaticTemplate(companyName string, lang i18n.Language) *image.RGBA {
	tmplMu.RLock()
	if img, ok := templateCache[lang]; ok {
		defer tmplMu.RUnlock()
		return img
	}
	tmplMu.RUnlock()

	tmplMu.Lock()
	defer tmplMu.Unlock()
	if img, ok := templateCache[lang]; ok {
		return img
	}

	// Language-specific template path
	tmplFilename := fmt.Sprintf("receipt_template_%s.png", strings.ToLower(string(lang)))
	tmplPath := filepath.Join(GetAssetsPath(), "img", tmplFilename)
	versionPath := tmplPath + ".version"

	// Version includes company name so changing it invalidates the cached template
	versionKey := TemplateVersion + ":" + strings.ToUpper(companyName)

	needsRefresh := false
	if vData, err := os.ReadFile(versionPath); err != nil || string(vData) != versionKey {
		needsRefresh = true
	}

	if _, err := os.Stat(tmplPath); os.IsNotExist(err) || needsRefresh {
		logger.Info().Str("lang", string(lang)).Str("company", companyName).Msg("Regenerating language-specific template...")
		if err := generateStaticTemplate(tmplPath, companyName, lang); err != nil {
			logger.Error().Err(err).Msg("Failed to generate static template")
			return nil
		}
		_ = os.WriteFile(versionPath, []byte(versionKey), 0644)
	}

	img, err := gg.LoadImage(tmplPath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load static template")
		return nil
	}

	// BRUTAL OPTIMIZATION 2: Ensure template is cached as *image.RGBA
	// This avoids any conversion overhead during the render path.
	var rgba *image.RGBA
	if r, ok := img.(*image.RGBA); ok {
		rgba = r
	} else {
		b := img.Bounds()
		rgba = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
	}

	templateCache[lang] = rgba
	return rgba
}

func generateStaticTemplate(path, companyName string, lang i18n.Language) error {
	dc := gg.NewContext(Width, Height)
	dc.SetHexColor(ColorBgHex)
	dc.Clear()

	drawNoisePro(dc)
	drawGuillochePatterns(dc)
	drawFoldLines(dc)

	if err := LoadFont(dc, FontArialBold, 240); err == nil {
		dc.Push()
		dc.SetRGBA255(139, 0, 0, 50)
		dc.RotateAbout(gg.Radians(-30), Width/2, Height/2)
		dc.DrawStringAnchored("ORIGINAL", Width/2, Height/2, 0.5, 0.5)
		dc.Pop()
	}

	margin := 20.0
	yH := 160.0
	if logoImg := loadCompanyLogo(); logoImg != nil {
		dc.DrawImageAnchored(logoImg, int(margin+10), 210, 0, 0.5)
	}

	dc.SetColor(ColorBurgundy)
	if err := LoadFont(dc, FontArialBold, 80); err == nil {
		text := "AIRWAY BILL"
		if companyName != "" {
			text = strings.ToUpper(companyName)
		}
		dc.DrawStringAnchored(text, Width/2, yH-30, 0.5, 0.5)
	}

	dc.SetColor(ColorBurgundy)
	dc.DrawRectangle(Width/2-400, yH+30, 800, 45)
	dc.Fill()

	if err := LoadFont(dc, FontArialBold, 26); err == nil {
		dc.SetColor(color.White)
		dc.DrawStringAnchored("INTERNATIONAL SPECIAL DELIVERY SERVICE", Width/2, yH+55, 0.5, 0.5)
	}

	drawStaticGridLines(dc)

	gX, gY := 20.0, 420.0
	gW := Width - 40.0
	c1W := gW * 0.30
	c2W := gW * 0.22
	c3W := gW * 0.22
	c4W := gW - c1W - c2W - c3W
	rowH := 104.0

	// GRID HEADERS & SELECTOR STATIC PARTS
	drawStaticSelectorParts(dc, gX+c1W, gY, c2W, rowH*4, i18n.T(lang, "receipt_service"), []string{"EXPRESS", "DIPLOMATIC", "DOMESTIC", "OVERNIGHT"})
	drawStaticSelectorParts(dc, gX+c1W+c2W, gY, c3W, rowH*4, i18n.T(lang, "receipt_payment"), []string{"CASH", "CHEQUE", "ACCOUNT", "BILLED"})

	drawWarningV9(dc, gX+c1W+c2W+c3W, gY+rowH*4, c4W, rowH*2)

	// Ultra-Fast Pre-render: Add visual elements to template
	drawLinearBarcodePro(dc, Width/2, 160.0+165, 520, 70)
	drawQRCodePro(dc, Width-margin-80, 1250+10, 80)
	drawSecurityFoilPro(dc, Width-100, Height-100)

	// Internal drawV11Stamps to match legacy exactly
	dc.Push()
	dc.RotateAbout(gg.Radians(-20), Width/2, Height/2+100)
	dc.SetColor(ColorStampRed)
	dc.SetLineWidth(6)
	if err := LoadFont(dc, FontArialBold, 120); err == nil {
		dc.DrawStringAnchored("DISPATCHED", Width/2+100, Height/2+100, 0.5, 0.5)
	}
	dc.Pop()

	if stampImg := loadApprovedStamp(); stampImg != nil {
		dc.DrawImageAnchored(stampImg, int(Width-margin-50), 260, 1, 0.5)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return dc.SavePNG(path)
}

func drawStaticGridLines(dc *gg.Context) {
	gX, gY := 20.0, 420.0
	gW := Width - 40.0
	c1W := gW * 0.30
	c2W := gW * 0.22
	c3W := gW * 0.22
	c4W := gW - c1W - c2W - c3W
	rowH := 104.0
	nameH := rowH * 2
	selectorH := rowH * 4
	addressH := 200.0
	margin := 20.0

	dc.SetColor(color.Black)
	// Grid components
	// LEFT column (5 rectangles)
	dc.DrawRectangle(gX, gY, c1W, rowH)              // Destination
	dc.DrawRectangle(gX, gY+rowH, c1W, nameH)        // Receiver
	dc.DrawRectangle(gX, gY+rowH+nameH, c1W, rowH)   // Email
	dc.DrawRectangle(gX, gY+rowH*2+nameH, c1W, rowH) // Content
	dc.DrawRectangle(gX, gY+rowH*3+nameH, c1W, rowH) // Weight

	// CENTRAL columns
	dc.DrawRectangle(gX+c1W, gY, c2W, selectorH)          // Service
	dc.DrawRectangle(gX+c1W+c2W, gY, c3W, selectorH)      // Payment
	dc.DrawRectangle(gX+c1W, gY+selectorH, c2W, rowH)     // Dep Date
	dc.DrawRectangle(gX+c1W+c2W, gY+selectorH, c3W, rowH) // Arr Date

	// RIGHT column (Ensuring perfect alignment with a single vertical outer line)
	dc.DrawRectangle(gX+c1W+c2W+c3W, gY, c4W, rowH)            // Origin
	dc.DrawRectangle(gX+c1W+c2W+c3W, gY+rowH, c4W, nameH)      // Sender
	dc.DrawRectangle(gX+c1W+c2W+c3W, gY+rowH+nameH, c4W, rowH) // Gap

	// FOOTER ROW
	dc.DrawRectangle(gX, gY+624.0, c1W+c2W+c3W, addressH)     // Address
	dc.DrawRectangle(gX+c1W+c2W+c3W, gY+624.0, c4W, addressH) // Phone

	dc.Stroke()

	// AUTH LINE
	yA := 1250.0
	dc.SetLineWidth(2.5)
	dc.DrawLine(margin+40, yA+90, margin+640, yA+90)
	dc.Stroke()
}

func drawStaticSelectorParts(dc *gg.Context, x, y, w, h float64, title string, opts []string) {
	// Header bar
	dc.SetRGBA255(20, 20, 20, 255)
	dc.DrawRectangle(x, y, w, 40)
	dc.Fill()

	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(color.White)
		dc.DrawStringAnchored(title, x+w/2, y+25, 0.5, 0.5)
	}

	for i, opt := range opts {
		oy := y + 90 + float64(i*85)
		// Box
		dc.SetColor(color.Black)
		dc.SetLineWidth(1.5)
		dc.DrawRectangle(x+30, oy-18, 36, 36)
		dc.Stroke()

		// Label
		if err := LoadFont(dc, FontArialBold, 17); err == nil {
			dc.SetColor(ColorDarkText)
			dc.DrawString(opt, x+80, oy+12)
		}
	}
}

func drawDynamicFields(dc *gg.Context, s shipment.Shipment, companyName string, lang i18n.Language) {
	margin := 20.0
	yH := 160.0

	dc.SetColor(ColorDarkText)
	if err := LoadFont(dc, FontArialBold, 26); err == nil {
		dc.DrawStringAnchored(s.TrackingID, Width/2, yH+100, 0.5, 0.5)
	}
	if err := LoadFont(dc, FontArialBold, 40); err == nil {
		dc.SetHexColor("#cc0000")
		dc.DrawString(fmt.Sprintf("№ 00%s", s.TrackingID), Width-margin-420, 105)
	}

	gX, gY := 20.0, 420.0
	gW := Width - 40.0
	c1W := gW * 0.30
	c2W := gW * 0.22
	c3W := gW * 0.22
	c4W := gW - c1W - c2W - c3W
	rowH := 104.0
	nameH := rowH * 2

	drawSmartCellV10OnlyText(dc, gX, gY, c1W, rowH, i18n.T(lang, "receipt_destination"), s.Destination)
	drawSmartCellV10OnlyText(dc, gX+c1W+c2W+c3W, gY, c4W, rowH, i18n.T(lang, "receipt_origin"), s.Origin)
	drawSmartCellV10OnlyText(dc, gX, gY+rowH, c1W, nameH, i18n.T(lang, "receipt_receiver"), s.RecipientName)
	drawSmartCellV10OnlyText(dc, gX+c1W+c2W+c3W, gY+rowH, c4W, nameH, i18n.T(lang, "receipt_sender"), s.SenderName)
	drawSmartCellV10OnlyText(dc, gX, gY+rowH+nameH, c1W, rowH, i18n.T(lang, "receipt_email"), s.RecipientEmail)

	cargoType := s.CargoType
	if cargoType == "" {
		cargoType = "Consignment Box"
	}
	drawSmartCellV10OnlyText(dc, gX, gY+rowH*2+nameH, c1W, rowH, i18n.T(lang, "receipt_content"), cargoType)
	drawSmartCellV10OnlyText(dc, gX, gY+rowH*3+nameH, c1W, rowH, i18n.T(lang, "receipt_weight"), fmt.Sprintf("%.2f KGS", s.Weight))

	dateFormat := i18n.GetDateFormat(lang)
	var depStr, arrStr string
	if s.ScheduledTransitTime != nil {
		departure := *s.ScheduledTransitTime
		origLoc, _ := time.LoadLocation(s.SenderTimezone)
		if origLoc == nil {
			origLoc = time.UTC
		}
		depStr = departure.In(origLoc).Format(dateFormat)
	} else {
		depStr = "TBD"
	}

	if s.ExpectedDeliveryTime != nil {
		arrival := *s.ExpectedDeliveryTime
		destLoc, _ := time.LoadLocation(s.RecipientTimezone)
		if destLoc == nil {
			destLoc = time.UTC
		}
		arrStr = arrival.In(destLoc).Format(dateFormat)
	} else {
		arrStr = "TBD"
	}

	drawSmartCellV10OnlyText(dc, gX+c1W, gY+rowH*4, c2W, rowH, i18n.T(lang, "receipt_dep_date"), depStr)
	drawSmartCellV10OnlyText(dc, gX+c1W+c2W, gY+rowH*4, c3W, rowH, i18n.T(lang, "receipt_arr_date"), arrStr)
	drawSmartCellV10OnlyText(dc, gX, gY+624.0, gW-c4W, 200.0, i18n.T(lang, "receipt_address"), s.RecipientAddress)
	drawSmartCellV10OnlyText(dc, gX+c1W+c2W+c3W, gY+624.0, c4W, 200.0, i18n.T(lang, "receipt_phone"), s.RecipientPhone)

	drawSelectorV10OnlyText(dc, gX+c1W, gY, c2W, rowH*4, i18n.T(lang, "receipt_service"), "DIPLOMATIC", []string{"EXPRESS", "DIPLOMATIC", "DOMESTIC", "OVERNIGHT"})
	drawSelectorV10OnlyText(dc, gX+c1W+c2W, gY, c3W, rowH*4, i18n.T(lang, "receipt_payment"), "ACCOUNT", []string{"CASH", "CHEQUE", "ACCOUNT", "BILLED"})

	drawAuthSignature(dc, s.SenderName)
	drawSecurityFooter(dc, s, companyName)
}

func drawSmartCellV10OnlyText(dc *gg.Context, x, y, w, h float64, label, value string) {
	padding := 24.0
	availW, availH := w-(padding*2), h-45.0
	if err := LoadFont(dc, FontArialBold, 15); err == nil {
		dc.SetColor(ColorLabel)
		dc.DrawString(strings.ToUpper(label), x+padding, y+32)
	}
	if value == "" {
		value = "---"
	}
	dc.SetColor(ColorDarkText)

	// Flow text to fill horizontal space better (especially for addresses)
	flowedValue := strings.ReplaceAll(strings.ToUpper(value), "\n", " ")

	// BRUTAL OPTIMIZATION 3: Binary Search for Font Size
	// Instead of linear search (36, 35, 34...), we find the optimal size in ~5 steps.
	low, high := 8.0, 36.0
	optimalSize := 8.0
	var finalLines []string

	for (high - low) > 0.5 {
		mid := (low + high) / 2
		if err := LoadFont(dc, FontArialBold, mid); err != nil {
			high = mid
			continue
		}

		wrapped := dc.WordWrap(flowedValue, availW)
		var testLines []string
		for _, wLine := range wrapped {
			lw, _ := dc.MeasureString(wLine)
			if lw > availW {
				testLines = append(testLines, charWrap(dc, wLine, availW)...)
			} else {
				testLines = append(testLines, wLine)
			}
		}

		totalH := float64(len(testLines)) * mid * 1.15
		if totalH <= availH {
			optimalSize = mid
			finalLines = testLines
			low = mid
		} else {
			high = mid
		}
	}

	// Render the best result found
	if len(finalLines) > 0 {
		_ = LoadFont(dc, FontArialBold, optimalSize)
		totalH := float64(len(finalLines)) * optimalSize * 1.15
		yTarget := y + 45 + (availH-totalH)/2 + optimalSize*0.8
		dc.Push()
		dc.RotateAbout(gg.Radians((rand.Float64()*0.02)-0.01), x+padding, yTarget)
		for i, line := range finalLines {
			dc.DrawString(line, x+padding, yTarget+float64(i)*optimalSize*1.15)
		}
		dc.Pop()
	}
}

func drawSelectorV10OnlyText(dc *gg.Context, x, y, w, h float64, title, active string, opts []string) {
	for i, opt := range opts {
		if strings.EqualFold(active, opt) {
			oy := y + 90 + float64(i*85)
			if err := LoadFont(dc, FontArialBold, 28); err == nil {
				dc.SetColor(ColorDarkText)
				dc.DrawStringAnchored("X", x+48, oy-1, 0.5, 0.5)
			}
		}
	}
}

func drawAuthSignature(dc *gg.Context, name string) {
	yA := 1250.0
	margin := 20.0
	if err := LoadFont(dc, FontSignature, 52); err == nil {
		dc.SetColor(ColorInkBlue)
		parts := strings.Fields(name)
		if len(parts) >= 3 {
			name = string(parts[0][0]) + string(parts[1][0]) + parts[2]
		} else if len(parts) == 2 {
			name = string(parts[0][0]) + parts[1]
		}
		name = strings.ReplaceAll(name, " ", "")
		sigX, sigY := margin+120, yA+70
		dc.Push()
		dc.RotateAbout(gg.Radians(-5), sigX, sigY)
		dc.SetLineWidth(2.5)
		currentX := sigX
		for i, char := range name {
			charStr := string(char)
			w, _ := dc.MeasureString(charStr)
			yOffset := math.Sin(float64(i)*0.8) * 3
			dc.DrawString(charStr, currentX, sigY+yOffset)
			currentX += w * 0.65
		}
		totalW := currentX - sigX
		dc.MoveTo(sigX-15, sigY+18)
		dc.QuadraticTo(sigX+totalW/2, sigY+30, sigX+totalW+25, sigY+12)
		dc.QuadraticTo(sigX+totalW+35, sigY+5, sigX+totalW+40, sigY+8)
		dc.Stroke()
		dc.Pop()
	}

	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(ColorSlate)
		text := "This document serves as an official diplomatic air-freight manifest.\nAll contents have been verified for secure international transit."
		dc.DrawStringAnchored(text, Width-20-120, yA+45, 1.0, 0.5)
	}
}

