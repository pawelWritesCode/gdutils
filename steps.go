package gdutils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/moul/http2curl"
	"github.com/pawelWritesCode/qjson"
)

const (
	typeJSON = "JSON"
)

//bodyHeaders is entity that holds information about request body and request headers
type bodyHeaders struct {
	Body    interface{}
	Headers map[string]string
}

//ISendRequestToWithBodyAndHeaders sends HTTP request with provided body and headers.
//Argument method indices HTTP request method for example: "POST", "GET" etc.
//Argument urlTemplate should be full url path. May include template values.
//Argument bodyTemplate should be slice of bytes marshallable on bodyHeaders struct
func (s *Scenario) ISendRequestToWithBodyAndHeaders(method, urlTemplate string, bodyTemplate *godog.DocString) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	input, err := s.replaceTemplatedValue(bodyTemplate.Content)
	if err != nil {
		return err
	}

	url, err := s.replaceTemplatedValue(urlTemplate)
	if err != nil {
		return err
	}

	var bodyAndHeaders bodyHeaders
	err = json.Unmarshal([]byte(input), &bodyAndHeaders)
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(bodyAndHeaders.Body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	for headerName, headerValue := range bodyAndHeaders.Headers {
		req.Header.Set(headerName, headerValue)
	}

	if s.isDebug {
		command, _ := http2curl.GetCurlCommand(req)
		fmt.Println(command)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	s.lastResponse = resp
	//err = s.saveLastResponseCredentials(resp)
	if s.isDebug {
		fmt.Printf("Response body:\n\n")
		_ = s.IPrintLastResponseBody()
		fmt.Printf("\n")
	}

	return err
}

//TheResponseStatusCodeShouldBe compare last response status code with given in argument.
func (s *Scenario) TheResponseStatusCodeShouldBe(code uint16) error {
	if s.lastResponse.StatusCode != int(code) {
		return fmt.Errorf("%w, expected: %d, actual: %d",
			ErrResponseCode, code, s.lastResponse.StatusCode)
	}

	return nil
}

//TheResponseShouldBeIn checks whether last response body has given data type
//available data types are listed in switch section
func (s *Scenario) TheResponseShouldBeIn(dataType string) error {
	switch dataType {
	case typeJSON:
		return s.theResponseShouldBeInJSON()
	default:
		return fmt.Errorf("unknown data type, available values: %s", typeJSON)
	}
}

//ISaveFromTheLastResponseJSONNodeAs saves from last response json node under given variableName.
func (s *Scenario) ISaveFromTheLastResponseJSONNodeAs(node, variableName string) error {
	iVal, err := qjson.Resolve(node, s.GetLastResponseBody())

	if err != nil {
		if s.isDebug {
			_ = s.IPrintLastResponseBody()
		}

		return err
	}

	s.Save(variableName, iVal)

	return nil
}

//IGenerateARandomIntInTheRangeToAndSaveItAs generates random integer from provided range and preserve it under given name in cache
func (s *Scenario) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, name string) error {
	s.Save(name, randomInt(from, to))

	return nil
}

//IGenerateARandomFloatInTheRangeToAndSaveItAs generates random float from provided range and preserve it under given name in cache
func (s *Scenario) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, name string) error {
	randInt := randomInt(from, to)
	float01 := rand.Float64()

	strFloat := fmt.Sprintf("%.2f", float01*float64(randInt))
	floatVal, err := strconv.ParseFloat(strFloat, 64)
	if err != nil {
		return err
	}

	s.Save(name, floatVal)

	return nil
}

//IGenerateARandomStringOfLengthWithoutUnicodeCharactersAndSaveItAs generates random string of given length without unicode characters
func (s *Scenario) IGenerateARandomStringOfLengthWithoutUnicodeCharactersAndSaveItAs(strLength int, key string) error {
	if strLength <= 0 {
		return fmt.Errorf("provided string length %d can't be less than 1", strLength)
	}

	s.Save(key, s.stringWithCharset(strLength, charsetLettersOnly))

	return nil
}

//IGenerateARandomStringOfLengthWithUnicodeCharactersAndSaveItAs generates random string of given length with unicode characters
func (s *Scenario) IGenerateARandomStringOfLengthWithUnicodeCharactersAndSaveItAs(strLength int, key string) error {
	if strLength <= 0 {
		return fmt.Errorf("provided string length %d can't be less than 1", strLength)
	}

	s.Save(key, s.stringWithCharset(strLength, charsetUnicodeCharacters))

	return nil
}

