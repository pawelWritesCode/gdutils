package gdutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"text/template"
	"time"
)

const (
	//charset represents set of string characters of letters and numbers
	charsetUnicodeCharacters = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" + "🤡🤖🧟🏋🥇☟💄🐲🌓🌪🇵🇱⚥❄☠⌘©®💵⓵ "

	//charsetLettersOnly represents set of string characters including only letters
	charsetLettersOnly = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//replaceTemplatedValue accept as input string, within which search for values
//between two brackets {{ }} preceded with dot, for example: {{.NAME}}
//and replace them with corresponding preserved values, if they are previously cache.
//
//returns input string with replaced values.
func (s *Scenario) replaceTemplatedValue(inputString string) (string, error) {
	templ := template.Must(template.New("abc").Parse(inputString))
	var buff bytes.Buffer
	err := templ.Execute(&buff, s.cache)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

//stringWithCharset returns random string of given length.
//Argument length indices length of output string.
//Argument charset indices input charset from which output string will be composed
func (s *Scenario) stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

//randomInt returns random int from provided range
//"from" should be less or equal than "to" otherwise func will panic
func randomInt(from, to int) int {
	if to < from {
		panic(fmt.Sprintf("could not generate random int because %d is less than %d", from, to))
	}

	rand.Seed(time.Now().UnixNano())
	return rand.Intn(to-from+1) + from
}

//valueIsNil checks whether provided Value is nil
func valueIsNil(v reflect.Value) bool {
	nodeKind := v.Kind()
	if nodeKind == reflect.Ptr || nodeKind == reflect.Map || nodeKind == reflect.Array ||
		nodeKind == reflect.Chan || nodeKind == reflect.Slice {
		return v.IsNil()
	}

	return false
}

//theResponseShouldBeInJSON checks if last response body is in JSON format.
func (s *Scenario) theResponseShouldBeInJSON() error {
	var js map[string]interface{}
	var js2 []map[string]interface{}

	if json.Unmarshal(s.GetLastResponseBody(), &js) == nil || json.Unmarshal(s.GetLastResponseBody(), &js2) == nil {
		return nil
	}

	return fmt.Errorf("response has %w", ErrJson)
}
