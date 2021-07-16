package gdutils

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"text/template"
	"time"
)

type bodyHeaders struct {
	Body    interface{}
	Headers map[string]string
}

const (
	//charset represents set of string characters of letters and numbers
	charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	//charsetLettersOnly represents set of string characters including only letters
	charsetLettersOnly = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//ResetApiFeature resets ApiFeature struct instance to default values.
func (af *ApiFeature) ResetApiFeature(isDebug bool) {
	af.saved = map[string]interface{}{}
	af.lastResponseBody = []byte{}
	af.lastResponse = &http.Response{}
	af.isDebug = isDebug
}

//Save preserve value under given key.
func (af *ApiFeature) Save(key string, value interface{}) {
	af.saved[key] = value
}

//GetSaved returns preserved value if present, error otherwise.
func (af *ApiFeature) GetSaved(key string) (interface{}, error) {
	val, ok := af.saved[key]

	if ok == false {
		return val, ErrPreservedData
	}

	return val, nil
}

//replaceTemplatedValue accept as input string, within which search for values
//between two brackets {{ }} preceded with dot, for example: {{.NAME}}
//and replace them with corresponding preserved values, if they are previously saved.
//
//returns input string with replaced values.
func (af *ApiFeature) replaceTemplatedValue(inputString string) (string, error) {
	templ := template.Must(template.New("abc").Parse(inputString))
	var buff bytes.Buffer
	err := templ.Execute(&buff, af.saved)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

//stringWithCharset returns random string of given length.
//Argument length indices length of output string.
//Argument charset indices input charset from which output string will be composed
func (af *ApiFeature) stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

//saveLastResponseCredentials preserve last HTTP request's response and it's response body.
//Returns error if present.
func (af *ApiFeature) saveLastResponseCredentials(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	af.lastResponse = resp
	af.lastResponseBody = body

	return err
}
