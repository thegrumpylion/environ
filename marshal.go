package environ

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Marshal marshals input map or struct to a string slice of form key=value
func Marshal(i interface{}) (out []string, err error) {
	return marshal(i, "")
}

// MarshalMap marshals input map or struct to a map
func MarshalMap(i interface{}) (m map[string]string, err error) {
	m = map[string]string{}
	env, err := marshal(i, "")
	for _, e := range env {
		idx := strings.Index(e, "=")
		key := e[:idx]
		val := e[idx+1:]
		m[key] = val
	}
	return
}

// MarshalAndSet marshals input map or struct and the sets all values to the environment
func MarshalAndSet(i interface{}) error {
	m, err := MarshalMap(i)
	if err != nil {
		return err
	}
	for k, v := range m {
		err = os.Setenv(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalPfx marshals input map or struct to a string slice of form key=value. the keys are prefixed with pfx.
func MarshalPfx(i interface{}, pfx string) (out []string, err error) {
	return marshal(i, pfx)
}

// MarshalMapPfx marshals input map or struct to a map. the keys are prefixed with pfx.
func MarshalMapPfx(i interface{}, pfx string) (m map[string]string, err error) {
	m = map[string]string{}
	env, err := marshal(i, pfx)
	for _, e := range env {
		idx := strings.Index(e, "=")
		key := e[:idx]
		val := e[idx+1:]
		m[key] = val
	}
	return
}

// MarshalPfxAndSet marshals input map or struct and the sets all values to the environment. the keys are prefixed with pfx.
func MarshalPfxAndSet(i interface{}, pfx string) error {
	m, err := MarshalMapPfx(i, pfx)
	if err != nil {
		return err
	}
	for k, v := range m {
		err = os.Setenv(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func marshal(i interface{}, pfx string) (out []string, err error) {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if !isPtr(t) {
		return nil, errors.New("i must be a pointer")
	}

	// dereference
	t = t.Elem()

	ret := []string{}
	switch {
	case isStruct(t):
		ret, err = marshalStruct(v, "")
	case isMap(t):
		ret, err = marshalMap(v, "")
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
