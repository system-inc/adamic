// Executed through a Go overlay. Production cohere is unchanged.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	cfg "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	mode := symbol
	if mode == "pop" && !cfg.AdamicEmptyPopPanics() {
		panic("Go empty pop unexpectedly succeeds")
	}
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E]." + map[string]string{"push": "pushJump", "pop": "popJump", "break": "makeBreak"}[mode]

	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(err)
	var readiness struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &readiness))
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
	consumers := map[string]bool{}
	for _, r := range readiness.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("no consumers")
	}
	sources := map[string]bool{}
	counts := map[string]int{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, path := range r.Tests.Files {
			file, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			must(err)
			goast.Inspect(file, func(n goast.Node) bool {
				lit, ok := n.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				source, err := strconv.Unquote(lit.Value)
				must(err)
				if len(source) > 0 {
					sources[source] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no source strings for " + r.Name)
		}
	}

	controls := []string{
		"outer: inner: while(true) { break outer; }",
		"A: for(;;) { a: { break A; } break; }",
		"未: while(true) { break 未; }",
		"switch(x) { case 1: break; }",
		"' '; ''; A; a; 未; absent;",
	}
	for _, s := range controls {
		sources[s] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	nodes := []*ast.Node{}
	labels := []*ast.Node{nil}
	seenLabels := map[string]bool{}
	for _, source := range ordered {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/probe.tsx", Path: tspath.Path("/probe.tsx")}, source, core.ScriptKindTSX)
		nodes = append(nodes, file.AsNode())
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			switch n.Kind {
			case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindSwitchStatement:
				nodes = append(nodes, n)
			case ast.KindBreakStatement:
				label := n.AsBreakStatement().Label
				if label != nil && !seenLabels[label.Text()] {
					labels = append(labels, label)
					seenLabels[label.Text()] = true
				}
			case ast.KindExpressionStatement:
				e := n.AsExpressionStatement().Expression
				if source == controls[4] && (e.Kind == ast.KindIdentifier || e.Kind == ast.KindStringLiteral) && !seenLabels[e.Text()] {
					labels = append(labels, e)
					seenLabels[e.Text()] = true
				}
			}
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
	}
	rows := []cfg.AdamicSample{}
	var expected strings.Builder
	observe := func(n, label *ast.Node, row cfg.AdamicSample) {
		row.HasLabel = label != nil
		if label != nil {
			row.Text = label.Text()
		}
		sample, want := cfg.AdamicJumpObserve(mode, n, label, row)
		rows = append(rows, sample)
		expected.WriteString(want)
	}
	stack := func(depth, mask int) []cfg.AdamicJump {
		jumps := []cfg.AdamicJump{}
		for i := 0; i < depth; i++ {
			names := [][]string{{"A", "a"}, {"a", "未"}, {"outer", "inner"}, {"A"}, {"absent"}, {}}
			dest := 1 + i%4
			if mask&32 != 0 {
				dest = -1
			}
			lid := -1
			if mask&64 != 0 {
				lid = 1
			}
			jumps = append(jumps, cfg.AdamicJump{Labels: names[i%len(names)], Break: dest, Continue: -1, Loop: lid, Breakable: mask&(1<<i) != 0, Broken: mask&128 != 0})
		}
		return jumps
	}
	if mode == "push" {
		for _, n := range nodes {
			for _, depth := range []int{0, 1, 4} {
				for bits := 0; bits < 8; bits++ {
					bt, ct, lid := 1, 2, 1
					if bits&1 != 0 {
						bt = -1
					}
					if bits&2 != 0 {
						ct = -1
					}
					if bits&4 != 0 {
						lid = -1
					}
					observe(n, nil, cfg.AdamicSample{Jumps: stack(depth, bits), Labels: []string{}, Break: bt, Continue: ct, Loop: lid, Reachable: bits&1 != 0})
				}
			}
		}
	} else if mode == "pop" {
		for depth := 1; depth <= 32; depth++ {
			for bits := 0; bits < 256; bits++ {
				observe(nodes[0], nil, cfg.AdamicSample{Jumps: stack(depth, bits), Labels: []string{}, Reachable: bits&1 != 0})
			}
		}
	} else {
		for depth := 0; depth <= 5; depth++ {
			for bits := 0; bits < 256; bits++ {
				for _, label := range labels {
					for _, reachable := range []bool{false, true} {
						observe(nodes[0], label, cfg.AdamicSample{Jumps: stack(depth, bits), Labels: []string{}, Reachable: reachable})
					}
				}
			}
		}
	}
	data, err = json.Marshal(struct {
		Mode string             `json:"mode"`
		Rows []cfg.AdamicSample `json:"rows"`
	}{mode, rows})
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	coverage := struct {
		Consumers                       map[string]int `json:"consumer_literal_occurrences"`
		Sources, Nodes, Labels, Queries int
	}{counts, len(sources), len(nodes), len(labels), len(rows)}
	data, err = json.MarshalIndent(coverage, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%s: %d queries, %d sources, %d parsed nodes, %d label controls\n", mode, len(rows), len(sources), len(nodes), len(labels))
}
