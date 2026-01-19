package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // PNG decoder
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"webtracker-bot/internal/config"
	"webtracker-bot/internal/models"

	"github.com/fogleman/gg"
)

// =============================================================================
// MANUAL EDIT SECTION: DIMENSIONS & QUALITY
// =============================================================================
const (
	Width   = 1800 // Canvas width in pixels
	Height  = 1400 // Canvas height in pixels (Increased for more vertical space)
	Quality = 95   // JPEG output quality (1-100)
)

// =============================================================================
// MANUAL EDIT SECTION: COLOR PALETTE
// =============================================================================
var (
	ColorBurgundy = color.RGBA{139, 0, 0, 255}  // Unified base red
	ColorDarkText = color.RGBA{45, 45, 45, 235} // More "faded" professional look
	ColorLabel    = color.RGBA{70, 70, 70, 220} // Softened label color
	ColorSlate    = color.RGBA{60, 60, 60, 255}
	ColorInkBlue  = color.RGBA{0, 0, 139, 230}
	ColorStampRed = color.RGBA{139, 0, 0, 160} // Unified red with transparency
)

// InitReceiptRenderer ensures fonts are ready
func InitReceiptRenderer() error {
	return EnsureFontsDownloader()
}

// RenderReceipt generates the official receipt image
func RenderReceipt(shipment models.Shipment) ([]byte, error) {
	companyName := config.Load().CompanyName
	dc := gg.NewContext(Width, Height)

	// 1. Background (Set page color)
	dc.SetHexColor("#f4f2eb")
	dc.Clear()
	drawNoisePro(dc)          // Adds paper texture noise
	drawGuillochePatterns(dc) // NEW: High-security background pattern
	drawFoldLines(dc)         // NEW: Subtle paper fold simulation

	// 2. Watermark Visibility
	if err := LoadFont(dc, FontArialBold, 240); err == nil {
		dc.Push()
		dc.SetRGBA255(139, 0, 0, 50) // Change '16' to adjust watermark opacity
		dc.RotateAbout(gg.Radians(-30), Width/2, Height/2)
		dc.DrawStringAnchored("ORIGINAL", Width/2, Height/2, 0.5, 0.5)
		dc.Pop()
	}

	// 3. Document Border Removed

	// 4. Header Elements
	drawV11Header(dc, shipment, companyName)

	// 5. Main Data Grid
	drawV11Grid(dc, shipment)

	// 6. Validation/Signature Area & Security Footer
	drawV11AuthArea(dc, shipment)

	// 7. Micro Security Details
	drawSecurityFooter(dc, shipment, companyName)
	drawSecurityFoilPro(dc, Width-100, Height-100) // NEW: Security foil seal

	// 8. Background Stamps
	drawV11Stamps(dc)

	// Final Output Processing
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: Quality})
	return buf.Bytes(), err
}

// -----------------------------------------------------------------------------
// HEADER DRAWING LOGIC (Logo, Title, Approved Stamp)
// -----------------------------------------------------------------------------
func drawV11Header(dc *gg.Context, shipment models.Shipment, companyName string) {
	margin := 20.0
	yH := 150.0

	// Draw Company Logo (Left)
	if logoImg := loadCompanyLogo(); logoImg != nil {
		// Align vertically with the stamp (moved up as requested)
		dc.DrawImage(logoImg, int(margin+10), int(yH-120))
	} else {
		// Fallback to vector logo if image fails
		drawCargoPlaneLogoPro(dc, margin+20, yH-60)
	}

	// Main Title
	dc.SetColor(ColorBurgundy)
	if err := LoadFont(dc, FontArialBold, 72); err == nil {
		text := "AIRWAY BILL"
		if companyName != "" {
			text = strings.ToUpper(companyName)
		}
		dc.DrawStringAnchored(text, Width/2, yH, 0.5, 0.5)
	}

	// Service Subtitle with Professional Banner
	dc.SetColor(ColorBurgundy)
	dc.DrawRectangle(Width/2-400, yH+30, 800, 45) // Professional banner bar
	dc.Fill()

	if err := LoadFont(dc, FontArialBold, 26); err == nil {
		dc.SetColor(color.White) // Contrast text on banner
		dc.DrawStringAnchored("INTERNATIONAL SPECIAL DELIVERY SERVICE", Width/2, yH+55, 0.5, 0.5)
	}

	dc.SetColor(ColorDarkText)
	if err := LoadFont(dc, FontArialBold, 26); err == nil {
		dc.DrawStringAnchored(shipment.TrackingNumber, Width/2, yH+100, 0.5, 0.5)
	}

	// Serial Number (Top Rightish)
	if err := LoadFont(dc, FontArialBold, 40); err == nil {
		dc.SetHexColor("#cc0000")
		dc.DrawString(fmt.Sprintf("№ 00%s", shipment.TrackingNumber), Width-margin-420, 105)
	}

	// Barcode (Center)
	drawLinearBarcodePro(dc, Width/2, yH+165, 520, 70)

	// Approved Stamp in Header Position
	// drawApprovedStampV10(dc, Width-margin-150, 240)
	if stampImg := loadApprovedStamp(); stampImg != nil {
		dc.DrawImage(stampImg, int(Width-margin-120-100), 160) // Adjusted x,y for stamp image
	} else {
		drawApprovedStampV10(dc, Width-margin-150, 240) // Fallback
	}
}

