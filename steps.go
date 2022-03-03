package gdutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/moul/http2curl"

	"github.com/pawelWritesCode/gdutils/pkg/format"
	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
	"github.com/pawelWritesCode/gdutils/pkg/mathutils"
	"github.com/pawelWritesCode/gdutils/pkg/reflectutils"
	"github.com/pawelWritesCode/gdutils/pkg/stringutils"
	"github.com/pawelWritesCode/gdutils/pkg/timeutils"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

// BodyHeaders is entity that holds information about request body and request headers.
type BodyHeaders struct {

	// Body should contain HTTP(s) request body
	Body interface{}

	// Headers should contain HTTP(s) request headers
	Headers map[string]string
}

/*
	ISendRequestToWithFormatBodyAndHeaders sends HTTP(s) requests with provided body and headers.

	Argument "method" indices HTTP request method for example: "POST", "GET" etc.
 	Argument "urlTemplate" should be full valid URL. May include template values.
	Argument "bodyTemplate" should contain data (may include template values)
	in JSON or YAML format with keys "body" and "headers".
*/
func (s *State) ISendRequestToWithBodyAndHeaders(method, urlTemplate string, bodyTemplate string) error {
	input, err := s.TemplateEngine.Replace(bodyTemplate, s.Cache.All())
	if err != nil {
		return err
	}

	url, err := s.TemplateEngine.Replace(urlTemplate, s.Cache.All())
	if err != nil {
		return err
	}

	var bodyAndHeaders BodyHeaders
	var dataFormat format.DataFormat
	if format.IsJSON([]byte(input)) {
		dataFormat = format.JSON
	} else if format.IsYAML([]byte(input)) {
		dataFormat = format.YAML
	} else {
		return fmt.Errorf("%w: data is passed in unknown format", ErrGdutils)
	}

	switch dataFormat {
	case format.JSON:
		err = s.Formatters.JSON.Deserialize([]byte(input), &bodyAndHeaders)
	case format.YAML:
		err = s.Formatters.YAML.Deserialize([]byte(input), &bodyAndHeaders)
	default:
		err = fmt.Errorf("unknown data format: '%s'", dataFormat)
	}

	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	var reqBody []byte
	switch dataFormat {
	case format.JSON:
		reqBody, err = s.Formatters.JSON.Serialize(bodyAndHeaders.Body)
	case format.YAML:
		reqBody, err = s.Formatters.YAML.Serialize(bodyAndHeaders.Body)
	default:
		err = fmt.Errorf("unknown data format: '%s'", dataFormat)
	}

	if err != nil {
		return fmt.Errorf("%w: %s", ErrJson, err.Error())
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPReqRes, err.Error())
	}

	for headerName, headerValue := range bodyAndHeaders.Headers {
		req.Header.Set(headerName, headerValue)
	}

	if s.Debugger.IsOn() {
		command, _ := http2curl.GetCurlCommand(req)
		s.Debugger.Print(command.String())
	}

	s.Cache.Save(httpcache.LastHTTPRequestTimestamp, time.Now())

	resp, err := s.RequestDoer.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPReqRes, err.Error())
	}

	s.Cache.Save(httpcache.LastHTTPResponseTimestamp, time.Now())
	s.Cache.Save(httpcache.LastHTTPResponseCacheKey, resp)

	if s.Debugger.IsOn() {
		respBody, _ := s.GetLastResponseBody()
		s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", respBody))
	}

	return nil
}

// IPrepareNewRequestToAndSaveItAs prepares new request and saves it in cache under cacheKey
func (s *State) IPrepareNewRequestToAndSaveItAs(method, urlTemplate, cacheKey string) error {
	url, err := s.TemplateEngine.Replace(urlTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("%w: could not create new request, reason: %s", ErrHTTPReqRes, err.Error())
	}

	s.Cache.Save(cacheKey, req)

	return nil
}

