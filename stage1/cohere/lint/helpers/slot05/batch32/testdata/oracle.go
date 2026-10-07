package main

import (
	"encoding/json"
	"fmt"
	tsast "github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	structure "github.com/system-inc/cohere/internal/lint/rules/structure"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Node struct {
	Element       bool `json:"element"`
	Fragment      bool `json:"fragment"`
	SelfClosing   bool `json:"selfClosing"`
	Parenthesized bool `json:"parenthesized"`
	Conditional   bool `json:"conditional"`
	Expression    int  `json:"expression"`
	WhenTrue      int  `json:"whenTrue"`
	WhenFalse     int  `json:"whenFalse"`
}
type Row struct {
	Nodes   []Node `json:"nodes"`
	Queries []int  `json:"queries"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/structure." + map[string]string{"value": "isJsxValue", "after": "jsxAfterParentheses", "returns": "returnArgumentLooksLikeJsx"}[mode]

	data, e := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(e)
	var ready struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &ready))
	consumers := map[string]bool{}
	for _, r := range ready.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("missing consumers")
	}
	data, e = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(e)
	var inventory struct {
		Rules []struct {
			Name         string `json:"name"`
			Dependencies []struct {
				Symbol string `json:"symbol"`
			} `json:"dependencies"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	// Include all inventory consumers, even those already ported outside the blocked cohort.
	for _, r := range inventory.Rules {
		for _, d := range r.Dependencies {
			if d.Symbol == symbol {
				consumers[r.Name] = true
			}
		}
	}
	counts := map[string]int{}
	sources := map[string]bool{}
	files := map[string][]string{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, p := range r.Tests.Files {
			files[r.Name] = append(files[r.Name], p)
			tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(root, p), nil, 0)
			must(e)
			ast.Inspect(tree, func(n ast.Node) bool {
				v, ok := n.(*ast.BasicLit)
				if ok && v.Kind == token.STRING {
					s, e := strconv.Unquote(v.Value)
					must(e)
					sources[s] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no strings for " + r.Name)
		}
	}
	captured := len(sources)

	controls := []string{
		"<A/>; <A></A>; <></>; (<A/>); (((<A></A>))); (((<>x</>)));",
		"function A(){return;} const a = () => flag ? (<A/>) : null; const b = () => flag ? null : ((<B/>));",
		"const a = flag ? x : y; const b = flag ? (<A/>) : (<B/>); const c = (<A/>) ? x : y;",
		"const a = flag ? (other ? <A/> : null) : null; const b = flag ? null : (other ? <A/> : null);",
		"const a = ((<A/> as unknown)); const b = ((<A/> satisfies unknown)); const c = ((<A/>)!);",
		"const a = x && <A/>; const b = x || <A/>; const c = [<A/>]; const d = {x:<A/>};",
		strings.Repeat("(", 128) + "<A/>" + strings.Repeat(")", 128) + ";",
		"const value = (<T>x);",
	}
	for _, source := range controls {
		sources[source] = true
	}
	ordered := []string{}
	for source := range sources {
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	kinds := map[string]int{}
	positive, queries := 0, 0
	for _, source := range ordered {
		kind := core.ScriptKindTSX
		if source == controls[len(controls)-1] {
			kind = core.ScriptKindTS
		}
		file := tsparser.ParseSourceFile(tsast.SourceFileParseOptions{FileName: "/probe.tsx", Path: tspath.Path("/probe.tsx")}, source, kind)
		actual := []*tsast.Node{}
		ids := map[*tsast.Node]int{}
		var visit func(*tsast.Node)
		visit = func(node *tsast.Node) {
			if node == nil {
				return
			}
			if _, exists := ids[node]; exists {
				return
			}
			ids[node] = len(actual)
			actual = append(actual, node)
			node.ForEachChild(func(child *tsast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		id := func(node *tsast.Node) int {
			if node == nil {
				return -1
			}
			result, exists := ids[node]
			if !exists {
				panic("unvisited parser edge")
			}
			return result
		}
		row := Row{Nodes: []Node{}, Queries: []int{-1}}
		fmt.Fprintln(&want, structure.AdamicJsx(mode, nil))
		queries++
		for i, node := range actual {
			kinds[node.Kind.String()]++
			projection := Node{Element: node.Kind == tsast.KindJsxElement, Fragment: node.Kind == tsast.KindJsxFragment, SelfClosing: node.Kind == tsast.KindJsxSelfClosingElement, Parenthesized: node.Kind == tsast.KindParenthesizedExpression, Conditional: node.Kind == tsast.KindConditionalExpression, Expression: -1, WhenTrue: -1, WhenFalse: -1}
			if projection.Parenthesized {
				projection.Expression = id(node.Expression())
			}
			if projection.Conditional {
				projection.WhenTrue = id(node.AsConditionalExpression().WhenTrue)
				projection.WhenFalse = id(node.AsConditionalExpression().WhenFalse)
			}
			row.Nodes = append(row.Nodes, projection)
			row.Queries = append(row.Queries, i)
			verdict := structure.AdamicJsx(mode, node)
			if verdict {
				positive++
			}
			fmt.Fprintln(&want, verdict)
			queries++
		}
		rows = append(rows, row)
	}
	for _, kind := range []string{"KindJsxElement", "KindJsxFragment", "KindJsxSelfClosingElement", "KindParenthesizedExpression", "KindConditionalExpression", "KindAsExpression", "KindSatisfiesExpression", "KindNonNullExpression", "KindTypeAssertionExpression"} {
		if kinds[kind] == 0 {
			panic("missing control " + kind)
		}
	}
	if positive == 0 {
		panic("missing positive verdict")
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "queries": queries, "positive": positive, "kinds": kinds, "boundary": "all inventory consumer strings parsed by pinned typescript-go; every AST node and nil queried; JSX kinds, parentheses to depth 128, conditional branches and non-parenthesis wrapper controls"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
