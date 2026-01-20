package utils

import (
	"regexp"
	"strings"
)

// SplitIntoSyllables tries to split a word into its phonetic syllables using a basic heuristic.
// Note: This is an approximation for English/International names.
func SplitIntoSyllables(word string) []string {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil
	}

	// Remove non-alphabetic characters for syllable counting
	reg := regexp.MustCompile("[^a-zA-Z]")
	cleanWord := reg.ReplaceAllString(word, "")

	if len(cleanWord) <= 3 {
		return []string{cleanWord}
	}

	// Vowel-based heuristic:
	// A syllable is usually a group of characters containing one vowel sound.
	// We look for vowel-consonant clusters.
	vowels := "aeiouyAEIOUY"
	isVowel := func(c rune) bool {
		return strings.ContainsRune(vowels, c)
	}

	var syllables []string
	start := 0
	runes := []rune(cleanWord)

	for i := 0; i < len(runes); i++ {
		// If we find a vowel followed by a consonant, we might have a boundary
		if isVowel(runes[i]) {
			// Look ahead
			if i+1 < len(runes) && !isVowel(runes[i+1]) {
				// If there's another vowel later, we split after the consonant or vowel
				hasMoreVowels := false
				for j := i + 1; j < len(runes); j++ {
					if isVowel(runes[j]) {
						hasMoreVowels = true
						break
					}
				}

				if hasMoreVowels {
					// Basic split rule: VCV -> V-CV, VCCV -> VC-CV
					if i+2 < len(runes) && !isVowel(runes[i+2]) {
						// VCCV
						syllables = append(syllables, string(runes[start:i+2]))
						start = i + 2
						i = i + 1
					} else {
						// VCV
						syllables = append(syllables, string(runes[start:i+1]))
						start = i + 1
					}
				}
			}
		}
	}

	// Add the remaining part
	if start < len(runes) {
		syllables = append(syllables, string(runes[start:]))
	}

	// Post-process: Merge very short leftovers if they aren't the only ones
	if len(syllables) > 1 && len(syllables[len(syllables)-1]) < 2 {
		last := syllables[len(syllables)-1]
		syllables = syllables[:len(syllables)-1]
		syllables[len(syllables)-1] += last
	}

	return syllables
}