// ISetFollowingHeadersForPreparedRequest sets provided headers for previously prepared request.
// incoming data should be in JSON or YAML format
func (s *State) ISetFollowingHeadersForPreparedRequest(cacheKey, headersTemplate string) error {
	headers, err := s.TemplateEngine.Replace(headersTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	var headersMap map[string]string
	headersBytes := []byte(headers)
	if format.IsJSON(headersBytes) {
		if err = s.Formatters.JSON.Deserialize(headersBytes, &headersMap); err != nil {
			return fmt.Errorf("%w: could not parse provided headers: \n\n%s\n\nerr: %s", ErrGdutils, headers, err.Error())
		}
	} else if format.IsYAML(headersBytes) {
		if err = s.Formatters.YAML.Deserialize(headersBytes, &headersMap); err != nil {
			return fmt.Errorf("%w: could not parse provided headers: \n\n%s\n\nerr: %s", ErrGdutils, headers, err.Error())
		}
	} else {
		return fmt.Errorf("%w: unknown data format", ErrGdutils)
	}

	req, err := s.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	for hName, hValue := range headersMap {
		req.Header.Set(hName, hValue)
	}

	s.Cache.Save(cacheKey, req)

	return nil
}

// ISetFollowingBodyForPreparedRequest sets body for previously prepared request
// bodyTemplate may be in any format and accepts template values
func (s *State) ISetFollowingBodyForPreparedRequest(cacheKey, bodyTemplate string) error {
	body, err := s.TemplateEngine.Replace(bodyTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	req, err := s.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	req.Body = ioutil.NopCloser(bytes.NewReader([]byte(body)))
	s.Cache.Save(cacheKey, req)

	return nil
}

// ISetFollowingCookiesForPreparedRequest sets cookies for previously prepared request
// cookies template should be YAML or JSON deserializable on []http.Cookie.
func (s *State) ISetFollowingCookiesForPreparedRequest(cacheKey, cookiesTemplate string) error {
	var cookies []http.Cookie

	userCookies, err := s.TemplateEngine.Replace(cookiesTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPReqRes, err.Error())
	}

	req, err := s.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPReqRes, err.Error())
	}

	userCookiesBytes := []byte(userCookies)
	if format.IsJSON(userCookiesBytes) {
		if err = s.Formatters.JSON.Deserialize(userCookiesBytes, &cookies); err != nil {
			return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
		}
	} else if format.IsYAML(userCookiesBytes) {
		if err = s.Formatters.YAML.Deserialize(userCookiesBytes, &cookies); err != nil {
			return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
		}
	} else {
		return fmt.Errorf("%w: unknown data format", ErrGdutils)
	}

	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}

	s.Cache.Save(cacheKey, req)

	return nil
}

// ISendRequest sends previously prepared HTTP(s) request.
func (s *State) ISendRequest(cacheKey string) error {
	req, err := s.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if s.Debugger.IsOn() {
		command, _ := http2curl.GetCurlCommand(req)
		s.Debugger.Print(command.String())
	}

	s.Cache.Save(httpcache.LastHTTPRequestTimestamp, time.Now())

	resp, err := s.RequestDoer.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPReqRes, err.Error())
	}

	s.Cache.Save(httpcache.LastHTTPResponseTimestamp, time.Now())
	s.Cache.Save(httpcache.LastHTTPResponseCacheKey, resp)

	if s.Debugger.IsOn() {
		respBody, _ := s.GetLastResponseBody()
		s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", respBody))
	}

	return nil
}

// TheResponseStatusCodeShouldBe compare last response status code with given in argument.
func (s *State) TheResponseStatusCodeShouldBe(code int) error {
	lastResponse, err := s.GetLastResponse()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if lastResponse.StatusCode != code {
		return fmt.Errorf("%w: expected status code %d, but got %d", ErrHTTPReqRes, code, lastResponse.StatusCode)
	}

	return nil
}

// TheResponseBodyShouldHaveFormat checks whether last response body has given data format.
// Available data formats are listed in format package.
func (s *State) TheResponseBodyShouldHaveFormat(dataFormat format.DataFormat) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	switch dataFormat {
	case format.JSON:
		isJSON := format.IsJSON(body)
		if isJSON {
			return nil
		}

		return fmt.Errorf("%w: response body is not %s", ErrGdutils, format.JSON)
	case format.YAML:
		isYAML := format.IsYAML(body)
		if isYAML {
			return nil
		}

		return fmt.Errorf("%w: response body is not %s", ErrGdutils, format.YAML)
	//This case checks whether last response body is not any of known types - then it assumes it is plain text
	case format.PlainText:
		if !(format.IsJSON(body) || format.IsYAML(body)) {
			return nil
		}

		return fmt.Errorf("%w: response body is one of: %s or %s", ErrGdutils, format.JSON, format.YAML)
	default:
		return fmt.Errorf("%w: unknown last response body data type, available values: %s, %s, %s",
			ErrHTTPReqRes, format.JSON, format.YAML, format.PlainText)
	}
}

