# CSV reader to structs

[![GoDoc](https://godoc.org/github.com/smacker/structcsv?status.png)](https://godoc.org/github.com/smacker/structcsv)

Very simple deserialization of csv file to slice of structs using struct tags.

If you also need serialization, take a look at [GoCSV](https://github.com/gocarina/gocsv).

## Installation

```go get -u github.com/smacker/structcsv```

## Example

```go
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

```

Output:

```
$ go run showcase.go
[client_id client_name age]
{ID:1 Name:Jose Age:28}
{ID:2 Name:Daniel Age:10}
{ID:3 Name:Vincent Age:54}
```