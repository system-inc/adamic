package main

import (
	"encoding/json"
	engine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

func main() {
	path := os.Args[len(os.Args)-1]
	if path == "--full" {
		engine.AdamicFull()
		return
	}
	inputs := []engine.AdamicCase{}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &inputs); err != nil {
		panic(err)
	}
	rows := []engine.AdamicCase{}
	for _, in := range inputs {
		if in.Helper != "" {
			rows = append(rows, in)
		} else {
			c := in
			c.Helper = "value"
			for _, b := range []byte(in.Source) {
				c.Data = append(c.Data, int(b))
			}
			rows = append(rows, c)
		}
	}
	adapted := len(os.Args) > 2 && os.Args[1] == "--cases"
	if adapted {
		for i, c := range rows {
			if c.Helper != "union" {
				rows[i].Adaptation = engine.AdamicAdapt(c)
			}
		}
		if err = json.NewEncoder(os.Stdout).Encode(rows); err != nil {
			panic(err)
		}
	} else {
		for _, c := range rows {
			engine.AdamicObserve(c)
		}
	}
}
