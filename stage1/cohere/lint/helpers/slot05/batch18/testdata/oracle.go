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
	react "github.com/system-inc/cohere/internal/lint/rules/react"
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
	if mode == "strict" {
		symbol = "github.com/system-inc/cohere/internal/lint/rules/react.isEs5ComponentCallStrict"
	} else {
		symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph.*Builder[E]." + map[string]string{"reads": "patternReads", "writes": "patternWrites"}[mode]
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

	controls := []string{"((x))!; (x as T); (x satisfies T); (<T>x); (((x as T)!) satisfies T); undefined; arguments; obj[k];", "const {p:q=foo()}=obj; function f(a,b){a++; b.p++;} label: while(true){break label;} <Box prop={x}/>;", "createReactClass({}); createClass({}); React.createReactClass({}); React.createClass({}); other.createReactClass({}); (createReactClass)({}); (((React.createReactClass)))({}); React['createReactClass']({}); CreateReactClass({}); createReactClassX({}); React.createReactClass?.({});"}
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
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/probe.tsx", Path: tspath.Path("/probe.tsx")}, source, kind)
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
	queries := 0
	throwableQueries, positiveQueries := 0, 0
	if mode == "strict" {
		rows := []react.AdamicStrictRow{}
		for _, n := range nodes {
			row, want := react.AdamicStrictObserve(n)
			if strings.HasPrefix(want, "true|") {
				positiveQueries++
			}
			rows = append(rows, row)
			expected.WriteString(want)
		}
		queries = len(rows)
		payload = struct {
			Mode string                  `json:"mode"`
			Rows []react.AdamicStrictRow `json:"rows"`
		}{mode, rows}
	} else {
		rows := []cfg.AdamicPatternRow{}
		for _, n := range nodes {
			row, want := cfg.AdamicPatternObserve(mode, n)
			if strings.Contains(want, "fork;") {
				throwableQueries++
			}
			rows = append(rows, row)
			expected.WriteString(want)
		}
		queries = len(rows)
		payload = struct {
			Mode string                 `json:"mode"`
			Rows []cfg.AdamicPatternRow `json:"rows"`
		}{mode, rows}
	}
	data, err = json.Marshal(payload)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	if mode == "reads" && throwableQueries == 0 {
		panic("throwable fork path is not covered")
	}
	for _, kind := range []string{"KindAsExpression", "KindNonNullExpression", "KindParenthesizedExpression", "KindSatisfiesExpression", "KindTypeAssertionExpression"} {
		if wrapperKinds[kind] == 0 {
			panic("missing wrapper " + kind)
		}
	}
	data, err = json.MarshalIndent(struct {
		Consumers                         map[string]int
		Sources, Nodes, Queries           int
		ThrowableQueries, PositiveQueries int
		WrapperKinds                      map[string]int
	}{counts, len(sources), len(nodes), queries, throwableQueries, positiveQueries, wrapperKinds}, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%s: %d queries, %d sources, %d parser nodes\n", mode, queries, len(sources), len(nodes))
}
