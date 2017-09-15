package main

import (
	"encoding"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type StructReader struct {
	csv *csv.Reader

	headers []string
}

func NewStructReader(r *csv.Reader) *StructReader {
	return &StructReader{csv: r}
}

func (r *StructReader) Headers() ([]string, error) {
	if r.headers != nil {
		return r.headers, nil
	}
	err := r.readHeaders()
	return r.headers, err
}

func (r *StructReader) Read(v interface{}) error {
	if v == nil {
		return nil
	}

	rType := reflect.TypeOf(v)
	if rType.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer %s", rType)
	}
	st := reflect.ValueOf(v).Elem()
	rKind := st.Kind()
	if rKind != reflect.Struct && rKind != reflect.Ptr {
		return fmt.Errorf("can't read to type %s", rType)
	}

	if r.headers == nil {
		if err := r.readHeaders(); err != nil {
			return err
		}
	}

	elType := getNonPtrElemType(rType)
	m := make(map[string]fieldPath)
	fillStructColumns(m, elType, nil)
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			st.Set(reflect.New(elType))
		}
		st = st.Elem()
	}

	record, err := r.csv.Read()
	if err != nil {
		return err
	}
	for i, h := range r.headers {
		fieldIdx, ok := m[strings.ToLower(h)]
		if !ok {
			continue
		}
		field := fieldIdx.Field(st)
		s := record[i]
		if err := set(field, h, s); err != nil {
			return err
		}
	}

	return nil
}

func getNonPtrElemType(t reflect.Type) reflect.Type {
	elType := t.Elem()
	if elType.Kind() != reflect.Ptr {
		return elType
	}
	return elType.Elem()
}

func fillStructColumns(m map[string]fieldPath, rType reflect.Type, path fieldPath) {
	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)
		// composed struct
		if field.Anonymous {
			elType := getNonPtrElemType(reflect.New(field.Type).Type())
			fillStructColumns(m, elType, append(path, i))
			continue
		}
		tag := parseTag(field)
		if tag == "" {
			continue
		}
		m[tag] = append(path, i)
	}
}

type fieldPath []int

func (f fieldPath) Field(st reflect.Value) reflect.Value {
	v := st
	for _, idx := range f {
		if v.Kind() == reflect.Ptr {
			if !v.Elem().IsValid() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(idx)
	}
	return v
}

func parseTag(field reflect.StructField) string {
	tag := field.Tag.Get("csv")
	if tag == "-" {
		return ""
	}
	if tag == "" {
		tag = field.Name
	}
	return strings.ToLower(tag)
}

func set(field reflect.Value, column, s string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(s)
	case reflect.Bool:
		v, err := toBool(s)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := toInt(s)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := toUint(s)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := toFloat(s)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Struct:
		newPtr := reflect.New(field.Type())
		unmarshaler, ok := newPtr.Interface().(encoding.TextUnmarshaler)
		if !ok {
			return fmt.Errorf("%s: struct doesn't implement encoding.TextUnmarshaler", column)
		}
		if err := unmarshaler.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		field.Set(newPtr.Elem())
	case reflect.Ptr:
		if s == "" {
			return nil
		}
		if field.IsValid() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		if err := set(field.Elem(), column, s); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unsupported type", column)
	}
	return nil
}

func toBool(s string) (bool, error) {
	switch s {
	case "yes":
		return true, nil
	case "no", "":
		return false, nil
	default:
		return strconv.ParseBool(s)
	}
}

func toInt(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 0, 64)
}

func toUint(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 0, 64)
}

func toFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func (r *StructReader) ReadAll(v interface{}) error {
	if v == nil {
		return nil
	}
	rType := reflect.TypeOf(v)
	if rType.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer %s", rType)
	}

	slicePtr := reflect.ValueOf(v)
	slice := slicePtr.Elem()

	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("non-slice %s", rType)
	}

	elType := reflect.TypeOf(v).Elem().Elem()
	for {
		el := reflect.New(elType)
		err := r.Read(el.Interface())
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		slice.Set(reflect.Append(slice, el.Elem()))
	}
	return nil
}

func (r *StructReader) readHeaders() error {
	headers, err := r.csv.Read()
	if err != nil {
		return err
	}
	dupMap := make(map[string]bool)
	for i, h := range headers {
		h = strings.TrimSpace(h)
		headers[i] = h
		if _, ok := dupMap[h]; ok {
			return fmt.Errorf("csv contains duplicated header %s", h)
		}
		dupMap[h] = true
	}

	r.headers = headers
	return nil
}
