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
	react "github.com/system-inc/cohere/internal/lint/ecmascript/react"
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
	if mode == "hook" {
		symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/react.IsHookCall"
	} else {
		symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E]." + map[string]string{"visit": "visitUnknown", "chain": "forkOptionalChain"}[mode]
	}

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

	controls := []string{"a?.b; a?.[c]; a?.(); a?.b?.c; a.b();", "function f(){if(x){return y;} while(a){break;} }", "useState(); used(); use2(); useÉ(); useΩ(); use𐐀(); React.useState(); other.useState(); React['useState'](); (useState)(); React?.useState(); React.useState?.();"}
	for _, s := range controls {
		sources[s] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	nodes := []*ast.Node{nil}
	chainNodes := []*ast.Node{}
	calls := []*ast.CallExpression{nil, {}}
	for _, source := range ordered {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/probe.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/probe.tsx"))}, source, core.ScriptKindTSX)
		chainNodes = append(chainNodes, file.AsNode())
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			nodes = append(nodes, n)
			if ast.IsOptionalChainRoot(n) {
				chainNodes = append(chainNodes, n)
			}
			if n.Kind == ast.KindCallExpression {
				calls = append(calls, n.AsCallExpression())
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	var expected strings.Builder
	var payload any
	queries := 0
	switch mode {
	case "chain":
		rows := []cfg.AdamicChainRow{}
		for _, n := range chainNodes {
			for depth := 0; depth <= 3; depth++ {
				for _, reachable := range []bool{false, true} {
					row, want := cfg.AdamicChainObserve(n, depth, reachable)
					rows = append(rows, row)
					expected.WriteString(want)
				}
			}
		}
		queries = len(rows)
		payload = struct {
			Mode string               `json:"mode"`
			Rows []cfg.AdamicChainRow `json:"rows"`
		}{mode, rows}
	case "hook":
		rows := []react.AdamicHookRow{}
		for _, call := range calls {
			row, want := react.AdamicHookObserve(call)
			rows = append(rows, row)
			expected.WriteString(want)
		}
		queries = len(rows)
		payload = struct {
			Mode string                `json:"mode"`
			Rows []react.AdamicHookRow `json:"rows"`
		}{mode, rows}
	case "visit":
		rows := []cfg.AdamicVisitRow{}
		for _, n := range nodes {
			row, want := cfg.AdamicVisitObserve(n)
			rows = append(rows, row)
			expected.WriteString(want)
		}
		queries = len(rows)
		payload = struct {
			Mode string               `json:"mode"`
			Rows []cfg.AdamicVisitRow `json:"rows"`
		}{mode, rows}
	default:
		panic("bad mode")
	}
	data, err = json.Marshal(payload)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(struct {
		Consumers               map[string]int
		Sources, Nodes, Queries int
	}{counts, len(sources), len(nodes), queries}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%s: %d queries, %d sources, %d parser nodes\n", mode, queries, len(sources), len(nodes))
}
