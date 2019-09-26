package environ

import (
	"reflect"
)

func isBool(t reflect.Type) bool {
	return t.Kind() == reflect.Bool
}

func isInt(t reflect.Type) bool {
	return t.Kind() == reflect.Int ||
		t.Kind() == reflect.Int8 ||
		t.Kind() == reflect.Int16 ||
		t.Kind() == reflect.Int32 ||
		t.Kind() == reflect.Int64 ||
		t.Kind() == reflect.Uint ||
		t.Kind() == reflect.Uint8 ||
		t.Kind() == reflect.Uint16 ||
		t.Kind() == reflect.Uint32 ||
		t.Kind() == reflect.Uint64
}

func isString(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isMap(t reflect.Type) bool {
	return t.Kind() == reflect.Map
}

func isSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice
}

func isInterface(t reflect.Type) bool {
	return t.Kind() == reflect.Interface
}

func isPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

func isScalar(t reflect.Type) bool {
	return isBool(t) ||
		isInt(t) || isString(t)
}

func isValid(t reflect.Type) bool {
	if isPtr(t) {
		t = t.Elem()
	}
	return isScalar(t) ||
		isStruct(t) ||
		isSlice(t) ||
		isMap(t)
}
