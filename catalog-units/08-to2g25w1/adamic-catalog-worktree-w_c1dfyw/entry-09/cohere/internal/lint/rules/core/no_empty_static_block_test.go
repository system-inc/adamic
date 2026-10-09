package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const staticBlockFile = "/repository/source/Thing.ts"

func TestNoEmptyStaticBlockFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an empty block", "export class Thing {\n    static {}\n}\n"},
		{"an empty block spanning lines", "export class Thing {\n    static {\n    }\n}\n"},
		{"one of two blocks empty", "export class Thing {\n    static { Thing.ready = true; }\n    static {}\n    static ready = false;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoEmptyStaticBlock, staticBlockFile, testCase.sourceText),
				"unexpectedEmptyStaticBlock")
		})
	}
}

func TestNoEmptyStaticBlockStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a block with a statement", "export class Thing {\n    static ready = false;\n    static { Thing.ready = true; }\n}\n"},
		// The decision boundary, and the reason it sits here rather than at "is the block empty".
		// A comment is what lets a reader tell a deliberate no-op from a body that was deleted, so
		// it satisfies the rule exactly as it does for no-empty. The two rules agree on purpose.
		{"a block holding a line comment", "export class Thing {\n    static {\n        // nothing to initialize yet, see #123\n    }\n}\n"},
		{"a block holding a block comment", "export class Thing {\n    static {\n        /* deliberately empty */\n    }\n}\n"},
		// A comment marker inside a string earlier in the class must not read as a comment. This is
		// what the parser-trivia helper buys over a text search.
		{"a comment marker inside a string", "export class Thing {\n    static label = '/* not a comment */';\n    static { Thing.label = 'x'; }\n}\n"},
		{"a class with no static block", "export class Thing {\n    ready = false;\n}\n"},
		{"an empty class", "export class Thing {}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoEmptyStaticBlock, staticBlockFile, testCase.sourceText))
		})
	}
}
