// Oracle-only program over the pinned Go parser and real helper functions.
package main

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rules/structure"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

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
	sources := []string{"", "page.tsx", "/page.tsx", "/page.jsx", "/route.ts", "/route.jsx", "C:\\app\\page.tsx", "/page.TSX", "/source/services/network/x.ts", "/services/local-storage/LocalStorageService.ts.bak"}
	for _, prefix := range []string{"", "/", "/repo", "C:\\repo", "/é/𐐀", "/repo/../repo"} {
		for _, base := range []string{"page", "layout", "error", "global-error", "not-found", "loading", "route", "Page", "component"} {
			for _, ext := range []string{"ts", "tsx", "jsx", "js", "TSX", "tsx.bak"} {
				sources = append(sources, prefix+"/"+base+"."+ext)
			}
		}
		for _, suffix := range []string{"/source/services/network/entry.ts", "/components/navigation/Link.tsx", "/components/navigation/Link.jsx", "/components/layout/HorizontalRule.tsx", "/components/layout/HorizontalRule.jsx", "/libraries/structure/entry.ts", "/services/local-storage/LocalStorageService.ts", "/services/local-storage/internal/LocalStorageServiceUtilities.ts"} {
			for _, tail := range []string{"", ".bak", "/nested", "?query"} {
				path := prefix + suffix + tail
				sources = append(sources, path, strings.ReplaceAll(path, "/", `\`), strings.ToLower(path))
			}
		}
	}
	sources = append(sources, "Component", "PureComponent", "component", "pureComponent", "PureComponentX", "Component.PureComponent", " Component", "Component ", "Ｃomponent")
	sources = append(sources, `class Plain {} class A extends Component {} class B extends React.PureComponent {} const E = class extends Component {}; class Wrong implements Component {} class Lower extends react.Component {} class Computed extends React["Component"] {} class P extends (React.Component) {} class R extends (React).Component {}`)
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
	if os.Args[4] == "literal" {
		sources = append(sources, "'", "\"", "`", "  '", "  `", "''", "  /* trivia */ 'flex'", "const x = 'é𐐀';", "const x = '\\n';", "const x = `flex`;", "const x = `unterminated", "const x = '\\x41';")
	}
	for _, path := range sources {
		if os.Args[4] == "literal" {
			writeLiteral(path, encode, expected)
			continue
		}
		if os.Args[4] == "memo" {
			must(encode.Encode(path))
			for _, verdict := range tailwind.AdamicMemoVerdicts(path) {
				fmt.Fprintln(expected, verdict)
			}
			continue
		}
		if os.Args[4] == "factory" {
			writeFactory(path, encode, expected)
			continue
		}
		if os.Args[4] == "class" {
			writeClass(path, encode, expected)
			continue
		}
		must(encode.Encode(path))
		if os.Args[4] == "defaults" {
			first := tailwind.DefaultClassLiteralSettings()
			printSettings(expected, first)
			first.AttributeNames[0] = "mutated attribute"
			first.CalleeNames[0] = "mutated callee"
			first.VariablePatterns[0] = "mutated pattern"
			printSettings(expected, first)
			printSettings(expected, tailwind.DefaultClassLiteralSettings())
			continue
		}
		c := structure.FileContextFor(path)
		fmt.Fprintf(expected, "%t %t %t %t %t %t %t %t %t\n", c.IsReactFile, c.IsSpecialNextJsFile, c.IsNetworkServiceFile, c.IsLinkComponentFile, c.IsHorizontalRuleComponentFile, c.IsInLibrariesStructure, c.IsPageFile, c.IsLayoutFile, c.IsLocalStorageServiceFile)
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d helper input batches (%s)\n", fixtureCount, len(sources), os.Args[4])
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func writeClass(source string, encode *json.Encoder, expected *os.File) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.tsx", Path: tspath.Path("/fixture.tsx")}, source, core.ScriptKindTSX)
	if file == nil {
		panic("parser returned nil")
	}
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
	id := func(node *ast.Node) int {
		index, present := ids[node]
		if !present {
			panic("parser named field absent from child traversal")
		}
		return index
	}
	list := func(raw *ast.NodeList) []int {
		result := []int{}
		if raw != nil {
			for _, node := range raw.Nodes {
				result = append(result, id(node))
			}
		}
		return result
	}
	positive := 0
	rows := [][]any{}
	for _, node := range nodes {
		kind := "Other"
		expression := -1
		heritage, types := []int{}, []int{}
		switch node.Kind {
		case ast.KindClassDeclaration:
			kind = "ClassDeclaration"
			heritage = list(node.AsClassDeclaration().HeritageClauses)
		case ast.KindClassExpression:
			kind = "ClassExpression"
			heritage = list(node.AsClassExpression().HeritageClauses)
		case ast.KindHeritageClause:
			kind = "HeritageClause"
			types = list(node.AsHeritageClause().Types)
		case ast.KindExpressionWithTypeArguments:
			kind = "ExpressionWithTypeArguments"
			expression = id(node.AsExpressionWithTypeArguments().Expression)
		}
		rows = append(rows, []any{kind, expression, heritage, types, react.AdamicComponentBase(node)})
	}
	must(encode.Encode(rows))
	fmt.Fprintln(expected, react.IsEs6ComponentClass(nil))
	for _, node := range nodes {
		answer := react.IsEs6ComponentClass(node)
		if answer {
			positive++
		}
		fmt.Fprintln(expected, answer)
	}
	if positive > 0 {
		fmt.Fprintf(os.Stderr, "class batch: %d nodes, %d positive answers\n", len(nodes), positive)
	}
}

func printSettings(output *os.File, settings tailwind.ClassLiteralSettings) {
	fmt.Fprintf(output, "%s;%s;%s\n", strings.Join(settings.AttributeNames, "|"), strings.Join(settings.CalleeNames, "|"), strings.Join(settings.VariablePatterns, "|"))
}

func writeFactory(source string, encode *json.Encoder, expected *os.File) {
	settings := tailwind.ClassLiteralSettings{AttributeNames: []string{"class", "className", source, source, ""}, CalleeNames: []string{"mergeClassNames", "createVariantClassNames", source, source, ""}, VariablePatterns: []string{`.*[Cc]lassName$`, `.*[Cc]lassNames$`, source, "[", `^custom$`, `^custom$`}}
	names := []string{"", source, "class", "className", "buttonClassName", "buttonClassNames", "mergeClassNames", "custom", "Custom", "prefixcustomsuffix", "notClasses", "éClassName"}
	compiled := [][]any{}
	for _, source := range settings.VariablePatterns {
		pattern, err := regexp.Compile(source)
		matches := []bool{}
		if err == nil {
			for _, name := range names {
				matches = append(matches, pattern.MatchString(name))
			}
		}
		compiled = append(compiled, []any{source, err == nil, matches})
	}
	must(encode.Encode([]any{settings.AttributeNames, settings.CalleeNames, settings.VariablePatterns, compiled, names}))
	for _, answer := range tailwind.AdamicFactoryVerdicts(settings, names) {
		fmt.Fprintln(expected, answer)
	}
}

func writeLiteral(source string, encode *json.Encoder, expected *os.File) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.tsx", Path: tspath.Path("/fixture.tsx")}, source, core.ScriptKindTSX)
	id := 0
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		current := id
		id++
		if node.Kind == ast.KindStringLiteral || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
			for _, origin := range []string{"Attribute", "Callee", "Variable", "", "custom"} {
				pos, end, verdict := tailwind.AdamicLiteralVerdicts(node, origin)
				must(encode.Encode([]any{current, node.Text(), origin, pos, end}))
				fmt.Fprintln(expected, current)
				fmt.Fprintln(expected, verdict)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
}
