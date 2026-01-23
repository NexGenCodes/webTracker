package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/models"

	"github.com/fogleman/gg"
)

const (
	Width   = 1800
	Height  = 1400
	Quality = 100
)

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
	logoCache  image.Image
	stampCache image.Image
	cacheMu    sync.RWMutex
)

func InitReceiptRenderer() error {
	rand.Seed(time.Now().UnixNano())
	return EnsureFontsDownloader()
}

func RenderReceipt(shipment models.Shipment, companyName string, lang i18n.Language) ([]byte, error) {
	dc := ctxPool.Get().(*gg.Context)
	defer ctxPool.Put(dc)

	// Reset context for reuse
	dc.SetHexColor("#f4f2eb")
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

	drawV11Header(dc, shipment, companyName, lang)
	drawV11Grid(dc, shipment, lang)
	drawV11AuthArea(dc, shipment)
	drawSecurityFooter(dc, shipment, companyName)
	drawSecurityFoilPro(dc, Width-100, Height-100)
	drawV11Stamps(dc)

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: 80})
	return buf.Bytes(), err
}

func drawV11Header(dc *gg.Context, shipment models.Shipment, companyName string, lang i18n.Language) {
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

	dc.SetColor(ColorDarkText)
	if err := LoadFont(dc, FontArialBold, 26); err == nil {
		dc.DrawStringAnchored(shipment.TrackingNumber, Width/2, yH+100, 0.5, 0.5)
	}

	if err := LoadFont(dc, FontArialBold, 40); err == nil {
		dc.SetHexColor("#cc0000")
		dc.DrawString(fmt.Sprintf("№ 00%s", shipment.TrackingNumber), Width-margin-420, 105)
	}

	drawLinearBarcodePro(dc, Width/2, yH+165, 520, 70)
	if stampImg := loadApprovedStamp(); stampImg != nil {
		dc.DrawImageAnchored(stampImg, int(Width-margin-50), 260, 1, 0.5)
	}
}

func drawV11Grid(dc *gg.Context, shipment models.Shipment, lang i18n.Language) {
	gX, gY := 20.0, 420.0
	gW := Width - 40.0
	gH := 624.0

	c1W := gW * 0.30
	c2W := gW * 0.22
	c3W := gW * 0.22
	c4W := gW - c1W - c2W - c3W

	rowH := 104.0

	dc.SetColor(color.Black)
	dc.SetLineWidth(2)
	dc.DrawRectangle(gX, gY, gW, gH)
	dc.Stroke()

	dc.DrawLine(gX+c1W, gY, gX+c1W, gY+gH)
	dc.DrawLine(gX+c1W+c2W, gY, gX+c1W+c2W, gY+rowH*4)
	dc.DrawLine(gX+c1W+c2W+c3W, gY, gX+c1W+c2W+c3W, gY+gH)
	dc.Stroke()

	drawSmartCellV10(dc, gX, gY, c1W, rowH, i18n.T(lang, "receipt_destination"), shipment.ReceiverCountry)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY, c4W, rowH, i18n.T(lang, "receipt_origin"), shipment.SenderCountry)

	nameH := rowH * 2
	drawSmartCellV10(dc, gX, gY+rowH, c1W, nameH, i18n.T(lang, "receipt_receiver"), shipment.ReceiverName)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY+rowH, c4W, nameH, i18n.T(lang, "receipt_sender"), shipment.SenderName)

	drawSmartCellV10(dc, gX, gY+rowH+nameH, c1W, rowH, i18n.T(lang, "receipt_email"), shipment.ReceiverEmail)

	drawSmartCellV10(dc, gX, gY+rowH*2+nameH, c1W, rowH, i18n.T(lang, "receipt_content"), "DIPLOMATIC CONSIGNMENT")
	drawSmartCellV10(dc, gX, gY+rowH*3+nameH, c1W, rowH, i18n.T(lang, "receipt_weight"), "15.00 KGS")

	selectorH := rowH * 4
	drawSelectorV10(dc, gX+c1W, gY, c2W, selectorH, i18n.T(lang, "receipt_service"), []string{"EXPRESS", "DIPLOMATIC", "DOMESTIC", "OVERNIGHT"}, "DIPLOMATIC")
	drawSelectorV10(dc, gX+c1W+c2W, gY, c3W, selectorH, i18n.T(lang, "receipt_payment"), []string{"CASH", "CHEQUE", "ACCOUNT", "BILLED"}, "ACCOUNT")

	now := time.Now()
	departure := now
	if now.Hour() >= 11 {
		departure = now.AddDate(0, 0, 1)
	}
	arrival := departure.AddDate(0, 0, 1)

	dateFormat := i18n.GetDateFormat(lang)
	depStr := departure.Format(dateFormat)
	arrStr := arrival.Format(dateFormat)

	drawSmartCellV10(dc, gX+c1W, gY+selectorH, c2W, rowH, i18n.T(lang, "receipt_dep_date"), depStr)
	drawSmartCellV10(dc, gX+c1W+c2W, gY+selectorH, c3W, rowH, i18n.T(lang, "receipt_arr_date"), arrStr)

	drawWarningV9(dc, gX+c1W+c2W+c3W, gY+rowH*3+rowH, c4W, rowH*2)

	addressH := 200.0
	drawSmartCellV10(dc, gX, gY+gH, c1W+c2W+c3W, addressH, i18n.T(lang, "receipt_address"), shipment.ReceiverAddress)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY+gH, c4W, addressH, i18n.T(lang, "receipt_phone"), shipment.ReceiverPhone)
}