// ISaveAs saves into cache arbitrary passed value.
func (s *State) ISaveAs(value, cacheKey string) error {
	if len(value) == 0 {
		return fmt.Errorf("%w: pass any value", ErrGdutils)
	}

	if len(cacheKey) == 0 {
		return fmt.Errorf("%w: cacheKey should be not empty value", ErrGdutils)
	}

	s.Cache.Save(cacheKey, value)

	return nil
}

// ISaveFromTheLastResponseNodeAs saves from last response body node under given cacheKey key.
// expr should be valid according to injected PathResolver of given data type
func (s *State) ISaveFromTheLastResponseNodeAs(dataFormat format.DataFormat, expr, cacheKey string) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w, err: %s", ErrGdutils, err.Error())
	}

	var iVal interface{}
	switch dataFormat {
	case format.JSON:
		iVal, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iVal, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		if s.Debugger.IsOn() {
			s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", body))
		}

		return fmt.Errorf("%w, err: %s", ErrGdutils, err.Error())
	}

	s.Cache.Save(cacheKey, iVal)

	return nil
}

// IGenerateARandomIntInTheRangeToAndSaveItAs generates random integer from provided range
// and preserve it under given cacheKey key.
func (s *State) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, cacheKey string) error {
	randomInteger, err := mathutils.RandomInt(from, to)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	s.Cache.Save(cacheKey, randomInteger)

	return nil
}

// IGenerateARandomFloatInTheRangeToAndSaveItAs generates random float from provided range
// and preserve it under given cacheKey key.
func (s *State) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, cacheKey string) error {
	randInt, err := mathutils.RandomInt(from, to)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	float01 := rand.Float64()
	strFloat := fmt.Sprintf("%.2f", float01*float64(randInt))
	floatVal, err := strconv.ParseFloat(strFloat, 64)
	if err != nil {
		return fmt.Errorf("%w: error during parsing float, err: %s", ErrGdutils, err.Error())
	}

	s.Cache.Save(cacheKey, floatVal)

	return nil
}

// IGenerateARandomRunesInTheRangeToAndSaveItAs creates random runes generator func using provided charset
// return func creates runes from provided range and preserve it under given cacheKey
func (s *State) IGenerateARandomRunesInTheRangeToAndSaveItAs(charset string) func(from, to int, cacheKey string) error {
	return func(from, to int, cacheKey string) error {
		randInt, err := mathutils.RandomInt(from, to)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
		}

		s.Cache.Save(cacheKey, string(stringutils.RunesFromCharset(randInt, []rune(charset))))

		return nil
	}
}

/*
	IGenerateARandomSentenceInTheRangeFromToWordsAndSaveItAs creates generator func for creating random sentences
	each sentence has length from - to as provided in params and is saved in provided cacheKey
*/
func (s *State) IGenerateARandomSentenceInTheRangeFromToWordsAndSaveItAs(charset string, wordMinLength, wordMaxLength int) func(from, to int, cacheKey string) error {
	return func(from, to int, cacheKey string) error {
		if from > to {
			return fmt.Errorf("%w could not generate sentence because of invalid range provided, from '%d' should not be greater than to: '%d'", ErrGdutils, from, to)
		}

		if wordMinLength > wordMaxLength {
			return fmt.Errorf("%w could not generate sentence because of invalid range provided, wordMinLength '%d' should not be greater than wordMaxLength '%d'", ErrGdutils, wordMinLength, wordMaxLength)
		}

		numberOfWords, err := mathutils.RandomInt(from, to)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrGdutils, err)
		}

		sentence := ""
		for i := 0; i < numberOfWords; i++ {
			lengthOfWord, err := mathutils.RandomInt(wordMinLength, wordMaxLength)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrGdutils, err)
			}

			word := stringutils.RunesFromCharset(lengthOfWord, []rune(charset))
			if i == numberOfWords-1 {
				sentence += string(word)
			} else {
				sentence += string(word) + " "
			}
		}

		s.Cache.Save(cacheKey, sentence)

		return nil
	}
}

