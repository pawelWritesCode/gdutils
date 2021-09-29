package gdutils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestState_ResetState(t *testing.T) {
	s := NewDefaultState(true)
	s.Cache.Save("test", 1)

	s.ResetState(false)

	if s.IsDebug {
		t.Errorf("IsDebug property did not change")
	}

	if !reflect.DeepEqual(s.Cache.All(), map[string]interface{}{}) {
		t.Errorf("cache did not reset")
	}
}

func TestState_GetLastResponse(t *testing.T) {
	s := NewDefaultState(false)

	resp := &http.Response{Header: map[string][]string{"Content-Type": []string{"application/json"}}}

	s.Cache.Save(lastResponseKey, resp)

	obtainedResp, err := s.GetLastResponse()
	if err != nil {
		t.Errorf("could not get last response, err: %v", err)
	}

	if !reflect.DeepEqual(resp, obtainedResp) {
		t.Errorf("obtained response is not equal to saved one")
	}
}

func TestState_GetLastResponseBody(t *testing.T) {
	s := NewDefaultState(false)

	body := []byte(`{"a": "b"}`)
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(body)),
	}

	s.Cache.Save(lastResponseKey, resp)

	obtainedRespBody := s.GetLastResponseBody()

	if !reflect.DeepEqual(body, obtainedRespBody) {
		t.Errorf("obtained response body is not equal to saved one")
	}
}
