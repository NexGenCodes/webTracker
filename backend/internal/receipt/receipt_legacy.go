package receipt

import (
	"bytes"
	"fmt"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	"webtracker-bot/internal/i18n"
	"webtracker-bot/internal/shipment"

	"github.com/fogleman/gg"
)

func RenderReceiptLegacy(s shipment.Shipment, companyName string, lang i18n.Language) ([]byte, error) {
	dc := ctxPool.Get().(*gg.Context)
	defer ctxPool.Put(dc)

	// Reset context for reuse
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

	drawV11Header(dc, s, companyName, lang)
	drawV11Grid(dc, s, lang)
	drawV11AuthArea(dc, s)
	drawSecurityFooter(dc, s, companyName)
	drawV11Stamps(dc)
	drawSecurityFoilPro(dc, Width-100, Height-100)

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: 80})
	return buf.Bytes(), err
}

func drawV11Header(dc *gg.Context, shipment shipment.Shipment, companyName string, lang i18n.Language) {
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
		dc.DrawStringAnchored(shipment.TrackingID, Width/2, yH+100, 0.5, 0.5)
	}

	if err := LoadFont(dc, FontArialBold, 40); err == nil {
		dc.SetHexColor("#cc0000")
		dc.DrawString(fmt.Sprintf("№ 00%s", shipment.TrackingID), Width-margin-420, 105)
	}

	drawLinearBarcodePro(dc, Width/2, yH+165, 520, 70)
	if stampImg := loadApprovedStamp(); stampImg != nil {
		dc.DrawImageAnchored(stampImg, int(Width-margin-50), 260, 1, 0.5)
	}
}

func drawV11Grid(dc *gg.Context, shipment shipment.Shipment, lang i18n.Language) {
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

	drawSmartCellV10(dc, gX, gY, c1W, rowH, i18n.T(lang, "receipt_destination"), shipment.Destination)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY, c4W, rowH, i18n.T(lang, "receipt_origin"), shipment.Origin)

	nameH := rowH * 2
	drawSmartCellV10(dc, gX, gY+rowH, c1W, nameH, i18n.T(lang, "receipt_receiver"), shipment.RecipientName)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY+rowH, c4W, nameH, i18n.T(lang, "receipt_sender"), shipment.SenderName)

	drawSmartCellV10(dc, gX, gY+rowH+nameH, c1W, rowH, i18n.T(lang, "receipt_email"), shipment.RecipientEmail)

	cargoType := shipment.CargoType
	if cargoType == "" {
		cargoType = "Consignment Box"
	}
	drawSmartCellV10(dc, gX, gY+rowH*2+nameH, c1W, rowH, i18n.T(lang, "receipt_content"), cargoType)
	drawSmartCellV10(dc, gX, gY+rowH*3+nameH, c1W, rowH, i18n.T(lang, "receipt_weight"), fmt.Sprintf("%.2f KGS", shipment.Weight))

	selectorH := rowH * 4
	drawSelectorV10(dc, gX+c1W, gY, c2W, selectorH, i18n.T(lang, "receipt_service"), []string{"EXPRESS", "DIPLOMATIC", "DOMESTIC", "OVERNIGHT"}, "DIPLOMATIC")
	drawSelectorV10(dc, gX+c1W+c2W, gY, c3W, selectorH, i18n.T(lang, "receipt_payment"), []string{"CASH", "CHEQUE", "ACCOUNT", "BILLED"}, "ACCOUNT")

	// Use stored timestamps from database
	var depStr, arrStr string
	dateFormat := i18n.GetDateFormat(lang)

	if shipment.ScheduledTransitTime != nil {
		departure := *shipment.ScheduledTransitTime
		origLoc, _ := time.LoadLocation(shipment.SenderTimezone)
		if origLoc == nil {
			origLoc = time.UTC
		}
		depStr = departure.In(origLoc).Format(dateFormat)
	} else {
		depStr = "TBD"
	}

	if shipment.ExpectedDeliveryTime != nil {
		arrival := *shipment.ExpectedDeliveryTime
		destLoc, _ := time.LoadLocation(shipment.RecipientTimezone)
		if destLoc == nil {
			destLoc = time.UTC
		}
		arrStr = arrival.In(destLoc).Format(dateFormat)
	} else {
		arrStr = "TBD"
	}

	drawSmartCellV10(dc, gX+c1W, gY+selectorH, c2W, rowH, i18n.T(lang, "receipt_dep_date"), depStr)
	drawSmartCellV10(dc, gX+c1W+c2W, gY+selectorH, c3W, rowH, i18n.T(lang, "receipt_arr_date"), arrStr)

	drawWarningV9(dc, gX+c1W+c2W+c3W, gY+rowH*3+rowH, c4W, rowH*2)

	addressH := 200.0
	// Use RecipientAddress explicitly
	addrVal := shipment.RecipientAddress
	drawSmartCellV10(dc, gX, gY+gH, c1W+c2W+c3W, addressH, i18n.T(lang, "receipt_address"), addrVal)
	drawSmartCellV10(dc, gX+c1W+c2W+c3W, gY+gH, c4W, addressH, i18n.T(lang, "receipt_phone"), shipment.RecipientPhone)
}

func drawV11AuthArea(dc *gg.Context, shipment shipment.Shipment) {
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

func drawSecurityFooter(dc *gg.Context, shipment shipment.Shipment, companyName string) {
	if err := LoadFont(dc, FontArialBold, 10); err == nil {
		dc.SetRGBA255(20, 20, 20, 100)
		footerText := fmt.Sprintf("SECURE DOCUMENT ID: %s | %s | VERIFIED BY  %s",
			shipment.TrackingID, time.Now().Format("2006-01-02"), strings.ToUpper(companyName))
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

	fontSize := 36.0
	dc.SetColor(ColorDarkText)

	// Flow text to fill horizontal space better (especially for addresses)
	flowedValue := strings.ReplaceAll(strings.ToUpper(value), "\n", " ")

	for fontSize >= 8 {
		if err := LoadFont(dc, FontArialBold, fontSize); err == nil {
			wrapped := dc.WordWrap(flowedValue, availW)
			var allLines []string
			for _, wLine := range wrapped {
				lw, _ := dc.MeasureString(wLine)
				if lw > availW {
					allLines = append(allLines, charWrap(dc, wLine, availW)...)
				} else {
					allLines = append(allLines, wLine)
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

