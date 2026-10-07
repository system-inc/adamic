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
	Element         bool   `json:"element"`
	Fragment        bool   `json:"fragment"`
	SelfClosing     bool   `json:"selfClosing"`
	Parenthesized   bool   `json:"parenthesized"`
	Conditional     bool   `json:"conditional"`
	Expression      int    `json:"expression"`
	WhenTrue        int    `json:"whenTrue"`
	WhenFalse       int    `json:"whenFalse"`
	FunctionLike    bool   `json:"functionLike"`
	Identifier      bool   `json:"identifier"`
	TypeReference   bool   `json:"typeReference"`
	Call            bool   `json:"call"`
	Property        bool   `json:"property"`
	Block           bool   `json:"block"`
	ReturnStatement bool   `json:"returnStatement"`
	Text            string `json:"text"`
	Name            int    `json:"name"`
	TypeName        int    `json:"typeName"`
	Body            int    `json:"body"`
	Parameters      []int  `json:"parameters"`
	Statements      []int  `json:"statements"`
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
	symbol := "github.com/system-inc/cohere/internal/lint/rules/structure." + map[string]string{"component": "IsLikelyReactComponent", "type": "typeReferenceName", "network": "isNetworkServiceHookCall"}[mode]

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
		"function A(props){return 0;} function B(properties){} function C(x,props){} function D({props}){} function E(Props){}; const a=(props)=>null; const b=function(properties){}; class C { method(props){} get props(){return null;} }",
		"function A(){return (<A/>);} function B(){if(x){return <A/>;}} function C(){return flag ? <A/> : null;} const a=()=>((<A/>)); const b=()=>flag ? null : <B/>; const c=()=>{function inner(){return <A/>;}}; function D(){return;} function E(){return flag ? (other ? <A/> : null) : null;} function F(){ (<A/>); }",
		"const a = ((<A/> as unknown)); const b = ((<A/> satisfies unknown)); const c = ((<A/>)!); const d = () => <></>; const e=()=> <A></A>;",
		"type A = Plain<T>; type B = Namespace.Plain<T>; type C = typeof Plain; type D = {x:string}; type E = string[]; type F=InferUseGraphQlQueryOptions;",
		"networkService.useGraphQlQuery(); networkService.useGraphQlMutation(); networkService.graphQlRequest(); networkService.useSuspenseGraphQlQuery(); networkService.other(); other.useGraphQlQuery(); networkService['useGraphQlQuery'](); networkService?.useGraphQlQuery(); networkService.useGraphQlQuery?.();",
		"(((networkService)).useGraphQlQuery)(); ((networkService.useGraphQlMutation))(); NetworkService.useGraphQlQuery(); obj.networkService.useGraphQlQuery(); (networkService as T).useGraphQlQuery(); networkService.UseGraphQlQuery(); networkService.useGraphQlQueryX(); networkService.__false_entry__();",
		strings.Repeat("(", 64) + "networkService" + strings.Repeat(")", 64) + ".useGraphQlQuery();",
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
	structure.NetworkServiceHookMethods["__false_entry__"] = false
	rows := []Row{}
	var want strings.Builder
	kinds := map[string]int{}
	positive, queries := 0, 0
	for _, source := range ordered {
		kind := core.ScriptKindTSX
		if source == controls[len(controls)-1] {
			kind = core.ScriptKindTS
		}
		file := tsparser.ParseSourceFile(tsast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/probe.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/probe.tsx"))}, source, kind)
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
		row := Row{Nodes: []Node{}, Queries: []int{}}
		if mode != "network" {
			row.Queries = append(row.Queries, -1)
			fmt.Fprintln(&want, structure.AdamicObserve(mode, nil))
			queries++
		}
		for i, node := range actual {
			kinds[node.Kind.String()]++

			projection := Node{Element: node.Kind == tsast.KindJsxElement, Fragment: node.Kind == tsast.KindJsxFragment, SelfClosing: node.Kind == tsast.KindJsxSelfClosingElement, Parenthesized: node.Kind == tsast.KindParenthesizedExpression, Conditional: node.Kind == tsast.KindConditionalExpression, Identifier: node.Kind == tsast.KindIdentifier, TypeReference: node.Kind == tsast.KindTypeReference, Call: node.Kind == tsast.KindCallExpression, Property: node.Kind == tsast.KindPropertyAccessExpression, Block: node.Kind == tsast.KindBlock, ReturnStatement: node.Kind == tsast.KindReturnStatement, Expression: -1, WhenTrue: -1, WhenFalse: -1, Name: -1, TypeName: -1, Body: -1, Parameters: []int{}, Statements: []int{}}
			switch node.Kind {
			case tsast.KindFunctionDeclaration, tsast.KindFunctionExpression, tsast.KindArrowFunction, tsast.KindMethodDeclaration:
				projection.FunctionLike = true
			}
			if projection.Identifier || node.Kind == tsast.KindPrivateIdentifier {
				projection.Text = node.Text()
			}
			if projection.Parenthesized || projection.Call || projection.Property || projection.ReturnStatement || node.Kind == tsast.KindExpressionStatement {
				projection.Expression = id(node.Expression())
			}
			if projection.Conditional {
				projection.WhenTrue = id(node.AsConditionalExpression().WhenTrue)
				projection.WhenFalse = id(node.AsConditionalExpression().WhenFalse)
			}
			if projection.TypeReference {
				projection.TypeName = id(node.AsTypeReferenceNode().TypeName)
			}
			if projection.Property {
				projection.Name = id(node.AsPropertyAccessExpression().Name())
			}
			if node.Kind == tsast.KindParameter {
				projection.Name = id(node.Name())
			}
			if projection.FunctionLike {
				projection.Body = id(node.Body())
				for _, v := range node.Parameters() {
					projection.Parameters = append(projection.Parameters, id(v))
				}
			}
			if projection.Block && node.AsBlock().Statements != nil {
				for _, v := range node.AsBlock().Statements.Nodes {
					projection.Statements = append(projection.Statements, id(v))
				}
			}
			row.Nodes = append(row.Nodes, projection)
			if mode == "network" && !projection.Call {
				continue
			}
			row.Queries = append(row.Queries, i)
			verdict := structure.AdamicObserve(mode, node)
			if verdict == "true" || mode == "type" && verdict != "" {
				positive++
			}
			fmt.Fprintln(&want, verdict)
			queries++
		}
		rows = append(rows, row)
	}
	for _, kind := range []string{"KindJsxElement", "KindJsxFragment", "KindJsxSelfClosingElement", "KindParenthesizedExpression", "KindConditionalExpression", "KindAsExpression", "KindSatisfiesExpression", "KindNonNullExpression", "KindTypeAssertionExpression", "KindTypeReference", "KindQualifiedName", "KindCallExpression", "KindPropertyAccessExpression", "KindElementAccessExpression"} {
		if kinds[kind] == 0 {
			panic("missing control " + kind)
		}
	}
	if positive == 0 {
		panic("missing positive verdict")
	}
	methods := []map[string]any{}
	names := []string{}
	for name := range structure.NetworkServiceHookMethods {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		methods = append(methods, map[string]any{"name": name, "value": structure.NetworkServiceHookMethods[name]})
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows, "methods": methods})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "queries": queries, "positive": positive, "kinds": kinds, "boundary": "every inventory consumer string parsed by pinned typescript-go; component/type every node and nil; network every actual call expression; parameter, return scope, type-name, receiver/callee and method-map controls"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
