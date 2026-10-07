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

type arenaNode struct {
	Kind         string `json:"kind"`
	Property     string `json:"property"`
	Value        string `json:"value"`
	ValuePresent bool   `json:"valuePresent"`
	Nodes        []int  `json:"nodes"`
}
type sortCase struct {
	Arena []arenaNode `json:"arena"`
	Roots []int       `json:"roots"`
}
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
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.propertySort"
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
	properties := []string{}
	for key := range collapse.PropertyOrder {
		properties = append(properties, key)
	}
	sort.Strings(properties)
	makeNode := func(kind, property, value string, present bool, children ...int) arenaNode {
		if children == nil {
			children = []int{}
		}
		return arenaNode{kind, property, value, present, children}
	}
	cases := []sortCase{{[]arenaNode{}, []int{}}}
	for i, source := range append(ordered, "", "é😀\ue000\x00", "--tw-sort") {
		a, b := properties[i%len(properties)], properties[(i+1)%len(properties)]
		for _, present := range []bool{false, true} {
			cases = append(cases, sortCase{[]arenaNode{
				makeNode("rule", "", "", false, 2, 3),
				makeNode("declaration", b, source, present),
				makeNode("declaration", "--tw-sort", a, true),
				makeNode("at-rule", "", "", false, 4, 5),
				makeNode("declaration", "--tw-sort", b, true),
				makeNode("declaration", source, "", true),
			}, []int{0, 1}})
			cases = append(cases, sortCase{[]arenaNode{
				makeNode("context", "", "", false, 4),
				makeNode("at-root", "", "", false, 4),
				makeNode("comment", a, source, true, 4),
				makeNode("declaration", "--tw-sort", source, true),
				makeNode("declaration", a, "", present),
				makeNode("declaration", a, source, present),
				makeNode("declaration", b, source, true),
			}, []int{0, 1, 2, 3, 4, 5, 6}})
		}
	}
	// Explicit latch controls and repeated roots establish unknown-value fallthrough,
	// first-known wins, value absence, deduplication and count after the latch.
	for _, value := range []string{"", "missing", "display", "--tw-sort"} {
		cases = append(cases, sortCase{[]arenaNode{
			makeNode("declaration", "--tw-sort", value, true),
			makeNode("declaration", "--tw-sort", "color", true),
			makeNode("declaration", "display", "", true),
			makeNode("declaration", "color", "", false),
		}, []int{0, 1, 2, 2, 3}})
	}
	names := []string{}
	for name := range collapse.FrameworkStaticDeclarations {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := sortCase{[]arenaNode{}, []int{}}
		for _, d := range collapse.FrameworkStaticDeclarations[name] {
			c.Roots = append(c.Roots, len(c.Arena))
			c.Arena = append(c.Arena, makeNode("declaration", d.Property, d.Value, d.ValuePresent))
		}
		cases = append(cases, c)
	}
	var want strings.Builder
	emit := func(result collapse.Sort, visited []string) {
		fmt.Fprintln(&want, result.Count)
		fmt.Fprintln(&want, len(result.Order))
		for _, x := range result.Order {
			fmt.Fprintln(&want, x)
		}
		fmt.Fprintln(&want, len(visited))
		for _, x := range visited {
			fmt.Fprintln(&want, x)
		}
	}
	for _, c := range cases {
		arena := []*collapse.Node{}
		for _, n := range c.Arena {
			arena = append(arena, &collapse.Node{Kind: collapse.NodeKind(n.Kind), Property: n.Property, Value: n.Value, ValuePresent: n.ValuePresent})
		}
		for i, n := range c.Arena {
			for _, index := range n.Nodes {
				arena[i].Nodes = append(arena[i].Nodes, arena[index])
			}
		}
		roots := []*collapse.Node{}
		for _, index := range c.Roots {
			roots = append(roots, arena[index])
		}
		result, visited := collapse.AdamicPropertySort(roots)
		emit(result, visited)
		result, visited = collapse.AdamicPropertySort(roots)
		emit(result, visited)
		fmt.Fprintln(&want, len(c.Roots))
		for _, index := range c.Roots {
			fmt.Fprintln(&want, index)
		}
		for _, n := range c.Arena {
			fmt.Fprintln(&want, len(n.Nodes))
			for _, index := range n.Nodes {
				fmt.Fprintln(&want, index)
			}
		}
	}
	corpus := map[string]any{"propertyOrder": collapse.PropertyOrder, "cases": cases}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, err = json.MarshalIndent(map[string]any{"consumers": counts, "uniqueStrings": len(sources), "frameworkStatics": len(names), "cases": len(cases)}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct literals, %d framework statics, %d cases\n", len(consumers), len(sources), len(names), len(cases))
}
