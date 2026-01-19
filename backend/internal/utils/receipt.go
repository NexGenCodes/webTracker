package utils

import (
	"bytes"
	"image/color"
	"image/jpeg"
	"math/rand"
	"strings"

	"webtracker-bot/internal/models"

	"github.com/fogleman/gg"
)

// Receipt Dimensions (Pro proportions based on reference)
const (
	Width  = 1200
	Height = 1600
)

// InitReceiptRenderer ensures fonts are ready
func InitReceiptRenderer() error {
	return EnsureFontsDownloader()
}

// RenderReceipt generates a 100% exact professional receipt image natively
func RenderReceipt(shipment models.Shipment, companyName string) ([]byte, error) {
	dc := gg.NewContext(Width, Height)

	// 1. Background (Tan Paper Texture)
	dc.SetHexColor("#e6e0d0")
	dc.Clear()
	drawNoise(dc)
	drawPaperWear(dc)

	// 2. Large Background Watermark
	if err := LoadFont(dc, FontArialBold, 240); err == nil {
		dc.Push()
		dc.SetRGBA255(122, 42, 28, 12)
		dc.RotateAbout(gg.Radians(-30), Width/2, Height/2)
		dc.DrawStringAnchored("CERTIFIED", Width/2, Height/2, 0.5, 0.5)
		dc.Pop()
	}

	// 3. Header
	drawProfessionalHeader(dc, shipment, companyName)

	// 4. Main Data Grid (Complex 3-Column Style)
	drawProfessionalGrid(dc, shipment)

	// 5. Footer (Fingerprint, Warnings, UN Sign)
	drawProfessionalFooter(dc, shipment)

	// 6. DISPATCHED Stamp Override (Over Grid)
	dc.Push()
	dc.RotateAbout(gg.Radians(-10), Width/2, Height/2)
	dc.SetColor(color.RGBA{160, 30, 30, 160})
	dc.SetLineWidth(8)
	dc.DrawRoundedRectangle(Width/2-280, Height/2-70, 560, 140, 20)
	dc.Stroke()
	if err := LoadFont(dc, FontArialBold, 90); err == nil {
		dc.DrawStringAnchored("DISPATCHED", Width/2, Height/2, 0.5, 0.5)
	}
	dc.Pop()

	// 7. Output
	var buf bytes.Buffer
	quality := 90
	err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: quality})
	return buf.Bytes(), err
}

func drawProfessionalHeader(dc *gg.Context, shipment models.Shipment, companyName string) {
	// Top Left Plane/Globe Graphic
	drawPlaneLogo(dc, 60, 60)

	// Center Titles
	dc.SetHexColor("#7a2a1c")
	if err := LoadFont(dc, FontArialBold, 72); err == nil {
		text := "AIRWAY BILL"
		if companyName != "" {
			text = strings.ToUpper(companyName)
		}
		dc.DrawStringAnchored(text, Width/2, 100, 0.5, 0.5)
	}
	dc.SetColor(color.Black)
	if err := LoadFont(dc, FontArialBold, 24); err == nil {
		dc.Push()
		dc.SetFillStyle(gg.NewSolidPattern(color.Black))
		dc.DrawStringAnchored("International Special Delivery", Width/2, 150, 0.5, 0.5)
		dc.Pop()
	}

	// Ref No below title
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		refText := "No : 00486/LRG/VIP/00233" // Mock serial format from image
		dc.DrawStringAnchored(refText, Width/2-100, 195, 0.5, 0.5)
	}

	// Top Right UN Seal
	drawStandardUNSeal(dc, Width-180, 100, 70)

	// Barcode in middle of header
	drawLinearBarcode(dc, Width/2, 230, 400, 80)
	if err := LoadFont(dc, FontArialBold, 12); err == nil {
		dc.DrawStringAnchored(shipment.TrackingNumber, Width/2, 320, 0.5, 0.5)
	}

	// Far Right Origin Box
	drawOriginBoxRefined(dc, Width-250, 205, shipment.SenderCountry)
}

