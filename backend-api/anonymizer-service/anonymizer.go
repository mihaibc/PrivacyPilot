package main

import "regexp"

// anonymizeText replaces capitalized words (simulating names) with "[REDACTED]".
// In production, integrate an AI engine or more sophisticated detection.
func anonymizeText(text string) string {
	re := regexp.MustCompile(`\b[A-Z][a-z]+\b`)
	return re.ReplaceAllString(text, "[REDACTED]")
}
