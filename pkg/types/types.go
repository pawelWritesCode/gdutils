// Package types holds utilities for working with different formats data types.
package types

// DataType represents data type.
type DataType string

const (
	Array    DataType = "array"
	Bool     DataType = "bool"
	Boolean  DataType = "boolean"
	DateTime DataType = "dateTime"
	Float    DataType = "float"
	Int      DataType = "int"
	Integer  DataType = "integer"
	Map      DataType = "map"
	Mapping  DataType = "mapping"
	Nil      DataType = "nil"
	Null     DataType = "null"
	Number   DataType = "number"
	Object   DataType = "object"
	Scalar   DataType = "scalar"
	Sequence DataType = "sequence"
	Slice    DataType = "slice"
	String   DataType = "string"
)

const (
	// Unknown represents unknown data type.
	Unknown DataType = "unknown"

	// Any represents any data type
	Any DataType = "any"
)

// Mapper is entity that has ability to map data's type into corresponding DataType of given format.
type Mapper interface {
	// Map maps data type.
	Map(data any) DataType
}

// IsValidJSONDataType checks whether is valid JSON data type.
func (dt DataType) IsValidJSONDataType() bool {
	dts := []DataType{Null, Array, Object, Number, Boolean, String}

	return isValidDataType(dts, dt)
}

// IsValidYAMLDataType checks whether is valid YAML data type.
func (dt DataType) IsValidYAMLDataType() bool {
	dts := []DataType{Scalar, Sequence, Mapping, Null}
	return isValidDataType(dts, dt)
}

// IsValidGoDataType checks whether is valid Go-like data type.
func (dt DataType) IsValidGoDataType() bool {
	dts := []DataType{String, Int, Float, Bool, Map, Slice, Nil}

	return isValidDataType(dts, dt)
}

// IsValidXMLDataType checks whether is valid XML data type.
func (dt DataType) IsValidXMLDataType() bool {
	dts := []DataType{String, Boolean, Float, Integer, DateTime}

	return isValidDataType(dts, dt)
}

func isValidDataType(set []DataType, in DataType) bool {
	for _, dt := range set {
		if in == dt {
			return true
		}
	}

	return false
}
