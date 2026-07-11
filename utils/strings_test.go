package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskRandomWord(t *testing.T) {
	t.Run("should replace one word with ____", func(t *testing.T) {
		sentence := "The quick brown fox jumps"
		masked := MaskRandomWord(sentence)

		// Verifica se a máscara está presente
		assert.Contains(t, masked, "____")

		// Verifica se o tamanho das palavras (número de espaços + 1) foi mantido
		originalWords := strings.Fields(sentence)
		maskedWords := strings.Fields(masked)
		assert.Equal(t, len(originalWords), len(maskedWords))
	})

	t.Run("should handle single word sentences", func(t *testing.T) {
		sentence := "Hello"
		masked := MaskRandomWord(sentence)

		assert.Equal(t, "____", masked)
	})

	t.Run("should return original string if empty", func(t *testing.T) {
		sentence := ""
		masked := MaskRandomWord(sentence)

		assert.Equal(t, "", masked)
	})
}
