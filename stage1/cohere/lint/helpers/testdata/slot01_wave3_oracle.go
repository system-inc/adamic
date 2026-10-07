// Oracle-only capture for slot 01's third helper batch.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	texthelpers "github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
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
	sources := []string{"", `const a = <div className="flex" class="block" title="hidden" CLASSNAME="grid" data-class="gap-2" className />;`, `const a=<div className={(('p-2'))} />;`, `const a=<div className={flag ? 'flex' : 'grid'} />;`, `const a=<div className={['a', 'b']} />;`, "const a=<div className={`flex${flag ? ' hidden' : ''}`} />;"}
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

	if os.Args[4] == "entity" {
		if !texthelpers.AdamicWave3EmptyPanics() {
			panic("Go empty entity no longer panics")
		}
		items := []string{"#", "#x", "#X41", "#0", "#x0", "#x10ffff", "#1114112", "#x110000", "#000000000000000000001", "#x000000000000000000041", "#9999999999999999999999999999999999999999999", "#xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", "#999999999999999999999999999999z", "#xFFFFFFFFFFFFFFFFFFFFFFFFFFFFz", "unknownName", "AMP", "amp", "a_b", "#-1", "# 1", "#١", "#xé", "#x😀"}
		for _, source := range sources {
			if source != "" {
				items = append(items, source)
			}
			for start := 0; start < len(source); start++ {
				if source[start] != '&' {
					continue
				}
				end := strings.IndexByte(source[start+1:], ';')
				if end > 0 {
					items = append(items, source[start+1:start+1+end])
				}
			}
		}
		for _, pair := range texthelpers.AdamicWave3Entities() {
			items = append(items, pair[0], strings.ToUpper(pair[0]), strings.ToLower(pair[0]))
		}
		for value := 0; value <= 0x110000; value += 997 {
			items = append(items, "#"+strconv.Itoa(value), "#x"+strconv.FormatInt(int64(value), 16))
		}
		for value := 0xd7ff; value <= 0xe000; value++ {
			items = append(items, "#"+strconv.Itoa(value), "#x"+strconv.FormatInt(int64(value), 16))
		}
		for _, value := range []int{0, 9, 10, 13, 32, 127, 128, 255, 256, 0xffff, 0x10000, 0x10ffff, 0x110000} {
			items = append(items, "#"+strconv.Itoa(value), "#x"+strconv.FormatInt(int64(value), 16))
		}
		for _, item := range items {
			must(encode.Encode(item))
			replacement, ok := texthelpers.AdamicWave3Decode(item)
			fmt.Fprintln(expected, ok)
			units := []string{}
			for _, unit := range utf16.Encode([]rune(replacement)) {
				units = append(units, strconv.Itoa(int(unit)))
			}
			fmt.Fprintln(expected, strings.Join(units, ","))
			fmt.Fprintln(expected, replacement)
		}
		fmt.Fprintf(os.Stderr, "%d fixture strings; %d entity queries; %d pinned named entities\n", fixtureCount, len(items), len(texthelpers.AdamicWave3Entities()))
		return
	}
	queries := 0
	for _, source := range sources {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.tsx", Path: tspath.Path("/fixture.tsx")}, source, core.ScriptKindTSX)
		nodes := []*ast.Node{}
		ids := map[*ast.Node]int{nil: -1}
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if _, ok := ids[node]; ok {
				return
			}
			ids[node] = len(nodes)
			nodes = append(nodes, node)
			node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		for _, node := range nodes {
			if node.Kind != ast.KindJsxAttribute {
				continue
			}
			attribute := node.AsJsxAttribute()
			name := ""
			present := attribute.Name() != nil
			if present {
				name = attribute.Name().Text()
			}
			for _, settings := range []map[string]bool{{"class": true, "className": true}, {"className": false, "data-class": true, "CLASSNAME": true}, {}} {
				pairs := [][]any{}
				// Stable order, including explicitly false membership.
				for _, key := range []string{"class", "className", "data-class", "CLASSNAME"} {
					if value, ok := settings[key]; ok {
						pairs = append(pairs, []any{key, value})
					}
				}
				init := attribute.Initializer
				delegated := tailwind.AdamicWave3Under(init, "Attribute", ids)
				must(encode.Encode([]any{present, name, ids[init], pairs, delegated}))
				tailwind.AdamicWave3Print(expected, tailwind.AdamicWave3Attribute(node, settings, ids))
				queries++
			}
		}
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d source batches; %d helper queries\n", fixtureCount, len(sources), queries)
}
