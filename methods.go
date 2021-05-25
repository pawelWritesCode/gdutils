package gdutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//ApiFeature struct represents data shared across one feature.
//Field saved holds preserved values,
//Field lastResponse holds last HTTP response,
//Field lastResponseBody holds last HTTP's response body
//Field baseUrl indices url for HTTP requests.
type ApiFeature struct {
	saved            map[string]interface{}
	lastResponse     *http.Response
	lastResponseBody []byte
	baseUrl          string
	isDebug          bool
}

//ErrJson tells that value has invalid JSON format.
var ErrJson = errors.New("invalid JSON format")

//ErrResponseCode tells that response had invalid response code.
var ErrResponseCode = errors.New("invalid response code")

//ErrJsonNode tells that there is some kind of error with json node.
var ErrJsonNode = errors.New("invalid JSON node")

//ErrPreservedData tells indices that there is some kind of error with feature preserved data.
var ErrPreservedData = errors.New("preserved data error")

type bodyHeaders struct {
	Body    interface{}
	Headers map[string]string
}

//ISendAModifiedRequestToWithData sends HTTP request with body.
//Argument method indices HTTP request method for example: "POST", "GET" etc.
//Argument urlTemplate should be relative path starting from baseUrl. May include template value.
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

	req, err := http.NewRequest(method, af.baseUrl+url, bytes.NewBuffer([]byte(reqBody)))

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
//Argument urlTemplate should be relative path starting from baseUrl. May include template value.
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

	req, err := http.NewRequest(method, af.baseUrl+url, bytes.NewBuffer(reqBody))
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

func (af *ApiFeature) ISendRequestWithTokenToWithData(method, tokenTemplated, urlTemplate string, bodyTemplate *godog.DocString) error {
	client := &http.Client{}
	reqBody, err := af.replaceTemplatedValue(bodyTemplate.Content)

	if err != nil {
		return err
	}

	url, err := af.replaceTemplatedValue(urlTemplate)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, af.baseUrl+url, bytes.NewBuffer([]byte(reqBody)))

	if err != nil {
		return err
	}

	token, err := af.replaceTemplatedValue(tokenTemplated)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	return af.saveLastResponseCredentials(resp)
}

func (af *ApiFeature) ISendRequestWithTokenTo(method, tokenTemplated, urlTemplated string) error {
	client := &http.Client{}

	url, err := af.replaceTemplatedValue(urlTemplated)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, af.baseUrl+url, bytes.NewBuffer([]byte("")))

	if err != nil {
		return err
	}

	token, err := af.replaceTemplatedValue(tokenTemplated)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	return af.saveLastResponseCredentials(resp)
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
	iVal, err := Resolve(node, af.lastResponseBody)

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

func (af *ApiFeature) TheJSONResponseShouldHaveKey(key string) error {
	_, err := Resolve(key, af.lastResponseBody)
	if err != nil {
		if af.isDebug {
			_ = af.IPrintLastResponse()
		}

		return fmt.Errorf("%v, missing key '%s'", ErrJsonNode, key)
	}

	return nil
}

func (af *ApiFeature) ICreateData(data *godog.DocString) error {
	fakeData := &Data{}
	err := json.Unmarshal([]byte(data.Content), fakeData)

	if err != nil {
		return err
	}

	fakeData.Generate(af)

	for _, user := range fakeData.Users {
		requestBody := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			string(user.Username),
			string(user.Password),
		}
		reqBody, err := json.Marshal(requestBody)

		if err != nil {
			return err
		}

		_, err = af.iSendInternalRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))

		if err != nil {
			return err
		}

		if af.lastResponse.StatusCode != 201 {
			return fmt.Errorf("cannot create user %s", reqBody)
		}
		var responseContentRegister map[string]interface{}
		err = json.Unmarshal(af.lastResponseBody, &responseContentRegister)

		if err != nil {
			return err
		}

		id, ok := responseContentRegister["_id"]

		if !ok {
			return fmt.Errorf("missing _id")
		}

		af.Save(user.Alias+".id", id)

		_, err = af.iSendInternalRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))

		if err != nil {
			return err
		}

		if af.lastResponse.StatusCode != 200 {
			return fmt.Errorf("cannot login user %s, err: %s", reqBody, string(af.lastResponseBody))
		}

		var responseContent map[string]interface{}
		err = json.Unmarshal(af.lastResponseBody, &responseContent)

		if err != nil {
			return err
		}

		token, ok := responseContent["token"]

		if !ok {
			return fmt.Errorf("missing token")
		}

		af.Save(user.Alias+".token", token)
		af.Save(user.Alias+".username", requestBody.Username)
		af.Save(user.Alias+".password", requestBody.Password)
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
	iValue, err := Resolve(expr, af.lastResponseBody)
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

	iValue, err := Resolve(expr, af.lastResponseBody)
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

//TheJSONResponseShouldHaveKeys checks whether last request body has keys defined in string separated by comma
func (af *ApiFeature) TheJSONResponseShouldHaveKeys(keys string) error {
	keysSlice := strings.Split(keys, ",")

	errs := make([]error, 0, len(keysSlice))
	for _, key := range keysSlice {
		trimmedKey := strings.TrimSpace(key)
		_, err := Resolve(trimmedKey, af.lastResponseBody)

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

//IWait waits for given timeInterval amout of time
func (af *ApiFeature) IWait(timeInterval string) error {
	duration, err := time.ParseDuration(timeInterval)
	if err != nil {
		return err
	}
	time.Sleep(duration)
	return nil
}
