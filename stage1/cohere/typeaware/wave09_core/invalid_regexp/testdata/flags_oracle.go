package main

import (
	"encoding/json"
	"fmt"
	core "github.com/system-inc/cohere/internal/lint/rules/core"
	"os"
)

type row struct {
	Flags  string   `json:"flags"`
	Extra  []string `json:"extra"`
	Quoted string   `json:"quoted"`
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []row
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, row := range rows {
		value := core.Wave09FlagsMessage(row.Flags, row.Extra)
		fmt.Println(value)
	}
}
