package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// A class member with a computed name lowers rather than crashing the file.
//
// `Node.Text()` panics on a `KindComputedPropertyName` rather than returning anything, and
// `functionName` read it unguarded. One method spelled `[Symbol.for('x')]() {}` took down the whole
// file: the linter recovers per file, so nothing in it was checked by any rule.
//
// Found by running the binary over a repository, on
// `libraries/structure/libraries/nexus/source/security/secrets/Secret.ts`, which spells its inspect
// hook `[Symbol.for('nodejs.util.inspect.custom')]()`. That file had been silently unchecked.
//
// The empty name is the right answer rather than a fallback. The name feeds `classifyFunction`,
// which asks whether it reads as a component or a hook, and a computed key is neither.
func TestLoweringAClassMemberWithAComputedName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			name:       "the shape Secret.ts writes",
			sourceText: "class Secret { [Symbol.for('nodejs.util.inspect.custom')]() { return 'redacted'; } }",
		},
		{
			name:       "a computed key naming a variable",
			sourceText: "const key = 'k'; class Holder { [key]() { return 1; } }",
		},
		{
			name:       "a computed key that is a template",
			sourceText: "class Holder { [`inspect`]() { return 1; } }",
		},
		{
			name:       "a getter with a computed name",
			sourceText: "class Holder { get [Symbol.toStringTag]() { return 'Holder'; } }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/test.tsx",
				PathKey:  "/test.tsx",
			}, testCase.sourceText, core.ScriptKindTSX)

			// Every function-like node in the file, which is what the linter walks and what reached
			// the panic. A crash here fails the test by taking the process down, so the assertion is
			// that this returns at all.
			var lowered int
			var walk func(node *ast.Node) bool
			walk = func(node *ast.Node) bool {
				if ast.IsFunctionLike(node) && functionBody(node) != nil {
					if Lower(node, nil) != nil {
						lowered++
					}
				}
				node.ForEachChild(walk)
				return false
			}
			source.AsNode().ForEachChild(walk)

			if lowered == 0 {
				t.Fatal("no function lowered, so this case did not reach the code it is about")
			}
		})
	}
}

// A computed name lowers to no name at all, which is what keeps it out of the component classifier.
//
// Asserted separately from the crash above, because a fix that returned the source text would stop
// the panic and start classifying `[Symbol.for('Foo')]` as a component.
func TestAComputedMemberNameLowersToNoName(t *testing.T) {
	t.Parallel()

	source := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx",
		PathKey:  "/test.tsx",
	}, "class Holder { [Symbol.for('Foo')]() { return 1; } }", core.ScriptKindTSX)

	var checked int
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) && functionBody(node) != nil {
			if name := functionName(node); name != "" {
				t.Errorf("a computed member name lowered to %q; a name here reaches classifyFunction "+
					"and `Symbol.for('Foo')` would read as a component", name)
			}
			checked++
		}
		node.ForEachChild(walk)
		return false
	}
	source.AsNode().ForEachChild(walk)

	if checked == 0 {
		t.Fatal("no function was checked, so this test measured nothing")
	}
}
