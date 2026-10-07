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
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E]." + map[string]string{"for": "forStatement", "inof": "forInOfStatement", "access": "accessOrCall"}[mode]

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

	controls := []string{"for(;;){continue;} for(;;){if(a)return;else throw a;} for(let x=foo();;){break;} for(x=0;true;x++){continue;} for(;false;foo()){return;} for(;x && y;bar()){break;} for(;;foo()){continue;} for(;x;){} for(let x=0,y=1;x<y;x++){}", "for(const x of xs){continue;} for(x in obj){break;} for(const {p:q=foo()} of xs){if(q)return;} for([obj[k],...rest] of xs){break;} outer: for(x of xs){break outer;} for(let a,b in xs){}", "obj.x; obj[k]; obj?.x; obj?.[k]; obj.x?.(a,b); obj?.x?.(a,b); f<T>(a,b,c); new C; new C<T>(a,b); tag<T>`x${y}`; obj?.x.y?.[k]?.(a,b); (((obj?.x)))?.y;"}

	for _, s := range controls {
		sources[s] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	nodes := []*ast.Node{nil}
	wrapperKinds := map[string]int{}
	for _, source := range ordered {
		kind := core.ScriptKindTSX
		if source == controls[0] {
			kind = core.ScriptKindTS
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/probe.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/probe.tsx"))}, source, kind)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			nodes = append(nodes, n)
			switch n.Kind {
			case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
				wrapperKinds[n.Kind.String()]++
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	var expected strings.Builder
	var payload any
	rows := []cfg.AdamicRow{}
	for _, n := range nodes {
		if n == nil {
			continue
		}
		relevant := mode == "for" && n.Kind == ast.KindForStatement || mode == "inof" && (n.Kind == ast.KindForInStatement || n.Kind == ast.KindForOfStatement) || mode == "access" && (n.Kind == ast.KindPropertyAccessExpression || n.Kind == ast.KindElementAccessExpression || n.Kind == ast.KindCallExpression || n.Kind == ast.KindNewExpression || n.Kind == ast.KindTaggedTemplateExpression)
		if !relevant {
			continue
		}
		for _, reachable := range []bool{false, true} {
			for prefix := 0; prefix < 3; prefix++ {
				depthLimit := 1
				if mode == "access" {
					depthLimit = 3
				}
				for depth := 0; depth < depthLimit; depth++ {
					for _, handler := range []bool{false, true} {
						row, want := cfg.AdamicObserve(mode, n, reachable, prefix, depth, handler)
						rows = append(rows, row)
						expected.WriteString(want)
					}
				}
			}
		}
	}
	queries := len(rows)
	if queries == 0 {
		panic("no relevant nodes")
	}
	payload = struct {
		Mode string          `json:"mode"`
		Rows []cfg.AdamicRow `json:"rows"`
	}{mode, rows}

	data, err = json.Marshal(payload)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))

	data, err = json.MarshalIndent(struct {
		Consumers               map[string]int
		Sources, Nodes, Queries int
		WrapperKinds            map[string]int
	}{counts, len(sources), len(nodes), queries, wrapperKinds}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%s: %d queries, %d sources, %d parser nodes\n", mode, queries, len(sources), len(nodes))
}
