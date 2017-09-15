package main

import (
	"bytes"
	"encoding/csv"
	"testing"
)

var csvBytes []byte

func init() {
	csvBytes = []byte("a,b,c,d,e,f,g\n")
	for i := 0; i < 1000; i++ {
		csvBytes = append(csvBytes, []byte("a,b,c,d,e,f,g\n")...)
	}
}

func BenchmarkStringsStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := bytes.NewBuffer(csvBytes)
		r := NewStructReader(csv.NewReader(input))
		var s []Strings
		_ = r.ReadAll(&s)
	}
}

func BenchmarkStringsStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := bytes.NewBuffer(csvBytes)
		recs, _ := csv.NewReader(input).ReadAll()
		s := make([]Strings, len(recs))
		for i, r := range recs {
			s[i] = Strings{
				A: r[0],
				B: r[1],
				C: r[2],
				D: r[3],
				E: r[4],
				F: r[5],
				G: r[6],
			}
		}
	}
}

type Strings struct {
	A string
	B string
	C string
	D string
	E string
	F string
	G string
}