// IGetTimeAndTravelByAndSaveItAs accepts time object, move timeDuration in time and
// save it in cache under given cacheKey.
func (s *State) IGetTimeAndTravelByAndSaveItAs(t time.Time, timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error {
	var newTime time.Time
	switch timeDirection {
	case timeutils.TimeDirectionBackward:
		newTime = t.Add(-timeDuration)
	case timeutils.TimeDirectionForward:
		newTime = t.Add(timeDuration)
	default:
		return fmt.Errorf("unknown timeDirection: %s, allowed: %s, %s", timeDirection, timeutils.TimeDirectionForward, timeutils.TimeDirectionBackward)
	}

	s.Cache.Save(cacheKey, newTime)

	return nil
}

// IGenerateCurrentTimeAndTravelByAndSaveItAs creates current time object, move timeDuration in time and
// save it in cache under given cacheKey.
func (s *State) IGenerateCurrentTimeAndTravelByAndSaveItAs(timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error {
	return s.IGetTimeAndTravelByAndSaveItAs(time.Now(), timeDirection, timeDuration, cacheKey)
}

// TheResponseShouldHaveNode checks whether last response body contains given node.
// expr should be valid according to injected PathFinder for given data format
func (s *State) TheResponseShouldHaveNode(dataFormat format.DataFormat, expr string) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	switch dataFormat {
	case format.JSON:
		_, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		_, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		if s.Debugger.IsOn() {
			s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", body))
		}

		return fmt.Errorf("%w, node '%s', err: %s", ErrGdutils, expr, err.Error())
	}

	return nil
}

// TheNodeShouldNotBe checks whether node from last response body is not of provided type.
// goType may be one of: nil, string, int, float, bool, map, slice,
// expr should be valid according to injected PathFinder for given data format.
func (s *State) TheNodeShouldNotBe(dataFormat format.DataFormat, expr string, goType string) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	var iNodeVal interface{}
	switch dataFormat {
	case format.JSON:
		iNodeVal, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iNodeVal, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		return fmt.Errorf("%w, node '%s', err: %s", ErrGdutils, expr, err.Error())
	}

	vNodeVal := reflect.ValueOf(iNodeVal)
	errInvalidType := fmt.Errorf("%w: %s value is \"%s\", but expected not to be", ErrGdutils, expr, goType)
	switch goType {
	case "nil":
		if !vNodeVal.IsValid() || reflectutils.IsValueNil(vNodeVal) {
			return errInvalidType
		}
	case "string":
		if vNodeVal.Kind() == reflect.String {
			return errInvalidType
		}
	case "int":
		if vNodeVal.Kind() == reflect.Int64 || vNodeVal.Kind() == reflect.Int32 || vNodeVal.Kind() == reflect.Int16 ||
			vNodeVal.Kind() == reflect.Int8 || vNodeVal.Kind() == reflect.Int || vNodeVal.Kind() == reflect.Uint ||
			vNodeVal.Kind() == reflect.Uint8 || vNodeVal.Kind() == reflect.Uint16 || vNodeVal.Kind() == reflect.Uint32 ||
			vNodeVal.Kind() == reflect.Uint64 {
			return errInvalidType
		}

		if vNodeVal.Kind() == reflect.Float64 {
			_, frac := math.Modf(vNodeVal.Float())
			if frac == 0 {
				return errInvalidType
			}
		}
	case "float":
		if vNodeVal.Kind() == reflect.Float64 || vNodeVal.Kind() == reflect.Float32 {
			_, frac := math.Modf(vNodeVal.Float())
			if frac == 0 {
				return nil
			}

			return errInvalidType
		}
	case "bool":
		if vNodeVal.Kind() == reflect.Bool {
			return errInvalidType
		}
	case "map":
		if vNodeVal.Kind() == reflect.Map {
			return errInvalidType
		}
	case "slice":
		if vNodeVal.Kind() == reflect.Slice {
			return errInvalidType
		}
	default:
		return fmt.Errorf("%w: %s is unknown type for this step", ErrGdutils, goType)
	}

	return nil
}

// TheNodeShouldBe checks whether node from last response body is of provided type
// goType may be one of: nil, string, int, float, bool, map, slice
// expr should be valid according to injected PathResolver
func (s *State) TheNodeShouldBe(dataFormat format.DataFormat, expr string, goType string) error {
	errInvalidType := fmt.Errorf("%w: %s value is not \"%s\", but expected to be", ErrGdutils, expr, goType)

	switch goType {
	case "nil", "string", "int", "float", "bool", "map", "slice":
		if err := s.TheNodeShouldNotBe(dataFormat, expr, goType); err == nil {
			return errInvalidType
		}

		return nil
	default:
		return fmt.Errorf("%w: %s is unknown type for this step", ErrGdutils, goType)
	}
}

