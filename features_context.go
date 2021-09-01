package gdutils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
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

//ApiFeature struct represents data shared across one feature.
//Field saved holds preserved values,
//Field lastResponse holds last HTTP response,
//Field lastResponseBody holds last HTTP's response body
type ApiFeature struct {
	saved            map[string]interface{}
	lastResponse     *http.Response
	lastResponseBody []byte
	isDebug          bool
}

//ISendRequestToWithBodyAndHeaders sends HTTP request with body and headers.
//Argument method indices HTTP request method for example: "POST", "GET" etc.
//Argument urlTemplate should be full url path. May include template value.
//Argument bodyTemplate is string representing json request body from test suite.
//
//Response and response body will be saved and available in next steps.
func (af *ApiFeature) ISendRequestToWithBodyAndHeaders(method, urlTemplate string, bodyTemplate *godog.DocString) error {
	client := &http.Client{}
	input, err := af.replaceTemplatedValue(bodyTemplate.Content)
	if err != nil {
		return err
	}

	url, err := af.replaceTemplatedValue(urlTemplate)
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

	if af.isDebug {
		command, _ := http2curl.GetCurlCommand(req)
		fmt.Println(command)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	err = af.saveLastResponseCredentials(resp)
	if af.isDebug {
		fmt.Printf("Response body:\n\n")
		_ = af.IPrintLastResponseBody()
		fmt.Printf("\n")
	}

	return err
}

//TheResponseStatusCodeShouldBe compare last response status code with given in argument.
func (af *ApiFeature) TheResponseStatusCodeShouldBe(code int) error {
	if af.lastResponse.StatusCode != code {
		return fmt.Errorf("%w, expected: %d, actual: %d",
			ErrResponseCode, code, af.lastResponse.StatusCode)
	}

	return nil
}

//TheResponseShouldBeInJSON checks if last response body is in JSON format.
func (af *ApiFeature) TheResponseShouldBeInJSON() error {
	var js map[string]interface{}
	var js2 []map[string]interface{}

	if json.Unmarshal(af.lastResponseBody, &js) == nil || json.Unmarshal(af.lastResponseBody, &js2) == nil {
		return nil
	}

	return fmt.Errorf("response has %w", ErrJson)
}

//ISaveFromTheLastResponseJSONNodeAs saves from last response json node under given variableName.
func (af *ApiFeature) ISaveFromTheLastResponseJSONNodeAs(node, variableName string) error {
	iVal, err := qjson.Resolve(node, af.lastResponseBody)

	if err != nil {
		if af.isDebug {
			_ = af.IPrintLastResponseBody()
		}

		return err
	}

	af.Save(variableName, iVal)

	return nil
}

//IGenerateARandomInt generates random integer and preserve it under given name.
//internally it picks random integer from <1, 200000)
func (af *ApiFeature) IGenerateARandomInt(name string) error {
	af.Save(name, randomInt(1, 200000))

	return nil
}

//IGenerateARandomIntInTheRangeToAndSaveItAs generates random integer from provided range and preserve it under given name
func (af *ApiFeature) IGenerateARandomIntInTheRangeToAndSaveItAs(from, to int, name string) error {
	af.Save(name, randomInt(from, to))

	return nil
}

//IGenerateARandomFloat generates random float and preserve it under given name.
//internally it picks random float from <1, 200000) and round it to 2 decimal points
func (af *ApiFeature) IGenerateARandomFloat(name string) error {
	randInt := randomInt(1, 200000)
	float01 := rand.Float64()

	strFloat := fmt.Sprintf("%.2f", float01*float64(randInt))
	floatVal, err := strconv.ParseFloat(strFloat, 64)
	if err != nil {
		return err
	}

	af.Save(name, floatVal)

	return nil
}

//IGenerateARandomFloatInTheRangeToAndSaveItAs generates random float from provided range and preserve it under given name
func (af *ApiFeature) IGenerateARandomFloatInTheRangeToAndSaveItAs(from, to int, name string) error {
	randInt := randomInt(from, to)
	float01 := rand.Float64()

	strFloat := fmt.Sprintf("%.2f", float01*float64(randInt))
	floatVal, err := strconv.ParseFloat(strFloat, 64)
	if err != nil {
		return err
	}

	af.Save(name, floatVal)

	return nil
}

//IGenerateARandomString generates random string and save it under key
func (af *ApiFeature) IGenerateARandomString(key string) error {
	af.Save(key, af.stringWithCharset(15, charsetLettersOnly))

	return nil
}

//TheJSONResponseShouldHaveKey checks whether last response body contains given key
func (af *ApiFeature) TheJSONResponseShouldHaveKey(key string) error {
	_, err := qjson.Resolve(key, af.lastResponseBody)
	if err != nil {
		if af.isDebug {
			_ = af.IPrintLastResponseBody()
		}

		return fmt.Errorf("%v, missing key '%s'", ErrJsonNode, key)
	}

	return nil
}

//TheJSONResponseShouldHaveKeys checks whether last request body has keys defined in string separated by comma
func (af *ApiFeature) TheJSONResponseShouldHaveKeys(keys string) error {
	keysSlice := strings.Split(keys, ",")

	errs := make([]error, 0, len(keysSlice))
	for _, key := range keysSlice {
		trimmedKey := strings.TrimSpace(key)
		_, err := qjson.Resolve(trimmedKey, af.lastResponseBody)

		if err != nil {
			errs = append(errs, fmt.Errorf("missing key %s", trimmedKey))
		}
	}

	if len(errs) > 0 {
		var errString string
		for _, err := range errs {
			errString += fmt.Sprintf("%s\n", err)
		}

		if af.isDebug {
			_ = af.IPrintLastResponseBody()
		}

		return errors.New(errString)
	}

	return nil
}

//IPrintLastResponseBody prints last response from request
func (af *ApiFeature) IPrintLastResponseBody() error {
	var tmp map[string]interface{}
	err := json.Unmarshal(af.lastResponseBody, &tmp)

	if err != nil {
		fmt.Println(string(af.lastResponseBody))
		return nil
	}

	indentedRespBody, err := json.MarshalIndent(tmp, "", "\t")

	if err != nil {
		fmt.Println(string(af.lastResponseBody))
		return nil
	}

	fmt.Println(string(indentedRespBody))
	return nil
}

//TheJSONNodeShouldBeSliceOfLength checks whether given key is slice and has given length
func (af *ApiFeature) TheJSONNodeShouldBeSliceOfLength(expr string, length int) error {
	iValue, err := qjson.Resolve(expr, af.lastResponseBody)
	if err != nil {
		if af.isDebug {
			_ = af.IPrintLastResponseBody()
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
	if af.isDebug {
		_ = af.IPrintLastResponseBody()
	}

	return fmt.Errorf("%s is not slice", expr)
}

//TheJSONNodeShouldBeOfValue compares json node value from expression to expected by user dataValue of given by user dataType
//available data types are listed in switch section in each case directive
func (af *ApiFeature) TheJSONNodeShouldBeOfValue(expr, dataType, dataValue string) error {
	nodeValueReplaced, err := af.replaceTemplatedValue(dataValue)
	if err != nil {
		return err
	}

	if af.isDebug {
		fmt.Printf("Replaced value: %s\n", nodeValueReplaced)
	}

	iValue, err := qjson.Resolve(expr, af.lastResponseBody)
	if err != nil {
		if af.isDebug {
			_ = af.IPrintLastResponseBody()
		}

		return err
	}

	switch dataType {
	case "string":
		strVal, ok := iValue.(string)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		if strVal != nodeValueReplaced {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s string value: %s is not equal to expected string value: %s", expr, strVal, nodeValueReplaced)
		}
	case "int":
		floatVal, ok := iValue.(float64)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		intVal := int(floatVal)

		intNodeValue, err := strconv.Atoi(nodeValueReplaced)

		if err != nil {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to int", expr, nodeValueReplaced)
		}

		if intVal != intNodeValue {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s int value: %d is not equal to expected int value: %d", expr, intVal, intNodeValue)
		}
	case "float":
		floatVal, ok := iValue.(float64)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		floatNodeValue, err := strconv.ParseFloat(nodeValueReplaced, 64)
		if err != nil {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to float64", expr, nodeValueReplaced)
		}

		if floatVal != floatNodeValue {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s float value %f is not equal to expected float value %f", expr, floatVal, floatNodeValue)
		}
	case "bool":
		boolVal, ok := iValue.(bool)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		boolNodeValue, err := strconv.ParseBool(nodeValueReplaced)
		if err != nil {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to bool", expr, nodeValueReplaced)
		}

		if boolVal != boolNodeValue {
			if af.isDebug {
				_ = af.IPrintLastResponseBody()
			}
			return fmt.Errorf("node %s bool value %t is not equal to expected bool value %t", expr, boolVal, boolNodeValue)
		}
	}

	return nil
}

//IWait waits for given timeInterval amount of time
func (af *ApiFeature) IWait(timeInterval string) error {
	duration, err := time.ParseDuration(timeInterval)
	if err != nil {
		return err
	}
	time.Sleep(duration)
	return nil
}

//TheResponseShouldBeInXML checks whether last respone body is in XML format
func (af *ApiFeature) TheResponseShouldBeInXML() error {
	var data interface{}
	return xml.Unmarshal(af.lastResponseBody, &data)
}

//TheJSONNodeShouldNotBe checks whether JSON node from last response body is not of provided type
//goType may be one of: nil, string, int, float, bool, map, slice
//node should be expression acceptable by qjson package against JSON node from last response body
func (af *ApiFeature) TheJSONNodeShouldNotBe(node string, goType string) error {
	iNodeVal, err := qjson.Resolve(node, af.GetLastResponseBody())
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
func (af *ApiFeature) TheJSONNodeShouldBe(node string, goType string) error {
	errInvalidType := fmt.Errorf("%s value is not \"%s\", but expected to be", node, goType)

	switch goType {
	case "nil", "string", "int", "float", "bool", "map", "slice":
		if err := af.TheJSONNodeShouldNotBe(node, goType); err == nil {
			return errInvalidType
		}

		return nil
	default:
		return fmt.Errorf("%s is unknown type for this step", goType)
	}
}
