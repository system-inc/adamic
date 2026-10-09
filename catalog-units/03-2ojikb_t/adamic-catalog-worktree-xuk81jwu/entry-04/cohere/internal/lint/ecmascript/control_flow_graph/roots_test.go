package control_flow_graph

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// TestIndexRootsRecordsWhatEachRootsOwnCodeHolds pins the two things a rule's gate reads from the
// index: every root is listed, in the order a preorder walk for IsRoot meets them, and a bare return
// or a loop is recorded on its nearest enclosing root and never on the root around that one.
func TestIndexRootsRecordsWhatEachRootsOwnCodeHolds(t *testing.T) {
	t.Parallel()
	code := "while (ready) {}\n" +
		"function outer() {\n" +
		"  for (const item of items) {}\n" +
		"  const inner = () => { do {} while (false); return; };\n" +
		"  return 1;\n" +
		"}\n" +
		"class Thing {\n" +
		"  field = () => { for (;;) {} };\n" +
		"  static { for (const key in object) {} }\n" +
		"  method() { if (done) { return; } }\n" +
		"}\n"
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts"}, code, core.ScriptKindTS)

	type want struct {
		kind       ast.Kind
		bareReturn bool
		loops      []ast.Kind
	}
	wants := []want{
		{ast.KindSourceFile, false, []ast.Kind{ast.KindWhileStatement}},
		{ast.KindFunctionDeclaration, false, []ast.Kind{ast.KindForOfStatement}},
		{ast.KindArrowFunction, true, []ast.Kind{ast.KindDoStatement}},
		{ast.KindPropertyDeclaration, false, nil},
		{ast.KindArrowFunction, false, []ast.Kind{ast.KindForStatement}},
		{ast.KindClassStaticBlockDeclaration, false, []ast.Kind{ast.KindForInStatement}},
		{ast.KindMethodDeclaration, true, nil},
	}
	roots := IndexRoots(source.AsNode())
	if len(roots) != len(wants) {
		t.Fatalf("indexed %d roots, want %d", len(roots), len(wants))
	}
	for index, root := range roots {
		if root.Node.Kind != wants[index].kind || root.BareReturn != wants[index].bareReturn ||
			!slices.Equal(root.Loops, wants[index].loops) {
			t.Errorf("root %d is %v, bare return %t, loops %v; want %v, %t, %v", index, root.Node.Kind,
				root.BareReturn, root.Loops, wants[index].kind, wants[index].bareReturn, wants[index].loops)
		}
	}
}