func drawProfessionalGrid(dc *gg.Context, shipment models.Shipment) {
	gridX := 40.0
	gridY := 350.0
	gridW := Width - 80.0
	col1W := 320.0
	col2W := 280.0
	col3W := gridW - col1W - col2W

	dc.SetColor(color.Black)
	dc.SetLineWidth(2)

	// Column 1: Destination, Sender, Receiver
	drawGridCell(dc, gridX, gridY, col1W, 120, "1  DESTINATION COUNTRY", shipment.ReceiverCountry, true)
	drawGridCell(dc, gridX, gridY+120, col1W, 120, "SENDER'S NAME", shipment.SenderName, true)
	drawGridCell(dc, gridX, gridY+240, col1W, 120, "RECEIVER'S NAME", shipment.ReceiverName, true)
	drawGridCell(dc, gridX, gridY+360, col1W, 120, "RECEIVER'S EMAIL", shipment.ReceiverEmail, true)
	drawGridCell(dc, gridX, gridY+480, col1W, 120, "CONTENT OF ITEM", "Consignment box", true)

	// Column 2: Service Type List
	drawServiceTypeBox(dc, gridX+col1W, gridY, col2W, 600)

	// Column 3: Dates, Weight, Customs
	drawGridCell(dc, gridX+col1W+col2W, gridY, col3W, 120, "DEPARTURE DATE", shipment.CreatedAt.Format("7/1/2006"), true)
	drawGridCell(dc, gridX+col1W+col2W, gridY+120, col3W, 120, "VOLUME TRIC CHARGED WEIGHT:", "15kgs", true)

	// Complex Customs/Charges area
	drawCustomsArea(dc, gridX+col1W+col2W, gridY+240, col3W, 360)

	// Bottom Spanning Rows
	drawGridCell(dc, gridX, gridY+600, col3W+col2W, 180, "DELIVERING ADDRESS", shipment.ReceiverAddress, true)

	// Sub-grid for Phone & Arrival Date (Bottom Center)
	midX := gridX + col1W + col2W
	drawGridCell(dc, midX-col2W, gridY+780, col2W, 120, "RECEIVER'S PHONE NUMBER", shipment.ReceiverPhone, true)
	drawGridCell(dc, midX-col2W, gridY+900, col2W, 120, "DATE OF ARRIVAL", "12/1/2026", true)

	// Warning Box (Yellow Tint)
	drawWarningBox(dc, gridX+col1W+col2W, gridY+780, col3W, 240)
}

func drawProfessionalFooter(dc *gg.Context, shipment models.Shipment) {
	footerY := 1320.0

	// Fingerprint (Bottom Left)
	drawFingerprint(dc, 150, footerY+150)
	if err := LoadFont(dc, FontArialBold, 14); err == nil {
		dc.SetColor(color.Black)
		dc.DrawStringAnchored("SENDER AUTHORIZED", 150, footerY+50, 0.5, 0.5)
	}

	// UN SIGN Seal (Bottom Right)
	drawComplexUNSignSeal(dc, Width-220, footerY+120)
}

// Sub-Graphics Helpers

func drawPlaneLogo(dc *gg.Context, x, y float64) {
	dc.Push()
	dc.Translate(x, y)

	// Blue background circle Gradient
	grad := gg.NewRadialGradient(100, 100, 20, 100, 100, 100)
	grad.AddColorStop(0, color.RGBA{100, 150, 255, 100})
	grad.AddColorStop(1, color.Transparent)
	dc.SetFillStyle(grad)
	dc.DrawCircle(100, 100, 100)
	dc.Fill()

	// Plane Silhouette (approximate)
	dc.SetColor(color.Black)
	dc.SetLineWidth(2)
	dc.DrawCircle(100, 100, 70) // Earth orbit
	dc.Stroke()

	// Simple Plane vector
	dc.Translate(100, 100)
	dc.Rotate(gg.Radians(-30))
	dc.MoveTo(-40, 0)
	dc.LineTo(40, 0)
	dc.LineTo(30, -10)
	dc.LineTo(10, -5)
	dc.LineTo(0, -30)
	dc.LineTo(-10, -5)
	dc.LineTo(-30, -10)
	dc.ClosePath()
	dc.Fill()

	dc.Pop()
	if err := LoadFont(dc, FontArialBold, 12); err == nil {
		dc.SetHexColor("#7a2a1c")
		dc.DrawStringAnchored("GLOBAL LOGISTICS", x+100, y+210, 0.5, 0.5)
	}
}

