package main

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"
)

func TestSimpleHeaders(t *testing.T) {
	r := open("testdata/simple.csv")
	h, err := r.Headers()
	if err != nil {
		t.Errorf("headers returned error %s", err)
	}
	expected := []string{"client_id", "client_name"}
	if !reflect.DeepEqual(h, expected) {
		t.Errorf("Wrong headers: %v, expected %v", h, expected)
	}
}

func TestSimpleRead(t *testing.T) {
	var actual Simple
	r := open("testdata/simple.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	if !reflect.DeepEqual(actual, simpleExpected[0]) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", simpleExpected[0], actual)
	}
}

func TestSimpleReadPtr(t *testing.T) {
	var actual *Simple
	r := open("testdata/simple.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	if !reflect.DeepEqual(actual, simpleExpectedPtr[0]) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", simpleExpected[0], actual)
	}
}

func TestSimpleReadAll(t *testing.T) {
	var actual []Simple
	r := open("testdata/simple.csv")
	if err := r.ReadAll(&actual); err != nil {
		t.Errorf("readall returned error %s", err)
	}
	if !reflect.DeepEqual(actual, simpleExpected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", simpleExpected, actual)
	}
}

func TestSimpleReadAllPtr(t *testing.T) {
	var actual []*Simple
	r := open("testdata/simple.csv")
	if err := r.ReadAll(&actual); err != nil {
		t.Errorf("readall returned error %s", err)
	}
	if len(simpleExpectedPtr) != len(actual) {
		t.Error("wrong size")
	}
	for i, expected := range simpleExpectedPtr {
		if !reflect.DeepEqual(actual[i], expected) {
			t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual[i])
		}
	}
}

