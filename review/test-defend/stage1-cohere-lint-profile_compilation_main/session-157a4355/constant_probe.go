package main

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
)

func main() {
	p, err := load.Load([]string{"stage1/cohere/lint/main.ts"})
	if err != nil {
		panic(err)
	}
	value, err := lower.Lower(context.Background(), p)
	if err != nil {
		panic(err)
	}
	unique := map[string]bool{}
	for _, s := range value.Strings {
		unique[s] = true
	}
	json.NewEncoder(os.Stdout).Encode(map[string]int{"string_constants": len(value.Strings), "distinct_values": len(unique), "functions": len(value.Functions)})
}