func drawGridCell(dc *gg.Context, x, y, w, h float64, label, value string, border bool) {
	if border {
		dc.SetColor(color.Black)
		dc.SetLineWidth(1.5)
		dc.DrawRectangle(x, y, w, h)
		dc.Stroke()
	}

	// Label (Gray/Small)
	if err := LoadFont(dc, FontArialBold, 14); err == nil {
		dc.SetHexColor("#555")
		dc.DrawString(label, x+15, y+30)
	}

	// Value (Bold/Center)
	if err := LoadFont(dc, FontArialBold, 28); err == nil {
		dc.SetColor(color.Black)
		dc.DrawStringAnchored(strings.ToUpper(value), x+w/2, y+h*0.65, 0.5, 0.5)
	}
}

func drawServiceTypeBox(dc *gg.Context, x, y, w, h float64) {
	dc.SetColor(color.Black)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.SetHexColor("#FFFFFF")
		dc.DrawRectangle(x, y, w, 40)
		dc.SetHexColor("#666")
		dc.Fill()
		dc.SetColor(color.White)
		dc.DrawStringAnchored("SERVICE TYPE", x+w/2, y+25, 0.5, 0.5)
	}

	options := []string{"WORLD WIDE EXPRESS", "DIPLOMATIC DELIVERY", "DOMESTIC EXPRESS", "SPECIAL SERVICE", "WORLD OVERNIGHT EXPRESS", "REGULAR"}
	dc.SetColor(color.Black)
	for i, opt := range options {
		oy := y + 80 + float64(i*80)
		dc.DrawRectangle(x+20, oy-15, 30, 30) // Checkbox
		dc.Stroke()

		if opt == "DIPLOMATIC DELIVERY" {
			// Draw '*' in box
			if err := LoadFont(dc, FontArialBold, 24); err == nil {
				dc.DrawStringAnchored("*", x+35, oy+5, 0.5, 0.5)
			}
		}

		if err := LoadFont(dc, FontArialBold, 14); err == nil {
			dc.DrawString(opt, x+65, oy+10)
		}
	}
}

func drawCustomsArea(dc *gg.Context, x, y, w, h float64) {
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	// Labels
	if err := LoadFont(dc, FontArialBold, 12); err == nil {
		dc.DrawString("CODE:", x+10, y+25)
		dc.DrawString("SERVICES", x+w-80, y+25)
		dc.DrawString("CHARGES", x+w-80, y+45)
		dc.DrawStringAnchored("(16258/Bco/TG011)", x+w/2, y+70, 0.5, 0.5)
	}

	// CUSTOMS DUTY Blue Box
	dc.SetRGBA255(200, 230, 255, 100)
	dc.DrawRectangle(x+5, y+100, w-10, 40)
	dc.Fill()
	dc.SetColor(color.Black)
	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.DrawStringAnchored("CUSTOMS DUTY", x+w/2, y+125, 0.5, 0.5)
	}

	legalText := "Non Inspection Clearance Charge\nshould be paid from the receiver\nto customs/immigrations."
	if err := LoadFont(dc, FontArialBold, 10); err == nil {
		dc.DrawStringAnchored(legalText, x+w/2, y+180, 0.5, 0.5)
	}

	note := "NOTE: RULES AND REGULATIONS\nVARIES FROM ONE JURISDICTION\nTO ANOTHER AND NOT LIMITED TO\nTHE CONCLUSION TERMS OF THIS"
	if err := LoadFont(dc, FontArialBold, 12); err == nil {
		dc.DrawStringAnchored(note, x+w/2, y+280, 0.5, 0.5)
	}
}

func drawWarningBox(dc *gg.Context, x, y, w, h float64) {
	dc.SetRGBA255(255, 240, 150, 150) // Yellow tint
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()
	dc.SetColor(color.Black)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	if err := LoadFont(dc, FontArialBold, 22); err == nil {
		dc.SetHexColor("#7a2a1c")
		dc.DrawStringAnchored("WARNING !!", x+w/2, y+50, 0.5, 0.5)
	}
	if err := LoadFont(dc, FontArialBold, 14); err == nil {
		dc.SetColor(color.Black)
		dc.DrawStringAnchored("CONFIDENTIAL,\nTHE PACKAGE\nCAN BE OPEN BY\nRECEIVER ONLY", x+w/2, y+130, 0.5, 0.5)
	}
}

