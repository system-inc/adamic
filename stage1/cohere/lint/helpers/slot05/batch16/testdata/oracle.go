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
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E]." + map[string]string{"parameter": "parameter", "update": "updateExpression", "fork": "firstThrowableFork"}[mode]

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

	controls := []string{"function f(a: typeof T,b=foo(),{x}=obj,...r: typeof R[]) { x++; --a; obj[k]++; }", "function f(未: string='a', z?: number) {}", "let x=0; x++; ++x; a.b--; a[c()]++;"}
	for _, s := range controls {
		sources[s] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	nodes := []*ast.Node{}
	if mode == "update" {
		nodes = append(nodes, nil)
	}
	for _, source := range ordered {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/probe.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/probe.tsx"))}, source, core.ScriptKindTSX)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if mode == "parameter" && n.Kind == ast.KindParameter {
				nodes = append(nodes, n)
			}
			if mode == "update" {
				if n.Kind == ast.KindPostfixUnaryExpression {
					nodes = append(nodes, n.AsPostfixUnaryExpression().Operand)
				}
				if n.Kind == ast.KindPrefixUnaryExpression {
					u := n.AsPrefixUnaryExpression()
					if u.Operator == ast.KindPlusPlusToken || u.Operator == ast.KindMinusMinusToken {
						nodes = append(nodes, u.Operand)
					}
				}
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	var expected strings.Builder
	var payload any
	queries := 0
	if mode == "fork" {
		rows := []cfg.AdamicForkRow{}
		for depth := 0; depth <= 4; depth++ {
			for mask := 0; mask < 256; mask++ {
				for _, reachable := range []bool{false, true} {
					frames := []cfg.AdamicFrame{}
					for i := 0; i < depth; i++ {
						frames = append(frames, cfg.AdamicFrame{Position: (mask >> i) & 1, HasFinally: mask&16 != 0, Forked: mask&32 != 0 && i%2 == 0, Any: mask&64 != 0, Catch: 1 + i, Finally: 5})
						if mask&128 != 0 {
							frames[i].Catch = -1
							frames[i].Finally = -1
						}
					}
					row, want := cfg.AdamicForkObserve(cfg.AdamicForkRow{Reachable: reachable, Frames: frames}, 2)
					rows = append(rows, row)
					expected.WriteString(want)
				}
			}
		}
		queries = len(rows)
		payload = struct {
			Mode string              `json:"mode"`
			Rows []cfg.AdamicForkRow `json:"rows"`
		}{mode, rows}
	} else {
		rows := []cfg.AdamicNodeRow{}
		for _, n := range nodes {
			row, want := cfg.AdamicNodeObserve(mode, n)
			rows = append(rows, row)
			expected.WriteString(want)
		}
		queries = len(rows)
		payload = struct {
			Mode string              `json:"mode"`
			Rows []cfg.AdamicNodeRow `json:"rows"`
		}{mode, rows}
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
