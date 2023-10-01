package cache

import (
	"reflect"
	"testing"
)

func TestConcurrentCache_Reset(t *testing.T) {
	c := NewConcurrentCache()
	c.Save("test1", 1)
	c.Save("test2", 2)

	expected := map[string]any{"test1": 1, "test2": 2}

	if !reflect.DeepEqual(c.All(), expected) {
		t.Errorf("all does not returns all cached values")
	}

	c.Reset()

	if !reflect.DeepEqual(c.All(), map[string]any{}) {
		t.Errorf("reset does not work")
	}
}

func TestConcurrentCache_SaveAndGetValue(t *testing.T) {
	c := NewConcurrentCache()
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

func TestConcurrentCache_GetAllValues(t *testing.T) {
	c := NewConcurrentCache()
	c.Save("test1", 1)
	c.Save("test2", 2)

	expected := map[string]any{"test1": 1, "test2": 2}

	if !reflect.DeepEqual(c.All(), expected) {
		t.Errorf("all does not returns all cached values")
	}
}
