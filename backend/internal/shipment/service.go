package shipment

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

var (
	locationCache = make(map[string]*time.Location)
	locMutex      sync.RWMutex
)

func loadLocation(name string) (*time.Location, error) {
	locMutex.RLock()
	loc, ok := locationCache[name]
	locMutex.RUnlock()
	if ok {
		return loc, nil
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}

	locMutex.Lock()
	locationCache[name] = loc
	locMutex.Unlock()
	return loc, nil
}

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
	loc, err := loadLocation(adminTZ)
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

// CalculateArrival determines the final delivery window.
// Rule: Arrival is always the day following the Departure Date.
func (c *Calculator) CalculateArrival(departure time.Time, senderCountry, receiverCountry string) (time.Time, time.Time) {
	// Resolve destination timezone
	recipientTZ := c.ResolveTimezone(receiverCountry)
	loc, err := loadLocation(recipientTZ)
	if err != nil {
		loc = time.UTC
	}

	// 1. Set baseline arrival to the day after departure
	arrivalDate := departure.In(loc).Add(24 * time.Hour)

	// 2. Set arrival time to a random window between 9:00 AM and 4:00 PM local time
	hour := 9 + rand.Intn(7) // 9, 10, 11, 12, 13, 14, 15
	minute := rand.Intn(60)
	arrival := time.Date(arrivalDate.Year(), arrivalDate.Month(), arrivalDate.Day(), hour, minute, 0, 0, loc).UTC()

	// 3. Out for delivery is always 3-6 hours before final arrival
	outfordelivery := arrival.Add(-time.Duration(3+rand.Intn(3)) * time.Hour)

	return arrival, outfordelivery
}
