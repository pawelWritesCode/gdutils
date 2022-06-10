package types

import (
	"reflect"

	"github.com/pawelWritesCode/gdutils/pkg/reflectutils"
)

// JSONTypeMapper is entity that has ability to map underlying data type into corresponding JSON data type.
type JSONTypeMapper struct{}

func NewJSONTypeMapper() JSONTypeMapper {
	return JSONTypeMapper{}
}

// Map maps underlying data type into corresponding JSON data type.
func (J JSONTypeMapper) Map(data any) DataType {
	if data == nil {
		return Null
	}

	v := reflect.ValueOf(data)

	if v.Kind() == reflect.String {
		return String
	}

	if reflectutils.IsValueNil(v) {
		return Null
	}

	if v.Kind() == reflect.Int64 || v.Kind() == reflect.Int32 || v.Kind() == reflect.Int16 ||
		v.Kind() == reflect.Int8 || v.Kind() == reflect.Int || v.Kind() == reflect.Uint ||
		v.Kind() == reflect.Uint8 || v.Kind() == reflect.Uint16 || v.Kind() == reflect.Uint32 ||
		v.Kind() == reflect.Uint64 || v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64 {
		return Number
	}

	if v.Kind() == reflect.Bool {
		return Boolean
	}

	if v.Kind() == reflect.Map || v.Kind() == reflect.Struct {
		return Object
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return Array
	}

	if !v.IsValid() {
		return Unknown
	}

	return Unknown
}
