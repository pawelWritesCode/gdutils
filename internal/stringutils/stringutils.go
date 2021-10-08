package stringutils

import (
	"math/rand"
	"time"
)

const (
	//CharsetUnicode represents set of Unicode characters
	CharsetUnicode = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" + "🤡🤖🧟🏋🥇☟💄🐲🌓🌪🇵🇱⚥❄☠⌘©®💵⓵ " + "ęśćżźłóń"

	//CharsetASCII represents set of ASCII characters
	CharsetASCII = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//StringWithCharset returns random string of given length.
//Argument length indices length of output string.
//Argument charset indices input charset from which output string will be composed
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