//TheJSONResponseShouldHaveKey checks whether last response body contains given key
func (s *Scenario) TheJSONResponseShouldHaveKey(key string) error {
	_, err := qjson.Resolve(key, s.GetLastResponseBody())
	if err != nil {
		if s.isDebug {
			_ = s.IPrintLastResponseBody()
		}

		return fmt.Errorf("%v, missing key '%s'", ErrJsonNode, key)
	}

	return nil
}

//TheJSONNodeShouldNotBe checks whether JSON node from last response body is not of provided type
//goType may be one of: nil, string, int, float, bool, map, slice
//node should be expression acceptable by qjson package against JSON node from last response body
func (s *Scenario) TheJSONNodeShouldNotBe(node string, goType string) error {
	iNodeVal, err := qjson.Resolve(node, s.GetLastResponseBody())
	if err != nil {
		return err
	}

	vNodeVal := reflect.ValueOf(iNodeVal)
	errInvalidType := fmt.Errorf("%s value is \"%s\", but expected not to be", node, goType)
	switch goType {
	case "nil":
		if !vNodeVal.IsValid() || valueIsNil(vNodeVal) {
			return errInvalidType
		}

		return nil
	case "string":
		if vNodeVal.Kind() == reflect.String {
			return errInvalidType
		}

		return nil
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

		return nil
	case "float":
		if vNodeVal.Kind() == reflect.Float64 || vNodeVal.Kind() == reflect.Float32 {
			_, frac := math.Modf(vNodeVal.Float())
			if frac == 0 {
				return nil
			}

			return errInvalidType
		}

		return nil
	case "bool":
		if vNodeVal.Kind() == reflect.Bool {
			return errInvalidType
		}

		return nil
	case "map":
		if vNodeVal.Kind() == reflect.Map {
			return errInvalidType
		}

		return nil
	case "slice":
		if vNodeVal.Kind() == reflect.Slice {
			return errInvalidType
		}

		return nil
	default:
		return fmt.Errorf("%s is unknown type for this step", goType)
	}
}

//TheJSONNodeShouldBe checks whether JSON node from last response body is of provided type
//goType may be one of: nil, string, int, float, bool, map, slice
//node should be expression acceptable by qjson package against JSON node from last response body
func (s *Scenario) TheJSONNodeShouldBe(node string, goType string) error {
	errInvalidType := fmt.Errorf("%s value is not \"%s\", but expected to be", node, goType)

	switch goType {
	case "nil", "string", "int", "float", "bool", "map", "slice":
		if err := s.TheJSONNodeShouldNotBe(node, goType); err == nil {
			return errInvalidType
		}

		return nil
	default:
		return fmt.Errorf("%s is unknown type for this step", goType)
	}
}

//TheJSONResponseShouldHaveKeys checks whether last request body has keys defined in string separated by comma
func (s *Scenario) TheJSONResponseShouldHaveKeys(keys string) error {
	keysSlice := strings.Split(keys, ",")

	errs := make([]error, 0, len(keysSlice))
	for _, key := range keysSlice {
		trimmedKey := strings.TrimSpace(key)
		_, err := qjson.Resolve(trimmedKey, s.GetLastResponseBody())

		if err != nil {
			errs = append(errs, fmt.Errorf("missing key %s", trimmedKey))
		}
	}

	if len(errs) > 0 {
		var errString string
		for _, err := range errs {
			errString += fmt.Sprintf("%s\n", err)
		}

		if s.isDebug {
			_ = s.IPrintLastResponseBody()
		}

		return errors.New(errString)
	}

	return nil
}

//IPrintLastResponseBody prints last response from request
func (s *Scenario) IPrintLastResponseBody() error {
	var tmp map[string]interface{}
	err := json.Unmarshal(s.GetLastResponseBody(), &tmp)

	if err != nil {
		fmt.Println(string(s.GetLastResponseBody()))
		return nil
	}

	indentedRespBody, err := json.MarshalIndent(tmp, "", "\t")

	if err != nil {
		fmt.Println(string(s.GetLastResponseBody()))
		return nil
	}

	fmt.Println(string(indentedRespBody))
	return nil
}

