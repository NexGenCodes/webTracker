package shipment

import (
	"strings"
)

// Service defines the interface for shipment utilities.
// Scheduling logic has been moved to the PostgreSQL trigger fn_shipment_auto_schedule().
type Service interface {
	ResolveTimezone(country string) string
}

// Calculator handles timezone resolution for shipment metadata.
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
