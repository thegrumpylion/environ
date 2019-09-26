package environ

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func MarshalEnv(i interface{}, pfx string) (out []string, err error) {

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if !isPtr(t) {
		panic("i should be a pointer. try with &")
	}

	// dereference
	t = t.Elem()

	ret := []string{}
	switch {
	case isStruct(t):
		ret, err = marshalStruct(v, pfx)
	case isMap(t):
		ret, err = marshalMap(v, pfx)
	default:
		return nil, errors.New("not struct or map")
	}

	if err != nil {
		return nil, err
	}

	if pfx != "" {
		for _, r := range ret {
			out = append(out, pfx+r)
		}
		return
	}

	return ret, nil
}

func marshalStruct(v reflect.Value, name string) (out []string, err error) {

	t := v.Type()
	if isPtr(t) {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		fld := t.Field(i)
		fv := v.Field(i)
		ft := fld.Type
		fn := name + strings.ToUpper(fld.Name)

		// ignore private members
		if unicode.IsLower(rune(fld.Name[0])) {
			continue
		}

		if isPtr(ft) {
			if fv.IsNil() { // nil
				continue
			}
			fv = fv.Elem()
			ft = ft.Elem()
		}

		ret := []string{}
		switch {
		case isScalar(ft):
			s := marshalScalar(fv, fn)
			if s != "" {
				ret = []string{s}
			}
		case isSlice(ft):
			ret, err = marshalSlice(fv, fn)
		case isMap(ft):
			ret, err = marshalMap(fv, fn+".")
		case isStruct(ft):
			ret, err = marshalStruct(fv, fn+".")
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ret...)
	}
	return
}

func marshalMap(v reflect.Value, name string) (out []string, err error) {

	// map type
	t := v.Type()
	if isPtr(t) {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	// map element type
	te := t.Elem()
	if isPtr(te) {
		te = te.Elem()
	}

	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		newName := name + strings.ToUpper(fmt.Sprint(key.Interface()))

		ret := []string{}
		switch {
		case isScalar(te):
			s := marshalScalar(val, newName)
			if s != "" {
				ret = []string{s}
			}
		case isSlice(te):
			ret, err = marshalSlice(val, newName)
		case isMap(te):
			ret, err = marshalMap(val, newName+".")
		case isStruct(te):
			ret, err = marshalStruct(val, newName+".")
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ret...)
	}
	return
}

func marshalSlice(v reflect.Value, name string) (out []string, err error) {

	// slice type
	t := v.Type()
	if isPtr(t) {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
		t = t.Elem()
	}

	// slice element type
	te := t.Elem()
	if isPtr(te) {
		te = te.Elem()
	}

	if isScalar(te) {
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i)
			if isPtr(val.Type()) {
				val = val.Elem()
			}
			out = append(out, fmt.Sprintf("%v", val.Interface()))
		}
		buf := &bytes.Buffer{}
		w := csv.NewWriter(buf)
		w.Write(out)
		w.Flush()
		s := buf.String()
		valOut := reflect.ValueOf(s[:len(s)-1])
		return []string{marshalScalar(valOut, name)}, nil
	}

	for i := 0; i < v.Len(); i++ {
		pfx := name + "." + strconv.Itoa(i)
		val := v.Index(i)

		ret := []string{}
		switch {
		case isSlice(te):
			ret, err = marshalSlice(val, pfx)
		case isMap(te):
			ret, err = marshalMap(val, pfx+".")
		case isStruct(te):
			ret, err = marshalStruct(val, pfx+".")
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ret...)
	}

	return
}

func marshalScalar(v reflect.Value, name string) string {
	t := v.Type()
	if isPtr(t) {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	name = strings.ToUpper(name)
	return fmt.Sprintf("%s=%v", name, v.Interface())
}
