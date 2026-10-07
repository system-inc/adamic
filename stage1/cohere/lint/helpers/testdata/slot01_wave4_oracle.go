package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strconv"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	var paths []string
	data, err := os.ReadFile(os.Args[1])
	must(err)
	must(json.Unmarshal(data, &paths))
	cases, err := os.Create(os.Args[2])
	must(err)
	defer cases.Close()
	expected, err := os.Create(os.Args[3])
	must(err)
	defer expected.Close()
	encode := json.NewEncoder(cases)
	sources := []string{"", `const e=<div href={id} href="later" />;`, `const e=<div href="" HREF="&#47;about" {...props} xlink:href="other" />;`, `const e=<div href href="later" />;`, `const e=<div href={"expression"} />;`}
	fixtureCount := 0
	for _, path := range paths {
		file, err := goparser.ParseFile(token.NewFileSet(), path, nil, 0)
		must(err)
		constants := map[string]goast.Expr{}
		goast.Inspect(file, func(node goast.Node) bool {
			if declaration, ok := node.(*goast.ValueSpec); ok && len(declaration.Names) == len(declaration.Values) {
				for i, name := range declaration.Names {
					constants[name.Name] = declaration.Values[i]
				}
			}
			return true
		})
		var value func(goast.Expr, int) (string, bool)
		value = func(expression goast.Expr, depth int) (string, bool) {
			if depth > 32 {
				return "", false
			}
			switch e := expression.(type) {
			case *goast.BasicLit:
				if e.Kind == token.STRING {
					s, err := strconv.Unquote(e.Value)
					return s, err == nil
				}
			case *goast.BinaryExpr:
				if e.Op == token.ADD {
					a, ok := value(e.X, depth+1)
					b, ok2 := value(e.Y, depth+1)
					return a + b, ok && ok2
				}
			case *goast.Ident:
				if e, ok := constants[e.Name]; ok {
					return value(e, depth+1)
				}
			case *goast.ParenExpr:
				return value(e.X, depth+1)
			}
			return "", false
		}
		found := 0
		seen := map[string]bool{}
		goast.Inspect(file, func(node goast.Node) bool {
			expression, ok := node.(goast.Expr)
			if !ok {
				return true
			}
			source, ok := value(expression, 0)
			if ok {
				if source != "" && !seen[source] {
					sources = append(sources, source)
					seen[source] = true
					found++
				}
				return false // Preserve complete concatenations rather than parsing their fragments.
			}
			return true
		})
		if found == 0 {
			panic("no string inputs in " + path)
		}
		fmt.Fprintf(os.Stderr, "%s: %d string inputs\n", path, found)
		fixtureCount += found
	}

	queries := 0
	must(encode.Encode([]any{"Other", [][]any{}, "href"}))
	fmt.Fprintln(expected, false)
	fmt.Fprintln(expected, "")
	for _, source := range sources {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.tsx", Path: tspath.Path("/fixture.tsx")}, source, core.ScriptKindTSX)
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindJsxAttributes {
				attrs := node.AsJsxAttributes()
				names := []string{"href", "src", "rel", "async", "className", "", "HREF"}
				if attrs.Properties != nil {
					for _, p := range attrs.Properties.Nodes {
						name, named := jsx.AttributeName(p)
						if named {
							names = append(names, name)
						}
					}
				}
				for _, target := range names {
					for _, fold := range []bool{false, true} {
						matcher := jsx.MatchExactly
						if fold {
							matcher = jsx.MatchIgnoringCase
						}
						rows := [][]any{}
						if attrs.Properties != nil {
							for _, p := range attrs.Properties.Nodes {
								kind := "Other"
								initKind := "Missing"
								raw := ""
								name, named := jsx.AttributeName(p)
								if p.Kind == ast.KindJsxAttribute {
									kind = "JsxAttribute"
									init := p.AsJsxAttribute().Initializer
									if init != nil {
										initKind = "Other"
										if init.Kind == ast.KindStringLiteral {
											initKind = "StringLiteral"
											raw = init.Text()
										}
									}
								}
								rows = append(rows, []any{kind, named, name, initKind, raw, text.UnescapeStringLiteralText(raw), matcher(name, target)})
							}
						}
						must(encode.Encode([]any{"JsxAttributes", rows, target}))
						value, found := jsx.StringAttributeValue(node, target, matcher)
						fmt.Fprintln(expected, found)
						fmt.Fprintln(expected, value)
						queries++
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d batches; %d string-value queries\n", fixtureCount, len(sources), queries)
}
