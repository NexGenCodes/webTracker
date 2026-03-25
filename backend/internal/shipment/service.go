package shipment

import (
	"math/rand"
	"strings"
	"time"
)

// Service defines the interface for shipment utilities.
// Scheduling logic is fully handled here in pure Go, ensuring
// maintainable and typed code without relying on implicit DB triggers.
type Service interface {
	ResolveTimezone(country string) string
	CalculateDeparture(now time.Time, adminTZ string) time.Time
	CalculateArrival(departure time.Time, senderCountry, receiverCountry string) (arrival, outfordelivery time.Time)
}

// Calculator handles timezone resolution and timeline calculations.
type Calculator struct{}

// Ensure Calculator implements Service
var _ Service = (*Calculator)(nil)

// CountryTimezoneMap maps country names to IANA timezone identifiers.
var CountryTimezoneMap = map[string]string{
	"nigeria":        "Africa/Lagos",
	"usa":            "America/New_York",
	"united states":  "America/New_York",
	"uk":             "Europe/London",
	"united kingdom": "Europe/London",
	"germany":        "Europe/Berlin",
	"spain":          "Europe/Madrid",
	"mexico":         "America/Mexico_City",
	"argentina":      "America/Argentina/Buenos_Aires",
	"colombia":       "America/Bogota",
	"chile":          "America/Santiago",
	"peru":           "America/Lima",
	"venezuela":      "America/Caracas",
	"china":          "Asia/Shanghai",
	"dubai":          "Asia/Dubai",
	"uae":            "Asia/Dubai",
	"ghana":          "Africa/Accra",
	"afghanistan":    "Asia/Kabul",
	"honduras":       "America/Tegucigalpa",
	"guatemala":      "America/Guatemala",
	"ecuador":        "America/Guayaquil",
	"bolivia":        "America/La_Paz",
	"paraguay":       "America/Asuncion",
	"uruguay":        "America/Montevideo",
	"panama":         "America/Panama",
	"benin":          "Africa/Porto-Novo",
	"togo":           "Africa/Lome",
	"pakistan":       "Asia/Karachi",
	"turkey":         "Europe/Istanbul",
}

// TransitDurationMap defines baseline international transit durations.
var TransitDurationMap = map[string]int{
	"pakistan":       20,
	"uae":            22,
	"turkey":         24,
	"united kingdom": 28,
	"uk":             28,
	"germany":        28,
	"nigeria":        30,
	"ghana":          32,
	"china":          26,
	"usa":            34,
	"united states":  34,
	"colombia":       36,
}

// ResolveTimezone attempts to find a valid timezone for a country name with fuzzy matching
func (c *Calculator) ResolveTimezone(country string) string {
	country = strings.ToLower(strings.TrimSpace(country))
	if country == "" {
		return "UTC"
	}

	// 1. Exact match for speed
	if tz, ok := CountryTimezoneMap[country]; ok {
		return tz
	}

	// 2. Fuzzy match (starts with or contains)
	for name, tz := range CountryTimezoneMap {
		if strings.Contains(country, name) || strings.Contains(name, country) {
			return tz
		}
	}

	return "UTC" // Safe fallback
}

// CalculateDeparture (Algorithm A) determines when the package officially goes "In Transit".
// Respects Nigerian Warehouse Hours (8:00 AM - 10:00 PM).
func (c *Calculator) CalculateDeparture(now time.Time, adminTZ string) time.Time {
	if adminTZ == "" {
		adminTZ = "Africa/Lagos"
	}
	loc, err := time.LoadLocation(adminTZ)
	if err != nil {
		loc = time.FixedZone("WAT", 3600) // Nigeria fallback
	}

	// Transit starts exactly 1 hour after creation
	transit := now.In(loc).Add(1 * time.Hour)

	// Boundary Check
	if transit.Hour() < 8 {
		// Before 8am: Push to 8:00 AM today
		transit = time.Date(transit.Year(), transit.Month(), transit.Day(), 8, 0, 0, 0, loc)
	} else if transit.Hour() >= 22 {
		// After 10pm: Push to 8:00 AM tomorrow
		tomorrow := transit.Add(24 * time.Hour)
		transit = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, loc)
	}

	return transit.UTC()
}

var CountryRegionMap = map[string]string{
	"nigeria":        "africa",
	"ghana":          "africa",
	"benin":          "africa",
	"togo":           "africa",
	"usa":            "north_america",
	"united states":  "north_america",
	"canada":         "north_america",
	"uk":             "europe",
	"united kingdom": "europe",
	"germany":        "europe",
	"spain":          "europe",
	"turkey":         "europe",
	"uae":            "middle_east",
	"dubai":          "middle_east",
	"pakistan":       "middle_east",
	"china":          "asia",
	"india":          "asia",
	"colombia":       "south_america",
	"argentina":      "south_america",
	"chile":          "south_america",
	"peru":           "south_america",
	"venezuela":      "south_america",
	"mexico":         "north_america",
}

// CalculateArrival (Algorithm B) determines the final delivery window relative to Departure.
func (c *Calculator) CalculateArrival(departure time.Time, senderCountry, receiverCountry string) (time.Time, time.Time) {
	sender := strings.ToLower(strings.TrimSpace(senderCountry))
	dest := strings.ToLower(strings.TrimSpace(receiverCountry))
	
	// Base baseline
	baseDuration := 26
	if d, ok := TransitDurationMap[dest]; ok {
		baseDuration = d
	}

	// Proximity Multiplier
	multiplier := 1.0
	senderRegion := CountryRegionMap[sender]
	destRegion := CountryRegionMap[dest]

	if senderRegion != "" && destRegion != "" {
		if senderRegion == destRegion {
			multiplier = 0.6 // Same region (e.g. West Africa)
		} else if (senderRegion == "europe" && destRegion == "middle_east") || (senderRegion == "middle_east" && destRegion == "europe") {
			multiplier = 0.8 // Close regions
		} else if (senderRegion == "africa" && (destRegion == "north_america" || destRegion == "asia")) {
			multiplier = 1.3 // Long haul
		}
	}

	finalDuration := float64(baseDuration) * multiplier
	jitter := rand.Intn(4)
	arrival := departure.Add(time.Duration(int(finalDuration)+jitter) * time.Hour)

	recipientTZ := c.ResolveTimezone(receiverCountry)
	loc, err := time.LoadLocation(recipientTZ)
	if err != nil {
		loc = time.UTC
	}

	localArrival := arrival.In(loc)
	if localArrival.Hour() < 9 {
		arrival = time.Date(localArrival.Year(), localArrival.Month(), localArrival.Day(), 9, 0, 0, 0, loc).UTC()
	} else if localArrival.Hour() >= 16 {
		arrival = time.Date(localArrival.Year(), localArrival.Month(), localArrival.Day(), 16, 0, 0, 0, loc).UTC()
	}

	outfordelivery := arrival.Add(-time.Duration(3+rand.Intn(3)) * time.Hour)
	return arrival, outfordelivery
}
