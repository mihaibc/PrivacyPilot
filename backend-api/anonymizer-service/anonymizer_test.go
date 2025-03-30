package main

import "testing"

func TestAnonymizeText(t *testing.T) {
	input := "Alice went to Wonderland"
	expected := "[REDACTED] went to [REDACTED]"
	result := anonymizeText(input)
	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}
