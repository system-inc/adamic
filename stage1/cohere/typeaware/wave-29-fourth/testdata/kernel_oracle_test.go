package react

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestWave29FourthKernels(t *testing.T) {
	dir := os.Getenv("ADAMIC_FOURTH_ARTIFACTS")
	if dir == "" {
		t.Fatal("artifact directory required")
	}
	kinds := map[string]int{"Identifier": int(ast.KindIdentifier), "PropertyAccessExpression": int(ast.KindPropertyAccessExpression), "ImportSpecifier": int(ast.KindImportSpecifier), "NamedImports": int(ast.KindNamedImports), "ImportClause": int(ast.KindImportClause), "ImportDeclaration": int(ast.KindImportDeclaration), "BindingElement": int(ast.KindBindingElement), "ObjectBindingPattern": int(ast.KindObjectBindingPattern), "VariableDeclaration": int(ast.KindVariableDeclaration), "CallExpression": int(ast.KindCallExpression), "StringLiteral": int(ast.KindStringLiteral), "NoSubstitutionTemplateLiteral": int(ast.KindNoSubstitutionTemplateLiteral), "ObjectLiteralExpression": int(ast.KindObjectLiteralExpression), "ArrayLiteralExpression": int(ast.KindArrayLiteralExpression), "ArrowFunction": int(ast.KindArrowFunction), "FunctionExpression": int(ast.KindFunctionExpression), "FunctionDeclaration": int(ast.KindFunctionDeclaration), "ClassExpression": int(ast.KindClassExpression), "NewExpression": int(ast.KindNewExpression), "RegularExpressionLiteral": int(ast.KindRegularExpressionLiteral), "JsxElement": int(ast.KindJsxElement), "JsxFragment": int(ast.KindJsxFragment), "JsxOpeningElement": int(ast.KindJsxOpeningElement), "JsxSelfClosingElement": int(ast.KindJsxSelfClosingElement), "ParenthesizedExpression": int(ast.KindParenthesizedExpression), "AsExpression": int(ast.KindAsExpression), "SatisfiesExpression": int(ast.KindSatisfiesExpression), "ConditionalExpression": int(ast.KindConditionalExpression), "BinaryExpression": int(ast.KindBinaryExpression), "And": int(ast.KindAmpersandAmpersandToken), "Or": int(ast.KindBarBarToken), "Nullish": int(ast.KindQuestionQuestionToken), "FirstAssignment": int(ast.KindFirstAssignment), "LastAssignment": int(ast.KindLastAssignment)}
	b, _ := json.MarshalIndent(kinds, "", "  ")
	os.WriteFile(filepath.Join(dir, "kinds.json"), b, 0644)
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/input.tsx", Path: "/input.tsx"}, "", core.ScriptKindTSX)
	listeners := map[string][]string{}
	for _, r := range []rule.Rule{JsxFragments, JsxNoConstructedContextValues, JsxNoUndef} {
		for k := range r.Run(rule.Context{SourceFile: source, TypeChecker: &checker.Checker{}}, nil) {
			listeners[r.Name] = append(listeners[r.Name], strings.TrimPrefix(k.String(), "Kind"))
		}
		sort.Strings(listeners[r.Name])
	}
	b, _ = json.MarshalIndent(listeners, "", "  ")
	os.WriteFile(filepath.Join(dir, "listeners.json"), b, 0644)
	type example struct {
		mode int
		code string
	}
	cases := []example{}
	for _, code := range []string{"import {Fragment} from 'react';", "import {Fragment as F} from 'react';", "import {Fragment as F} from 'preact';", "import {Other as F} from 'react';", "const F=React.Fragment;", "const F=React.Other;", "const F=A.React.Fragment;", "const F=React;", "const F=require('react');", "const F=require('preact');", "const F=require();", "const F=Require('react');", "const {Fragment}=React;", "const {Fragment:F}=require('react');", "const {Other}=React;", "const [F]=React;", "const {Fragment}={};", "let F;", "const F=(React.Fragment);"} {
		cases = append(cases, example{0, code})
	}
	for _, tag := range []string{"div", "x-gif", "Foo-bar", "Foo", "_foo", "$foo", "테스트", "Ä", "app.Foo", "A.B.C", "this.props.Tag", "this", "A:Foo", "React.Fragment", "A.React.Fragment"} {
		cases = append(cases, example{1, "const element=<" + tag + "/>;"})
	}
	for _, expr := range []string{"{}", "[]", "()=>0", "function(){}", "class{}", "new Object()", "/x/", "<div/>", "<div></div>", "<></>", "({})", "({} as any)", "true?{}:[]", "false?0:[]", "null||{}", "false&&[]", "null??(()=>0)", "0+1", "(0,{})", "v={}", "v+=[]", "({}).x", "([]).length", "0", "'s'", "`text`", "null", "make()"} {
		cases = append(cases, example{2, "const value=" + expr + ";"})
	}
	for _, mode := range []int{3, 4} {
		for _, code := range []string{"const v=<React.Fragment>hello</React.Fragment>;", "const v=<React.Fragment/>;", "const v=<React.Fragment key=\"x\"/>;", "const v=<A.React.Fragment/>;", "const v=<Other.Fragment/>;", "const v=<>hello</>;"} {
			cases = append(cases, example{mode, code})
		}
	}
	var frames, expected strings.Builder
	frame := func(v any) {
		text := fmt.Sprint(v)
		fmt.Fprintf(&frames, "%d\n%s", len(utf16.Encode([]rune(text))), text)
	}
	frame(len(cases))
	for index, c := range cases {
		sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/input.tsx", Path: "/input.tsx"}, c.code, core.ScriptKindTSX)
		if len(sf.Diagnostics()) != 0 {
			t.Fatalf("bad fixture %s", c.code)
		}
		nodes := []*ast.Node{}
		ids := map[*ast.Node]int{}
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			ids[n] = len(nodes)
			nodes = append(nodes, n)
			n.ForEachChild(walk)
			return false
		}
		walk(sf.AsNode())
		id := func(n *ast.Node) int {
			if n == nil {
				return -1
			}
			return ids[n]
		}
		queries := []*ast.Node{}
		for _, n := range nodes {
			if c.mode >= 3 && (n.Kind == ast.KindJsxElement || n.Kind == ast.KindJsxSelfClosingElement || n.Kind == ast.KindJsxFragment) {
				queries = append(queries, n)
			}
			if c.mode == 0 && (n.Kind == ast.KindImportSpecifier || n.Kind == ast.KindBindingElement || n.Kind == ast.KindVariableDeclaration) {
				queries = append(queries, n)
			}
			if c.mode == 1 && n.Kind == ast.KindJsxSelfClosingElement {
				queries = append(queries, n)
			}
			if c.mode == 2 && n.Kind == ast.KindVariableDeclaration {
				queries = append(queries, n.AsVariableDeclaration().Initializer)
			}
		}
		frame(c.mode)
		frame(len(nodes))
		for _, n := range nodes {
			var a, b *ast.Node
			op := 0
			text := ""
			switch n.Kind {
			case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				text = n.Text()
			case ast.KindPropertyAccessExpression:
				a = n.AsPropertyAccessExpression().Expression
				b = n.AsPropertyAccessExpression().Name()
			case ast.KindImportSpecifier:
				a = n.AsImportSpecifier().PropertyName
				b = n.AsImportSpecifier().Name()
			case ast.KindImportDeclaration:
				a = n.AsImportDeclaration().ModuleSpecifier
			case ast.KindVariableDeclaration:
				a = n.AsVariableDeclaration().Initializer
				b = n.AsVariableDeclaration().Name()
			case ast.KindCallExpression:
				a = n.AsCallExpression().Expression
				if n.AsCallExpression().Arguments != nil && len(n.AsCallExpression().Arguments.Nodes) > 0 {
					b = n.AsCallExpression().Arguments.Nodes[0]
				}
			case ast.KindJsxSelfClosingElement:
				a = n.AsJsxSelfClosingElement().TagName
				if attrs := n.AsJsxSelfClosingElement().Attributes; attrs != nil && attrs.AsJsxAttributes().Properties != nil {
					op = len(attrs.AsJsxAttributes().Properties.Nodes)
				}
			case ast.KindJsxOpeningElement:
				a = n.AsJsxOpeningElement().TagName
				if attrs := n.AsJsxOpeningElement().Attributes; attrs != nil && attrs.AsJsxAttributes().Properties != nil {
					op = len(attrs.AsJsxAttributes().Properties.Nodes)
				}
			case ast.KindJsxElement:
				a = n.AsJsxElement().OpeningElement
			case ast.KindParenthesizedExpression:
				a = n.AsParenthesizedExpression().Expression
			case ast.KindAsExpression:
				a = n.AsAsExpression().Expression
			case ast.KindSatisfiesExpression:
				a = n.AsSatisfiesExpression().Expression
			case ast.KindConditionalExpression:
				a = n.AsConditionalExpression().WhenTrue
				b = n.AsConditionalExpression().WhenFalse
			case ast.KindBinaryExpression:
				a = n.AsBinaryExpression().Left
				b = n.AsBinaryExpression().Right
				op = int(n.AsBinaryExpression().OperatorToken.Kind)
			}
			frame(int(n.Kind))
			frame(text)
			frame(id(n.Parent))
			frame(id(a))
			frame(id(b))
			frame(op)
		}
		frame(len(queries))
		for _, n := range queries {
			frame(id(n))
			switch c.mode {
			case 3, 4:
				options := DefaultJsxFragmentsOptions()
				if c.mode == 4 {
					options.Mode = JsxFragmentsElement
				}
				answer := ""
				ctx := rule.Context{SourceFile: sf, TypeChecker: &checker.Checker{}, Report: func(d rule.Diagnostic) { answer = d.Message.Id }}
				callbacks := JsxFragments.Run(ctx, options)
				callbacks[n.Kind](n)
				fmt.Fprintf(&expected, "%d\t%d\t%s\n", index, id(n), answer)
			case 0:
				fmt.Fprintf(&expected, "%d\t%d\t%t\n", index, id(n), jsxFragmentsDeclarationIsFragment(n))
			case 1:
				tag := n.AsJsxSelfClosingElement().TagName
				fmt.Fprintf(&expected, "%d\t%d\t%d\n", index, id(n), id(resolvableJsxReference(tag)))
			case 2:
				found := jsxNoConstructedContextValuesConstructionOf(rule.Context{}, n, 0)
				if found == nil {
					fmt.Fprintf(&expected, "%d\t%d\t-\n", index, id(n))
					continue
				}
				message := jsxNoConstructedContextValuesMessage(*found, 3, 5, "v")
				fmt.Fprintf(&expected, "%d\t%d\t%s\t%d\t%d\t%s\t%s\n", index, id(n), found.kind, id(found.node), id(found.usage), message.Id, message.Description)
			}
		}
	}
	os.WriteFile(filepath.Join(dir, "cases.frames"), []byte(frames.String()), 0644)
	os.WriteFile(filepath.Join(dir, "go.expected"), []byte(expected.String()), 0644)
	t.Logf("%d parsed TSX cases; %d compared rows; %d bytes", len(cases), strings.Count(expected.String(), "\n"), expected.Len())
}
