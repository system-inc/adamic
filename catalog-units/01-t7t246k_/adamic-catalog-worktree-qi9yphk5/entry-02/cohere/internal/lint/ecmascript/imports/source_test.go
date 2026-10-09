package imports

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// This file had no tests before the lift.
//
// It was proven only through the three boundary rules that called it, and its own doc comment says
// what is at stake: a rule that misses one specifier shape guards nothing, because the shape it
// misses is the one used to route around it. That claim was unasserted, and four `structure` rules
// were hand-rolling the same concern without being able to see this at all.

// collect walks source text with the visitors and returns every specifier reported.
func collect(t *testing.T, sourceText string) []string {
	t.Helper()
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/repository/source/Thing.ts",
		PathKey:  "/repository/source/Thing.ts",
	}, sourceText, core.ScriptKindTS)
	if file == nil {
		t.Fatal("the parser returned no source file")
	}

	var sources []string
	listeners := SourceVisitors(func(source string, node *ast.Node) {
		if node == nil {
			t.Fatal("want a node to blame for every specifier")
		}
		sources = append(sources, source)
	})

	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if listener, watched := listeners[node.Kind]; watched {
			listener(node)
		}
		node.ForEachChild(walk)
		return false
	}
	file.AsNode().ForEachChild(walk)
	return sources
}

// All three shapes, which is the whole contract. A rule seeing only the static form guards nothing,
// because the dynamic and require forms are what routes around it.
func TestSourceVisitorsSeeEveryShape(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a static default import", "import Thing from '@project/Thing';\n", "@project/Thing"},
		{"a static named import", "import { Thing } from '@project/Thing';\n", "@project/Thing"},
		{"a namespace import", "import * as Thing from '@project/Thing';\n", "@project/Thing"},
		{"a side-effect import", "import '@project/Thing';\n", "@project/Thing"},
		{"a type-only import", "import type { Thing } from '@project/Thing';\n", "@project/Thing"},
		{"a dynamic import", "export async function run() {\n    await import('@project/Thing');\n}\n", "@project/Thing"},
		{"a require call", "const thing = require('@project/Thing');\n", "@project/Thing"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sources := collect(t, testCase.sourceText)
			if len(sources) != 1 {
				t.Fatalf("want one specifier, got %d: %v", len(sources), sources)
			}
			if sources[0] != testCase.want {
				t.Fatalf("want %q, got %q", testCase.want, sources[0])
			}
		})
	}
}

// A file mixing the shapes reports all of them, which is what a boundary rule actually meets.
func TestSourceVisitorsSeeAllShapesInOneFile(t *testing.T) {
	t.Parallel()
	sources := collect(t, "import A from 'a';\nconst b = require('b');\nexport async function run() {\n    await import('c');\n}\n")
	if len(sources) != 3 {
		t.Fatalf("want three specifiers, got %d: %v", len(sources), sources)
	}
}

// A specifier that is not a plain string cannot be resolved by reading it, so it is declined rather
// than guessed at. Reporting a computed specifier would mean reporting on a value the rule has not
// actually seen.
func TestSourceVisitorsDeclineNonLiteralSpecifiers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
	}{
		{"a dynamic import of a variable", "export async function run(name: string) {\n    await import(name);\n}\n"},
		{"a require of a variable", "export function run(name: string) {\n    return require(name);\n}\n"},
		{"a template specifier", "export async function run(name: string) {\n    await import(`@project/${name}`);\n}\n"},
		{"a call that is not an import", "export function run() {\n    return other('@project/Thing');\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if sources := collect(t, testCase.sourceText); len(sources) != 0 {
				t.Fatalf("want nothing reported, got %v", sources)
			}
		})
	}
}

// The node handed to report differs by shape on purpose: a static import blames the declaration,
// while a dynamic or require call blames the call, since that is the expression a reader changes.
func TestSourceVisitorsBlameTheRightNode(t *testing.T) {
	t.Parallel()
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/repository/source/Thing.ts",
		PathKey:  "/repository/source/Thing.ts",
	}, "import Thing from '@project/Thing';\n", core.ScriptKindTS)

	var blamed *ast.Node
	listeners := SourceVisitors(func(source string, node *ast.Node) { blamed = node })
	for _, statement := range file.Statements.Nodes {
		if listener, watched := listeners[statement.Kind]; watched {
			listener(statement)
		}
	}

	if blamed == nil {
		t.Fatal("want a node for a static import")
	}
	if blamed.Kind != ast.KindImportDeclaration {
		t.Fatalf("want a static import to blame the declaration, got kind %v", blamed.Kind)
	}
}

// The specifier node is what a finding anchors on, and it must be the specifier rather than the
// statement, because a range starting before the specifier lands on a line a suppression comment
// cannot reach.
func TestSpecifierNodeReturnsTheSpecifier(t *testing.T) {
	t.Parallel()
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/repository/source/Thing.ts",
		PathKey:  "/repository/source/Thing.ts",
	}, "import Thing from '@project/Thing';\n", core.ScriptKindTS)

	statement := file.Statements.Nodes[0]
	specifier := SpecifierNode(statement)
	if specifier == nil {
		t.Fatal("want a specifier node")
	}
	if specifier.Text() != "@project/Thing" {
		t.Fatalf("want the specifier text, got %q", specifier.Text())
	}
}

// A nil node answers rather than panics. A panic in a shared package takes the whole run down
// instead of one rule's finding.
func TestHelpersSurviveNilInput(t *testing.T) {
	t.Parallel()
	if _, ok := CallExpressionSource(nil); ok {
		t.Fatal("want a nil call to report nothing")
	}
	if SpecifierNode(nil) != nil {
		t.Fatal("want a nil node to have no specifier")
	}
	var noCache *rule.FileCache
	_ = noCache
}

// TestHasPathSegment covers the question three rules were answering separately, two of them on the
// forward slash alone.
func TestHasPathSegment(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		path    string
		segment string
		want    bool
	}{
		{"a matching segment", "/a/b/internal/Thing.ts", "internal", true},
		{"a leading segment", "internal/Thing.ts", "internal", true},
		{"a trailing segment", "/a/b/internal", "internal", true},
		// The reason this is not a substring test. A boundary rule that fires on a name it never
		// meant to claim teaches people to disable it.
		{"a longer word containing it", "/a/internalization/Thing.ts", "internal", false},
		{"a word ending in it", "/a/notinternal/Thing.ts", "internal", false},
		{"an unrelated path", "/a/b/Thing.ts", "internal", false},
		// The Windows shape, which two of the three lifted implementations answered no to for every
		// question because they split on the forward slash alone.
		{"a Windows path", `\a\b\internal\Thing.ts`, "internal", true},
		{"a mixed path", `/a\b/internal\Thing.ts`, "internal", true},
		{"an empty segment", "/a/b/Thing.ts", "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := HasPathSegment(testCase.path, testCase.segment); got != testCase.want {
				t.Errorf("HasPathSegment(%q, %q) = %v, want %v",
					testCase.path, testCase.segment, got, testCase.want)
			}
		})
	}
}
