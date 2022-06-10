package types

import "reflect"

// YAMLTypeMapper is entity that has ability to map underlying data type into corresponding YAML data type.
type YAMLTypeMapper struct{}

func NewYAMLTypeMapper() YAMLTypeMapper {
	return YAMLTypeMapper{}
}

// Map maps underlying data type into corresponding YAML data type.
func (Y YAMLTypeMapper) Map(data any) DataType {
	if data == nil {
		return Null
	}

	v := reflect.ValueOf(data)

	if !v.IsValid() {
		return Unknown
	}

	k := v.Kind()

	if k == reflect.Int64 || k == reflect.Int32 || k == reflect.Int16 ||
		k == reflect.Int8 || k == reflect.Int || k == reflect.Uint ||
		k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 ||
		k == reflect.Uint64 || k == reflect.Float32 || k == reflect.Float64 ||
		k == reflect.String || k == reflect.Bool {
		return Scalar
	}

	if k == reflect.Array || k == reflect.Slice {
		return Sequence
	}

	if k == reflect.Map || k == reflect.Struct {
		return Mapping
	}

	return Unknown
}