// TheResponseShouldHaveNodes checks whether last request body has keys defined in string separated by comma
// nodeExprs should be valid according to injected PathFinder expressions separated by comma (,)
func (s *State) TheResponseShouldHaveNodes(dataFormat format.DataFormat, expressions string) error {
	keysSlice := strings.Split(expressions, ",")

	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	errs := make([]error, 0, len(keysSlice))
	for _, key := range keysSlice {
		trimmedKey := strings.TrimSpace(key)

		switch dataFormat {
		case format.JSON:
			_, err = s.PathFinders.JSON.Find(trimmedKey, body)
		case format.YAML:
			_, err = s.PathFinders.YAML.Find(trimmedKey, body)
		default:
			return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("%w, node '%s', err: %s", ErrGdutils, trimmedKey, err.Error()))
		}
	}

	if len(errs) > 0 {
		var errString string
		for _, err := range errs {
			errString += fmt.Sprintf("%s\n", err)
		}

		if s.Debugger.IsOn() {
			s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", body))
		}

		return errors.New(errString)
	}

	return nil
}

// TheNodeShouldBeSliceOfLength checks whether given key is slice and has given length
// expr should be valid according to injected PathFinder for provided dataFormat
func (s *State) TheNodeShouldBeSliceOfLength(dataFormat format.DataFormat, expr string, length int) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	var iValue interface{}
	switch dataFormat {
	case format.JSON:
		iValue, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iValue, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		if s.Debugger.IsOn() {
			s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", body))
		}

		return fmt.Errorf("%w, node '%s', err: %s", ErrGdutils, expr, err.Error())
	}

	v := reflect.ValueOf(iValue)
	if v.Kind() == reflect.Slice {
		if v.Len() != length {
			return fmt.Errorf("%w: %s slice has length: %d, expected: %d", ErrGdutils, expr, v.Len(), length)
		}

		return nil
	}

	if s.Debugger.IsOn() {
		s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", body))
	}

	return fmt.Errorf("%w: %s does not point at slice(array) in last HTTP(s) response body", ErrGdutils, expr)
}

