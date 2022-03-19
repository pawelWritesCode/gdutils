package gdutils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/pawelWritesCode/gdutils/pkg/osutils"
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
func (apiCtx *APIContext) ISendRequestToWithBodyAndHeaders(method, urlTemplate string, bodyTemplate string) error {
	input, err := apiCtx.TemplateEngine.Replace(bodyTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'headers and body' template, err: %w", err)
	}

	url, err := apiCtx.TemplateEngine.Replace(urlTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'url' template, err: %w", err)
	}

	var bodyAndHeaders BodyHeaders
	var dataFormat format.DataFormat
	if format.IsJSON([]byte(input)) {
		dataFormat = format.JSON
	} else if format.IsYAML([]byte(input)) {
		dataFormat = format.YAML
	} else if format.IsXML([]byte(input)) {
		return fmt.Errorf("this method does not support data in format: %s", format.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", format.JSON, format.YAML)
	}

	switch dataFormat {
	case format.JSON:
		err = apiCtx.Formatters.JSON.Deserialize([]byte(input), &bodyAndHeaders)
	case format.YAML:
		err = apiCtx.Formatters.YAML.Deserialize([]byte(input), &bodyAndHeaders)
	default:
		err = fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", format.JSON, format.YAML)
	}

	if err != nil {
		return fmt.Errorf("can't deserialize 'headers and body' data, err: %w", err)
	}

	var reqBody []byte
	switch dataFormat {
	case format.JSON:
		reqBody, err = apiCtx.Formatters.JSON.Serialize(bodyAndHeaders.Body)
	case format.YAML:
		reqBody, err = apiCtx.Formatters.YAML.Serialize(bodyAndHeaders.Body)
	default:
		err = fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", format.JSON, format.YAML)
	}

	if err != nil {
		return fmt.Errorf("can't serialize data to send it with HTTP(s) request, err: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("can't create request due to err: %w", err)
	}

	for headerName, headerValue := range bodyAndHeaders.Headers {
		req.Header.Set(headerName, headerValue)
	}

	if apiCtx.Debugger.IsOn() {
		command, _ := http2curl.GetCurlCommand(req)
		apiCtx.Debugger.Print(command.String())
	}

	apiCtx.Cache.Save(httpcache.LastHTTPRequestTimestamp, time.Now())

	resp, err := apiCtx.RequestDoer.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request %s %s, reason: %w", req.Method, req.URL.String(), err)
	}

	apiCtx.Cache.Save(httpcache.LastHTTPResponseTimestamp, time.Now())
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, resp)

	if apiCtx.Debugger.IsOn() {
		respBody, _ := apiCtx.GetLastResponseBody()
		apiCtx.Debugger.Print(fmt.Sprintf("%s %s response body (status code: %d):\n\n%s\n", req.Method, req.URL.String(), resp.StatusCode, respBody))
	}

	return nil
}

// IPrepareNewRequestToAndSaveItAs prepares new request and saves it in cache under cacheKey
func (apiCtx *APIContext) IPrepareNewRequestToAndSaveItAs(method, urlTemplate, cacheKey string) error {
	url, err := apiCtx.TemplateEngine.Replace(urlTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'url' template, err: %w", err)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("can't create request due to err: %w", err)
	}

	apiCtx.Cache.Save(cacheKey, req)

	return nil
}

// ISetFollowingHeadersForPreparedRequest sets provided headers for previously prepared request.
// incoming data should be in JSON or YAML format
func (apiCtx *APIContext) ISetFollowingHeadersForPreparedRequest(cacheKey, headersTemplate string) error {
	headers, err := apiCtx.TemplateEngine.Replace(headersTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'headers' template, err: %w", err)
	}

	var headersMap map[string]string
	headersBytes := []byte(headers)
	if format.IsJSON(headersBytes) {
		if err = apiCtx.Formatters.JSON.Deserialize(headersBytes, &headersMap); err != nil {
			return fmt.Errorf("could not deserialize provided headers, err: %w", err)
		}
	} else if format.IsYAML(headersBytes) {
		if err = apiCtx.Formatters.YAML.Deserialize(headersBytes, &headersMap); err != nil {
			return fmt.Errorf("could not deserialize provided headers, err: %w", err)
		}
	} else if format.IsXML(headersBytes) {
		return fmt.Errorf("this method does not support data in format: %s", format.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", format.JSON, format.YAML)
	}

	req, err := apiCtx.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("could not obtain prepared request, err: %w", err)
	}

	for hName, hValue := range headersMap {
		req.Header.Set(hName, hValue)
	}

	apiCtx.Cache.Save(cacheKey, req)

	return nil
}