// -----------------------------------------------------------------------------
// GRID DRAWING LOGIC (Cells, Spacing, Columns)
// -----------------------------------------------------------------------------
func drawV11Grid(dc *gg.Context, shipment models.Shipment) {
	gX, gY := 20.0, 420.0 // Expanded grid width, removed large margins
	gW := Width - 40.0    // Total grid width
	gH := 624.0           // Total grid body height (Increased to fit taller rows)

	c1W := gW * 0.30            // Column 1 width ratio
	c2W := gW * 0.22            // Column 2
	c3W := gW * 0.22            // Column 3
	c4W := gW - c1W - c2W - c3W // Column 4 (Auto)

	rowH := 104.0 // Standard row height

	// Draw Grid Outer Rectangle
	dc.SetColor(color.Black)
	dc.SetLineWidth(2)
	dc.DrawRectangle(gX, gY, gW, gH)
	dc.Stroke()

	// Draw Vertical Dividers
	dc.DrawLine(gX+c1W, gY, gX+c1W, gY+gH)
	dc.DrawLine(gX+c1W+c2W, gY, gX+c1W+c2W, gY+rowH*4) // Dividers for middle columns
	dc.DrawLine(gX+c1W+c2W+c3W, gY, gX+c1W+c2W+c3W, gY+gH)
	dc.Stroke()

	// --- MIRRORED DATA GRID (Logical Pairs) ---
	drawSmartCellV10(dc, gX, gY, c1W, rowH, "DESTINATION", shipment.ReceiverCountry)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY, c4W, rowH, "ORIGIN", shipment.SenderCountry)

	// Doubled row height for Names
	nameH := rowH * 2
	drawSmartCellV10(dc, gX, gY+rowH, c1W, nameH, "RECEIVER", shipment.ReceiverName)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY+rowH, c4W, nameH, "SENDER", shipment.SenderName)

	drawSmartCellV10(dc, gX, gY+rowH+nameH, c1W, rowH, "EMAIL", shipment.ReceiverEmail)

	drawSmartCellV10(dc, gX, gY+rowH*2+nameH, c1W, rowH, "CONTENT", "DIPLOMATIC CONSIGNMENT")
	drawSmartCellV10(dc, gX, gY+rowH*3+nameH, c1W, rowH, "WEIGHT", "15.00 KGS")

	// --- SELECTORS: Middle Columns (Top 4 Rows) ---
	selectorH := rowH * 4
	drawSelectorV10(dc, gX+c1W, gY, c2W, selectorH, "SERVICE MODE", []string{"EXPRESS", "DIPLOMATIC", "DOMESTIC", "OVERNIGHT"}, "DIPLOMATIC")
	drawSelectorV10(dc, gX+c1W+c2W, gY, c3W, selectorH, "PAYMENT METHOD", []string{"CASH", "CHEQUE", "ACCOUNT", "BILLED"}, "ACCOUNT")

	// --- DATES: Logic (11:00 AM Rule) ---
	now := time.Now()
	departure := now
	if now.Hour() >= 11 {
		departure = now.AddDate(0, 0, 1)
	}
	arrival := departure.AddDate(0, 0, 1)

	depStr := departure.Format("02/01/2006")
	arrStr := arrival.Format("02/01/2006")

	drawSmartCellV10(dc, gX+c1W, gY+selectorH, c2W, rowH, "DEPARTURE DATE", depStr)
	drawSmartCellV10(dc, gX+c1W+c2W, gY+selectorH, c3W, rowH, "ARRIVAL DATE", arrStr)

	// Security Warning Box (Col 4, Bottom Span)
	drawWarningV9(dc, gX+c1W+c2W+c3W, gY+rowH*3+rowH, c4W, rowH*2)

	// --- ADDRESS & PHONE (Side by Side) ---
	addressH := 200.0
	drawSmartCellV10(dc, gX, gY+gH, c1W+c2W+c3W, addressH, "DELIVERY ADDRESS", shipment.ReceiverAddress)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY+gH, c4W, addressH, "CONTACT PHONE", shipment.ReceiverPhone)
}