// TheNodeShouldBeOfValue compares json node value from expression to expected by user dataValue of given by user dataType
// Available data types are listed in switch section in each case directive.
// expr should be valid according to injected PathFinder for provided dataFormat.
func (s *State) TheNodeShouldBeOfValue(dataFormat format.DataFormat, expr, dataType, dataValue string) error {
	nodeValueReplaced, err := s.TemplateEngine.Replace(dataValue, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if s.Debugger.IsOn() {
		s.Debugger.Print(fmt.Sprintf("replaced value %s\n", nodeValueReplaced))
	}

	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if s.Debugger.IsOn() {
		defer func() {
			s.Debugger.Print(fmt.Sprintf("last response body:\n\n%s\n", body))
		}()
	}

	var iValue interface{}
	switch dataFormat {
	case format.JSON:
		iValue, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iValue, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		return fmt.Errorf("%w, node '%s', err: %s", ErrGdutils, expr, err.Error())
	}

	switch dataType {
	case "string":
		strVal, ok := iValue.(string)
		if !ok {
			return fmt.Errorf("%w: expected %s to be %s, got %v", ErrGdutils, expr, dataType, iValue)
		}

		if strVal != nodeValueReplaced {
			return fmt.Errorf("%w: node %s string value: %s is not equal to expected string value: %s", ErrGdutils, expr, strVal, nodeValueReplaced)
		}
	case "int":
		floatVal, ok := iValue.(float64)
		if !ok {
			uint64Val, ok := iValue.(uint64)
			if !ok {
				return fmt.Errorf("%w: expected %s to be %s, got %v", ErrGdutils, expr, dataType, iValue)
			}

			floatVal = float64(uint64Val)
		}

		intVal := int(floatVal)
		intNodeValue, err := strconv.Atoi(nodeValueReplaced)

		if err != nil {
			return fmt.Errorf("%w: replaced node %s value %s could not be converted to int", ErrGdutils, expr, nodeValueReplaced)
		}

		if intVal != intNodeValue {
			return fmt.Errorf("%w: node %s int value: %d is not equal to expected int value: %d", ErrGdutils, expr, intVal, intNodeValue)
		}
	case "float":
		floatVal, ok := iValue.(float64)
		if !ok {
			return fmt.Errorf("%w: expected %s to be %s, got %v", ErrGdutils, expr, dataType, iValue)
		}

		floatNodeValue, err := strconv.ParseFloat(nodeValueReplaced, 64)
		if err != nil {
			return fmt.Errorf("%w: replaced node %s value %s could not be converted to float64", ErrGdutils, expr, nodeValueReplaced)
		}

		if floatVal != floatNodeValue {
			return fmt.Errorf("%w: node %s float value %f is not equal to expected float value %f", ErrGdutils, expr, floatVal, floatNodeValue)
		}
	case "bool":
		boolVal, ok := iValue.(bool)
		if !ok {
			strVal, ok := iValue.(string)
			if !ok {
				return fmt.Errorf("%w: expected %s to be %s, got %v", ErrGdutils, expr, dataType, iValue)
			}

			boolVal, err = strconv.ParseBool(strVal)
			if err != nil {
				return fmt.Errorf("%w: expected %s to be %s, got %v", ErrGdutils, expr, dataType, iValue)
			}
		}

		boolNodeValue, err := strconv.ParseBool(nodeValueReplaced)
		if err != nil {
			return fmt.Errorf("%w: replaced node %s value %s could not be converted to bool", ErrGdutils, expr, nodeValueReplaced)
		}

		if boolVal != boolNodeValue {
			return fmt.Errorf("%w: node %s bool value %t is not equal to expected bool value %t", ErrGdutils, expr, boolVal, boolNodeValue)
		}
	}

	return nil
}

// TheNodeShouldMatchRegExp checks whether last response body node matches provided regExp.
func (s *State) TheNodeShouldMatchRegExp(dataFormat format.DataFormat, expr, regExpTemplate string) error {
	regExpString, err := s.TemplateEngine.Replace(regExpTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	var iValue interface{}
	switch dataFormat {
	case format.JSON:
		iValue, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iValue, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		return fmt.Errorf("%w, node '%s', err: %s", ErrGdutils, expr, err.Error())
	}

	jsonValue, err := s.Formatters.JSON.Serialize(iValue)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if s.Debugger.IsOn() {
		s.Debugger.Print(fmt.Sprintf("matching: %s, with regExp: %s", jsonValue, regExpString))
	}

	matched, err := regexp.Match(regExpString, jsonValue)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if !matched {
		return fmt.Errorf("%s does not contain any: %s", string(jsonValue), regExpString)
	}

	return nil
}

// TheResponseShouldHaveHeader checks whether last HTTP response has given header.
func (s *State) TheResponseShouldHaveHeader(name string) error {
	defer func() {
		if s.Debugger.IsOn() {
			lastResp, err := s.GetLastResponse()
			if err != nil {
				s.Debugger.Print("could not obtain last response headers")
			}

			s.Debugger.Print(fmt.Sprintf("last HTTP(s) response headers: %+v", lastResp.Header))
		}
	}()

	lastResp, err := s.GetLastResponse()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	header := lastResp.Header.Get(name)
	if header != "" {
		return nil
	}

	if s.Debugger.IsOn() {
		s.Debugger.Print(fmt.Sprintf("last HTTP(s) response headers: %+v", lastResp.Header))
	}

	return fmt.Errorf("%w: could not find header %s in last HTTP response", ErrHTTPReqRes, name)
}

// TheResponseShouldHaveHeaderOfValue checks whether last HTTP response has given header with provided valueTemplate.
func (s *State) TheResponseShouldHaveHeaderOfValue(name, valueTemplate string) error {
	defer func() {
		if s.Debugger.IsOn() {
			lastResp, err := s.GetLastResponse()
			if err != nil {
				s.Debugger.Print("could not obtain last response headers")
			}

			s.Debugger.Print(fmt.Sprintf("last HTTP(s) response headers: %+v", lastResp.Header))
		}
	}()

	lastResp, err := s.GetLastResponse()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrHTTPReqRes, err.Error())
	}

	header := lastResp.Header.Get(name)
	value, err := s.TemplateEngine.Replace(valueTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	if header == "" && value == "" {
		return fmt.Errorf("%w: could not find header %s in last HTTP response", ErrHTTPReqRes, name)
	}

	if header == value {
		return nil
	}

	return fmt.Errorf("%w: %s header exists but, expected value: %s, is not equal to actual: %s", ErrHTTPReqRes, name, value, header)
}

// IValidateLastResponseBodyWithSchemaReference validates last response body against schema as provided in reference.
// reference may be: URL or full/relative path
func (s *State) IValidateLastResponseBodyWithSchemaReference(reference string) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err)
	}

	return s.SchemaValidators.ReferenceValidator.Validate(string(body), reference)
}

