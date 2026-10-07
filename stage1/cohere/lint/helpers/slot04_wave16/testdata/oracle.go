package main

import (
	"encoding/json"
	engine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

func main() {
	data, e := os.ReadFile(os.Args[len(os.Args)-1])
	if e != nil {
		panic(e)
	}
	var inputs []engine.AdamicCase
	if e = json.Unmarshal(data, &inputs); e != nil {
		panic(e)
	}
	rows := []engine.AdamicCase{}
	for _, c := range inputs {
		rows = append(rows, engine.AdamicExpand(c)...)
	}
	if len(os.Args) > 2 && os.Args[1] == "--cases" {
		if e = json.NewEncoder(os.Stdout).Encode(rows); e != nil {
			panic(e)
		}
	} else {
		for _, c := range rows {
			engine.AdamicObserve(c)
		}
	}
}