// -----------------------------------------------------------------------------
// SIGNATURE & AUTHENTICATION AREA
// -----------------------------------------------------------------------------
func drawV11AuthArea(dc *gg.Context, shipment models.Shipment) {
	yA := 1250.0
	margin := 20.0

	// Signature Line
	dc.SetLineWidth(2.5)
	dc.SetColor(color.Black)
	dc.DrawLine(margin+40, yA+90, margin+640, yA+90)
	dc.Stroke()

	// Handwritten Signature (Illegible Cursive Style)
	if err := LoadFont(dc, FontSignature, 52); err == nil {
		dc.SetColor(ColorInkBlue)
		name := shipment.SenderName
		parts := strings.Fields(name)

		// Create illegible signature by removing spaces and condensing
		if len(parts) >= 3 {
			name = string(parts[0][0]) + string(parts[1][0]) + parts[2]
		} else if len(parts) == 2 {
			name = string(parts[0][0]) + parts[1]
		}
		// Remove all spaces for cursive flow
		name = strings.ReplaceAll(name, " ", "")

		sigX, sigY := margin+120, yA+70
		dc.Push()
		dc.RotateAbout(gg.Radians(-5), sigX, sigY) // More aggressive slant

		// Draw with character overlap for illegibility
		dc.SetLineWidth(2.5)
		currentX := sigX
		for i, char := range name {
			charStr := string(char)
			w, _ := dc.MeasureString(charStr)

			// Vary vertical position slightly for natural handwriting
			yOffset := math.Sin(float64(i)*0.8) * 3
			dc.DrawString(charStr, currentX, sigY+yOffset)

			// Overlap characters more (reduce spacing)
			currentX += w * 0.65
		}

		// Signature Flourish (More dramatic underline)
		totalW := currentX - sigX
		dc.SetLineWidth(2.5)
		dc.MoveTo(sigX-15, sigY+18)
		dc.QuadraticTo(sigX+totalW/2, sigY+30, sigX+totalW+25, sigY+12)
		dc.QuadraticTo(sigX+totalW+35, sigY+5, sigX+totalW+40, sigY+8)
		dc.Stroke()
		dc.Pop()
	}

	// Agent Verification Note
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(ColorSlate)
		text := "This document serves as an official diplomatic air-freight manifest.\nAll contents have been verified for secure international transit."
		dc.DrawStringAnchored(text, Width-margin-120, yA+45, 1.0, 0.5)
	}

	// QR Code Placeholder
	drawQRCodePro(dc, Width-margin-80, yA+10, 80)
}

// drawSecurityFooter adds a microscopic security string at the very bottom
func drawSecurityFooter(dc *gg.Context, shipment models.Shipment, companyName string) {
	if err := LoadFont(dc, FontArialBold, 10); err == nil {
		dc.SetRGBA255(20, 20, 20, 100)
		footerText := fmt.Sprintf("SECURE DOCUMENT ID: %s | %s | VERIFIED BY  %s",
			shipment.TrackingNumber, time.Now().Format("2006-01-02"), strings.ToUpper(companyName))
		dc.DrawString(footerText, 60, Height-25)
	}
}