func drawFingerprint(dc *gg.Context, x, y float64) {
	dc.Push()
	dc.SetRGBA255(100, 50, 150, 120) // Purple thumb ink
	for i := 0; i < 15; i++ {
		r := 40.0 - float64(i*2) + (rand.Float64() * 5)
		dc.DrawEllipse(x, y, r, r*1.3)
		dc.SetLineWidth(2)
		dc.SetDash(rand.Float64()*10, rand.Float64()*5)
		dc.Stroke()
	}
	dc.Pop()
}

func drawComplexUNSignSeal(dc *gg.Context, x, y float64) {
	dc.Push()
	dc.Translate(x, y)

	dc.SetColor(color.Black)
	dc.DrawCircle(0, 0, 100)
	dc.SetLineWidth(2)
	dc.Stroke()
	dc.DrawCircle(0, 0, 92)
	dc.Stroke()

	if err := LoadFont(dc, FontArialBold, 12); err == nil {
		dc.DrawStringAnchored("UNITED NATIONS", 0, -40, 0.5, 0.5)
		dc.DrawStringAnchored("AGENT REPRESENTATIVE", 0, -20, 0.5, 0.5)
		dc.DrawStringAnchored("SIGN", 0, 10, 0.5, 0.5)
	}

	// Globe/Branches
	dc.Scale(0.8, 0.8)
	dc.DrawCircle(0, 0, 15)
	dc.Stroke()
	dc.MoveTo(-50, 20)
	dc.LineTo(50, 20)
	dc.Stroke()
	dc.Pop()
}

func drawOriginBoxRefined(dc *gg.Context, x, y float64, country string) {
	dc.SetColor(color.Black)
	dc.DrawRectangle(x, y, 220, 80)
	dc.Stroke()
	dc.DrawLine(x+100, y, x+100, y+80)
	dc.Stroke()

	if err := LoadFont(dc, FontArialBold, 16); err == nil {
		dc.DrawStringAnchored("ORIGIN", x+50, y+45, 0.5, 0.5)
	}
	if err := LoadFont(dc, FontArialBold, 22); err == nil {
		dc.DrawStringAnchored(strings.ToUpper(country), x+160, y+45, 0.5, 0.5)
	}
}

func drawLinearBarcode(dc *gg.Context, x, y, w, h float64) {
	for i := 0.0; i < w; i += 5 {
		bw := 1.0 + float64(rand.Intn(4))
		dc.DrawRectangle(x-w/2+i, y-h/2, bw, h)
		dc.Fill()
	}
}

// Global drawing helpers for noise/wear (same as before)
func drawNoise(dc *gg.Context) {
	for i := 0; i < 20000; i++ {
		x, y := rand.Float64()*Width, rand.Float64()*Height
		dc.SetRGBA(0, 0, 0, rand.Float64()*0.02)
		dc.DrawPoint(x, y, 1)
		dc.Stroke()
	}
}

func drawPaperWear(dc *gg.Context) {
	dc.SetRGBA(0, 0, 0, 0.05)
	dc.DrawLine(Width/2, 0, Width/2, Height)
	dc.Stroke()
	dc.DrawLine(0, Height*0.25, Width, Height*0.25)
	dc.Stroke()
	dc.DrawLine(0, Height*0.75, Width, Height*0.75)
	dc.Stroke()
}

func drawStandardUNSeal(dc *gg.Context, x, y, r float64) {
	dc.Push()
	dc.Translate(x, y)
	dc.SetColor(color.Black)
	dc.DrawCircle(0, 0, r)
	dc.SetLineWidth(2)
	dc.Stroke()
	dc.DrawCircle(0, 0, r-8)
	dc.Stroke()
	dc.Scale(0.6, 0.6)
	dc.DrawCircle(0, 0, 20)
	dc.Stroke()
	dc.MoveTo(-40, 0)
	dc.QuadraticTo(0, -40, 40, 0)
	dc.QuadraticTo(0, 40, -40, 0)
	dc.Stroke()
	dc.Pop()
}
