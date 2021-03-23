package godog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"math/rand"
	"net/http"
	"strconv"
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
}

//ErrJson tells that value has invalid JSON format.
var ErrJson = errors.New("invalid JSON format")

//ErrResponseCode tells that response had invalid response code.
var ErrResponseCode = errors.New("invalid response code")

//ErrJsonNode tells that there is some kind of error with json node.
var ErrJsonNode = errors.New("invalid JSON node")

//ErrPreservedValue tells indices that there is some kind of error with feature preserved data.
var ErrPreservedData = errors.New("preserved data error")

//ISendAModifiedRequestToWithData sends HTTP request with body.
//Argument method indices HTTP request method for example: "POST", "GET" etc.
//Argument urlTemplate should be relative path starting from baseUrl. May include template value.
//Argument bodyTemplate is string representing json request body from test suite.
//
//Response and response body will be saved and available in next steps.
func (af *ApiFeature) ISendAModifiedRequestToWithData(method, urlTemplate string, bodyTemplate *godog.DocString) error {
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

func (af *ApiFeature) ISendAModifiedRequestWithTokenToWithData(method, tokenTemplated, urlTemplate string, bodyTemplate *godog.DocString) error {
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

func (af *ApiFeature) ISendAModifiedRequestWithTokenTo(method, tokenTemplated, urlTemplated string) error {
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

//ISendAModifiedRequestTo sends HTTP request without any body.
//Argument method indices HTTP request method for example: "POST", "GET" ...
//Argument url should be relative path starting from base url. May include template value
//
//Response and response body will be saved and available in next steps.
//
//example: ISendAModifiedRequestTo("POST", "/api/city/[cityId]"),
//This will send HTTP POST request to baseUrl/api/city/2 (assume [cityId] = 2),
func (af *ApiFeature) ISendAModifiedRequestTo(method, url string) error {
	return af.ISendAModifiedRequestToWithData(method, url, &godog.DocString{})
}

//TheJSONNodeShouldBeIntegerOfValue checks if json node is integer of given value.
func (af *ApiFeature) TheJSONNodeShouldBeIntegerOfValue(nodeName, nodeValue string) error {
	var data map[string]interface{}
	var nodeValueReplaced int
	valueTemp, err := af.replaceTemplatedValue(nodeValue)

	if err != nil {
		return err
	}

	temp, err := strconv.Atoi(valueTemp)
	if err != nil {
		return err
	}
	nodeValueReplaced = temp

	err = json.Unmarshal(af.lastResponseBody, &data)
	if err != nil {
		return err
	}

	nodeFloatValue, ok := data[nodeName].(float64)
	nodeIntValue := int(nodeFloatValue)

	if ok && nodeIntValue != nodeValueReplaced {
		return fmt.Errorf("%w '%s', expected: %d, actual: %d",
			ErrJsonNode, nodeName, nodeValueReplaced, nodeIntValue)
	}

	return err
}

//TheJSONNodeShouldBeStringOfValue checks if json node is string of given value.
func (af *ApiFeature) TheJSONNodeShouldBeStringOfValue(nodeName, nodeValue string) error {
	var data map[string]interface{}
	nodeValueReplaced, err := af.replaceTemplatedValue(nodeValue)

	if err != nil {
		return err
	}

	err = json.Unmarshal(af.lastResponseBody, &data)
	if err != nil {
		return err
	}

	nodeStringValue, ok := data[nodeName].(string)

	if ok && nodeStringValue != nodeValueReplaced {
		return fmt.Errorf("%w '%s', expected: %s, actual: %s",
			ErrJsonNode, nodeName, nodeValueReplaced, nodeStringValue)
	}

	return nil
}

//TheResponseStatusCodeShouldBe compare last response status code with given in argument.
func (af *ApiFeature) TheResponseStatusCodeShouldBe(code int) error {
	if af.lastResponse.StatusCode != code {
		responseBody := map[string]interface{}{}

		json.Unmarshal(af.lastResponseBody, &responseBody)
		return fmt.Errorf("%w, expected: %d, actual: %d\nResponseBody: %+v",
			ErrResponseCode, code, af.lastResponse.StatusCode, responseBody)
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

//TheJSONShouldBeValidAccordingToSchema checks if response body is valid according to given json schema.
//Argument path should be relative path starting from documentation/redoc/ folder.
//Argument should not contain json extension: .json
//
//Example TheJSONShouldBeValidAccordingToSchema("response/city")
//func (af *ApiFeature) TheJSONShouldBeValidAccordingToSchema(path string) error {
//	jsonSchemaBytes, err := af.getJsonSchemaBytes(path)
//
//	if err != nil {
//		return err
//	}
//
//	//rs := &jsonschema.
//	rs := &jsonschema.RootSchema{}
//	if err := json.Unmarshal(jsonSchemaBytes, rs); err != nil {
//		return err
//	}
//
//	if validationErrors, _ := rs.ValidateBytes(af.lastResponseBody); len(validationErrors) > 0 {
//		return validationErrors[0]
//	}
//
//	return nil
//}

//ISaveFromTheLastResponseJSONNodeAs saves from last response json node under given variableName.
func (af *ApiFeature) ISaveFromTheLastResponseJSONNodeAs(node, variableName string) error {
	var data map[string]interface{}

	err := json.Unmarshal(af.lastResponseBody, &data)

	if err != nil {
		return err
	}

	af.save(variableName, data[node])

	return nil
}

//IGenerateARandomInt generates random integer and preserve it under given name.
func (af *ApiFeature) IGenerateARandomInt(name string) error {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 200000
	af.save(name, rand.Intn(max-min+1)+min)

	return nil
}

//ListElementWithTheIdHasFieldWithStringValue compare value with those from last HTTP request's response.
func (af *ApiFeature) ListElementWithTheIdHasFieldWithStringValue(idTemplate, fieldName, valueTemplate string) error {
	idTemp, err := af.replaceTemplatedValue(idTemplate)

	if err != nil {
		return err
	}

	id, err := strconv.Atoi(idTemp)

	if err != nil {
		return err
	}

	value, err := af.replaceTemplatedValue(valueTemplate)

	if err != nil {
		return err
	}

	data := []map[string]interface{}{}

	err = json.Unmarshal(af.lastResponseBody, &data)

	if err != nil {
		return err
	}

	for _, val := range data {
		v, ok := val["id"]

		if ok == false {
			continue
		}

		intId := int(v.(float64))

		if id == intId {
			nodeValue, ok2 := val[fieldName]

			if ok2 == false {
				return fmt.Errorf("%w, list element with id %d is missing node: '%s'", ErrJsonNode, id, fieldName)
			}

			if nodeValue == value {
				return nil
			}

			return fmt.Errorf("%w, list element with id: %d has field '%s' with value: %s, but expected: %s",
				ErrJsonNode, id, fieldName, nodeValue, value)
		}
	}

	return fmt.Errorf("%w, list is missing element with id: %d", ErrJsonNode, id)
}

//TheJSONShouldBeValidAccordingToThisSchema validate last HTTP response against given by user schema.
//func (af *ApiFeature) TheJSONShouldBeValidAccordingToThisSchema(schema *godog.DocString) error {
//	rs := &jsonschema.RootSchema{}
//	if err := json.Unmarshal([]byte(schema.Content), rs); err != nil {
//		return err
//	}
//
//	if validationErrors, _ := rs.ValidateBytes(af.lastResponseBody); len(validationErrors) > 0 {
//		return validationErrors[0]
//	}
//
//	return nil
//}

//IGenerateARandomString generates random string and save it under key
func (af *ApiFeature) IGenerateARandomString(key string) error {
	af.save(key, af.stringWithCharset(15, charsetLettersOnly))

	return nil
}

func (af *ApiFeature) TheJSONResponseShouldHaveKey(key string) error {
	var data map[string]interface{}
	err := json.Unmarshal(af.lastResponseBody, &data)

	if err != nil {
		return err
	}

	_, ok := data[key]

	if !ok {
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

		af.save(user.Alias+".id", id)

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

		af.save(user.Alias+".token", token)
		af.save(user.Alias+".username", requestBody.Username)
		af.save(user.Alias+".password", requestBody.Password)
	}

	return nil
}
