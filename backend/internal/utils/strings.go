package utils

import (
	"regexp"
	"strings"
)

// GenerateAbbreviation creates a 3-letter abbreviation from a company name
// for use as a tracking ID prefix. Returns "AWB" for empty/invalid inputs.
func GenerateAbbreviation(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "AWB"
	}

	reg := regexp.MustCompile("[^a-zA-Z]")
	clean := reg.ReplaceAllString(name, "")
	if clean == "" {
		return "AWB"
	}

	syllables := SplitIntoSyllables(clean)
	count := len(syllables)

	var abbr string
	switch {
	case count == 1:
		s := syllables[0]
		if len(s) <= 3 {
			abbr = s
		} else {
			abbr = string(s[0]) + string(s[len(s)/2]) + string(s[len(s)-1])
		}
	case count == 2:
		s1 := syllables[0]
		s2 := syllables[1]
		p1 := s1
		if len(s1) > 2 {
			p1 = s1[:2]
		}
		p2 := string(s2[0])
		abbr = p1 + p2
	default:
		for i := 0; i < 3 && i < count; i++ {
			abbr += string(syllables[i][0])
		}
	}

	abbr = strings.ToUpper(abbr)
	if len(abbr) > 3 {
		abbr = abbr[:3]
	}
	for len(abbr) < 3 {
		abbr += "X"
	}

	return abbr
}

// GetBarePhone extracts only the digits before any device/suffix markers (@s.whatsapp.net, :1, etc).
func GetBarePhone(jid string) string {
	if jid == "" {
		return ""
	}
	if idx := strings.IndexByte(jid, '@'); idx != -1 {
		jid = jid[:idx]
	}
	if idx := strings.IndexByte(jid, ':'); idx != -1 {
		jid = jid[:idx]
	}
	return jid
}
