package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"io"
	"os"
	"regexp"
	"sort"
)

type testCase struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Order   int    `json:"order"`
	Last    int    `json:"last"`
	Present bool   `json:"present"`
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()
	zip, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer zip.Close()
	decoder := json.NewDecoder(zip)
	texts := map[string]bool{"": true, "😀": true, "\x00": true, "hover": true, "__proto__": true, "constructor": true}
	pattern := regexp.MustCompile(`[[:alnum:]_@.-]+`)
	for {
		var row struct{ Rule, File, Source string }
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		texts[row.Source] = true
		for _, text := range pattern.FindAllString(row.Source, -1) {
			texts[text] = true
		}
	}
	names := []string{}
	for name := range texts {
		names = append(names, name)
	}
	sort.Strings(names)
	cases := []testCase{}
	for _, name := range names {
		for _, present := range []bool{false, true} {
			for _, order := range []int{-17, 0, 1, 64, 9007199254740989} {
				for _, kind := range []string{"static", "functional", "compound", "arbitrary", ""} {
					cases = append(cases, testCase{name, kind, order, 82, present})
				}
			}
		}
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(out).Encode(map[string]any{"cases": cases}); err != nil {
		panic(err)
	}
	if err = out.Close(); err != nil {
		panic(err)
	}
	one := func(left, right collapse.ParsedVariant) int { return len(left.Root) - len(right.Root) }
	two := func(left, right collapse.ParsedVariant) int { return len(right.Root) - len(left.Root) + 7 }
	snapshot := func(state *collapse.VariantRegistry, name string, order int) {
		for _, line := range collapse.AdamicRegistrySnapshot(state, name, order) {
			fmt.Println(line)
		}
	}
	for _, c := range cases {
		state := collapse.NewVariantRegistry()
		other := collapse.NewVariantRegistry()
		snapshot(state, c.Name, c.Order)
		collapse.AdamicRegistryState(state, c.Last, c.Present, c.Order)
		state.Register(c.Name, collapse.ParsedVariantKind(c.Kind))
		snapshot(state, c.Name, c.Order)
		state.Register(c.Name, "replacement")
		snapshot(state, c.Name, c.Order)
		state.Register(c.Name+":other", "static")
		snapshot(state, c.Name+":other", c.Order)
		state.AttachComparison(c.Order, one)
		snapshot(state, c.Name, c.Order)
		state.AttachComparison(c.Order, nil)
		snapshot(state, c.Name, c.Order)
		state.AttachComparison(c.Order, two)
		snapshot(state, c.Name, c.Order)
		state.AttachComparison(c.Order+1, one)
		snapshot(state, c.Name, c.Order+1)
		snapshot(other, c.Name, c.Order)
	}
}