func TestTypes(t *testing.T) {
	var actual AllType
	r := open("testdata/types.csv")
	// ,,,,,,,,,,,
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := AllType{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	// string,t,1,1,1,1,1,1,1,1,1,1
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected = AllType{
		String:  "string",
		Bool:    true,
		Int:     1,
		Int8:    1,
		Int32:   1,
		Int64:   1,
		Uint:    1,
		Uint8:   1,
		Uint32:  1,
		Uint64:  1,
		Float32: 1,
		Float64: 1,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	// ,yes,,,,,,,,,,
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected = AllType{Bool: true}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	// ,no,,,,,,,,,,
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected = AllType{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	// ,,,,,,,,,,1.1,1.1
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected = AllType{Float32: 1.1, Float64: 1.1}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestTypesPtr(t *testing.T) {
	var actual AllTypePtr
	r := open("testdata/types.csv")
	// ,,,,,,,,,,,
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := AllTypePtr{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	actual = AllTypePtr{}
	// string,t,1,1,1,1,1,1,1,1,1,1
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	var (
		strVal             = "string"
		boolVal            = true
		intVal             = 1
		int8Val    int8    = 1
		int32Val   int32   = 1
		int64Val   int64   = 1
		uintVal    uint    = 1
		uint8Val   uint8   = 1
		uint32Val  uint32  = 1
		uint64Val  uint64  = 1
		float32Val float32 = 1
		float64Val float64 = 1
	)
	expected = AllTypePtr{
		String:  &strVal,
		Bool:    &boolVal,
		Int:     &intVal,
		Int8:    &int8Val,
		Int32:   &int32Val,
		Int64:   &int64Val,
		Uint:    &uintVal,
		Uint8:   &uint8Val,
		Uint32:  &uint32Val,
		Uint64:  &uint64Val,
		Float32: &float32Val,
		Float64: &float64Val,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	actual = AllTypePtr{}
	// ,yes,,,,,,,,,,
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected = AllTypePtr{Bool: &boolVal}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	actual = AllTypePtr{}
	// ,no,,,,,,,,,,
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	var falseVal = false
	expected = AllTypePtr{Bool: &falseVal}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
	actual = AllTypePtr{}
	// ,,,,,,,,,,1.1,1.1
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	var float32WithPoint float32 = 1.1
	var float64WithPoint float64 = 1.1
	expected = AllTypePtr{Float32: &float32WithPoint, Float64: &float64WithPoint}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestSimpleExtraColumn(t *testing.T) {
	var actual Simple
	r := open("testdata/clients.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	if !reflect.DeepEqual(actual, simpleExpected[0]) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", simpleExpected[0], actual)
	}
}

func TestSimpleIgnoreColumn(t *testing.T) {
	var actual SimpleIgnore
	r := open("testdata/clients.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := SimpleIgnore{
		Id:   1,
		Name: "Jose",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestSimpleNoTag(t *testing.T) {
	var actual SimpleNoTag
	r := open("testdata/clients.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := SimpleNoTag{
		Id:   1,
		Name: "Jose",
		Age:  28,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestSimpleEmbedded(t *testing.T) {
	var actual SimpleEmbedded
	r := open("testdata/simple.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := SimpleEmbedded{
		Id:   1,
		Name: NameEmbedded{Name: "Jose"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestSimpleEmbeddedPtr(t *testing.T) {
	var actual SimpleEmbeddedPtr
	r := open("testdata/simple.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := SimpleEmbeddedPtr{
		Id:   1,
		Name: &NameEmbedded{Name: "Jose"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestComposed(t *testing.T) {
	var actual Composed
	r := open("testdata/clients.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := Composed{
		Simple: Simple{
			Id:   1,
			Name: "Jose",
		},
		Age: 28,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected, actual)
	}
}

func TestComposedPtr(t *testing.T) {
	var actual ComposedPtr
	r := open("testdata/clients.csv")
	if err := r.Read(&actual); err != nil {
		t.Errorf("read returned error %s", err)
	}
	expected := ComposedPtr{
		Simple: &Simple{
			Id:   1,
			Name: "Jose",
		},
		Age: 28,
	}
	if actual.Age != expected.Age {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected.Age, actual.Age)
	}
	if !reflect.DeepEqual(actual.Simple, expected.Simple) {
		t.Errorf("expected:\n%+v\nactual:\n%+v", expected.Simple, actual.Simple)
	}
}

func TestDuplicatedHeaders(t *testing.T) {
	r := open("testdata/duplicated_headers.csv")
	_, err := r.Headers()
	if err == nil {
		t.Error("duplicated headers should return error")
	}
}

func TestReadNonStruct(t *testing.T) {
	r := open("testdata/simple.csv")
	if err := r.Read(nil); err != nil {
		t.Errorf("should not return err: %s", err)
	}
	var i int
	if err := r.Read(&i); err == nil {
		t.Error("should return err")
	}
	var s Simple
	if err := r.Read(s); err == nil {
		t.Error("should return err")
	}
}

func TestReadAllNonSliceStruct(t *testing.T) {
	r := open("testdata/simple.csv")
	if err := r.ReadAll(nil); err != nil {
		t.Errorf("should not return err: %s", err)
	}
	var i int
	if err := r.ReadAll(&i); err == nil {
		t.Error("should return err")
	}
	var s Simple
	if err := r.ReadAll(s); err == nil {
		t.Error("should return err")
	}
	var si []int
	if err := r.ReadAll(&si); err == nil {
		t.Error("should return err")
	}
}

func open(path string) *StructReader {
	in, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return NewStructReader(csv.NewReader(in))
}

type Simple struct {
	Id   int    `csv:"client_id"`
	Name string `csv:"client_name"`
}

type SimpleIgnore struct {
	Id   int    `csv:"client_id"`
	Name string `csv:"client_name"`
	Age  int    `csv:"-"`
}

type SimpleNoTag struct {
	Id   int    `csv:"client_id"`
	Name string `csv:"client_name"`
	Age  int
}

type NameEmbedded struct {
	Name string
}

func (n *NameEmbedded) UnmarshalText(b []byte) error {
	n.Name = string(b)
	return nil
}

type SimpleEmbedded struct {
	Id   int          `csv:"client_id"`
	Name NameEmbedded `csv:"client_name"`
}

type SimpleEmbeddedPtr struct {
	Id   int           `csv:"client_id"`
	Name *NameEmbedded `csv:"client_name"`
}

type Composed struct {
	Simple
	Age int `csv:"age"`
}

type ComposedPtr struct {
	*Simple
	Age int `csv:"age"`
}

var simpleExpected = []Simple{
	{
		Id:   1,
		Name: "Jose",
	},
	{
		Id:   2,
		Name: "Daniel",
	},
	{
		Id:   3,
		Name: "Vincent",
	},
}

var simpleExpectedPtr []*Simple

func init() {
	for i := range simpleExpected {
		simpleExpectedPtr = append(simpleExpectedPtr, &simpleExpected[i])
	}
}

type AllType struct {
	String  string  `csv:"string"`
	Bool    bool    `csv:"bool"`
	Int     int     `csv:"int"`
	Int8    int8    `csv:"int8"`
	Int32   int32   `csv:"int32"`
	Int64   int64   `csv:"int64"`
	Uint    uint    `csv:"uint"`
	Uint8   uint8   `csv:"uint8"`
	Uint32  uint32  `csv:"uint32"`
	Uint64  uint64  `csv:"uint64"`
	Float32 float32 `csv:"float32"`
	Float64 float64 `csv:"float64"`
}

type AllTypePtr struct {
	String  *string  `csv:"string"`
	Bool    *bool    `csv:"bool"`
	Int     *int     `csv:"int"`
	Int8    *int8    `csv:"int8"`
	Int32   *int32   `csv:"int32"`
	Int64   *int64   `csv:"int64"`
	Uint    *uint    `csv:"uint"`
	Uint8   *uint8   `csv:"uint8"`
	Uint32  *uint32  `csv:"uint32"`
	Uint64  *uint64  `csv:"uint64"`
	Float32 *float32 `csv:"float32"`
	Float64 *float64 `csv:"float64"`
}
