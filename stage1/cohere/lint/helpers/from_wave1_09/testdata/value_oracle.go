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
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.ParseValue"
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
	cases := append([]string{}, ordered...)
	// Capture input values and tree-node text from the actual pinned fixtures.
	fixtureCount := 0
	for _, name := range []string{"valueparser_fixtures.json", "wave2b_fixtures.json"} {
		data, err := os.ReadFile(filepath.Join(root, "cohere/internal/lint/rules/tailwind/collapse/testdata", name))
		must(err)
		var value any
		must(json.Unmarshal(data, &value))
		var capture func(any)
		capture = func(value any) {
			switch v := value.(type) {
			case map[string]any:
				for key, item := range v {
					if key == "input" || key == "value" {
						if text, ok := item.(string); ok {
							cases = append(cases, text)
							fixtureCount++
						}
					}
				}
				keys := []string{}
				for key := range v {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					capture(v[key])
				}
			case []any:
				for _, item := range v {
					capture(item)
				}
			}
		}
		capture(value)
	}
	// Deterministic cross-product separates accidental balanced-input agreement
	// from the pinned parser's malformed-input behavior.
	alphabet := []string{"a", "\\", "'", "\"", "(", ")", " ", ",", "/", "é", "😀"}
	controls := []string{"", "foo(bar", "a)b", "foo\\(bar)", "\\", "\"a b", "a\r\nb", "a\rb", "\x00é😀\ue000", "f()", "()", "(a)(b)"}
	cases = append(cases, controls...)
	for _, a := range alphabet {
		cases = append(cases, a)
		for _, b := range alphabet {
			cases = append(cases, a+b)
			for _, c := range alphabet {
				cases = append(cases, a+b+c)
			}
		}
	}
	cases = append(cases, strings.Repeat("f(", 64)+"x"+strings.Repeat(")", 64))
	// Sort so map iteration in the upstream fixture decoder cannot change hashes.
	sort.Strings(cases)
	var want strings.Builder
	var emit func([]collapse.ValueNode)
	emit = func(nodes []collapse.ValueNode) {
		for _, n := range nodes {
			fmt.Fprintln(&want, n.Kind)
			fmt.Fprintln(&want, n.Value)
			fmt.Fprintln(&want, n.Nodes != nil)
			fmt.Fprintln(&want, len(n.Nodes))
			emit(n.Nodes)
		}
	}
	for _, source := range cases {
		for repeat := 0; repeat < 2; repeat++ {
			nodes := collapse.ParseValue(source)
			fmt.Fprintln(&want, len(nodes))
			emit(nodes)
		}
	}
	corpus := map[string]any{"separators": collapse.AdamicValueSeparators(), "cases": cases}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, err = json.MarshalIndent(map[string]any{"consumers": counts, "uniqueStrings": len(sources), "upstreamFixtureStrings": fixtureCount, "cases": len(cases)}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct literals, %d upstream fixture strings, %d cases\n", len(consumers), len(sources), fixtureCount, len(cases))
}
