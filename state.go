package gdutils

import (
	"bytes"
	"crypto/tls"
	"fmt"
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
	//httpClient is cli that will send HTTP request
	httpClient *http.Client
}

//NewDefaultState returns *State with default http.Client, DefaultCache and provided debug mode
func NewDefaultState(isDebug bool) *State {
	return &State{
		IsDebug: isDebug,
		Cache:   NewDefaultCache(),
		httpClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}},
	}
}

//NewState returns *State with provided httpClient, cache and debug mode
func NewState(httpClient *http.Client, cache Cache, isDebug bool) *State {
	return &State{IsDebug: isDebug, Cache: cache, httpClient: httpClient}
}

//ResetState resets State struct instance to default values.
func (s *State) ResetState(isDebug bool) {
	s.Cache.Reset()
	s.IsDebug = isDebug
}

//GetLastResponse retrieve last response from cache
func (s *State) GetLastResponse() (*http.Response, error) {
	respInterface, err := s.Cache.GetSaved(lastResponseKey)
	if err != nil {
		return nil, fmt.Errorf("%w: missing last HTTP(s) response, err: %s", ErrHTTPReqRes, err.Error())
	}

	lastResp, ok := respInterface.(*http.Response)
	if !ok {
		if s.IsDebug {
			fmt.Println()
		}

		return nil, fmt.Errorf("%w: last HTTP(s) response data structure is not type *http.Response", ErrHTTPReqRes)
	}

	return lastResp, nil
}

//GetLastResponseBody returns last HTTP response body as slice of bytes
//method is safe for multiple use
func (s *State) GetLastResponseBody() []byte {
	lastResponse, err := s.GetLastResponse()
	if err != nil {
		if s.IsDebug {
			fmt.Println(err)
		}

		return []byte("")
	}

	var bodyBytes []byte

	if lastResponse != nil {
		bodyBytes, _ = ioutil.ReadAll(lastResponse.Body)
		defer lastResponse.Body.Close()
		lastResponse.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return bodyBytes
}

//GetLastResponseHeaders returns last HTTP response headers
func (s *State) GetLastResponseHeaders() http.Header {
	lastResponse, err := s.GetLastResponse()
	if err != nil {
		if s.IsDebug {
			fmt.Println(err)
		}

		return http.Header{}
	}

	return lastResponse.Header
}