// IValidateLastResponseBodyWithSchemaString validates last response body against schema.
func (s *State) IValidateLastResponseBodyWithSchemaString(schema string) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err)
	}

	return s.SchemaValidators.StringValidator.Validate(string(body), schema)
}

// IValidateNodeWithSchemaString validates last response body JSON node against schema
func (s *State) IValidateNodeWithSchemaString(dataFormat format.DataFormat, expr, schema string) error {
	return s.iValidateNodeWithSchemaGeneral(dataFormat, expr, schema, s.SchemaValidators.StringValidator)
}

// IValidateNodeWithSchemaReference validates last response body node against schema as provided in reference
func (s *State) IValidateNodeWithSchemaReference(dataFormat format.DataFormat, expr, reference string) error {
	return s.iValidateNodeWithSchemaGeneral(dataFormat, expr, reference, s.SchemaValidators.ReferenceValidator)
}

// TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo asserts that last HTTP request-response time
// is <= than expected timeInterval.
// timeInterval should be string acceptable by time.ParseDuration func
func (s *State) TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo(timeInterval time.Duration) error {
	lastReqTimestampI, err := s.Cache.GetSaved(httpcache.LastHTTPRequestTimestamp)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	lastReqTimestamp, ok := lastReqTimestampI.(time.Time)
	if !ok {
		return fmt.Errorf("%w: last request timestamp: '%+v' should be time.Time", ErrHTTPReqRes, lastReqTimestampI)
	}

	lastResTimestampI, err := s.Cache.GetSaved(httpcache.LastHTTPResponseTimestamp)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	lastResTimestamp, ok := lastResTimestampI.(time.Time)
	if !ok {
		return fmt.Errorf("%w: last response timestamp: '%+v' should be time.Time", ErrHTTPReqRes, lastResTimestampI)
	}

	timeBetweenReqRes := lastResTimestamp.Sub(lastReqTimestamp)
	if timeBetweenReqRes > timeInterval {
		return fmt.Errorf("%w: time between last request - response should be less than %+v, but it took %+v", ErrHTTPReqRes, timeInterval, timeBetweenReqRes)
	}

	return nil
}

// TheResponseShouldHaveCookie checks whether last HTTP(s) response has cookie of given name.
func (s *State) TheResponseShouldHaveCookie(name string) error {
	lastResp, err := s.GetLastResponse()
	if err != nil {
		return fmt.Errorf("%w: could not obtain last response, err: %s", ErrHTTPReqRes, err.Error())
	}

	defer func() {
		if s.Debugger.IsOn() {

			s.Debugger.Print(fmt.Sprintf("last HTTP(s) response cookies: %+v", lastResp.Cookies()))
		}
	}()

	for _, cookie := range lastResp.Cookies() {
		if cookie.Name == name {
			return nil
		}
	}

	return fmt.Errorf("%w: last HTTP(s) response does not have cookie with name '%s'", ErrHTTPReqRes, name)
}

// TheResponseShouldHaveCookieOfValue checks whether last HTTP(s) response has cookie of given name and value.
func (s *State) TheResponseShouldHaveCookieOfValue(name, valueTemplate string) error {
	lastResp, err := s.GetLastResponse()
	if err != nil {
		return fmt.Errorf("%w: could not obtain last response, err: %s", ErrHTTPReqRes, err.Error())
	}

	defer func() {
		if s.Debugger.IsOn() {
			s.Debugger.Print(fmt.Sprintf("last HTTP(s) response cookies: %+v", lastResp.Cookies()))
		}
	}()

	value, err := s.TemplateEngine.Replace(valueTemplate, s.Cache.All())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	for _, cookie := range lastResp.Cookies() {
		if cookie.Name == name && cookie.Value == value {
			return nil
		}
	}

	return fmt.Errorf("%w: last HTTP(s) response does not have cookie with name '%s' and value: '%s'", ErrHTTPReqRes, name, value)
}

