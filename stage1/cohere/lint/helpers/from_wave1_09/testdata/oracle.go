package main

import (
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type declaration struct {
	Property     string `json:"property"`
	Value        string `json:"value"`
	ValuePresent bool   `json:"valuePresent"`
	Important    bool   `json:"important"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output := os.Args[1], os.Args[2]
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(err)
	var readiness struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &readiness))
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.nodesFromStaticDeclarations"
	consumers := map[string]bool{}
	for _, r := range readiness.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) != 6 {
		panic("consumer count drift")
	}
	data, err = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(err)
	var inventory struct {
		Rules []struct {
			Name  string `json:"name"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	counts := map[string]int{}
	sources := map[string]bool{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, path := range r.Tests.Files {
			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			must(err)
			goast.Inspect(file, func(n goast.Node) bool {
				lit, ok := n.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				source, err := strconv.Unquote(lit.Value)
				must(err)
				if source != "" {
					sources[source] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("consumer omitted: " + r.Name)
		}
	}
	ordered := []string{}
	for source := range sources {
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	cases := [][]declaration{{}, {{"", "", false, false}}, {{"", "", true, true}, {"", "", true, true}}}
	for _, source := range append(ordered, "é\x00😀\ue000\n", "", "--tw-sort") {
		for _, present := range []bool{false, true} {
			for _, important := range []bool{false, true} {
				cases = append(cases, []declaration{{source, source + " value", present, important}, {"display", "", true, false}, {source, source + " value", present, important}})
			}
		}
	}
	names := []string{}
	for name := range collapse.FrameworkStaticDeclarations {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		list := []declaration{}
		for _, d := range collapse.FrameworkStaticDeclarations[name] {
			list = append(list, declaration{d.Property, d.Value, d.ValuePresent, d.Important})
		}
		cases = append(cases, list)
	}
	var want strings.Builder
	for _, list := range cases {
		input := []collapse.StaticDeclaration{}
		for _, d := range list {
			input = append(input, collapse.StaticDeclaration{Property: d.Property, Value: d.Value, ValuePresent: d.ValuePresent, Important: d.Important})
		}
		nodes := collapse.AdamicStaticNodes(input)
		again := collapse.AdamicStaticNodes(input)
		fmt.Fprintln(&want, len(nodes))
		for _, n := range nodes {
			for _, s := range []string{string(n.Kind), n.Selector, n.Name, n.Params, n.Property, n.Value} {
				fmt.Fprintln(&want, s)
			}
			fmt.Fprintln(&want, n.ValuePresent)
			fmt.Fprintln(&want, n.Important)
			fmt.Fprintln(&want, n.Context != nil)
			fmt.Fprintln(&want, n.Nodes != nil)
		}
		if len(nodes) > 0 {
			first, fresh := nodes[0], again[0]
			original := &input[0]
			fmt.Fprintln(&want, first == fresh)
			if len(nodes) > 1 {
				fmt.Fprintln(&want, first == nodes[1])
			}
			first.Property = original.Property + " mutated node"
			first.Value = original.Value + " mutated value"
			fmt.Fprintln(&want, original.Property)
			fmt.Fprintln(&want, original.Value)
			fmt.Fprintln(&want, fresh.Property)
			fmt.Fprintln(&want, fresh.Value)
			if len(nodes) > 1 {
				fmt.Fprintln(&want, nodes[1].Property)
				fmt.Fprintln(&want, nodes[1].Value)
			}
			original.Property += " mutated input"
			original.Value += " mutated input"
			fmt.Fprintln(&want, first.Property)
			fmt.Fprintln(&want, first.Value)
			fmt.Fprintln(&want, fresh.Property)
			fmt.Fprintln(&want, fresh.Value)
		}
	}
	data, err = json.Marshal(cases)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, err = json.MarshalIndent(map[string]any{"consumers": counts, "uniqueStrings": len(sources), "frameworkStatics": len(names), "cases": len(cases)}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct literals, %d framework statics, %d cases\n", len(consumers), len(sources), len(names), len(cases))
}
