package datatype

import (
	"testing"
)

func TestState_theResponseShouldBeInJSON(t *testing.T) {
	testCases := []string{
		`{"key": "value"}`,
		`{"key1": "value", "key2": "value"}`,
		`{"key1": []}`,
		`{"key1": ["aa"]}`,
	}

	for _, testCase := range testCases {
		if err := IsJSON([]byte(testCase)); err != nil {
			t.Errorf("response expected to be JSON, got err: %v", err)
		}
	}
}
