package gdutils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestState_theResponseShouldBeInJSON(t *testing.T) {
	s := NewDefaultState(false)

	testCases := []string{
		`{"key": "value"}`,
		`{"key1": "value", "key2": "value"}`,
		`{"key1": []}`,
		`{"key1": ["aa"]}`,
	}

	for _, testCase := range testCases {
		s.Cache.Save(lastResponseKey, &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(testCase))})

		if err := s.theResponseShouldBeInJSON(); err != nil {
			t.Errorf("response %+v to be JSON, got err: %v", s.GetLastResponseBody(), err)
		}
	}
}
