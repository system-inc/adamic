package main

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/bridge/tsgo/spec"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"strings"
)

func main() {
	loaded, err := load.Load([]string{"bridge/tsgo/testdata/queries.a"})
	if err != nil {
		panic(err)
	}
	loaded.EnableTSGo()
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		panic(err)
	}
	type body struct {
		Name       string
		Statements int
		Panics     int
	}
	bodies := []body{}
	for _, fn := range program.Functions {
		if !strings.HasPrefix(fn.Name, spec.Prefix) {
			continue
		}
		observation := body{Name: fn.Name, Statements: len(fn.Body)}
		for _, statement := range fn.Body {
			if _, ok := statement.(ir.Panic); ok {
				observation.Panics++
			}
		}
		bodies = append(bodies, observation)
	}
	if err := json.NewEncoder(os.Stdout).Encode(bodies); err != nil {
		panic(err)
	}
}
