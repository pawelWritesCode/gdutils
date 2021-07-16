package gdutils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
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

//ISendRequestToWithData sends HTTP request with body.
//Argument method indices HTTP request method for example: "POST", "GET" etc.
//Argument urlTemplate should be full url path. May include template value.
//Argument bodyTemplate is string representing json request body from test suite.
//
//Response and response body will be saved and available in next steps.
func (af *ApiFeature) ISendRequestToWithData(method, urlTemplate string, bodyTemplate *godog.DocString) error {
	client := &http.Client{}
	reqBody, err := af.replaceTemplatedValue(bodyTemplate.Content)

	if err != nil {
		return err
	}

	url, err := af.replaceTemplatedValue(urlTemplate)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(reqBody)))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	return af.saveLastResponseCredentials(resp)
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

	if af.isDebug {
		fmt.Printf("URL: %s\n", url)
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

	if af.isDebug {
		fmt.Printf("Request body:\n\n %s\n\n", string(reqBody))
		fmt.Printf("Request headers: %v\n", bodyAndHeaders.Headers)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	for headerName, headerValue := range bodyAndHeaders.Headers {
		req.Header.Set(headerName, headerValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	err = af.saveLastResponseCredentials(resp)
	if af.isDebug {
		fmt.Printf("Response body:\n\n")
		_ = af.IPrintLastResponse()
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
			_ = af.IPrintLastResponse()
		}

		return err
	}

	af.Save(variableName, iVal)

	return nil
}

//IGenerateARandomInt generates random integer and preserve it under given name.
func (af *ApiFeature) IGenerateARandomInt(name string) error {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 200000
	af.Save(name, rand.Intn(max-min+1)+min)

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
			_ = af.IPrintLastResponse()
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
			_ = af.IPrintLastResponse()
		}

		return errors.New(errString)
	}

	return nil
}

//IPrintLastResponse prints last response from request
func (af *ApiFeature) IPrintLastResponse() error {
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
			_ = af.IPrintLastResponse()
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
		_ = af.IPrintLastResponse()
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
			_ = af.IPrintLastResponse()
		}

		return err
	}

	switch dataType {
	case "string":
		strVal, ok := iValue.(string)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		if strVal != nodeValueReplaced {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("node %s string value: %s is not equal to expected string value: %s", expr, strVal, nodeValueReplaced)
		}
	case "int":
		floatVal, ok := iValue.(float64)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		intVal := int(floatVal)

		intNodeValue, err := strconv.Atoi(nodeValueReplaced)

		if err != nil {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to int", expr, nodeValueReplaced)
		}

		if intVal != intNodeValue {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("node %s int value: %d is not equal to expected int value: %d", expr, intVal, intNodeValue)
		}
	case "float":
		floatVal, ok := iValue.(float64)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		floatNodeValue, err := strconv.ParseFloat(nodeValueReplaced, 64)
		if err != nil {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to float64", expr, nodeValueReplaced)
		}

		if floatVal != floatNodeValue {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("node %s float value %f is not equal to expected float value %f", expr, floatVal, floatNodeValue)
		}
	case "bool":
		boolVal, ok := iValue.(bool)
		if !ok {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("expected %s to be %s, got %v", expr, dataType, iValue)
		}

		boolNodeValue, err := strconv.ParseBool(nodeValueReplaced)
		if err != nil {
			if af.isDebug {
				_ = af.IPrintLastResponse()
			}
			return fmt.Errorf("replaced node %s value %s could not be converted to bool", expr, nodeValueReplaced)
		}

		if boolVal != boolNodeValue {
			if af.isDebug {
				_ = af.IPrintLastResponse()
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
