package main

import (
	"encoding/json"
	"fmt"
)

type MyStruct struct {
	Data json.RawMessage `json:"data"`
}

func main() {
	s := MyStruct{
		Data: []byte(`["2026-04-22"]`),
	}
	b, _ := json.Marshal(s)
	fmt.Println(string(b))
}
