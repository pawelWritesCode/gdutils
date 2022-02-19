package gdutils

import (
	"reflect"
	"testing"
)

func TestState_ResetState(t *testing.T) {
	s := NewDefaultState(true, "")
	s.Cache.Save("test", 1)

	s.ResetState(false)

	if s.Debugger.IsOn() != false {
		t.Errorf("IsDebug property did not change")
	}

	if !reflect.DeepEqual(s.Cache.All(), map[string]interface{}{}) {
		t.Errorf("cache did not reset")
	}
}
