package gdutils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ItemType string

const (
	//TypeObject defines that item is object
	TypeObject ItemType = "object"
	//TypeArray defines that item is array or slice
	TypeArray ItemType = "array"
)

//Item represents part of expression
type Item struct {
	//Value represents json key name
	Value string
	//Type represents type of data structure
	Type ItemType
	//Index holds value of key for TypeArray data structure
	Index int
}

//Resolve returns data from expr of given respBody
func Resolve(expr string, respBody []byte) (interface{}, error) {
	var result interface{}

	var tmp map[string]interface{}
	err := json.Unmarshal(respBody, &tmp)
	if err != nil {
		var tmpSlice []interface{}
		err = json.Unmarshal(respBody, &tmpSlice)

		if err != nil {
			return result, fmt.Errorf("unmarshaling body error\nbody: %+v\nerr:%w", tmp, err)
		}

		newMap := map[string]interface{}{}
		newMap["root"] = tmpSlice

		exprParts, err := separate(expr)
		if err != nil {
			return result, err
		}

		return resolve(newMap, exprParts)
	}

	exprParts, err := separate(expr)
	if err != nil {
		return result, err
	}

	return resolve(tmp, exprParts)
}

//resolve holds main logic for resolving value of given expression
func resolve(data interface{}, items []Item) (interface{}, error) {
	var result interface{}

	if len(items) == 0 {
		return data, nil
	}

	item := items[0]
	if item.Type == TypeObject {
		res, err := resolveMapKey(data, item)
		if err != nil {
			return result, err
		}

		return resolve(res, items[1:])
	} else {
		res, err := resolveMapKey(data, item)
		if err != nil {
			return result, err
		}

		r, err := resolveSliceIndex(res, item)
		if err != nil {
			return result, err
		}

		return resolve(r, items[1:])
	}
}

//resolveSliceIndex resolve value from data being slice or array
func resolveSliceIndex(data interface{}, item Item) (interface{}, error) {
	val := reflect.ValueOf(data)
	valKind := val.Kind()
	if valKind == reflect.Slice || valKind == reflect.Array {
		if item.Index < val.Len() {
			fieldValue := val.Index(item.Index)
			return fieldValue.Interface(), nil
		} else {
			return nil, fmt.Errorf("slice %s does not have index %d, slice length: %d", item.Value, item.Index, val.Len())
		}
	}

	return nil, fmt.Errorf("data is not slice or array")
}

//resolveMapKey resolve value from data being map
func resolveMapKey(data interface{}, item Item) (interface{}, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			if item.Value == key.String() {
				res := val.MapIndex(key)
				return res.Interface(), nil
			}
		}
	} else {
		return nil, fmt.Errorf("data is not map")
	}

	return nil, fmt.Errorf("key %s does not exist", item.Value)
}

//separate separates expression into logic items
func separate(expr string) ([]Item, error) {
	items := []Item{}
	objects := strings.Split(expr, ".")

	for _, obj := range objects {
		leftBracketIndex := strings.Index(obj, "[")
		rightBracketIndex := strings.Index(obj, "]")

		if isIterable(leftBracketIndex, rightBracketIndex) {
			betweenBrackets := obj[leftBracketIndex+1 : rightBracketIndex]
			digit, err := strconv.Atoi(betweenBrackets)

			if err != nil {
				return items, fmt.Errorf("string between brackets does not contain digit in %s", obj)
			}

			items = append(items, Item{
				Value: obj[0:leftBracketIndex],
				Type:  TypeArray,
				Index: digit,
			})

			continue
		}

		items = append(items, Item{
			Value: obj,
			Type:  TypeObject,
			Index: 0,
		})
	}

	return items, nil
}

//isIterable checks whether expression is array or slice
func isIterable(leftBracketIndex, rightBracketIndex int) bool {
	return leftBracketIndex != -1 && rightBracketIndex != -1
}