// ISetFollowingBodyForPreparedRequest sets body for previously prepared request
// bodyTemplate may be in any format and accepts template values
func (apiCtx *APIContext) ISetFollowingBodyForPreparedRequest(cacheKey, bodyTemplate string) error {
	body, err := apiCtx.TemplateEngine.Replace(bodyTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'body' template, err: %w", err)
	}

	req, err := apiCtx.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("could not obtain prepared request, err: %w", err)
	}

	req.Body = ioutil.NopCloser(bytes.NewReader([]byte(body)))
	apiCtx.Cache.Save(cacheKey, req)

	return nil
}

// ISetFollowingCookiesForPreparedRequest sets cookies for previously prepared request.
// cookiesTemplate should be YAML or JSON deserializable on []http.Cookie.
func (apiCtx *APIContext) ISetFollowingCookiesForPreparedRequest(cacheKey, cookiesTemplate string) error {
	var cookies []http.Cookie

	userCookies, err := apiCtx.TemplateEngine.Replace(cookiesTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'cookies' template, err: %w", err)
	}

	req, err := apiCtx.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("could not obtain prepared request, err: %w", err)
	}

	userCookiesBytes := []byte(userCookies)
	if format.IsJSON(userCookiesBytes) {
		if err = apiCtx.Formatters.JSON.Deserialize(userCookiesBytes, &cookies); err != nil {
			return fmt.Errorf("could not deserialize provided cookies, err: %w", err)
		}
	} else if format.IsYAML(userCookiesBytes) {
		if err = apiCtx.Formatters.YAML.Deserialize(userCookiesBytes, &cookies); err != nil {
			return fmt.Errorf("could not deserialize provided cookies, err: %w", err)
		}
	} else if format.IsXML(userCookiesBytes) {
		return fmt.Errorf("this method does not support data in format: %s", format.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", format.JSON, format.YAML)
	}

	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}

	apiCtx.Cache.Save(cacheKey, req)

	return nil
}

