package types

import (
	"math"
	"reflect"
)

// GoTypeMapper is entity that has ability to map underlying data type into corresponding Go-like data type.
type GoTypeMapper struct{}

func NewGoTypeMapper() GoTypeMapper {
	return GoTypeMapper{}
}

// Map maps data underlying type into Go-like data type.
func (g GoTypeMapper) Map(data any) DataType {
	if data == nil {
		return Nil
	}

	v := reflect.ValueOf(data)

	k := v.Kind()

	if k == reflect.String {
		return String
	}

	if k == reflect.Int64 || k == reflect.Int32 || k == reflect.Int16 ||
		k == reflect.Int8 || k == reflect.Int || k == reflect.Uint ||
		k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 ||
		k == reflect.Uint64 {
		return Int
	}

	if k == reflect.Float32 || k == reflect.Float64 {
		_, frac := math.Modf(v.Float())
		if frac == 0 {
			return Int
		}

		return Float
	}

	if k == reflect.Bool {
		return Bool
	}

	if k == reflect.Map {
		return Map
	}

	if k == reflect.Slice || k == reflect.Array {
		return Slice
	}

	if !v.IsValid() {
		return Unknown
	}

	if v.IsNil() {
		return Nil
	}

	return Unknown
}
