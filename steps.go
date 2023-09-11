package gdutils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	ch "github.com/pawelWritesCode/charset"
	"github.com/pawelWritesCode/df"
	"moul.io/http2curl/v2"

	"github.com/pawelWritesCode/gdutils/pkg/httpcache"
	"github.com/pawelWritesCode/gdutils/pkg/mathutils"
	"github.com/pawelWritesCode/gdutils/pkg/osutils"
	"github.com/pawelWritesCode/gdutils/pkg/timeutils"
	"github.com/pawelWritesCode/gdutils/pkg/types"
	"github.com/pawelWritesCode/gdutils/pkg/validator"
)

// BodyHeaders is entity that holds information about request body and request headers.
type BodyHeaders struct {

	// Body should contain HTTP(s) request body
	Body any

	// Headers should contain HTTP(s) request headers
	Headers map[string]string
}

/*
		RequestSendWithBodyAndHeaders sends HTTP(s) requests with provided body and headers.

		Argument "method" indices HTTP request method for example: "POST", "GET" etc.
	 	Argument "urlTemplate" should be full valid URL. May include template values.
		Argument "bodyTemplate" should contain data (may include template values)
		in JSON or YAML format with keys "body" and "headers".
*/
func (apiCtx *APIContext) RequestSendWithBodyAndHeaders(method, urlTemplate string, bodyAndHeaderTemplate string) error {
	input, err := apiCtx.TemplateEngine.Replace(bodyAndHeaderTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'headers and body' template, err: %w", err)
	}

	url, err := apiCtx.TemplateEngine.Replace(urlTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'url' template, err: %w", err)
	}

	var bodyAndHeaders BodyHeaders
	var dataFormat df.DataFormat
	if df.IsJSON([]byte(input)) {
		dataFormat = df.JSON
	} else if df.IsYAML([]byte(input)) {
		dataFormat = df.YAML
	} else if df.IsXML([]byte(input)) {
		return fmt.Errorf("this method does not support data in format: %s", df.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", df.JSON, df.YAML)
	}

	switch dataFormat {
	case df.JSON:
		err = apiCtx.Formatters.JSON.Deserialize([]byte(input), &bodyAndHeaders)
	case df.YAML:
		err = apiCtx.Formatters.YAML.Deserialize([]byte(input), &bodyAndHeaders)
	default:
		err = fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", df.JSON, df.YAML)
	}

	if err != nil {
		return fmt.Errorf("can't deserialize 'headers and body' data, err: %w", err)
	}

	var reqBody []byte
	switch dataFormat {
	case df.JSON:
		reqBody, err = apiCtx.Formatters.JSON.Serialize(bodyAndHeaders.Body)
	case df.YAML:
		reqBody, err = apiCtx.Formatters.YAML.Serialize(bodyAndHeaders.Body)
	default:
		err = fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", df.JSON, df.YAML)
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
		apiCtx.Debugger.Print(fmt.Sprintf("%s %s (%d)", req.Method, req.URL.String(), resp.StatusCode))
		apiCtx.Debugger.Print(string(respBody))
	}

	return nil
}

// RequestPrepare prepares new request and saves it in cache under cacheKey
func (apiCtx *APIContext) RequestPrepare(method, urlTemplate, cacheKey string) error {
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

// RequestSetHeaders sets provided headers for previously prepared request.
// incoming data should be in JSON or YAML format
func (apiCtx *APIContext) RequestSetHeaders(cacheKey, headersTemplate string) error {
	headers, err := apiCtx.TemplateEngine.Replace(headersTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'headers' template, err: %w", err)
	}

	var headersMap map[string]string
	headersBytes := []byte(headers)
	if df.IsJSON(headersBytes) {
		if err = apiCtx.Formatters.JSON.Deserialize(headersBytes, &headersMap); err != nil {
			return fmt.Errorf("could not deserialize provided headers, err: %w", err)
		}
	} else if df.IsYAML(headersBytes) {
		if err = apiCtx.Formatters.YAML.Deserialize(headersBytes, &headersMap); err != nil {
			return fmt.Errorf("could not deserialize provided headers, err: %w", err)
		}
	} else if df.IsXML(headersBytes) {
		return fmt.Errorf("this method does not support data in format: %s", df.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", df.JSON, df.YAML)
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

// RequestSetBody sets body for previously prepared request
// bodyTemplate may be in any format and accepts template values
func (apiCtx *APIContext) RequestSetBody(cacheKey, bodyTemplate string) error {
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

// RequestSetCookies sets cookies for previously prepared request.
// cookiesTemplate should be YAML or JSON deserializable on []http.Cookie.
func (apiCtx *APIContext) RequestSetCookies(cacheKey, cookiesTemplate string) error {
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
	if df.IsJSON(userCookiesBytes) {
		if err = apiCtx.Formatters.JSON.Deserialize(userCookiesBytes, &cookies); err != nil {
			return fmt.Errorf("could not deserialize provided cookies, err: %w", err)
		}
	} else if df.IsYAML(userCookiesBytes) {
		if err = apiCtx.Formatters.YAML.Deserialize(userCookiesBytes, &cookies); err != nil {
			return fmt.Errorf("could not deserialize provided cookies, err: %w", err)
		}
	} else if df.IsXML(userCookiesBytes) {
		return fmt.Errorf("this method does not support data in format: %s", df.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", df.JSON, df.YAML)
	}

	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}

	apiCtx.Cache.Save(cacheKey, req)

	return nil
}

/*
RequestSetForm sets form for previously prepared request.
Internally method sets proper Content-Type: multipart/form-data header.
formTemplate should be YAML or JSON deserializable on map[string]string.
*/
func (apiCtx *APIContext) RequestSetForm(cacheKey, formTemplate string) error {
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
	if df.IsJSON(formBytes) {
		err = apiCtx.Formatters.JSON.Deserialize(formBytes, &formKeyVal)
	} else if df.IsYAML(formBytes) {
		err = apiCtx.Formatters.YAML.Deserialize(formBytes, &formKeyVal)
	} else if df.IsXML(formBytes) {
		return fmt.Errorf("this method does not support data in format: %s", df.XML)
	} else {
		return fmt.Errorf("could not recognize data format. Check your data, maybe you have typo somewhere or syntax error. Supported formats are: %s, %s", df.JSON, df.YAML)
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

// RequestSend sends previously prepared HTTP(s) request.
func (apiCtx *APIContext) RequestSend(cacheKey string) error {
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
		apiCtx.Debugger.Print(fmt.Sprintf("%s %s (%d)", req.Method, req.URL.String(), resp.StatusCode))
		apiCtx.Debugger.Print(string(respBody))
	}

	return nil
}

// GenerateRandomInt generates random integer from provided range
// and preserve it under given cacheKey key.
func (apiCtx *APIContext) GenerateRandomInt(from, to int, cacheKey string) error {
	randomInteger, err := mathutils.RandomInt(from, to)
	if err != nil {
		return fmt.Errorf("problem during generating pseudo random integer, err: %w", err)
	}

	apiCtx.Cache.Save(cacheKey, randomInteger)

	return nil
}

// GenerateFloat64 generates random float from provided range
// and preserve it under given cacheKey key.
func (apiCtx *APIContext) GenerateFloat64(from, to float64, cacheKey string) error {
	randFloat, err := mathutils.RandomFloat64(from, to)
	if err != nil {
		return fmt.Errorf("problem during generating pseudo random float, randomFloat err: %w", err)
	}

	apiCtx.Cache.Save(cacheKey, randFloat)
	return nil
}

// GeneratorRandomRunes creates random runes generator func using provided charset
// return func creates runes from provided range and preserve it under given cacheKey
func (apiCtx *APIContext) GeneratorRandomRunes(charset string) func(from, to int, cacheKey string) error {
	return func(from, to int, cacheKey string) error {
		randInt, err := mathutils.RandomInt(from, to)
		if err != nil {
			return fmt.Errorf("problem during generating random runes, randomInt err: %w", err)
		}

		apiCtx.Cache.Save(cacheKey, string(ch.RandomRunes(randInt, []rune(charset))))

		return nil
	}
}

/*
GeneratorRandomSentence creates generator func for creating random sentences
each sentence has length from - to as provided in params and is saved in provided cacheKey
*/
func (apiCtx *APIContext) GeneratorRandomSentence(charset string, wordMinLength, wordMaxLength int) func(from, to int, cacheKey string) error {
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

			word := ch.RandomRunes(lengthOfWord, []rune(charset))
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

// GetTimeAndTravel accepts time object, move timeDuration in time and
// save it in cache under given cacheKey.
func (apiCtx *APIContext) GetTimeAndTravel(t time.Time, timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error {
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

// GenerateTimeAndTravel creates current time object, move timeDuration in time and
// save it in cache under given cacheKey.
func (apiCtx *APIContext) GenerateTimeAndTravel(timeDirection timeutils.TimeDirection, timeDuration time.Duration, cacheKey string) error {
	return apiCtx.GetTimeAndTravel(time.Now(), timeDirection, timeDuration, cacheKey)
}

// AssertStatusCodeIs compare last response status code with given in argument.
func (apiCtx *APIContext) AssertStatusCodeIs(code int) error {
	lastResponse, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	if lastResponse.StatusCode != code {
		return fmt.Errorf("expected status code %d, but got %d", code, lastResponse.StatusCode)
	}

	return nil
}

// AssertStatusCodeIsNot asserts that last response status code is not provided.
func (apiCtx *APIContext) AssertStatusCodeIsNot(code int) error {
	lastResponse, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	if lastResponse.StatusCode != code {
		return nil
	}

	return fmt.Errorf("expected status code different than %d, but got %d", code, lastResponse.StatusCode)
}

// AssertResponseFormatIs checks whether last response body has given data format.
// Available data formats are listed in format package.
func (apiCtx *APIContext) AssertResponseFormatIs(dataFormat df.DataFormat) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	switch dataFormat {
	case df.JSON:
		if df.IsJSON(body) {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", df.JSON)
	case df.YAML:
		if df.IsYAML(body) {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", df.YAML)
	case df.XML:
		if df.IsXML(body) {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", df.XML)
	case df.HTML:
		if df.IsHTML(body) {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", df.HTML)
	case df.PlainText:
		if df.IsPlainText(body) {
			return nil
		}

		return fmt.Errorf("response body doesn't have format %s", df.PlainText)
	default:
		return fmt.Errorf("unknown last response body data format, available formats: %s, %s, %s, %s, %s",
			df.JSON, df.YAML, df.XML, df.HTML, df.PlainText)
	}
}

// AssertResponseFormatIsNot checks whether last response body has not given data format.
// Available data formats are listed in format package.
func (apiCtx *APIContext) AssertResponseFormatIsNot(dataFormat df.DataFormat) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	switch dataFormat {
	case df.JSON:
		if df.IsJSON(body) {
			return fmt.Errorf("response body has format %s", df.JSON)
		}

		return nil
	case df.YAML:
		if df.IsYAML(body) {
			return fmt.Errorf("response body has format %s", df.YAML)
		}

		return nil
	case df.XML:
		if df.IsXML(body) {
			return fmt.Errorf("response body has format %s", df.XML)
		}

		return nil
	case df.HTML:
		if df.IsHTML(body) {
			return fmt.Errorf("response body has format %s", df.HTML)
		}

		return nil
	case df.PlainText:
		if df.IsPlainText(body) {
			return fmt.Errorf("response body has format %s", df.PlainText)
		}

		return nil
	default:
		return fmt.Errorf("unknown last response body data format, available formats: %s, %s, %s, %s, %s",
			df.JSON, df.YAML, df.XML, df.HTML, df.PlainText)
	}
}

// AssertNodeExists checks whether last response body contains given node.
// expr should be valid according to injected PathFinder for given data format
func (apiCtx *APIContext) AssertNodeExists(dataFormat df.DataFormat, exprTemplate string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	_, err = apiCtx.getNode(body, expr, dataFormat, types.Any)
	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return fmt.Errorf("node '%s' could not be found within last response body, reason: %w", expr, err)
	}

	return nil
}

// AssertNodeNotExists checks whether last response body does not contain given node.
// expr should be valid according to injected PathFinder for given data format
func (apiCtx *APIContext) AssertNodeNotExists(dataFormat df.DataFormat, exprTemplate string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	if body == nil || len(body) == 0 {
		return fmt.Errorf("provided nil body")
	}

	_, err = apiCtx.getNode(body, expr, dataFormat, types.Any)
	if err != nil {
		return nil
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
	}

	return fmt.Errorf("%s node '%s' exists", dataFormat, expr)
}

// AssertNodesExist checks whether last request body has keys defined in string separated by comma
// nodeExprs should be valid according to injected PathFinder expressions separated by comma (,)
func (apiCtx *APIContext) AssertNodesExist(dataFormat df.DataFormat, expressionsTemplate string) error {
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

		_, err := apiCtx.getNode(body, trimmedKey, dataFormat, types.Any)
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

// AssertNodeIsType checks whether node from last response body is of provided type.
// available types are listed in types subpackage.
// expr should be valid according to injected PathResolver.
func (apiCtx *APIContext) AssertNodeIsType(dataFormat df.DataFormat, exprTemplate string, inType types.DataType) error {
	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	_, err = apiCtx.getNode(body, expr, dataFormat, inType)
	return err
}

// AssertNodeIsNotType checks whether node from last response body is of provided type.
// available types are listed in types subpackage.
// expr should be valid according to injected PathResolver.
func (apiCtx *APIContext) AssertNodeIsNotType(dataFormat df.DataFormat, exprTemplate string, inType types.DataType) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'form' template, err: %w", err)
	}

	var iNodeVal any
	switch dataFormat {
	case df.JSON:
		iNodeVal, err = apiCtx.PathFinders.JSON.Find(expr, body)
		if err != nil {
			return fmt.Errorf("could not find node using provided expression: '%s', err: %w", expr, err)
		}

		if !(inType.IsValidJSONDataType() || inType.IsValidGoDataType()) {
			return fmt.Errorf("%s is not any of JSON data types and is not any of Go Data types", inType)
		}

		if inType.IsValidJSONDataType() {
			recognizedDataType := apiCtx.TypeMappers.JSON.Map(iNodeVal)
			if recognizedDataType != inType {
				return nil
			}

			return fmt.Errorf("node '%s' has type '%s', but expected not to be", expr, inType)
		}

		recognizedDataType := apiCtx.TypeMappers.GO.Map(iNodeVal)
		if recognizedDataType != inType {
			return nil
		}

		return fmt.Errorf("node '%s' has type '%s', but expected not to be", expr, inType)
	case df.YAML:
		iNodeVal, err = apiCtx.PathFinders.YAML.Find(expr, body)
		if err != nil {
			return fmt.Errorf("could not find node using provided expression: '%s', err: %w", expr, err)
		}

		if !(inType.IsValidYAMLDataType() || inType.IsValidGoDataType()) {
			return fmt.Errorf("%s is not any of YAML data types and is not any of Go Data types", inType)
		}

		if inType.IsValidJSONDataType() {
			recognizedDataType := apiCtx.TypeMappers.YAML.Map(iNodeVal)
			if recognizedDataType != inType {
				return nil
			}

			return fmt.Errorf("node '%s' has type '%s', but expected not to be", expr, inType)
		}

		recognizedDataType := apiCtx.TypeMappers.GO.Map(iNodeVal)
		if recognizedDataType != inType {
			return nil
		}

		return fmt.Errorf("node '%s' has type '%s', but expected not to be", expr, inType)
	case df.XML:
		return fmt.Errorf("this method does not support data in format: %s", df.XML)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s",
			dataFormat, df.JSON, df.YAML)
	}
}

// AssertNodeIsTypeAndValue compares node value from expression to expected by user dataValue of given by user dataType
// Available data types are listed in switch section in each case directive.
// expr should be valid according to injected PathFinder for provided dataFormat.
func (apiCtx *APIContext) AssertNodeIsTypeAndValue(dataFormat df.DataFormat, exprTemplate string, dataType types.DataType, dataValue string) error {
	nodeValueReplaced, err := apiCtx.TemplateEngine.Replace(dataValue, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'value' template, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided expression template '%s' was replace to '%s'", exprTemplate, expr))
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	iValue, err := apiCtx.getNode(body, expr, dataFormat, dataType)
	if err != nil {
		return err
	}

	return assertNodeTypeAndValue(expr, dataType, iValue, nodeValueReplaced)
}

// AssertNodeIsTypeAndHasOneOfValues checks whether node value obtained using exprTemplate matches one of values held by
// valuesTemplates argument. Values should be separated by comma (,) and may contain template values.
func (apiCtx *APIContext) AssertNodeIsTypeAndHasOneOfValues(dataFormat df.DataFormat, exprTemplate string, dataType types.DataType, valuesTemplates string) error {
	values, err := apiCtx.TemplateEngine.Replace(valuesTemplates, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'valuesTemplates' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided values template: '%s' was replace to: '%s'", valuesTemplates, values))
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided expression template: '%s' was replace to: '%s'", exprTemplate, expr))
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	iValue, err := apiCtx.getNode(body, expr, dataFormat, dataType)
	if err != nil {
		return err
	}

	valuesSlice := strings.Split(values, ",")
	valuesSliceTrimmed := make([]string, 0, len(valuesSlice))
	for _, v := range valuesSlice {
		valuesSliceTrimmed = append(valuesSliceTrimmed, strings.TrimSpace(v))
	}

	for _, value := range valuesSliceTrimmed {
		if err = assertNodeTypeAndValue(expr, dataType, iValue, value); err == nil {
			return nil
		}
	}

	return fmt.Errorf("node '%s' doesn't contain any of: %#v", expr, valuesSliceTrimmed)
}

// AssertNodeContainsSubString AsserNodeContainsSubString checks whether value of last HTTP response node, obtained using exprTemplate
// is string type and contains given substring
func (apiCtx *APIContext) AssertNodeContainsSubString(dataFormat df.DataFormat, exprTemplate string, subTemplate string) error {
	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expr' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided expression template: '%s' was replace to: '%s'", exprTemplate, expr))
	}

	sub, err := apiCtx.TemplateEngine.Replace(subTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'sub' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided substring template: '%s' was replace to: '%s'", subTemplate, sub))
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	iValue, err := apiCtx.getNode(body, expr, dataFormat, types.String)
	if err != nil {
		return err
	}

	valueString, ok := iValue.(string)
	if !ok {
		return fmt.Errorf("node '%s' value detected as string, but can't be converted to string, err: %w", expr, err)
	}

	if !strings.Contains(valueString, sub) {
		return fmt.Errorf("node '%s' string value doesn't contain any occurrence of '%s'", expr, sub)
	}

	return nil
}

// AssertNodeNotContainsSubString AsserNodeNotContainsSubString checks whether value of last HTTP response node, obtained using exprTemplate
// is string type and doesn't contain given substring
func (apiCtx *APIContext) AssertNodeNotContainsSubString(dataFormat df.DataFormat, exprTemplate string, subTemplate string) error {
	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expr' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided expression template: '%s' was replace to: '%s'", exprTemplate, expr))
	}

	sub, err := apiCtx.TemplateEngine.Replace(subTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'sub' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided substring template: '%s' was replace to: '%s'", subTemplate, sub))
	}

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	iValue, err := apiCtx.getNode(body, expr, dataFormat, types.String)
	if err != nil {
		return err
	}

	valueString, ok := iValue.(string)
	if !ok {
		return fmt.Errorf("node '%s' value detected as string, but can't be converted to string, err: %w", expr, err)
	}

	if strings.Contains(valueString, sub) {
		return fmt.Errorf("node '%s' string value contain some '%s', but expected not to", expr, sub)
	}

	return nil
}

// AssertNodeSliceLengthIs checks whether given key is slice and has given length
// expr should be valid according to injected PathFinder for provided dataFormat
func (apiCtx *APIContext) AssertNodeSliceLengthIs(dataFormat df.DataFormat, exprTemplate string, length int) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	iValue, err := apiCtx.getNode(body, expr, dataFormat, types.Any)
	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return fmt.Errorf("node '%s', err: %s", expr, err.Error())
	}

	v := reflect.ValueOf(iValue)
	if v.Kind() == reflect.Slice {
		if v.Len() != length {
			return fmt.Errorf("node '%s' contains slice(array) which has length: %d, but expected: %d", expr, v.Len(), length)
		}

		return nil
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
	}

	return fmt.Errorf("%s does not point at slice(array) in last HTTP(s) response body", expr)
}

// AssertNodeSliceLengthIsNot checks whether given key is slice and has not given length
// expr should be valid according to injected PathFinder for provided dataFormat
func (apiCtx *APIContext) AssertNodeSliceLengthIsNot(dataFormat df.DataFormat, exprTemplate string, length int) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	iValue, err := apiCtx.getNode(body, expr, dataFormat, types.Any)
	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
		}

		return fmt.Errorf("node '%s', err: %s", expr, err.Error())
	}

	v := reflect.ValueOf(iValue)
	if v.Kind() == reflect.Slice {
		if v.Len() != length {
			return nil
		}

		return fmt.Errorf("node '%s' contains slice(array) which has length: %d, but expected not to have it", expr, v.Len())
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n%s", body))
	}

	return fmt.Errorf("%s does not point at slice(array) in last HTTP(s) response body", expr)
}

// AssertNodeMatchesRegExp checks whether last response body node matches provided regExp.
func (apiCtx *APIContext) AssertNodeMatchesRegExp(dataFormat df.DataFormat, exprTemplate, regExpTemplate string) error {
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

	iValue, err := apiCtx.getNode(body, expr, dataFormat, types.Any)
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
		return fmt.Errorf("node '%s' does not match regExp: '%s'", expr, regExpString)
	}

	return nil
}

// AssertNodeNotMatchesRegExp checks whether last response body node does not match provided regExp.
func (apiCtx *APIContext) AssertNodeNotMatchesRegExp(dataFormat df.DataFormat, exprTemplate, regExpTemplate string) error {
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

	iValue, err := apiCtx.getNode(body, expr, dataFormat, types.Any)
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
		return nil
	}

	return fmt.Errorf("node '%s' matches regExp: '%s', but expected not to", expr, regExpString)
}

// AssertResponseHeaderExists checks whether last HTTP response has given header.
func (apiCtx *APIContext) AssertResponseHeaderExists(name string) error {
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

	return fmt.Errorf("could not find header '%s' in last HTTP response", name)
}

// AssertResponseHeaderNotExists checks whether last HTTP response does not have given header.
func (apiCtx *APIContext) AssertResponseHeaderNotExists(name string) error {
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
	if header == "" {
		return nil
	}

	return fmt.Errorf("last HTTP(s) response has header '%s', but expected not to", name)
}

// AssertResponseHeaderValueIs checks whether last HTTP response has given header with provided valueTemplate.
func (apiCtx *APIContext) AssertResponseHeaderValueIs(name, valueTemplate string) error {
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

	return fmt.Errorf("last HTTP(s) response contains header '%s', but it's expected value: '%s', is not equal to actual value: '%s'", name, value, header)
}

// AssertResponseMatchesSchemaByReference validates last response body against schema as provided in referenceTemplate.
// referenceTemplate may be: URL or full/relative path
func (apiCtx *APIContext) AssertResponseMatchesSchemaByReference(referenceTemplate string) error {
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

// AssertResponseMatchesSchemaByString validates last response body against schema.
func (apiCtx *APIContext) AssertResponseMatchesSchemaByString(schema string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	return apiCtx.SchemaValidators.StringValidator.Validate(string(body), schema)
}

// AssertNodeMatchesSchemaByString validates last response body JSON node against schema
func (apiCtx *APIContext) AssertNodeMatchesSchemaByString(dataFormat df.DataFormat, exprTemplate, schemaTemplate string) error {
	return apiCtx.iValidateNodeWithSchemaGeneral(dataFormat, exprTemplate, schemaTemplate, apiCtx.SchemaValidators.StringValidator)
}

// AssertNodeMatchesSchemaByReference validates last response body node against schema as provided in referenceTemplate
func (apiCtx *APIContext) AssertNodeMatchesSchemaByReference(dataFormat df.DataFormat, exprTemplate, referenceTemplate string) error {
	return apiCtx.iValidateNodeWithSchemaGeneral(dataFormat, exprTemplate, referenceTemplate, apiCtx.SchemaValidators.ReferenceValidator)
}

// AssertTimeBetweenRequestAndResponseIs asserts that last HTTP request-response time
// is <= than expected timeInterval.
// timeInterval should be string acceptable by time.ParseDuration func
func (apiCtx *APIContext) AssertTimeBetweenRequestAndResponseIs(timeInterval time.Duration) error {
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

// AssertResponseCookieExists checks whether last HTTP(s) response has cookie of given name.
func (apiCtx *APIContext) AssertResponseCookieExists(name string) error {
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

// AssertResponseCookieNotExists checks whether last HTTP(s) response does not have cookie of given name.
func (apiCtx *APIContext) AssertResponseCookieNotExists(name string) error {
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
			return fmt.Errorf("last HTTP(s) response have cookie with name '%s'", name)
		}
	}

	return nil
}

// AssertResponseCookieValueIs checks whether last HTTP(s) response has cookie of given name and value.
func (apiCtx *APIContext) AssertResponseCookieValueIs(name, valueTemplate string) error {
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

// AssertResponseCookieValueMatchesRegExp checks whether last HTTP(s) response has cookie of given name and value matching regExp.
func (apiCtx *APIContext) AssertResponseCookieValueMatchesRegExp(name, regExpTemplate string) error {
	lastResp, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	defer func() {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response cookies: %#v", lastResp.Cookies()))
		}
	}()

	regExp, err := apiCtx.TemplateEngine.Replace(regExpTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'regExp' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided regExp template: '%s' was replace to: '%s'", regExpTemplate, regExp))
	}

	for _, cookie := range lastResp.Cookies() {
		if cookie.Name == name {
			matched, err := regexp.Match(regExp, []byte(cookie.Value))
			if err != nil {
				return fmt.Errorf("problem with regExp matching, err: %w", err)
			}

			if !matched {
				return fmt.Errorf("%s does not contain any: %s", cookie.Value, regExp)
			}

			return nil
		}
	}

	return fmt.Errorf("last HTTP(s) response does not have cookie with name '%s'", name)
}

// AssertResponseCookieValueNotMatchesRegExp checks whether last HTTP(s) response has cookie of given name and value
// is not matching provided regExp.
func (apiCtx *APIContext) AssertResponseCookieValueNotMatchesRegExp(name, regExpTemplate string) error {
	lastResp, err := apiCtx.GetLastResponse()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response, err: %w", err)
	}

	defer func() {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last HTTP(s) response cookies: %#v", lastResp.Cookies()))
		}
	}()

	regExp, err := apiCtx.TemplateEngine.Replace(regExpTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'regExp' template, err: %w", err)
	}

	if apiCtx.Debugger.IsOn() {
		apiCtx.Debugger.Print(fmt.Sprintf("provided regExp template: '%s' was replace to: '%s'", regExpTemplate, regExp))
	}

	for _, cookie := range lastResp.Cookies() {
		if cookie.Name == name {
			matched, err := regexp.Match(regExp, []byte(cookie.Value))
			if err != nil {
				return fmt.Errorf("problem with regExp matching, err: %w", err)
			}

			if !matched {
				return nil
			}

			return fmt.Errorf("%s contains some: %s, but expected not to", cookie.Value, regExp)
		}
	}

	return fmt.Errorf("last HTTP(s) response does not have cookie with name '%s'", name)
}

// Save saves into cache arbitrary passed data.
func (apiCtx *APIContext) Save(valueTemplate, cacheKey string) error {
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

// SaveNode saves from last response body node under given cache key.
// expr should be valid according to injected PathResolver of given data type
func (apiCtx *APIContext) SaveNode(dataFormat df.DataFormat, exprTemplate, cacheKey string) error {
	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	expr, err := apiCtx.TemplateEngine.Replace(exprTemplate, apiCtx.Cache.All())
	if err != nil {
		return fmt.Errorf("template engine has problem with 'expression' template, err: %w", err)
	}

	iVal, err := apiCtx.getNode(body, expr, dataFormat, types.Any)
	if err != nil {
		if apiCtx.Debugger.IsOn() {
			apiCtx.Debugger.Print(fmt.Sprintf("last response body:\n\n"))
			apiCtx.Debugger.Print(string(body))
		}

		return err
	}

	apiCtx.Cache.Save(cacheKey, iVal)

	return nil
}

// SaveHeader saves from last response header value under given cache key
func (apiCtx *APIContext) SaveHeader(name, cacheKey string) error {
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
	if header == "" {
		return fmt.Errorf("could not find header %s in last HTTP(s) response", name)
	}

	apiCtx.Cache.Save(cacheKey, header)

	return nil
}

// Wait waits for given timeInterval amount of time
func (apiCtx *APIContext) Wait(timeInterval time.Duration) error {
	time.Sleep(timeInterval)

	return nil
}

// DebugPrintResponseBody prints last response from request.
func (apiCtx *APIContext) DebugPrintResponseBody() error {
	var tmp any

	body, err := apiCtx.GetLastResponseBody()
	if err != nil {
		return fmt.Errorf("could not obtain last HTTP(s) response body, err: %w", err)
	}

	defer func() {
		if err != nil {
			apiCtx.Debugger.Print(string(body))
		}
	}()

	if df.IsYAML(body) {
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

	if df.IsJSON(body) {
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

	if df.IsXML(body) {
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

// DebugStart starts debugging mode
func (apiCtx *APIContext) DebugStart() error {
	apiCtx.Debugger.TurnOn()

	return nil
}

// DebugStop stops debugging mode
func (apiCtx *APIContext) DebugStop() error {
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
func (apiCtx *APIContext) iValidateNodeWithSchemaGeneral(dataFormat df.DataFormat, exprTemplate, referenceTemplate string, validator validator.SchemaValidator) error {
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

	var node any
	switch dataFormat {
	case df.JSON:
		node, err = apiCtx.PathFinders.JSON.Find(expr, body)
	case df.YAML:
		node, err = apiCtx.PathFinders.YAML.Find(expr, body)
	case df.XML:
		return fmt.Errorf("this method does not support data in format: %s", df.XML)
	default:
		return fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, df.JSON, df.YAML, df.XML)
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

// getNode returns node value, when dataFormat and dataType matches expectations.
func (apiCtx *APIContext) getNode(body []byte, expr string, dataFormat df.DataFormat, dataType types.DataType) (any, error) {
	if body == nil || len(body) == 0 {
		return nil, fmt.Errorf("provided nil body")
	}

	var iValue any
	var err error
	switch dataFormat {
	case df.JSON:
		iValue, err = apiCtx.PathFinders.JSON.Find(expr, body)
		if err != nil {
			return nil, fmt.Errorf("could not find node using provided expression: '%s', err: %w", expr, err)
		}

		if dataType == types.Any {
			return iValue, nil
		}

		if !(dataType.IsValidJSONDataType() || dataType.IsValidGoDataType()) {
			return nil, fmt.Errorf("%s is not any of JSON data types and is not any of Go Data types", dataType)
		}

		recognizedDataType := apiCtx.TypeMappers.JSON.Map(iValue)
		if recognizedDataType != dataType {
			tmpRecognizedDataType := recognizedDataType
			recognizedDataType = apiCtx.TypeMappers.GO.Map(iValue)

			if recognizedDataType != dataType {
				return nil, fmt.Errorf("expected node '%s' to be '%s', but node value is detected as '%s' in terms of JSON and '%s' in terms of Go",
					expr, dataType, tmpRecognizedDataType, recognizedDataType)
			}
		}
	case df.YAML:
		iValue, err = apiCtx.PathFinders.YAML.Find(expr, body)
		if err != nil {
			return nil, fmt.Errorf("could not find node using provided expression: '%s', err: %w", expr, err)
		}

		if dataType == types.Any {
			return iValue, nil
		}

		if !(dataType.IsValidYAMLDataType() || dataType.IsValidGoDataType()) {
			return nil, fmt.Errorf("%s is not any of YAML data types and is not any of Go Data types", dataType)
		}

		recognizedDataType := apiCtx.TypeMappers.YAML.Map(iValue)
		if recognizedDataType != dataType {
			tmpRecognizedDataType := recognizedDataType
			recognizedDataType = apiCtx.TypeMappers.GO.Map(iValue)

			if recognizedDataType != dataType {
				return nil, fmt.Errorf("expected node '%s' to be '%s', but node value is detected as '%s' in terms of YAML and '%s' in terms of Go",
					expr, dataType, tmpRecognizedDataType, recognizedDataType)
			}
		}
	case df.XML:
		iValue, err = apiCtx.PathFinders.XML.Find(expr, body)
		if err != nil {
			return nil, fmt.Errorf("node '%s', err: %s", expr, err.Error())
		}

		if dataType == types.Any {
			return iValue, nil
		}

		if !(dataType.IsValidXMLDataType() || dataType.IsValidGoDataType()) {
			return nil, fmt.Errorf("%s is not any of XML data types and is not any of Go Data types", dataType)
		}
	case df.HTML:
		iValue, err = apiCtx.PathFinders.HTML.Find(expr, body)
		if err != nil {
			return nil, fmt.Errorf("node '%s', err: %s", expr, err.Error())
		}

		if dataType == types.Any {
			return iValue, nil
		}

		if !(dataType.IsValidGoDataType()) {
			return nil, fmt.Errorf("%s is not any of Go Data types", dataType)
		}
	default:
		return nil, fmt.Errorf("provided unknown format: %s, format should be one of : %s, %s, %s",
			dataFormat, df.JSON, df.YAML, df.XML)
	}

	return iValue, nil
}

// TODO: Use following methods to create XML type checker. It is required because X-PATH engine returns everything as string
// TODO: XML types are listed here: https://www.w3.org/TR/xmlschema-2
// checkBool contains algorithm to compare nodeValue obtained from expr with expectedValue.
// internally it expects nodeValue to contain representation of bool data type.
func checkBool(nodeValue any, expectedValue, expr string) error {
	var err error
	boolVal, ok := nodeValue.(bool)
	if !ok {
		strVal, ok := nodeValue.(string)
		if !ok {
			return fmt.Errorf("expected '%s' to be '%s', got '%v'", expr, "bool", nodeValue)
		}

		boolVal, err = strconv.ParseBool(strVal)
		if err != nil {
			return fmt.Errorf("expected '%s' to be '%s', got '%v'", expr, "bool", nodeValue)
		}
	}

	boolNodeValue, err := strconv.ParseBool(expectedValue)
	if err != nil {
		return fmt.Errorf("node '%s' value '%s' could not be converted to bool", expr, expectedValue)
	}

	if boolVal != boolNodeValue {
		return fmt.Errorf("node '%s' has bool value '%t', but expected '%t'", expr, boolVal, boolNodeValue)
	}

	return nil
}

// checkString contains algorithm to compare nodeValue obtained from expr with expectedValue.
// internally it expects nodeValue to contain representation of string data type.
func checkString(nodeValue any, expectedValue, expr string) error {
	strVal, ok := nodeValue.(string)
	if !ok {
		return fmt.Errorf("expected '%s' to be '%s', got '%v'", expr, "string", nodeValue)
	}

	if strVal != expectedValue {
		return fmt.Errorf("node '%s' has string value: '%s', but expected: '%s'", expr, strVal, expectedValue)
	}

	return nil
}

// checkInt contains algorithm to compare nodeValue obtained from expr with expectedValue.
// internally it expects nodeValue to contain representation of int data type.
func checkInt(nodeValue any, expectedValue, expr string) error {
	intNodeValue, err := strconv.Atoi(expectedValue)
	if err != nil {
		return fmt.Errorf("node '%s' value '%s' could not be converted to int", expr, expectedValue)
	}

	if strVal, ok := nodeValue.(string); ok {
		if intVal, errAtoi := strconv.Atoi(strVal); errAtoi == nil {
			if intVal != intNodeValue {
				return fmt.Errorf("node '%s' has int value: '%d', but expected: '%d'", expr, intVal, intNodeValue)
			}

			return nil
		}
	}

	floatVal, ok := nodeValue.(float64)
	if !ok {
		uint64Val, ok := nodeValue.(uint64)
		if !ok {
			return fmt.Errorf("expected node '%s' to be %s, got %v", expr, "int", nodeValue)
		}

		floatVal = float64(uint64Val)
	}

	intVal := int(floatVal)
	if intVal != intNodeValue {
		return fmt.Errorf("node '%s' has int value: '%d', but expected: '%d'", expr, intVal, intNodeValue)
	}

	return nil
}

// checkFloat contains algorithm to compare nodeValue obtained from expr with expectedValue.
// internally it expects nodeValue to contain representation of float data type.
func checkFloat(nodeValue any, expectedValue, expr string) error {
	floatNodeValue, err := strconv.ParseFloat(expectedValue, 64)
	if err != nil {
		return fmt.Errorf("node '%s' value '%s' could not be converted to float64", expr, expectedValue)
	}

	if strVal, ok := nodeValue.(string); ok {
		if floatVal, errParseFloat := strconv.ParseFloat(strVal, 64); errParseFloat == nil {
			if floatVal != floatNodeValue {
				return fmt.Errorf("node '%s' has float value: '%f', but expected: '%f'", expr, floatVal, floatNodeValue)
			}

			return nil
		}
	}

	floatVal, ok := nodeValue.(float64)
	if !ok {
		return fmt.Errorf("expected '%s' to be '%s', got '%v'", expr, "float", nodeValue)
	}

	if floatVal != floatNodeValue {
		return fmt.Errorf("node '%s' has float value: '%f', but expected '%f'", expr, floatVal, floatNodeValue)
	}

	return nil
}

// assertNodeTypeAndValue checks node value against given dataType and its expectedValue yet represented as string
func assertNodeTypeAndValue(expr string, dataType types.DataType, nodeValue any, expectedValue string) error {
	switch dataType {
	case types.String:
		if err := checkString(nodeValue, expectedValue, expr); err != nil {
			return err
		}
	case types.Int, types.Integer:
		if err := checkInt(nodeValue, expectedValue, expr); err != nil {
			return err
		}
	case types.Float:
		if err := checkFloat(nodeValue, expectedValue, expr); err != nil {
			return err
		}
	case types.Number:
		if errInt, errFloat := checkInt(nodeValue, expectedValue, expr), checkFloat(nodeValue, expectedValue, expr); errInt != nil && errFloat != nil {
			return fmt.Errorf("intErr: %s\n, floatErr: %s", errInt, errFloat)
		}
	case types.Bool, types.Boolean:
		if err := checkBool(nodeValue, expectedValue, expr); err != nil {
			return err
		}
	case types.Scalar:
		if errBool, errString, errInt, errFloat := checkBool(nodeValue, expectedValue, expr), checkString(nodeValue, expectedValue, expr), checkInt(nodeValue, expectedValue, expr), checkFloat(nodeValue, expectedValue, expr); errBool != nil && errString != nil && errInt != nil && errFloat != nil {
			return fmt.Errorf("boolErr: %s\n, stringErr: %s\n, intErr: %s\n, floatErr: %s", errBool, errString, errInt, errFloat)
		}
	default:
		return fmt.Errorf("provided data type: '%s' can't be proceeded", dataType)
	}

	return nil
}
