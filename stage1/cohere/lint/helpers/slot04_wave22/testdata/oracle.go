package main

import (
	"encoding/json"
	"fmt"
	react "github.com/system-inc/cohere/internal/lint/rules/react"
	engine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

func main() {
	b, e := os.ReadFile(os.Args[len(os.Args)-1])
	if e != nil {
		panic(e)
	}
	var cs []engine.AdamicCase
	if e = json.Unmarshal(b, &cs); e != nil {
		panic(e)
	}
	if len(os.Args) > 2 && os.Args[1] == "--cases" {
		out := []engine.AdamicAdapted{}
		for _, c := range cs {
			out = append(out, engine.AdamicAdapt(c))
		}
		if e = json.NewEncoder(os.Stdout).Encode(out); e != nil {
			panic(e)
		}
		return
	}
	for _, c := range cs {
		engine.AdamicObserve(c)
		_, err := react.DecodeCompilerRuleOptions([]byte(c.OptionRaw))
		s := ""
		if err != nil {
			s = err.Error()
		}
		fmt.Println(s)
	}
}
