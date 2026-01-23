package shipment

import (
	"time"
)

// Calculator handles the business logic for shipment scheduling
type Calculator struct {
	// We can inject a timezone mapper interface here if needed later
}

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
	// Add more as default fallbacks
}

// ResolveTimezone attempts to find a valid timezone for a country name
func ResolveTimezone(country string) string {
	if tz, ok := CountryTimezoneMap[country]; ok {
		return tz
	}
	return "UTC" // Safe fallback
}

func CalculateSchedule(nowUTC time.Time, originCountry, destCountry string) (transitTimeUTC, outForDeliveryUTC, deliveryTimeUTC time.Time) {
	originTZ := ResolveTimezone(originCountry)
	loc, err := time.LoadLocation(originTZ)
	if err != nil {
		loc = time.UTC
	}

	// 1. Convert Now(UTC) to Origin Local Time
	localNow := nowUTC.In(loc)
	hour := localNow.Hour()

	// 2. Logic: Transit Start (Pending -> Intransit)
	var localTransitStart time.Time

	if hour < 8 {
		localTransitStart = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 8, 0, 0, 0, loc)
	} else if hour >= 21 {
		tomorrow := localNow.Add(24 * time.Hour)
		localTransitStart = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, loc)
	} else {
		localTransitStart = localNow.Add(1 * time.Hour)
	}

	// 3. Logic: Delivery Date (Intransit -> Delivered)
	transitDuration := 24 * time.Hour
	localDelivery := localTransitStart.Add(transitDuration)

	// Reset to 10:00 AM on delivery day
	localDelivery = time.Date(localDelivery.Year(), localDelivery.Month(), localDelivery.Day(), 10, 0, 0, 0, loc)

	// Out For Delivery: 8:00 AM on delivery day (2 hours before delivery)
	localOutForDelivery := time.Date(localDelivery.Year(), localDelivery.Month(), localDelivery.Day(), 8, 0, 0, 0, loc)

	return localTransitStart.UTC(), localOutForDelivery.UTC(), localDelivery.UTC()
}
