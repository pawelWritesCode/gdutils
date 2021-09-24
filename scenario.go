package gdutils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

//Scenario struct represents data shared across one scenario.
type Scenario struct {
	//cache is storage for scenario data. It may hold any value from scenario steps or globally available environment variables
	cache map[string]interface{}
	//lastResponse holds last HTTP response
	lastResponse *http.Response
	//isDebug determine whether scenario should be run under debug mode
	isDebug bool
}

//ResetScenario resets Scenario struct instance to default values.
func (s *Scenario) ResetScenario(isDebug bool) {
	s.cache = map[string]interface{}{}
	s.lastResponse = &http.Response{}
	s.isDebug = isDebug
}

//Save preserve value under given key in cache.
func (s *Scenario) Save(key string, value interface{}) {
	s.cache[key] = value
}

//GetSaved returns preserved value from cache if present, error otherwise.
func (s *Scenario) GetSaved(key string) (interface{}, error) {
	val, ok := s.cache[key]

	if ok == false {
		return val, ErrPreservedData
	}

	return val, nil
}

//GetLastResponseBody returns last HTTP response body as slice of bytes
//method is safe for multiple use
func (s *Scenario) GetLastResponseBody() []byte {
	var bodyBytes []byte

	if s.lastResponse != nil {
		bodyBytes, _ = ioutil.ReadAll(s.lastResponse.Body)
		defer s.lastResponse.Body.Close()
		s.lastResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return bodyBytes
}