// IWait waits for given timeInterval amount of time
func (s *State) IWait(timeInterval time.Duration) error {
	time.Sleep(timeInterval)

	return nil
}

// IPrintLastResponseBody prints last response from request.
func (s *State) IPrintLastResponseBody() error {
	var tmp interface{}

	body, err := s.GetLastResponseBody()
	if err != nil {
		return err
	}

	if format.IsYAML(body) {
		err = yaml.Unmarshal(body, &tmp)

		if err != nil {
			s.Debugger.Print(string(body))
			return nil
		}

		indentedRespBody, err := yaml.Marshal(tmp)
		if err != nil {
			s.Debugger.Print(string(body))
			return nil
		}

		s.Debugger.Print(string(indentedRespBody))
		return nil
	}

	if format.IsJSON(body) {
		err = json.Unmarshal(body, &tmp)

		if err != nil {
			s.Debugger.Print(string(body))
			return nil
		}

		indentedRespBody, err := json.MarshalIndent(tmp, "", "\t")
		if err != nil {
			s.Debugger.Print(string(body))
			return nil
		}

		s.Debugger.Print(string(indentedRespBody))
		return nil
	}

	s.Debugger.Print(string(body))

	return nil
}

// IStartDebugMode starts debugging mode
func (s *State) IStartDebugMode() error {
	s.Debugger.TurnOn()

	return nil
}

// IStopDebugMode stops debugging mode
func (s *State) IStopDebugMode() error {
	s.Debugger.TurnOff()

	return nil
}

// GetPreparedRequest returns prepared request from cache or error if failed
func (s *State) GetPreparedRequest(cacheKey string) (*http.Request, error) {
	reqInterface, err := s.Cache.GetSaved(cacheKey)
	if err != nil {
		return &http.Request{}, fmt.Errorf("%w: %s", ErrGdutils, err.Error())
	}

	req, ok := reqInterface.(*http.Request)
	if !ok {
		return &http.Request{}, fmt.Errorf("%w: value under key %s in cache doesn't contain *http.Request", ErrPreservedData, cacheKey)
	}

	return req, nil
}

// GetLastResponse returns last HTTP(s) response.
func (s *State) GetLastResponse() (*http.Response, error) {
	respInterface, err := s.Cache.GetSaved(httpcache.LastHTTPResponseCacheKey)
	if err != nil {
		return nil, fmt.Errorf("%w: missing last HTTP(s) response, err: %s", ErrHTTPReqRes, err.Error())
	}

	lastResp, ok := respInterface.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("%w: last HTTP(s) response data structure is not type *http.Response", ErrHTTPReqRes)
	}

	if lastResp == nil {
		return nil, fmt.Errorf("%w: missing last HTTP(s) response", ErrHTTPReqRes)
	}

	return lastResp, nil
}

// GetLastResponseBody returns last HTTP(s) response body.
// internally method creates new NoPCloser on last response so this method is safe to reuse many times
func (s *State) GetLastResponseBody() ([]byte, error) {
	lastResponse, err := s.GetLastResponse()
	if err != nil {
		return []byte(""), err
	}

	var bodyBytes []byte

	if lastResponse != nil {
		bodyBytes, _ = ioutil.ReadAll(lastResponse.Body)
		defer lastResponse.Body.Close()

		// last response body may be read again
		lastResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return bodyBytes, nil
}

// iValidateNodeWithSchemaGeneral validates last response body node against schema as provided in reference.
func (s *State) iValidateNodeWithSchemaGeneral(dataFormat format.DataFormat, expr, reference string, validator validator.SchemaValidator) error {
	body, err := s.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err)
	}

	var node interface{}
	switch dataFormat {
	case format.JSON:
		node, err = s.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		node, err = s.PathFinders.YAML.Find(expr, body)
	default:
		return fmt.Errorf("%w: unknown data format: %s", ErrGdutils, dataFormat)
	}

	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err)
	}

	jsonNode, err := s.Formatters.JSON.Serialize(node)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGdutils, err)
	}

	if s.Debugger.IsOn() {
		s.Debugger.Print(fmt.Sprintf("%+v\n\n was marshaled to:\n\n %s \n\nand passed for validation", node, jsonNode))
	}

	return validator.Validate(string(jsonNode), reference)
}
