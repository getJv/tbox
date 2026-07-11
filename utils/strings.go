package utils

import (
	"math/rand/v2"
	"strings"
)

// MaskRandomWord replaces a random word in the sentence with "____".
func MaskRandomWord(sentence string) string {
	words := strings.Fields(sentence)
	if len(words) == 0 {
		return sentence
	}

	// Select a random index using math/rand/v2
	randomIndex := rand.IntN(len(words))

	// Replace the word
	words[randomIndex] = "____"

	return strings.Join(words, " ")
}

// MaskString replaces shorter string (10 characters) to '******' and longer 'abcde...abcde'.
func MaskString(s string) string {
	if s == "" {
		return ""
	}
	n := len(s)
	if n <= 10 {
		return "******"
	}
	return s[:5] + "..." + s[n-5:]
}
