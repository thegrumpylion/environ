package environ

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type envMap struct {
	data  map[string]string
	left  map[string]string
	found map[string]string
	pfx   string
}

func (m *envMap) Get(key string) string {
	ret := m.left[key]
	m.found[key] = ret
	delete(m.left, key)
	return ret
}

func (m *envMap) Keys() []string {
	out := []string{}
	for k := range m.left {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (m *envMap) HasKeyPfx(key string) bool {
	for k := range m.left {
		if strings.HasPrefix(k, key) {
			return true
		}
	}
	return false
}

func (m *envMap) HasKey(key string) bool {
	_, ok := m.left[key]
	return ok
}

func (m *envMap) UnsetEnv() (err error) {
	for k := range m.found {
		err = os.Unsetenv(m.pfx + k)
	}
	// just the last error
	return err
}

func newEnvMap(env []string, pfx string) *envMap {
	out := &envMap{
		data:  map[string]string{},
		left:  map[string]string{},
		found: map[string]string{},
		pfx:   pfx,
	}
	for _, e := range env {
		idx := strings.Index(e, "=")
		if idx == -1 {
			continue
		}
		if strings.HasPrefix(e[:idx], pfx) {
			key := strings.TrimPrefix(e[:idx], pfx)
			out.data[key] = e[idx+1:]
			out.left[key] = e[idx+1:]
		}
	}
	return out
}

func UnmarshalEnv(i interface{}, env []string, pfx string) error {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if !isPtr(t) {
		return errors.New("i must be a pointer")
	}
	m := newEnvMap(env, pfx)
	switch t.Elem().Kind() {
	case reflect.Struct:
		return unmarshalStruct(v, m, "")
	case reflect.Map:
		return unmarshalMap(v, m, "")
	}

	return errors.New("not struct or map")
}

func unmarshalStruct(v reflect.Value, env *envMap, name string) error {
	t := v.Type()
	if isPtr(t) {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		fld := t.Field(i)
		fv := v.Field(i)
		ft := fld.Type
		fn := name + strings.ToUpper(fld.Name)
		// if a ptr check if valid & if nil.
		// if nil and is set in env allocate and dereference
		if isPtr(ft) {
			if isValid(ft) {
				if fv.IsNil() {
					if env.HasKeyPfx(fn) {
						fv.Set(reflect.New(fv.Type().Elem()))
						fv = fv.Elem()
						ft = ft.Elem()
					} else {
						continue
					}
				}
			}
		}

		var err error
		switch {
		case isScalar(ft):
			err = unmarshalScalar(fv, env.Get(fn))
		case isStruct(ft):
			err = unmarshalStruct(fv, env, fn+".")
		case isMap(ft):
			err = unmarshalMap(fv, env, fn+".")
		case isSlice(ft):
			err = unmarshalSlice(fv, env, fn)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalMap(v reflect.Value, env *envMap, name string) error {
	t := v.Type()
	m := reflect.MakeMap(t)
	if isPtr(t) {
		v = v.Elem()
		t = t.Elem()
	}

	for _, key := range env.Keys() {
		if strings.HasPrefix(key, name) {

			pfx := strings.TrimPrefix(key, name)
			ref := reflect.New(t.Elem()).Elem()
			te := t.Elem()

			if isPtr(te) {
				te = te.Elem()
			}
			// create key
			kv := reflect.New(t.Key()).Elem()
			err := unmarshalScalar(kv, pfx)
			if err != nil {
				return err
			}

			switch {
			case isScalar(te):
				err = unmarshalScalar(ref, env.Get(key))
			case isStruct(te):
				err = unmarshalStruct(ref, env, pfx+".")
			case isMap(te):
				err = unmarshalMap(ref, env, pfx+".")
			case isSlice(te):
				err = unmarshalSlice(ref, env, pfx)
			}
			if err != nil {
				return err
			}

			m.SetMapIndex(kv, ref)
		}
	}
	v.Set(m)
	return nil
}

func unmarshalSlice(v reflect.Value, env *envMap, name string) error {
	t := v.Type().Elem()
	ts := t
	slc := reflect.MakeSlice(v.Type(), 0, 10)

	if isPtr(t) {
		ts = t.Elem()
	}
	if isScalar(ts) {
		r := csv.NewReader(strings.NewReader(env.Get(name)))
		rec, err := r.Read()
		if err != nil {
			return err
		}
		for _, r := range rec {
			val := reflect.New(t).Elem()
			err := unmarshalScalar(val, r)
			if err != nil {
				return err
			}
			slc = reflect.Append(slc, val)
		}
		v.Set(slc)
		return nil
	}

	i := 0
	for {
		pfx := name + "." + strconv.Itoa(i)
		if env.HasKeyPfx(pfx) {
			var err error
			val := reflect.New(t).Elem()

			switch {
			case isStruct(ts):
				err = unmarshalStruct(val, env, pfx+".")
			case isMap(ts):
				err = unmarshalMap(val, env, pfx+".")
			case isSlice(ts):
				err = unmarshalSlice(val, env, pfx)
			}
			if err != nil {
				return err
			}

			slc = reflect.Append(slc, val)
			i++
			continue
		}
		break
	}
	v.Set(slc)

	return nil
}

func unmarshalScalar(v reflect.Value, sval string) error {
	t := v.Type()
	if isPtr(t) {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
		t = t.Elem()
	}
	switch {
	case isBool(t):
		if strings.ToLower(sval) == "true" {
			v.Set(reflect.ValueOf(true))
		}
	case isInt(t):
		i, err := strconv.Atoi(sval)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(i).Convert(v.Type()))
	case isString(t):
		v.Set(reflect.ValueOf(sval))
	default:
		return fmt.Errorf("unknown scalar %s", v.Kind().String())
	}
	return nil
}
