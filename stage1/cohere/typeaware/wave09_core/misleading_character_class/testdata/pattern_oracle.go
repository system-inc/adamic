package main

import (
	"encoding/json"
	"fmt"
	core "github.com/system-inc/cohere/internal/lint/rules/core"
	"os"
)

type input struct {
	Pattern, Flags string
	AllowEscape    bool
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []input
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for i, row := range rows {
		fmt.Printf("case %d\n", i)
		fmt.Print(core.Wave09PatternFindings(row.Pattern, row.Flags, row.AllowEscape))
	}
}