/*
	ISetFollowingFormForPreparedRequest sets form for previously prepared request.
	Internally method sets proper Content-Type: multipart/form-data header.
	formTemplate should be YAML or JSON deserializable on map[string]string.
*/
func (apiCtx *APIContext) ISetFollowingFormForPreparedRequest(cacheKey, formTemplate string) error {
	form, err := apiCtx.TemplateEngine.Replace(formTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	req, err := apiCtx.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("could not obtain prepared request, err: %w", err)
	}

	var formKeyVal map[string]string
	formBytes := []byte(form)
	if format.IsJSON(formBytes) {
		err = apiCtx.Formatters.JSON.Deserialize(formBytes, &formKeyVal)
	} else if format.IsYAML(formBytes) {
		err = apiCtx.Formatters.YAML.Deserialize(formBytes, &formKeyVal)
	} else if format.IsXML(formBytes) {
		return fmt.Errorf("this method does not support data in format: %s", format.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", format.JSON, format.YAML)
	}

	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range formKeyVal {
		reference, foundValidReference := apiCtx.fileRecognizer.Recognize(value)
		if foundValidReference {
			if reference.Reference.Type == osutils.ReferenceTypeOSPath {
				file, err := os.Open(reference.Reference.Value)
				if err != nil {
					return fmt.Errorf("could not open file with reference %s, err: %w", reference.Reference.Value, err)
				}

				part, err := writer.CreateFormFile(key, filepath.Base(file.Name()))
				if err != nil {
					return fmt.Errorf("writer could not create form file, err: %w", err)
				}

				_, err = io.Copy(part, file)
				if err != nil {
					return fmt.Errorf("internal problem with copying, err: %w", err)
				}
			}

			continue
		}

		if reference.IsFoundReference() && !foundValidReference {
			return fmt.Errorf("form field '%s' holds invalid reference to file", key)
		}

		fw, err := writer.CreateFormField(key)
		if err != nil {
			return fmt.Errorf("writer could not create form field, err: %w", err)
		}

		_, err = io.Copy(fw, strings.NewReader(value))
		if err != nil {
			return fmt.Errorf("internal problem with copying, err: %w", err)
		}
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("problem with closing writer, err: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Body = ioutil.NopCloser(bytes.NewReader(body.Bytes()))

	apiCtx.Cache.Save(cacheKey, req)

	return nil
}

// ISendRequest sends previously prepared HTTP(s) request.
func (apiCtx *APIContext) ISendRequest(cacheKey string) error {
	req, err := apiCtx.GetPreparedRequest(cacheKey)
	if err != nil {
		return fmt.Errorf("could not obtain prepared request, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		command, _ := http2curl.GetCurlCommand(req)
		apiCtx.Debugger.Print(command.String())
	}

	apiCtx.Cache.Save(httpcache.LastHTTPRequestTimestamp, time.Now())

	resp, err := apiCtx.RequestDoer.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request %s %s, reason: %w", req.Method, req.URL.String(), err)
	}

	apiCtx.Cache.Save(httpcache.LastHTTPResponseTimestamp, time.Now())
	apiCtx.Cache.Save(httpcache.LastHTTPResponseCacheKey, resp)

	if apiCtx.Debugger.IsOn() {
		respBody, _ := apiCtx.GetLastResponseBody()
		apiCtx.Debugger.Print(fmt.Sprintf("%s %s response body (status code: %d):\n\n%s\n", req.Method, req.URL.String(), resp.StatusCode, respBody))
	}

	return nil
}

// TheResponseStatusCodeShouldBe compare last response status code with given in argument.
func (apiCtx *APIContext) TheResponseStatusCodeShouldBe(code int) error {
	lastResponse, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	if lastResponse.StatusCode != code {
		return fmt.Errorf("expected status code %d, but got %d", code, lastResponse.StatusCode)
	}

	return nil
}

// TheResponseBodyShouldHaveFormat checks whether last response body has given data format.
// Available data formats are listed in format package.
func (apiCtx *APIContext) TheResponseBodyShouldHaveFormat(dataFormat format.DataFormat) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	switch dataFormat {
	case format.JSON:
		isJSON := format.IsJSON(body)
		if isJSON {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", format.JSON)
	case format.YAML:
		isYAML := format.IsYAML(body)
		if isYAML {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", format.YAML)
	case format.XML:
		isXml := format.IsXML(body)
		if isXml {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", format.XML)
	//This case checks whether last response body is not any of known types - then it assumes it is plain text
	case format.PlainText:
		if !(format.IsJSON(body) || format.IsYAML(body) || format.IsXML(body)) {
			return nil
		}

		return fmt.Errorf("response body is not plain text, it has one of following formats: %s, %s or %s", format.JSON, format.YAML, format.XML)
	default:
		return fmt.Errorf("unknown last response body data format, available formats: %s, %s, %s, %s",
			format.JSON, format.YAML, format.XML, format.PlainText)
	}
}

// IGenerateARandomIntInTheRangeToAndSaveItAs generates random integer from provided range
// and preserve it under given cacheKey key.
func (apiCtx *APIContext) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, cacheKey string) error {
	randomInteger, err := mathutils.RandomInt(from, to)
	if err != nil {
		return fmt.Errorf("problem during generating pseudo random integer, err: %w", err)
	}

	apiCtx.Cache.Save(cacheKey, randomInteger)

	return nil
}

// IGenerateARandomFloatInTheRangeToAndSaveItAs generates random float from provided range
// and preserve it under given cacheKey key.
func (apiCtx *APIContext) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, cacheKey string) error {
	randInt, err := mathutils.RandomInt(from, to)
	if err != nil {
		return fmt.Errorf("problem during generating pseudo random float, randomInt err: %w", err)
	}

	float01 := rand.Float64()
	strFloat := fmt.Sprintf("%.2f", float01*float64(randInt))
	floatVal, err := strconv.ParseFloat(strFloat, 64)
	if err != nil {
		return fmt.Errorf("problem during generating pseudo random float, parsing err: %w", err)
	}

	apiCtx.Cache.Save(cacheKey, floatVal)

	return nil
}

// IGenerateARandomRunesInTheRangeToAndSaveItAs creates random runes generator func using provided charset
// return func creates runes from provided range and preserve it under given cacheKey
func (apiCtx *APIContext) IGenerateARandomRunesInTheRangeToAndSaveItAs(charset string) func(from, to int, cacheKey string) error {
	return func(from, to int, cacheKey string) error {
		randInt, err := mathutils.RandomInt(from, to)
		if err != nil {
			return fmt.Errorf("problem during generating random runes, randomInt err: %w", err)
		}

		apiCtx.Cache.Save(cacheKey, string(stringutils.RunesFromCharset(randInt, []rune(charset))))

		return nil
	}
}

/*
	IGenerateARandomSentenceInTheRangeFromToWordsAndSaveItAs creates generator func for creating random sentences
	each sentence has length from - to as provided in params and is saved in provided cacheKey
*/
func (apiCtx *APIContext) IGenerateARandomSentenceInTheRangeFromToWordsAndSaveItAs(charset string, wordMinLength, wordMaxLength int) func(from, to int, cacheKey string) error {
	return func(from, to int, cacheKey string) error {
		if from > to {
			return fmt.Errorf("could not generate sentence because of invalid range provided, from '%d' should not be greater than to: '%d'", from, to)
		}

		if wordMinLength > wordMaxLength {
			return fmt.Errorf("could not generate sentence because of invalid range provided, wordMinLength '%d' should not be greater than wordMaxLength '%d'", wordMinLength, wordMaxLength)
		}

		numberOfWords, err := mathutils.RandomInt(from, to)
		if err != nil {
			return fmt.Errorf("problem during generating random sentence, randomInt err: %w", err)
		}

		sentence := ""
		for i := 0; i < numberOfWords; i++ {
			lengthOfWord, err := mathutils.RandomInt(wordMinLength, wordMaxLength)
			if err != nil {
				return fmt.Errorf("problem during generating random sentence, word randomInt err: %w", err)
			}

			word := stringutils.RunesFromCharset(lengthOfWord, []rune(charset))
			if i == numberOfWords-1 {
				sentence += string(word)
			} else {
				sentence += string(word) + " "
			}
		}

		apiCtx.Cache.Save(cacheKey, sentence)

		return nil
	}
}

// IGetTimeAndTravelByAndSaveItAs accepts time object, move timeDuration in time and
// save it in cache under given cacheKey.
func (apiCtx *APIContext) IGetTimeAndTravelByAndSaveItAs(t time.Time, timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error {
	var newTime time.Time
	switch timeDirection {
	case timeutils.TimeDirectionBackward:
		newTime = t.Add(-timeDuration)
	case timeutils.TimeDirectionForward:
		newTime = t.Add(timeDuration)
	default:
		return fmt.Errorf("unknown time direction: %s, allowed: %s, %s", timeDirection, timeutils.TimeDirectionForward, timeutils.TimeDirectionBackward)
	}

	apiCtx.Cache.Save(cacheKey, newTime)

	return nil
}

// IGenerateCurrentTimeAndTravelByAndSaveItAs creates current time object, move timeDuration in time and
// save it in cache under given cacheKey.
func (apiCtx *APIContext) IGenerateCurrentTimeAndTravelByAndSaveItAs(timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error {
	return apiCtx.IGetTimeAndTravelByAndSaveItAs(time.Now(), timeDirection, timeDuration, cacheKey)
}

// TheResponseShouldHaveNode checks whether last response body contains given node.
// expr should be valid according to injected PathFinder for given data format
func (apiCtx *APIContext) TheResponseShouldHaveNode(dataFormat format.DataFormat, exprTemplate string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	switch dataFormat {
	case format.JSON:
		_, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		_, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		_, err = apiCtx.PathFinders.XML.Find(expr, body)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return fmt.Errorf("node '%s', err: %w", expr, err)
	}

	return nil
}

// TheResponseShouldHaveNodes checks whether last request body has keys defined in string separated by comma
// nodeExprs should be valid according to injected PathFinder expressions separated by comma (,)
func (apiCtx *APIContext) TheResponseShouldHaveNodes(dataFormat format.DataFormat, expressionsTemplate string) error {
	expressions, err := apiCtx.TemplateEngine.Replace(expressionsTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	keysSlice := strings.Split(expressions, ",")

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	errs := make([]error, 0, len(keysSlice))
	for _, key := range keysSlice {
		trimmedKey := strings.TrimSpace(key)

		switch dataFormat {
		case format.JSON:
			_, err = apiCtx.PathFinders.JSON.Find(trimmedKey, body)
		case format.YAML:
			_, err = apiCtx.PathFinders.YAML.Find(trimmedKey, body)
		case format.XML:
			_, err = apiCtx.PathFinders.XML.Find(trimmedKey, body)
		default:
			return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
				dataFormat, format.JSON, format.YAML, format.XML)
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("node '%s', err: %s", trimmedKey, err.Error()))
		}
	}

	if len(errs) > 0 {
		var errString string
		for _, err := range errs {
			errString += fmt.Sprintf("%s\n", err)
		}

		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return errors.New(errString)
	}

	return nil
}

// TheNodeShouldNotBe checks whether node from last response body is not of provided type.
// goType may be one of: nil, string, int, float, bool, map, slice,
// expr should be valid according to injected PathFinder for given data format.
func (apiCtx *APIContext) TheNodeShouldNotBe(dataFormat format.DataFormat, exprTemplate string, goType string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	var iNodeVal interface{}
	switch dataFormat {
	case format.JSON:
		iNodeVal, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iNodeVal, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		return fmt.Errorf("this method does not support data in format: %s", format.XML)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		return err
	}

	vNodeVal := reflect.ValueOf(iNodeVal)
	errInvalidType := fmt.Errorf("%s value is \"%s\", but expected not to be", expr, goType)
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
		return fmt.Errorf("%s is not one of available types", goType)
	}

	return nil
}

// TheNodeShouldBe checks whether node from last response body is of provided type
// goType may be one of: nil, string, int, float, bool, map, slice
// expr should be valid according to injected PathResolver
func (apiCtx *APIContext) TheNodeShouldBe(dataFormat format.DataFormat, exprTemplate string, goType string) error {
	switch goType {
	case "nil", "string", "int", "float", "bool", "map", "slice":
		if err := apiCtx.TheNodeShouldNotBe(dataFormat, exprTemplate, goType); err == nil {
			return fmt.Errorf("%s value is not \"%s\", but expected to be", exprTemplate, goType)
		}

		return nil
	default:
		return fmt.Errorf("%s is not one of available types", goType)
	}
}

// TheNodeShouldBeSliceOfLength checks whether given key is slice and has given length
// expr should be valid according to injected PathFinder for provided dataFormat
func (apiCtx *APIContext) TheNodeShouldBeSliceOfLength(dataFormat format.DataFormat, exprTemplate string, length int) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	var iValue interface{}
	switch dataFormat {
	case format.JSON:
		iValue, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iValue, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		iValue, err = apiCtx.PathFinders.XML.Find(expr, body)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return fmt.Errorf("node '%s', err: %s", expr, err.Error())
	}

	v := reflect.ValueOf(iValue)
	if v.Kind() == reflect.Slice {
		if v.Len() != length {
			return fmt.Errorf("%s slice has length: %d, expected: %d", expr, v.Len(), length)
		}

		return nil
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
	}

	return fmt.Errorf("%s does not point at slice(array) in last HTTP(s) response body", expr)
}

// TheNodeShouldBeOfValue compares json node value from expression to expected by user dataValue of given by user dataType
// Available data types are listed in switch section in each case directive.
// expr should be valid according to injected PathFinder for provided dataFormat.
func (apiCtx *APIContext) TheNodeShouldBeOfValue(dataFormat format.DataFormat, exprTemplate, dataType, dataValue string) error {
	nodeValueReplaced, err := apiCtx.TemplateEngine.Replace(dataValue, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'value' template, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("expected value template '%s' was replace to '%s'\nprovided expression template '%s' was replace to '%s'", dataValue, nodeValueReplaced, exprTemplate, expr))
	}
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	var iValue interface{}
	switch dataFormat {
	case format.JSON:
		iValue, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iValue, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		iValue, err = apiCtx.PathFinders.XML.Find(expr, body)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		return fmt.Errorf("node '%s', err: %s", expr, err.Error())
	}

	switch dataType {
	case "string":
		strVal, ok := iValue.(string)
		if !ok {
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		if strVal != nodeValueReplaced {
			return fmt.Errorf("node %s string value: %s is not equal to expected string value: %s", expr, strVal, nodeValueReplaced)
		}
	case "int":
		intNodeValue, err := strconv.Atoi(nodeValueReplaced)
		if err != nil {
			return fmt.Errorf("replaced node %s value %s could not be converted to int", expr, nodeValueReplaced)
		}

		if strVal, ok := iValue.(string); ok {
			if intVal, errAtoi := strconv.Atoi(strVal); errAtoi == nil {
				if intVal != intNodeValue {
					return fmt.Errorf("node %s int value: %d is not equal to expected int value: %d", expr, intVal, intNodeValue)
				}

				return nil
			}
		}

		floatVal, ok := iValue.(float64)
		if !ok {
			uint64Val, ok := iValue.(uint64)
			if !ok {
				return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
			}

			floatVal = float64(uint64Val)
		}

		intVal := int(floatVal)
		if intVal != intNodeValue {
			return fmt.Errorf("node %s int value: %d is not equal to expected int value: %d", expr, intVal, intNodeValue)
		}
	case "float":
		floatNodeValue, err := strconv.ParseFloat(nodeValueReplaced, 64)
		if err != nil {
			return fmt.Errorf("replaced node %s value %s could not be converted to float64", expr, nodeValueReplaced)
		}

		if strVal, ok := iValue.(string); ok {
			if floatVal, errParseFloat := strconv.ParseFloat(strVal, 64); errParseFloat == nil {
				if floatVal != floatNodeValue {
					return fmt.Errorf("node %s float value: %f is not equal to expected int value: %f", expr, floatVal, floatNodeValue)
				}

				return nil
			}
		}

		floatVal, ok := iValue.(float64)
		if !ok {
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		if floatVal != floatNodeValue {
			return fmt.Errorf("node %s float value %f is not equal to expected float value %f", expr, floatVal, floatNodeValue)
		}
	case "bool":
		boolVal, ok := iValue.(bool)
		if !ok {
			strVal, ok := iValue.(string)
			if !ok {
				return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
			}

			boolVal, err = strconv.ParseBool(strVal)
			if err != nil {
				return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
			}
		}

		boolNodeValue, err := strconv.ParseBool(nodeValueReplaced)
		if err != nil {
			return fmt.Errorf("replaced node %s value %s could not be converted to bool", expr, nodeValueReplaced)
		}

		if boolVal != boolNodeValue {
			return fmt.Errorf("node %s bool value %t is not equal to expected bool value %t", expr, boolVal, boolNodeValue)
		}
	}

	return nil
}

// TheNodeShouldMatchRegExp checks whether last response body node matches provided regExp.
func (apiCtx *APIContext) TheNodeShouldMatchRegExp(dataFormat format.DataFormat, exprTemplate, regExpTemplate string) error {
	regExpString, err := apiCtx.TemplateEngine.Replace(regExpTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'regExp' template, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	var iValue interface{}
	switch dataFormat {
	case format.JSON:
		iValue, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iValue, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		iValue, err = apiCtx.PathFinders.XML.Find(expr, body)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		return fmt.Errorf("node '%s', err: %s", expr, err.Error())
	}

	jsonValue, err := apiCtx.Formatters.JSON.Serialize(iValue)
	if err != nil {
		return fmt.Errorf("problem during serialization, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("matching: %s, with regExp: %s", jsonValue, regExpString))
	}

	matched, err := regexp.Match(regExpString, jsonValue)
	if err != nil {
		return fmt.Errorf("problem with regExp matching, err: %w", err)
	}

	if !matched {
		return fmt.Errorf("%s does not contain any: %s", string(jsonValue), regExpString)
	}

	return nil
}

// TheResponseShouldHaveHeader checks whether last HTTP response has given header.
func (apiCtx *APIContext) TheResponseShouldHaveHeader(name string) error {
	defer func() {
		if apiCtx.Debugger.IsOn() {
			lastResp, err := apiCtx.GetLastResponse()
			if err != nil {
				apiCtx.Debugger.Print("could not obtain last response headers")
			}

			apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response headers: %#v", lastResp.Header))
		}
	}()

	lastResp, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	header := lastResp.Header.Get(name)
	if header != "" {
		return nil
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response headers: %#v", lastResp.Header))
	}

	return fmt.Errorf("could not find header %s in last HTTP response", name)
}

// TheResponseShouldHaveHeaderOfValue checks whether last HTTP response has given header with provided valueTemplate.
func (apiCtx *APIContext) TheResponseShouldHaveHeaderOfValue(name, valueTemplate string) error {
	defer func() {
		if apiCtx.Debugger.IsOn() {
			lastResp, err := apiCtx.GetLastResponse()
			if err != nil {
				apiCtx.Debugger.Print("could not obtain last response headers")
			}

			apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response headers: %#v", lastResp.Header))
		}
	}()

	lastResp, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	header := lastResp.Header.Get(name)
	value, err := apiCtx.TemplateEngine.Replace(valueTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'value' template, err: %w", err)
	}

	if header == "" || value == "" {
		return fmt.Errorf("could not find header %s in last HTTP(s) response", name)
	}

	if header == value {
		return nil
	}

	return fmt.Errorf("%s header exists but, expected value: %s, is not equal to actual: %s", name, value, header)
}

// IValidateLastResponseBodyWithSchemaReference validates last response body against schema as provided in referenceTemplate.
// referenceTemplate may be: URL or full/relative path
func (apiCtx *APIContext) IValidateLastResponseBodyWithSchemaReference(referenceTemplate string) error {
	reference, err := apiCtx.TemplateEngine.Replace(referenceTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'reference' template, err: %w", err)
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("'%s' was replaced as: '%s'\n", referenceTemplate, reference))
	}

	return apiCtx.SchemaValidators.ReferenceValidator.Validate(string(body), reference)
}

// IValidateLastResponseBodyWithSchemaString validates last response body against schema.
func (apiCtx *APIContext) IValidateLastResponseBodyWithSchemaString(schema string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	return apiCtx.SchemaValidators.StringValidator.Validate(string(body), schema)
}

// IValidateNodeWithSchemaString validates last response body JSON node against schema
func (apiCtx *APIContext) IValidateNodeWithSchemaString(dataFormat format.DataFormat, exprTemplate, schemaTemplate string) error {
	return apiCtx.iValidateNodeWithSchemaGeneral(dataFormat, exprTemplate, schemaTemplate, apiCtx.SchemaValidators.StringValidator)
}

// IValidateNodeWithSchemaReference validates last response body node against schema as provided in referenceTemplate
func (apiCtx *APIContext) IValidateNodeWithSchemaReference(dataFormat format.DataFormat, exprTemplate, referenceTemplate string) error {
	return apiCtx.iValidateNodeWithSchemaGeneral(dataFormat, exprTemplate, referenceTemplate, apiCtx.SchemaValidators.ReferenceValidator)
}

// TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo asserts that last HTTP request-response time
// is <= than expected timeInterval.
// timeInterval should be string acceptable by time.ParseDuration func
func (apiCtx *APIContext) TimeBetweenLastHTTPRequestResponseShouldBeLessThanOrEqualTo(timeInterval time.Duration) error {
	lastReqTimestampI, err := apiCtx.Cache.GetSaved(httpcache.LastHTTPRequestTimestamp)
	if err != nil {
		return fmt.Errorf("problem during obtaining last HTTP request timestamp, err: %w", err)
	}

	lastReqTimestamp, ok := lastReqTimestampI.(time.Time)
	if !ok {
		return fmt.Errorf("last request timestamp: '%+v' should be time.Time", lastReqTimestampI)
	}

	lastResTimestampI, err := apiCtx.Cache.GetSaved(httpcache.LastHTTPResponseTimestamp)
	if err != nil {
		return fmt.Errorf("problem during obtaining last HTTP response timestamp, err: %w", err)
	}

	lastResTimestamp, ok := lastResTimestampI.(time.Time)
	if !ok {
		return fmt.Errorf("last response timestamp: '%+v' should be time.Time", lastReqTimestampI)
	}

	timeBetweenReqRes := lastResTimestamp.Sub(lastReqTimestamp)
	if timeBetweenReqRes > timeInterval {
		return fmt.Errorf("time between last request - response should be less than %+v, but it took %+v", timeInterval, timeBetweenReqRes)
	}

	return nil
}

// TheResponseShouldHaveCookie checks whether last HTTP(s) response has cookie of given name.
func (apiCtx *APIContext) TheResponseShouldHaveCookie(name string) error {
	lastResp, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	defer func() {
		if apiCtx.Debugger.IsOn() {

			apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response cookies: %+v", lastResp.Cookies()))
		}
	}()

	for _, cookie := range lastResp.Cookies() {
		if cookie.Name == name {
			return nil
		}
	}

	return fmt.Errorf("last HTTP(s) response does not have cookie with name '%s'", name)
}

// TheResponseShouldHaveCookieOfValue checks whether last HTTP(s) response has cookie of given name and value.
func (apiCtx *APIContext) TheResponseShouldHaveCookieOfValue(name, valueTemplate string) error {
	lastResp, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	defer func() {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response cookies: %+v", lastResp.Cookies()))
		}
	}()

	value, err := apiCtx.TemplateEngine.Replace(valueTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'value' template, err: %w", err)
	}

	for _, cookie := range lastResp.Cookies() {
		if cookie.Name == name && cookie.Value == value {
			return nil
		}
	}

	return fmt.Errorf("last HTTP(s) response does not have cookie with name '%s' and value: '%s'", name, value)
}

// ISaveAs saves into cache arbitrary passed data.
func (apiCtx *APIContext) ISaveAs(valueTemplate, cacheKey string) error {
	if len(valueTemplate) == 0 {
		return fmt.Errorf("pass any value")
	}

	if len(cacheKey) == 0 {
		return fmt.Errorf("cacheKey should not be empty value")
	}

	value, err := apiCtx.TemplateEngine.Replace(valueTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'value' template, err: %w", err)
	}

	apiCtx.Cache.Save(cacheKey, value)

	return nil
}

// ISaveFromTheLastResponseNodeAs saves from last response body node under given cacheKey key.
// expr should be valid according to injected PathResolver of given data type
func (apiCtx *APIContext) ISaveFromTheLastResponseNodeAs(dataFormat format.DataFormat, exprTemplate, cacheKey string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	var iVal interface{}
	switch dataFormat {
	case format.JSON:
		iVal, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		iVal, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		iVal, err = apiCtx.PathFinders.XML.Find(expr, body)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return err
	}

	apiCtx.Cache.Save(cacheKey, iVal)

	return nil
}

// IWait waits for given timeInterval amount of time
func (apiCtx *APIContext) IWait(timeInterval time.Duration) error {
	time.Sleep(timeInterval)

	return nil
}

// IPrintLastResponseBody prints last response from request.
func (apiCtx *APIContext) IPrintLastResponseBody() error {
	var tmp interface{}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	defer func() {
		if err != nil {
			apiCtx.Debugger.Print(string(body))
		}
	}()

	if format.IsYAML(body) {
		err = apiCtx.Formatters.YAML.Deserialize(body, &tmp)
		if err != nil {
			return nil
		}

		indentedRespBody, err := yaml.Marshal(tmp)
		if err != nil {
			return nil
		}

		apiCtx.Debugger.Print(string(indentedRespBody))
		return nil
	}

	if format.IsJSON(body) {
		err = apiCtx.Formatters.JSON.Deserialize(body, &tmp)
		if err != nil {
			return nil
		}

		indentedRespBody, err := json.MarshalIndent(tmp, "", "\t")
		if err != nil {
			return nil
		}

		apiCtx.Debugger.Print(string(indentedRespBody))
		return nil
	}

	if format.IsXML(body) {
		err = apiCtx.Formatters.XML.Deserialize(body, &tmp)
		if err != nil {
			return nil
		}

		indentedRespBody, err := xml.MarshalIndent(tmp, "", "\t")
		if err != nil {
			return nil
		}

		apiCtx.Debugger.Print(string(indentedRespBody))
	}

	apiCtx.Debugger.Print(string(body))
	return nil
}

// IStartDebugMode starts debugging mode
func (apiCtx *APIContext) IStartDebugMode() error {
	apiCtx.Debugger.TurnOn()

	return nil
}

// IStopDebugMode stops debugging mode
func (apiCtx *APIContext) IStopDebugMode() error {
	apiCtx.Debugger.TurnOff()

	return nil
}

// GetPreparedRequest returns prepared request from cache or error if failed
func (apiCtx *APIContext) GetPreparedRequest(cacheKey string) (*http.Request, error) {
	reqInterface, err := apiCtx.Cache.GetSaved(cacheKey)
	if err != nil {
		return &http.Request{}, fmt.Errorf("could not obtain %s from cache, err: %w", cacheKey, err)
	}

	req, ok := reqInterface.(*http.Request)
	if !ok {
		return &http.Request{}, fmt.Errorf("value under key %s in cache doesn't contain *http.Request", cacheKey)
	}

	return req, nil
}

// GetLastResponse returns last HTTP(s) response.
func (apiCtx *APIContext) GetLastResponse() (*http.Response, error) {
	respInterface, err := apiCtx.Cache.GetSaved(httpcache.LastHTTPResponseCacheKey)
	if err != nil {
		return nil, fmt.Errorf("missing last HTTP(s) response, err: %s", err.Error())
	}

	lastResp, ok := respInterface.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("last HTTP(s) response data structure is not type *http.Response")
	}

	if lastResp == nil {
		return nil, fmt.Errorf("missing last HTTP(s) response")
	}

	return lastResp, nil
}

// GetLastResponseBody returns last HTTP(s) response body.
// internally method creates new NoPCloser on last response so this method is safe to reuse many times
func (apiCtx *APIContext) GetLastResponseBody() ([]byte, error) {
	lastResponse, err := apiCtx.GetLastResponse()
	if err != nil {
		return []byte(""), fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
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
func (apiCtx *APIContext) iValidateNodeWithSchemaGeneral(dataFormat format.DataFormat, exprTemplate, referenceTemplate string, validator validator.SchemaValidator) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	reference, err := apiCtx.TemplateEngine.Replace(referenceTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'reference' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("'%s' was replaced as: '%s'\n", referenceTemplate, reference))
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	var node interface{}
	switch dataFormat {
	case format.JSON:
		node, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case format.YAML:
		node, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case format.XML:
		return fmt.Errorf("this method does not support data in format: %s", format.XML)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, format.JSON, format.YAML, format.XML)
	}

	if err != nil {
		return err
	}

	jsonNode, err := apiCtx.Formatters.JSON.Serialize(node)
	if err != nil {
		return fmt.Errorf("problem during formatting data to JSON, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("%+v\n\n was marshaled to:\n\n %s \n\nand passed for validation", node, jsonNode))
	}

	return validator.Validate(string(jsonNode), reference)
}
