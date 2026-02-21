package shipment

import (
	"strings"
	"time"
)

// Service defines the interface for shipment calculations
type Service interface {
	CalculateSchedule(nowUTC time.Time, originCountry, destCountry string) (transitTimeUTC, outForDeliveryUTC, deliveryTimeUTC time.Time)
	ResolveTimezone(country string) string
}

// Calculator handles the business logic for shipment scheduling
type Calculator struct {
	// We can inject a timezone mapper interface here if needed later
}

// Ensure Calculator implements Service
var _ Service = (*Calculator)(nil)

// CountryTimezoneMap is a simple static map for MVP.
// In a refined version, this could be a proper database or external lib.
var CountryTimezoneMap = map[string]string{
	"nigeria":        "Africa/Lagos",
	"usa":            "America/New_York",
	"uk":             "Europe/London",
	"united kingdom": "Europe/London",
	"china":          "Asia/Shanghai",
	"dubai":          "Asia/Dubai",
	"uae":            "Asia/Dubai",
	"ghana":          "Africa/Accra",
	"afghanistan":    "Asia/Kabul",
	"honduras":       "America/Tegucigalpa",
	"mexico":         "America/Mexico_City",
}

// ResolveTimezone attempts to find a valid timezone for a country name
func (c *Calculator) ResolveTimezone(country string) string {
	country = strings.ToLower(strings.TrimSpace(country))
	if tz, ok := CountryTimezoneMap[country]; ok {
		return tz
	}
	return "UTC" // Safe fallback
}

func (c *Calculator) CalculateSchedule(nowUTC time.Time, originCountry, destCountry string) (transitTimeUTC, outForDeliveryUTC, deliveryTimeUTC time.Time) {
	// 1. Resolve Admin (Lagos) Time for Operational Rules
	adminLoc, _ := time.LoadLocation("Africa/Lagos")
	adminNow := nowUTC.In(adminLoc)
	hour := adminNow.Hour()

	// 2. Resolve Regional Timezones for localized display
	originTZ := c.ResolveTimezone(originCountry)
	destTZ := c.ResolveTimezone(destCountry)

	originLoc, err := time.LoadLocation(originTZ)
	if err != nil {
		originLoc = time.UTC
	}
	destLoc, err := time.LoadLocation(destTZ)
	if err != nil {
		destLoc = time.UTC
	}

	// 3. Logic: Transit Start (The Operational Gate)
	var localTransitStart time.Time

	if hour < 8 {
		// Before 8 AM Admin TZ: Process at 8 AM Today (in the Sender's local zone)
		senderNow := nowUTC.In(originLoc)
		localTransitStart = time.Date(senderNow.Year(), senderNow.Month(), senderNow.Day(), 8, 0, 0, 0, originLoc)
	} else if hour >= 21 {
		// After 9 PM Admin TZ: Process at 8 AM Tomorrow (in the Sender's local zone)
		senderNow := nowUTC.In(originLoc)
		tomorrow := senderNow.Add(24 * time.Hour)
		localTransitStart = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, originLoc)
	} else {
		// During the day Admin TZ: Start after a 1-hour buffer (The Holding Window)
		senderNow := nowUTC.In(originLoc)
		localTransitStart = senderNow.Add(1 * time.Hour)
	}

	// 4. Logic: Delivery Date (Standard 24-hour window)
	// We use the Receiver's locale for the arrival timestamp
	transitDuration := 24 * time.Hour
	localDelivery := localTransitStart.Add(transitDuration).In(destLoc)

	// Reset to 10:00 AM on delivery day in Receiver's zone
	localDelivery = time.Date(localDelivery.Year(), localDelivery.Month(), localDelivery.Day(), 10, 0, 0, 0, destLoc)

	// Out For Delivery: 8:00 AM on delivery day in Receiver's zone
	localOutForDelivery := time.Date(localDelivery.Year(), localDelivery.Month(), localDelivery.Day(), 8, 0, 0, 0, destLoc)

	return localTransitStart.UTC(), localOutForDelivery.UTC(), localDelivery.UTC()
}

// CalculateSchedule is a package-level helper that uses a default Calculator.
// This allows calling shipment.CalculateSchedule directly.
func CalculateSchedule(nowUTC time.Time, originCountry, destCountry string) (transitTimeUTC, outForDeliveryUTC, deliveryTimeUTC time.Time) {
	calc := &Calculator{}
	return calc.CalculateSchedule(nowUTC, originCountry, destCountry)
}
