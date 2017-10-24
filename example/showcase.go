package main

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/smacker/structcsv"
)

var csvContent = `client_id,client_name,age
1,Jose,28
2,Daniel,10
3,Vincent,54`

type Client struct {
	ID   int    `csv:"client_id"`
	Name string `csv:"client_name"`
	Age  int    `csv:"age"`
}

func main() {
	in := bytes.NewBufferString(csvContent)
	r := structcsv.NewStructReader(csv.NewReader(in))

	headers, err := r.Headers()
	if err != nil {
		panic(err)
	}
	fmt.Println(headers)

	var clients []Client
	if err := r.ReadAll(&clients); err != nil {
		panic(err)
	}
	for _, c := range clients {
		fmt.Printf("%+v\n", c)
	}
}
