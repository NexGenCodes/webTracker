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
	CalculateDeparture(now time.Time, adminTZ string) time.Time
	CalculateArrival(departure time.Time, senderCountry, receiverCountry string) (arrival, outfordelivery time.Time)
}

// Calculator handles timezone resolution and timeline calculations.
type Calculator struct{}

// Ensure Calculator implements Service
var _ Service = (*Calculator)(nil)

// CountryTimezoneMap maps country names to IANA timezone identifiers.
var CountryTimezoneMap = map[string]string{
	// Africa
	"nigeria":        "Africa/Lagos",
	"ghana":          "Africa/Accra",
	"benin":          "Africa/Porto-Novo",
	"togo":           "Africa/Lome",
	"south africa":   "Africa/Johannesburg",
	"kenya":          "Africa/Nairobi",
	"egypt":          "Africa/Cairo",
	"ethiopia":       "Africa/Addis_Ababa",
	"cameroon":       "Africa/Douala",
	"senegal":        "Africa/Dakar",
	"tanzania":       "Africa/Dar_es_Salaam",
	"morocco":        "Africa/Casablanca",
	"ivory coast":    "Africa/Abidjan",
	"cote d'ivoire":  "Africa/Abidjan",

	// Americas
	"usa":            "America/New_York",
	"united states":  "America/New_York",
	"canada":         "America/Toronto",
	"mexico":         "America/Mexico_City",
	"brazil":         "America/Sao_Paulo",
	"argentina":      "America/Argentina/Buenos_Aires",
	"colombia":       "America/Bogota",
	"chile":          "America/Santiago",
	"peru":           "America/Lima",
	"venezuela":      "America/Caracas",
	"honduras":       "America/Tegucigalpa",
	"guatemala":      "America/Guatemala",
	"ecuador":        "America/Guayaquil",
	"bolivia":        "America/La_Paz",
	"paraguay":       "America/Asuncion",
	"uruguay":        "America/Montevideo",
	"panama":         "America/Panama",
	"costa rica":     "America/Costa_Rica",
	"dominican republic": "America/Santo_Domingo",
	"jamaica":        "America/Jamaica",

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
	"china":          "Asia/Shanghai",
	"india":          "Asia/Kolkata",
	"japan":          "Asia/Tokyo",
	"south korea":    "Asia/Seoul",
	"indonesia":      "Asia/Jakarta",
	"malaysia":       "Asia/Kuala_Lumpur",
	"singapore":      "Asia/Singapore",
	"thailand":       "Asia/Bangkok",
	"vietnam":        "Asia/Ho_Chi_Minh",
	"philippines":    "Asia/Manila",
	"pakistan":        "Asia/Karachi",
	"bangladesh":     "Asia/Dhaka",
	"dubai":          "Asia/Dubai",
	"uae":            "Asia/Dubai",
	"saudi arabia":   "Asia/Riyadh",
	"qatar":          "Asia/Qatar",
	"afghanistan":    "Asia/Kabul",
	"israel":         "Asia/Jerusalem",
	"iraq":           "Asia/Baghdad",
	"iran":           "Asia/Tehran",

	// Oceania
	"australia":      "Australia/Sydney",
	"new zealand":    "Pacific/Auckland",
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
	hour := 9 + rand.IntN(7) // 9, 10, 11, 12, 13, 14, 15
	minute := rand.IntN(60)
	arrival := time.Date(arrivalDate.Year(), arrivalDate.Month(), arrivalDate.Day(), hour, minute, 0, 0, loc).UTC()

	// 3. Out for delivery is always 3-6 hours before final arrival
	outfordelivery := arrival.Add(-time.Duration(3+rand.IntN(3)) * time.Hour)

	return arrival, outfordelivery
}
