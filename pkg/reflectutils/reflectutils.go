// Package reflectutils holds utility methods related with reflect package.
package reflectutils

import "reflect"

// IsValueNil checks whether provided Value is nil.
func IsValueNil(v reflect.Value) bool {
	nodeKind := v.Kind()
	if nodeKind == reflect.Ptr || nodeKind == reflect.Map || nodeKind == reflect.Array ||
		nodeKind == reflect.Chan || nodeKind == reflect.Slice {
		return v.IsNil()
	}

	return false
}
