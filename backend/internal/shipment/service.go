package shipment

import (
	"math/rand/v2"
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
	CalculateDeparture(now time.Time, originCountry string) time.Time
	CalculateArrival(departure time.Time, destinationCountry string) (arrival, outfordelivery time.Time)
}

// Calculator handles timezone resolution and timeline calculations.
type Calculator struct{}

// CountryTimezoneMap maps country names to IANA timezone identifiers.
var CountryTimezoneMap = map[string]string{
	// Africa
	"nigeria":       "Africa/Lagos",
	"ghana":         "Africa/Accra",
	"benin":         "Africa/Porto-Novo",
	"togo":          "Africa/Lome",
	"south africa":  "Africa/Johannesburg",
	"kenya":         "Africa/Nairobi",
	"egypt":         "Africa/Cairo",
	"ethiopia":      "Africa/Addis_Ababa",
	"cameroon":      "Africa/Douala",
	"senegal":       "Africa/Dakar",
	"tanzania":      "Africa/Dar_es_Salaam",
	"morocco":       "Africa/Casablanca",
	"ivory coast":   "Africa/Abidjan",
	"cote d'ivoire": "Africa/Abidjan",

	// Americas
	"usa":                "America/New_York",
	"united states":      "America/New_York",
	"canada":             "America/Toronto",
	"mexico":             "America/Mexico_City",
	"brazil":             "America/Sao_Paulo",
	"argentina":          "America/Argentina/Buenos_Aires",
	"colombia":           "America/Bogota",
	"chile":              "America/Santiago",
	"peru":               "America/Lima",
	"venezuela":          "America/Caracas",
	"honduras":           "America/Tegucigalpa",
	"guatemala":          "America/Guatemala",
	"ecuador":            "America/Guayaquil",
	"bolivia":            "America/La_Paz",
	"paraguay":           "America/Asuncion",
	"uruguay":            "America/Montevideo",
	"panama":             "America/Panama",
	"costa rica":         "America/Costa_Rica",
	"dominican republic": "America/Santo_Domingo",
	"jamaica":            "America/Jamaica",

	// Europe
	"uk":             "Europe/London",
	"united kingdom": "Europe/London",
	"germany":        "Europe/Berlin",
	"france":         "Europe/Paris",
	"spain":          "Europe/Madrid",
	"italy":          "Europe/Rome",
	"netherlands":    "Europe/Amsterdam",
	"belgium":        "Europe/Brussels",
	"portugal":       "Europe/Lisbon",
	"switzerland":    "Europe/Zurich",
	"sweden":         "Europe/Stockholm",
	"norway":         "Europe/Oslo",
	"poland":         "Europe/Warsaw",
	"turkey":         "Europe/Istanbul",
	"russia":         "Europe/Moscow",
	"ukraine":        "Europe/Kiev",
	"ireland":        "Europe/Dublin",
	"greece":         "Europe/Athens",

	// Asia & Middle East
	"china":        "Asia/Shanghai",
	"india":        "Asia/Kolkata",
	"japan":        "Asia/Tokyo",
	"south korea":  "Asia/Seoul",
	"indonesia":    "Asia/Jakarta",
	"malaysia":     "Asia/Kuala_Lumpur",
	"singapore":    "Asia/Singapore",
	"thailand":     "Asia/Bangkok",
	"vietnam":      "Asia/Ho_Chi_Minh",
	"philippines":  "Asia/Manila",
	"pakistan":     "Asia/Karachi",
	"bangladesh":   "Asia/Dhaka",
	"dubai":        "Asia/Dubai",
	"uae":          "Asia/Dubai",
	"saudi arabia": "Asia/Riyadh",
	"qatar":        "Asia/Qatar",
	"afghanistan":  "Asia/Kabul",
	"israel":       "Asia/Jerusalem",
	"iraq":         "Asia/Baghdad",
	"iran":         "Asia/Tehran",

	// Oceania
	"australia":   "Australia/Sydney",
	"new zealand": "Pacific/Auckland",
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
// Resolves the timezone from the origin country — no hardcoded admin timezone needed.
// Rule: departure = now + 1 hour. If that falls between 11 PM and 8 AM, snap to 8 AM.
func (c *Calculator) CalculateDeparture(now time.Time, originCountry string) time.Time {
	tz := c.ResolveTimezone(originCountry)
	loc, err := loadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	// Transit starts exactly 1 hour after creation in origin local time
	transit := now.In(loc).Add(1 * time.Hour)

	// Cap: if departure falls between 11 PM and 8 AM, snap to next 8 AM
	if transit.Hour() >= 23 {
		// After 11 PM: Push to 8:00 AM next day
		next := transit.AddDate(0, 0, 1)
		transit = time.Date(next.Year(), next.Month(), next.Day(), 8, 0, 0, 0, loc)
	} else if transit.Hour() < 8 {
		// Before 8 AM: Push to 8:00 AM same day
		transit = time.Date(transit.Year(), transit.Month(), transit.Day(), 8, 0, 0, 0, loc)
	}

	return transit.UTC()
}

// CalculateArrival (Algorithm B) determines the delivery window.
// Rule: Arrival is always the next day after departure in the destination's timezone.
// Sunday skip: if arrival lands on Sunday, push to Monday.
// Out for Delivery: 7:30 AM - 9:00 AM local. Expected Delivery: 9:00 AM - 5:00 PM local.
func (c *Calculator) CalculateArrival(departure time.Time, destinationCountry string) (time.Time, time.Time) {
	tz := c.ResolveTimezone(destinationCountry)
	loc, err := loadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	// Project departure into destination timezone and add 1 day
	depLocal := departure.In(loc)
	arrivalDate := depLocal.AddDate(0, 0, 1)

	// Sunday skip: if arrival lands on Sunday, push to Monday
	if arrivalDate.Weekday() == time.Sunday {
		arrivalDate = arrivalDate.AddDate(0, 0, 1)
	}

	// Out for Delivery: random between 7:30 AM and 9:00 AM destination local
	ofdOffset := rand.IntN(91) // 0 to 90 minutes after 7:30
	totalMin := 7*60 + 30 + ofdOffset
	ofdHour := totalMin / 60
	ofdMin := totalMin % 60
	outForDelivery := time.Date(arrivalDate.Year(), arrivalDate.Month(), arrivalDate.Day(), ofdHour, ofdMin, 0, 0, loc).UTC()

	// Expected Delivery: random between 9:00 AM and 5:00 PM destination local
	hour := 9 + rand.IntN(8) // 9, 10, 11, 12, 13, 14, 15, 16
	minute := rand.IntN(60)
	arrival := time.Date(arrivalDate.Year(), arrivalDate.Month(), arrivalDate.Day(), hour, minute, 0, 0, loc).UTC()

	return arrival, outForDelivery
}
