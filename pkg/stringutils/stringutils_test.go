package stringutils

import (
	"math/rand"
	"testing"
)

func TestStringWithCharset(t *testing.T) {
	for i := 1; i < 10; i++ {
		randomNumberOfLetters := rand.Intn(50)
		randomStringASCII := StringWithCharset(randomNumberOfLetters, CharsetASCII)
		randomStringUnicode := StringWithCharset(randomNumberOfLetters, CharsetUnicode)
		if len(randomStringASCII) != randomNumberOfLetters {
			t.Errorf("expected string of lenght %d, got %d", randomNumberOfLetters, len(randomStringASCII))
		}

		if len(randomStringUnicode) != randomNumberOfLetters {
			t.Errorf("expected string of lenght %d, got %d", randomNumberOfLetters, len(randomStringUnicode))
		}
	}
}
