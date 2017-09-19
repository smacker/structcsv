package main

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type StructReader struct {
	csv CsvReader

	headers     []string
	typeColumns map[reflect.Type]map[string]fieldPath
}

type CsvReader interface {
	Read() (record []string, err error)
	ReadAll() (records [][]string, err error)
}

func NewStructReader(r CsvReader) *StructReader {
	return &StructReader{
		csv:         r,
		typeColumns: make(map[reflect.Type]map[string]fieldPath),
	}
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
	rValue := reflect.ValueOf(v).Elem()
	rKind := rValue.Kind()
	if rKind != reflect.Struct && rKind != reflect.Ptr {
		return fmt.Errorf("can't read to type %s", rType)
	}

	if r.headers == nil {
		if err := r.readHeaders(); err != nil {
			return err
		}
	}

	elType := getNonPtrElemType(rType)
	// input is empty struct pointer
	if rValue.Kind() == reflect.Ptr {
		if rValue.IsNil() {
			rValue.Set(reflect.New(elType))
		}
		rValue = rValue.Elem()
	}

	m := r.columnsFromType(elType)
	return r.read(m, rValue)
}

func (r *StructReader) ReadAll(v interface{}) error {
	if v == nil {
		return nil
	}
	sliceType := reflect.TypeOf(v)
	if sliceType.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer %s", sliceType)
	}

	slicePtr := reflect.ValueOf(v)
	slice := slicePtr.Elem()
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("non-slice %s", sliceType)
	}

	elKind := sliceType.Elem().Elem().Kind()
	if elKind != reflect.Struct && elKind != reflect.Ptr {
		return fmt.Errorf("can't read to type %s", slice)
	}

	sliceOfPtrs := elKind != reflect.Ptr
	elType := getNonPtrElemType(sliceType.Elem())
	m := r.columnsFromType(elType)

	if r.headers == nil {
		if err := r.readHeaders(); err != nil {
			return err
		}
	}

	for {
		rValue := reflect.New(elType)
		err := r.read(m, rValue.Elem())
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if sliceOfPtrs {
			rValue = rValue.Elem()
		}
		slice.Set(reflect.Append(slice, rValue))
	}
	return nil
}

func (r *StructReader) read(fieldMap map[string]fieldPath, rValue reflect.Value) error {
	record, err := r.csv.Read()
	if err != nil {
		return err
	}
	for i, h := range r.headers {
		fieldIdx, ok := fieldMap[h]
		if !ok {
			continue
		}
		field := fieldIdx.Field(rValue)
		if err := set(field, h, record[i]); err != nil {
			return err
		}
	}

	return nil
}

func (r *StructReader) columnsFromType(rType reflect.Type) map[string]fieldPath {
	m, ok := r.typeColumns[rType]
	if !ok {
		m = make(map[string]fieldPath)
		fillStructColumns(m, rType, nil)
		r.typeColumns[rType] = m
	}
	return r.typeColumns[rType]
}

func (r *StructReader) readHeaders() error {
	headers, err := r.csv.Read()
	if err != nil {
		return err
	}
	dupMap := make(map[string]bool)
	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		headers[i] = h
		if _, ok := dupMap[h]; ok {
			return fmt.Errorf("csv contains duplicated header %s", h)
		}
		dupMap[h] = true
	}

	r.headers = headers
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
