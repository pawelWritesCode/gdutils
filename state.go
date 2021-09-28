package gdutils

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

//State struct represents data shared across one scenario.
type State struct {
	//IsDebug determine whether scenario is in debug mode
	IsDebug bool
	//Cache is storage for scenario data.
	//It may hold any value from scenario steps or globally available environment variables
	Cache Cache
	//lastResponse holds last HTTP response
	lastResponse *http.Response
	//httpClient is cli that will send HTTP request
	httpClient *http.Client
}

//NewDefault returns *State with default httpClient, DefaultCache and provided debug mode
func NewDefault(isDebug bool) *State {
	return &State{
		IsDebug:      isDebug,
		Cache:        NewDefaultCache(),
		lastResponse: nil,
		httpClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}},
	}
}

//New returns *State with provided httpClient, DefaultCache and debug mode
func New(httpClient *http.Client, cache Cache, isDebug bool) *State {
	return &State{IsDebug: isDebug, Cache: cache, lastResponse: nil, httpClient: httpClient}
}

//ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.lastResponse = nil
	s.IsDebug = isDebug
}

//GetLastResponseBody returns last HTTP response body as slice of bytes
//method is safe for multiple use
func (s *State) GetLastResponseBody() []byte {
	var bodyBytes []byte

	if s.lastResponse != nil {
		bodyBytes, _ = ioutil.ReadAll(s.lastResponse.Body)
		defer s.lastResponse.Body.Close()
		s.lastResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return bodyBytes
}

//GetLastResponseHeaders returns last HTTP response headers
func (s *State) GetLastResponseHeaders() http.Header {
	return s.lastResponse.Header
}
