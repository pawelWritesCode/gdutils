package gdutils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//charset represents set of string characters of letters and numbers
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

//charsetLettersOnly represents set of string characters including only letters
const charsetLettersOnly = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//ResetApiFeature resets ApiFeature struct instance to default values.
func (af *ApiFeature) ResetApiFeature() {
	af.saved = map[string]interface{}{}
	af.lastResponseBody = []byte{}
	af.lastResponse = &http.Response{}
	af.baseUrl = ""
}

//save preserve value under given key.
func (af *ApiFeature) save(key string, value interface{}) {
	af.saved[key] = value
}

//getSaved returns preserved value if present, error otherwise.
func (af *ApiFeature) getSaved(key string) (interface{}, error) {
	val, ok := af.saved[key]

	if ok == false {
		return val, ErrPreservedData
	}

	return val, nil
}

//setBaseUrl sets base url for requests.
func (af *ApiFeature) setBaseUrl(url string) {
	af.baseUrl = url
}

//replaceTemplatedValue accept as input string, within which search for values
//between two square brackets, for example: [anything]
//and replace them with corresponding preserved values, if they are previously saved.
//
//returns input string with replaced values.
func (af *ApiFeature) replaceTemplatedValue(inputString string) (string, error) {
	result := inputString
	re := regexp.MustCompile(`\[([^\[\]]*)\]`)
	submatchall := re.FindAllString(inputString, -1)
	for _, element := range submatchall {
		extractedVariableName := element
		extractedVariableName = strings.Trim(extractedVariableName, "[")
		extractedVariableName = strings.Trim(extractedVariableName, "]")
		val, err := af.getSaved(extractedVariableName)

		if errors.Is(err, ErrPreservedData) {
			return "", fmt.Errorf("%w, missing value under key '%s'", err, extractedVariableName)
		}

		switch v := val.(type) {
		case string:
			result = strings.Replace(result, element, v, 1)
		case float64:
			result = strings.Replace(result, element, strconv.Itoa(int(v)), 1)
		case int:
			result = strings.Replace(result, element, strconv.Itoa(v), 1)
		default:
			break
		}
	}

	return result, nil
}

//getJsonSchemaBytes reads json schema from file and returns its bytes.
//Argument should be relative path starting from documentation/redoc/ folder.
//file path should NOT finish with .json
//
//example: getJsonSchemaBytes(response/city)
func (af *ApiFeature) getJsonSchemaBytes(name string) ([]byte, error) {
	return ioutil.ReadFile("../documentation/redoc/" + name + ".json")
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

func (af *ApiFeature) iSendInternalRequest(method, url string, reader io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, af.baseUrl+url, reader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	af.lastResponse = resp
	af.lastResponseBody = body

	return resp, err
}
