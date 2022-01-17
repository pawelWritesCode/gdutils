package stringutils

import (
	"math/rand"
	"testing"
)

func TestStringWithCharset(t *testing.T) {
	for i := 1; i < 10; i++ {
		randomNumberOfRunes := rand.Intn(50)
		randomRunesASCII := RunesFromCharset(randomNumberOfRunes, []rune(CharsetASCII))
		randomRunesUnicode := RunesFromCharset(randomNumberOfRunes, []rune(CharsetUnicode))

		if len(randomRunesASCII) != randomNumberOfRunes {
			t.Errorf("expected slice of ascii runes of lenght %d, got %d", randomNumberOfRunes, len(randomRunesASCII))
		}

		if len(randomRunesUnicode) != randomNumberOfRunes {
			t.Errorf("expected slice of unicode runes of lenght %d, got %d", randomNumberOfRunes, len(randomRunesUnicode))
		}
	}
}