// -----------------------------------------------------------------------------
// PROCESS STAMPS (Dispatched stamp over the document)
// -----------------------------------------------------------------------------
func drawV11Stamps(dc *gg.Context) {
	dc.Push()
	dc.RotateAbout(gg.Radians(-15), Width/2, Height/2+100)
	dc.SetColor(ColorStampRed)
	dc.SetLineWidth(6)
	if err := LoadFont(dc, FontArialBold, 120); err == nil {
		dc.DrawStringAnchored("DISPATCHED", Width/2+100, Height/2+100, 0.5, 0.5)
	}
	dc.Pop()
}

// =============================================================================
// HELPER DRAWING FUNCTIONS (Logo, Cells, Selectors)
// =============================================================================

// drawSmartCellV10: Draws a box with label/value and scales text to fit
func drawSmartCellV10(dc *gg.Context, x, y, w, h float64, label, value string) {
	dc.SetColor(color.Black)
	dc.SetLineWidth(1.5)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	padding := 24.0
	availW := w - (padding * 2)
	availH := h - 45.0

	// Draw Label
	if err := LoadFont(dc, FontArialBold, 15); err == nil {
		dc.SetColor(ColorLabel)
		dc.DrawString(strings.ToUpper(label), x+padding, y+32)
	}

	if value == "" {
		value = "---"
	}

	// Font Scaling Algorithm (Spill-Proof)
	fontSize := 42.0
	dc.SetColor(ColorDarkText)

	for fontSize >= 8 {
		if err := LoadFont(dc, FontArialBold, fontSize); err == nil {
			rawLines := strings.Split(strings.ToUpper(value), "\n")
			var allLines []string
			for _, rl := range rawLines {
				wrapped := dc.WordWrap(rl, availW)
				// Handle long strings without spaces (like emails)
				for _, wLine := range wrapped {
					lw, _ := dc.MeasureString(wLine)
					if lw > availW {
						// Fallback to character-level wrap for this long word
						allLines = append(allLines, charWrap(dc, wLine, availW)...)
					} else {
						allLines = append(allLines, wLine)
					}
				}
			}

			totalH := float64(len(allLines)) * fontSize * 1.15
			if totalH <= availH || fontSize == 8 {
				// Limit lines to prevent spill
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

// charWrap forces a wrap at the character level if a word is too long
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

// drawSelectorV10: Draws a group of checkboxes with one selected
func drawSelectorV10(dc *gg.Context, x, y, w, h float64, title string, opts []string, active string) {
	dc.SetColor(color.Black)
	dc.SetLineWidth(1.5)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	// Option Header Box
	dc.SetRGBA255(20, 20, 20, 255)
	dc.DrawRectangle(x, y, w, 40)
	dc.Fill()
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetColor(color.White)
		dc.DrawStringAnchored(title, x+w/2, y+25, 0.5, 0.5)
	}

	// Draw Options
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

// drawApprovedStampV10: Draws a high-security polished approval mark
func drawApprovedStampV10(dc *gg.Context, x, y float64) {
	dc.Push()

	// Slight random jitter for authenticity
	dc.RotateAbout(gg.Radians((rand.Float64()*0.06)-0.03), x, y)

	dc.SetColor(ColorStampRed)

	// Inner Circle
	dc.SetLineWidth(4)
	dc.DrawCircle(x, y, 90)
	dc.Stroke()

	// Outer Dashed Circle
	dc.SetLineWidth(2)
	dc.SetDash(8, 4)
	dc.DrawCircle(x, y, 100)
	dc.Stroke()
	dc.SetDash() // Reset dash

	// Arched Text - TOP (Simulated)
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.DrawStringAnchored("CERTIFIED", x, y-65, 0.5, 0.5)
	}

	// Main Text - CENTER
	if err := LoadFont(dc, FontArialBold, 28); err == nil {
		dc.DrawStringAnchored("APPROVED", x, y, 0.5, 0.5)
	}

	// Arched Text - BOTTOM (Simulated)
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.DrawStringAnchored("FOR TRANSIT", x, y+65, 0.5, 0.5)
	}

	// Micro Security Stars REMOVED

	dc.Pop()
}

// drawQRCodePro: Simulates a professional QR code
func drawQRCodePro(dc *gg.Context, x, y, size float64) {
	dc.Push()
	dc.SetColor(color.White)
	dc.DrawRectangle(x, y, size, size)
	dc.Fill()

	dc.SetColor(color.Black)
	dc.SetLineWidth(1)
	dc.DrawRectangle(x, y, size, size)
	dc.Stroke()

	// Draw QR patterns (Randomly generated dots/blocks)
	dotSize := size / 10
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			// Always draw corner squares (finding patterns)
			if (i < 3 && j < 3) || (i > 6 && j < 3) || (i < 3 && j > 6) {
				if i != 1 || j != 1 { // Hollow center for corner patterns
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

// drawCargoPlaneLogoPro: Draws a high-end elite professional cargo plane vector logo
func drawCargoPlaneLogoPro(dc *gg.Context, x, y float64) {
	dc.Push()
	dc.Translate(x, y)
	dc.Scale(1.3, 1.3)

	// Colors
	ColorPrimary := ColorBurgundy
	ColorShadow := color.RGBA{100, 0, 0, 255}
	ColorHighlight := color.RGBA{200, 50, 50, 255}
	ColorGlass := color.RGBA{220, 230, 255, 255}

	// 1. Double Orbital Rings (Global Reach)
	dc.SetLineWidth(1.5)
	dc.SetRGBA255(139, 0, 0, 60)
	dc.DrawEllipse(75, 40, 100, 30) // Horizontal orbit
	dc.Stroke()
	dc.DrawEllipse(75, 40, 40, 90) // Vertical orbit
	dc.Stroke()

	// 2. Fuselage Main Body (Lower Half Shaded)
	dc.SetColor(ColorPrimary)
	dc.MoveTo(10, 35)
	dc.LineTo(110, 35)
	dc.QuadraticTo(135, 35, 140, 25) // Nose top
	dc.LineTo(140, 35)
	dc.QuadraticTo(135, 50, 110, 50) // Nose bottom
	dc.LineTo(10, 50)
	dc.ClosePath()
	dc.Fill()

	// Fuselage Top Highlight
	dc.SetColor(ColorHighlight)
	dc.DrawRectangle(10, 35, 100, 3)
	dc.Fill()

	// Fuselage Bottom Shadow
	dc.SetColor(ColorShadow)
	dc.DrawRectangle(10, 47, 100, 3)
	dc.Fill()

	// 3. Panoramic Glass Cockpit
	dc.SetColor(ColorGlass)
	dc.MoveTo(122, 36)
	dc.LineTo(134, 36)
	dc.LineTo(131, 28)
	dc.ClosePath()
	dc.Fill()

	// 4. Swept-Back Professional Wings (Dual-Tone)
	dc.SetColor(ColorPrimary)
	dc.MoveTo(55, 45)
	dc.LineTo(20, 100)
	dc.LineTo(50, 100)
	dc.LineTo(85, 45)
	dc.ClosePath()
	dc.Fill()

	dc.SetColor(ColorShadow)
	dc.DrawLine(55, 45, 20, 100)
	dc.Stroke()

	// 5. Elite Vertical Stabilizer (Tail)
	dc.SetColor(ColorPrimary)
	dc.MoveTo(20, 35)
	dc.LineTo(5, 5)
	dc.LineTo(30, 5) // Rear tilt
	dc.LineTo(45, 35)
	dc.ClosePath()
	dc.Fill()

	dc.SetColor(ColorHighlight)
	dc.DrawLine(5, 5, 20, 35)
	dc.Stroke()

	// 6. Professional Engine Cowlings (Technical Detail)
	// Engine 1
	dc.SetColor(color.Black)
	dc.DrawRoundedRectangle(35, 95, 20, 12, 3)
	dc.Fill()
	dc.SetColor(ColorSlate)
	dc.DrawCircle(52, 101, 3) // Fan intake highlight
	dc.Fill()

	// Engine 2
	dc.SetColor(color.Black)
	dc.DrawRoundedRectangle(60, 95, 16, 10, 2)
	dc.Fill()

	// 7. Cargo Side Door Outline (Subtle)
	dc.SetRGBA255(255, 255, 255, 40)
	dc.DrawRectangle(40, 42, 15, 6)
	dc.Stroke()

	dc.Pop()
}

// drawFoldLines: Simulates physical paper folds
func drawFoldLines(dc *gg.Context) {
	dc.Push()
	dc.SetRGBA255(0, 0, 0, 15) // Extremely faint fold
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
	logoPath := filepath.Join(GetAssetsPath(), "img", "logo.png")
	if img, err := gg.LoadImage(logoPath); err == nil {
		// Scale image to fit logo area (350x350)
		bounds := img.Bounds()
		width, height := bounds.Dx(), bounds.Dy()
		scale := 350.0 / float64(max(width, height))
		newWidth, newHeight := int(float64(width)*scale), int(float64(height)*scale)

		dc := gg.NewContext(newWidth, newHeight)
		dc.Scale(scale, scale)
		dc.DrawImage(img, 0, 0)
		return dc.Image()
	}
	return nil
}

func loadApprovedStamp() image.Image {
	stampPath := filepath.Join(GetAssetsPath(), "img", "approved_stamp.png")
	if img, err := gg.LoadImage(stampPath); err == nil {
		// Scale stamp to approx 200px width
		bounds := img.Bounds()
		width := bounds.Dx()
		scale := 200.0 / float64(width)
		newWidth, newHeight := int(float64(width)*scale), int(float64(img.Bounds().Dy())*scale)

		dc := gg.NewContext(newWidth, newHeight)
		dc.Scale(scale, scale)
		dc.DrawImage(img, 0, 0)
		return dc.Image()
	}
	return nil
}

// drawGuillochePatterns: Adds a professional high-security wavy background
func drawGuillochePatterns(dc *gg.Context) {
	dc.Push()
	dc.SetRGBA255(139, 0, 0, 8) // Extremely subtle security line
	dc.SetLineWidth(0.4)
	for y := 450.0; y < 1050.0; y += 45 {
		dc.MoveTo(40, y)
		for x := 40.0; x < float64(Width)-40; x += 15 {
			dy := math.Sin(x/60) * 12
			dc.LineTo(x, y+dy)
		}
		dc.Stroke()
	}
	dc.Pop()
}

// drawSecurityFoilPro: Adds a high-contrast holographic metallic seal
func drawSecurityFoilPro(dc *gg.Context, x, y float64) {
	dc.Push()
	dc.Translate(x, y)
	dc.Rotate(gg.Radians(15))

	// 1. Foil Outer (Solid Gold with darkened edge)
	dc.SetRGBA255(212, 175, 55, 220) // High-opacity gold
	dc.DrawRegularPolygon(12, 0, 0, 46, 0)
	dc.Fill()

	dc.SetRGBA255(130, 100, 30, 255) // Dark bronze border
	dc.SetLineWidth(1.5)
	dc.DrawRegularPolygon(12, 0, 0, 46, 0)
	dc.Stroke()

	// 2. Inner Shimmer Pattern (High contrast white)
	dc.SetRGBA255(255, 255, 255, 140)
	dc.SetLineWidth(1)
	for i := 0; i < 8; i++ {
		dc.Rotate(gg.Radians(45))
		dc.DrawLine(-35, 0, 35, 0)
		dc.Stroke()
	}

	// 3. High-Contrast Text
	if err := LoadFont(dc, FontArialBold, 11); err == nil {
		dc.SetColor(color.Black) // Pure black for maximum legibility
		dc.DrawStringAnchored("SAFETY", 0, -3, 0.5, 0.5)
		dc.DrawStringAnchored("GUARANTEED", 0, 12, 0.5, 0.5)
	}

	// Micro-security ring around text
	dc.SetRGBA255(0, 0, 0, 40)
	dc.DrawCircle(0, 4, 34)
	dc.Stroke()

	dc.Pop()
}

// drawLinearBarcodePro: Simulates a linear barcode
func drawLinearBarcodePro(dc *gg.Context, x, y, w, h float64) {
	dc.SetColor(ColorDarkText)
	for i := 0.0; i < w; i += 6 {
		bw := 1.0 + float64(rand.Intn(4))
		dc.DrawRectangle(x-w/2+i, y-h/2, bw, h)
		dc.Fill()
	}
}

// drawNoisePro: Adds randomized grain to paper for realism
func drawNoisePro(dc *gg.Context) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 40000; i++ {
		dc.SetRGBA(0, 0, 0, 0.02)
		dc.DrawPoint(rand.Float64()*Width, rand.Float64()*Height, 1)
		dc.Stroke()
	}
}

// drawWarningV9: Draws the security warning box
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