func drawV11AuthArea(dc *gg.Context, shipment models.Shipment) {
	yA := 1250.0
	margin := 20.0

	dc.SetLineWidth(2.5)
	dc.SetColor(color.Black)
	dc.DrawLine(margin+40, yA+90, margin+640, yA+90)
	dc.Stroke()

	if err := LoadFont(dc, FontSignature, 52); err == nil {
		dc.SetColor(ColorInkBlue)
		name := shipment.SenderName
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
		dc.SetLineWidth(2.5)
		dc.MoveTo(sigX-15, sigY+18)
		dc.QuadraticTo(sigX+totalW/2, sigY+30, sigX+totalW+25, sigY+12)
		dc.QuadraticTo(sigX+totalW+35, sigY+5, sigX+totalW+40, sigY+8)
		dc.Stroke()
		dc.Pop()
	}

	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(ColorSlate)
		text := "This document serves as an official diplomatic air-freight manifest.\nAll contents have been verified for secure international transit."
		dc.DrawStringAnchored(text, Width-margin-120, yA+45, 1.0, 0.5)
	}

	drawQRCodePro(dc, Width-margin-80, yA+10, 80)
}

func drawSecurityFooter(dc *gg.Context, shipment models.Shipment, companyName string) {
	if err := LoadFont(dc, FontArialBold, 10); err == nil {
		dc.SetRGBA255(20, 20, 20, 100)
		footerText := fmt.Sprintf("SECURE DOCUMENT ID: %s | %s | VERIFIED BY  %s",
			shipment.TrackingNumber, time.Now().Format("2006-01-02"), strings.ToUpper(companyName))
		dc.DrawString(footerText, 60, Height-25)
	}
}

func drawV11Stamps(dc *gg.Context) {
	dc.Push()
	dc.RotateAbout(gg.Radians(-20), Width/2, Height/2+100)
	dc.SetColor(ColorStampRed)
	dc.SetLineWidth(6)
	if err := LoadFont(dc, FontArialBold, 120); err == nil {
		dc.DrawStringAnchored("DISPATCHED", Width/2+100, Height/2+100, 0.5, 0.5)
	}
	dc.Pop()
}

func drawSmartCellV10(dc *gg.Context, x, y, w, h float64, label, value string) {
	dc.SetColor(color.Black)
	dc.SetLineWidth(1.5)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	padding := 24.0
	availW := w - (padding * 2)
	availH := h - 45.0

	if err := LoadFont(dc, FontArialBold, 15); err == nil {
		dc.SetColor(ColorLabel)
		dc.DrawString(strings.ToUpper(label), x+padding, y+32)
	}

	if value == "" {
		value = "---"
	}

	fontSize := 42.0
	dc.SetColor(ColorDarkText)

	for fontSize >= 8 {
		if err := LoadFont(dc, FontArialBold, fontSize); err == nil {
			rawLines := strings.Split(strings.ToUpper(value), "\n")
			var allLines []string
			for _, rl := range rawLines {
				wrapped := dc.WordWrap(rl, availW)
				for _, wLine := range wrapped {
					lw, _ := dc.MeasureString(wLine)
					if lw > availW {
						allLines = append(allLines, charWrap(dc, wLine, availW)...)
					} else {
						allLines = append(allLines, wLine)
					}
				}
			}

			totalH := float64(len(allLines)) * fontSize * 1.15
			if totalH <= availH || fontSize == 8 {
				maxLines := int(availH / (fontSize * 1.15))
				if len(allLines) > maxLines && maxLines > 0 {
					allLines = allLines[:maxLines]
				}

				yTarget := y + 45 + (availH-totalH)/2 + fontSize*0.8
				if yTarget < y+45+fontSize*0.8 {
					yTarget = y + 45 + fontSize*0.8
				}

				dc.Push()
				dc.RotateAbout(gg.Radians((rand.Float64()*0.02)-0.01), x+padding, yTarget)
				for i, line := range allLines {
					dc.DrawString(line, x+padding, yTarget+float64(i)*fontSize*1.15)
				}
				dc.Pop()
				return
			}
		}
		fontSize -= 1
	}
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

func drawSelectorV10(dc *gg.Context, x, y, w, h float64, title string, opts []string, active string) {
	dc.SetColor(color.Black)
	dc.SetLineWidth(1.5)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	dc.SetRGBA255(20, 20, 20, 255)
	dc.DrawRectangle(x, y, w, 40)
	dc.Fill()
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(color.White)
		dc.DrawStringAnchored(title, x+w/2, y+25, 0.5, 0.5)
	}

	for i, opt := range opts {
		oy := y + 90 + float64(i*85)
		dc.SetColor(color.Black)
		dc.DrawRectangle(x+30, oy-18, 36, 36)
		dc.Stroke()
		if strings.EqualFold(active, opt) {
			if err := LoadFont(dc, FontArialBold, 28); err == nil {
				dc.SetColor(ColorDarkText)
				dc.DrawStringAnchored("X", x+48, oy-1, 0.5, 0.5)
			}
		}
		if err := LoadFont(dc, FontArialBold, 17); err == nil {
			dc.SetColor(ColorDarkText)
			dc.DrawString(opt, x+80, oy+12)
		}
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
		bw := 1.0 + float64(rand.Intn(4))
		dc.DrawRectangle(x-w/2+i, y-h/2, bw, h)
		dc.Fill()
	}
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
