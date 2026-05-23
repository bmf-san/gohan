package processor

import (
	"math"
	"unicode"
)

// readingSpeedWPM is the assumed reading speed in words per minute, used for
// reading-time estimates. Each CJK character is treated as one "word" for the
// purposes of this calculation.
const readingSpeedWPM = 250

// countWords returns the number of words in src. ASCII/Latin words are
// counted as whitespace-separated tokens; each CJK character contributes one
// to the count as well. Markdown syntax characters are not stripped; the
// resulting count is approximate but stable.
func countWords(src string) int {
	var count int
	inWord := false
	for _, r := range src {
		if isCJK(r) {
			if inWord {
				count++
				inWord = false
			}
			count++
			continue
		}
		if unicode.IsSpace(r) {
			if inWord {
				count++
				inWord = false
			}
			continue
		}
		inWord = true
	}
	if inWord {
		count++
	}
	return count
}

// isCJK reports whether r is a CJK ideograph, hiragana, katakana, or hangul
// character. Punctuation and full-width Latin are not included.
func isCJK(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified Ideographs
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Extension A
		return true
	case r >= 0x3040 && r <= 0x309F: // Hiragana
		return true
	case r >= 0x30A0 && r <= 0x30FF: // Katakana
		return true
	case r >= 0xAC00 && r <= 0xD7A3: // Hangul Syllables
		return true
	}
	return false
}

// readingTimeMinutes returns the estimated reading time in minutes for the
// given word count. The result is always at least 1 when the count is positive.
func readingTimeMinutes(words int) int {
	if words <= 0 {
		return 0
	}
	m := int(math.Ceil(float64(words) / float64(readingSpeedWPM)))
	if m < 1 {
		m = 1
	}
	return m
}
