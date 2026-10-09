package nextjs

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// parseSource builds a source file the way the harness does, so the predicate is tested against a
// real parse rather than a hand-built node.
func parseSource(t *testing.T, sourceText string) *ast.SourceFile {
	t.Helper()
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/repository/app/Component.tsx",
		PathKey:  "/repository/app/Component.tsx",
	}, sourceText, core.ScriptKindTSX)
}

// The position half, which is the whole reason this walks rather than asking the shim predicate.
//
// `ast.IsPrologueDirective` is purely local and answers true for the string statement in the last
// three cases below. Only the walk knows that a prologue ends at the first statement that is not a
// directive, and oxc is silent on all three. That is the defect this function exists to prevent, so
// it is asserted here rather than only through a rule fixture.
func TestHasFileDirective(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		sourceText string
		want       bool
	}{
		{"double quotes", "\"use client\"\nexport const x = 1\n", true},
		{"single quotes", "'use client'\nexport const x = 1\n", true},
		{"a trailing comment", "'use client'; // uses client-only features\nexport const x = 1\n", true},
		{"a use strict directive above it", "\"use strict\"\n'use client'\nexport const x = 1\n", true},
		{"three directives deep", "\"use strict\"\n\"use whatever\"\n'use client'\nexport const x = 1\n", true},
		{"no directive at all", "export const x = 1\n", false},
		{"a different directive", "\"use server\"\nexport const x = 1\n", false},
		{"a backtick-quoted string", "`use client`\nexport const x = 1\n", false},
		{"a value rather than a directive", "const x = 'use client'\n", false},
		{"after an import", "import * as react from \"react\"\n'use client'\n", false},
		{"after a statement", "const z = 1\n'use client'\n", false},
		{"inside a function body", "function f() { 'use client' }\n", false},
		{"an empty file", "", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := HasFileDirective(parseSource(t, testCase.sourceText), "use client")
			if got != testCase.want {
				t.Errorf("HasFileDirective = %v, want %v", got, testCase.want)
			}
		})
	}
}

// A nil source file answers false rather than panicking, since a rule may run before a parse.
func TestHasFileDirectiveToleratesNil(t *testing.T) {
	t.Parallel()
	if HasFileDirective(nil, "use client") {
		t.Error("a nil source file must answer false")
	}
}

// The directive argument is honoured rather than hardcoded, so `use server` works too.
func TestHasFileDirectiveReadsTheGivenDirective(t *testing.T) {
	t.Parallel()
	sourceFile := parseSource(t, "\"use server\"\nexport const x = 1\n")
	if !HasFileDirective(sourceFile, "use server") {
		t.Error("use server must be found when asked for")
	}
	if HasFileDirective(sourceFile, "use client") {
		t.Error("use client must not be found in a use server file")
	}
}
