// Package stringutils holds utility methods for working with strings.
package stringutils

import (
	"math/rand"
	"time"
)

const (
	// CharsetUnicode represents set of Unicode characters (contain multi byte runes).
	CharsetUnicode = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" + "🤡🤖🧟🏋🥇☟💄🐲🌓🌪🇵🇱⚥❄☠⌘©®💵⓵ " + "ęśćżźłóń"

	// CharsetASCII represents set containing only ASCII characters.
	CharsetASCII = " !\"#$%&\\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// RunesFromCharset returns random slice of runes of given length.
// Argument length indices length of output slice.
// Argument charset indices input charset from which output slice will be composed.
func RunesFromCharset(length int, charset []rune) []rune {
	output := make([]rune, 0, length)
	charsetR := charset

	for i := 0; i < length; i++ {
		output = append(output, charsetR[seededRand.Intn(len(charsetR))])
	}

	return output
}