//TheJSONNodeShouldBeSliceOfLength checks whether given key is slice and has given length
func (s *Scenario) TheJSONNodeShouldBeSliceOfLength(expr string, length int) error {
	iValue, err := qjson.Resolve(expr, s.GetLastResponseBody())
	if err != nil {
		if s.isDebug {
			_ = s.IPrintLastResponseBody()
		}

		return err
	}

	v := reflect.ValueOf(iValue)
	if v.Kind() == reflect.Slice {
		if v.Len() != length {
			return fmt.Errorf("%s slice has length: %d, expected: %d", expr, v.Len(), length)
		}

		return nil
	}
	if s.isDebug {
		_ = s.IPrintLastResponseBody()
	}

	return fmt.Errorf("%s is not slice", expr)
}

//TheJSONNodeShouldBeOfValue compares json node value from expression to expected by user dataValue of given by user dataType
//available data types are listed in switch section in each case directive
func (s *Scenario) TheJSONNodeShouldBeOfValue(expr, dataType, dataValue string) error {
	nodeValueReplaced, err := s.replaceTemplatedValue(dataValue)
	if err != nil {
		return err
	}

	if s.isDebug {
		fmt.Printf("Replaced value: %s\n", nodeValueReplaced)
	}

	iValue, err := qjson.Resolve(expr, s.GetLastResponseBody())
	if err != nil {
		if s.isDebug {
			_ = s.IPrintLastResponseBody()
		}

		return err
	}

	switch dataType {
	case "string":
		strVal, ok := iValue.(string)
		if !ok {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		if strVal != nodeValueReplaced {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s string value: %s is not equal to expected string value: %s", expr, strVal, nodeValueReplaced)
		}
	case "int":
		floatVal, ok := iValue.(float64)
		if !ok {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		intVal := int(floatVal)

		intNodeValue, err := strconv.Atoi(nodeValueReplaced)

		if err != nil {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to int", expr, nodeValueReplaced)
		}

		if intVal != intNodeValue {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s int value: %d is not equal to expected int value: %d", expr, intVal, intNodeValue)
		}
	case "float":
		floatVal, ok := iValue.(float64)
		if !ok {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		floatNodeValue, err := strconv.ParseFloat(nodeValueReplaced, 64)
		if err != nil {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to float64", expr, nodeValueReplaced)
		}

		if floatVal != floatNodeValue {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s float value %f is not equal to expected float value %f", expr, floatVal, floatNodeValue)
		}
	case "bool":
		boolVal, ok := iValue.(bool)
		if !ok {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		boolNodeValue, err := strconv.ParseBool(nodeValueReplaced)
		if err != nil {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to bool", expr, nodeValueReplaced)
		}

		if boolVal != boolNodeValue {
			if s.isDebug {
				_ = s.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s bool value %t is not equal to expected bool value %t", expr, boolVal, boolNodeValue)
		}
	}

	return nil
}

//IWait waits for given timeInterval amount of time
//timeInterval should be string valid for time.ParseDuration func
func (s *Scenario) IWait(timeInterval string) error {
	duration, err := time.ParseDuration(timeInterval)
	if err != nil {
		return err
	}
	time.Sleep(duration)
	return nil
}

// TheResponseShouldHaveHeader checks whether last HTTP response has given header
func (s *Scenario) TheResponseShouldHaveHeader(name string) error {
	headers := s.lastResponse.Header

	header := headers.Get(name)
	if header != "" {
		return nil
	}

	if s.isDebug {
		fmt.Printf("last HTTP response headers: %+v", headers)
	}

	return fmt.Errorf("could not find header %s in last HTTP response", name)
}

// TheResponseShouldHaveHeaderOfValue checks whether last HTTP response has given header with provided value
func (s *Scenario) TheResponseShouldHaveHeaderOfValue(name, value string) error {
	headers := s.lastResponse.Header

	header := headers.Get(name)

	if header == "" && value == "" {
		return fmt.Errorf("could not find header %s", name)
	}

	if header == value {
		return nil
	}

	if s.isDebug {
		fmt.Printf("last HTTP response headers: %+v", headers)
	}

	return fmt.Errorf("could not find header %s in last HTTP response", name)
}
