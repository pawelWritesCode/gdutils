package gdutils

import (
	"reflect"
	"testing"
)

func TestDefaultCache_SaveAndGetValue(t *testing.T) {
	c := NewDefaultCache()
	c.Save("test", 1)
	val, err := c.GetSaved("test")
	if err != nil {
		t.Errorf("could not obtain saved value %v", err)
	}

	iVal, ok := val.(int)
	if !ok {
		t.Errorf("cache changed preserved item type")
	}

	if iVal != 1 {
		t.Errorf("cache changed preserved item value")
	}
}

func TestDefaultCache_GetAllValues(t *testing.T) {
	c := NewDefaultCache()
	c.Save("test1", 1)
	c.Save("test2", 2)

	expected := map[string]interface{}{"test1": 1, "test2": 2}

	if !reflect.DeepEqual(c.All(), expected) {
		t.Errorf("all does not returns all cached values")
	}
}

func TestDefaultCache_Reset(t *testing.T) {
	c := NewDefaultCache()
	c.Save("test1", 1)
	c.Save("test2", 2)

	expected := map[string]interface{}{"test1": 1, "test2": 2}

	if !reflect.DeepEqual(c.All(), expected) {
		t.Errorf("all does not returns all cached values")
	}

	c.Reset()

	if !reflect.DeepEqual(c.All(), map[string]interface{}{}) {
		t.Errorf("reset does not work")
	}
}
