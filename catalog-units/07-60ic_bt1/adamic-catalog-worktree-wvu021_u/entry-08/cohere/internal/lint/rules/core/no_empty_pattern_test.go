package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoEmptyPatternReportsEmptyPatterns is the fixture that must fire.
//
// Both shapes, at top level and nested, because the nested form is the one the rule is really
// about: `{ a: {} }` looks so much like `{ a = {} }` that it survives review, and it is the case
// ESLint's own documentation leads with.
func TestNoEmptyPatternReportsEmptyPatterns(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		wantId string
	}{
		{"const {} = foo;", "unexpectedObject"},
		{"const [] = foo;", "unexpectedArray"},
		{"const { a: {} } = foo;", "unexpectedObject"},
		{"const { a: [] } = foo;", "unexpectedArray"},
		{"function foo({}) {}", "unexpectedObject"},
		{"function bar([]) {}", "unexpectedArray"},
		{"function baz({ a: {} }) {}", "unexpectedObject"},
		{"function qux({ a: [] }) {}", "unexpectedArray"},
		{"const f = ({}) => {};", "unexpectedObject"},
		{"const g = ([]) => {};", "unexpectedArray"},
		{"let {} = foo;", "unexpectedObject"},
		{"var [] = foo;", "unexpectedArray"},
		{"for (const {} of list) {}", "unexpectedObject"},
	} {
		result := rule_testing.Run(t, NoEmptyPattern, "pattern.ts", testCase.source)
		rule_testing.ExpectFindings(t, result, testCase.wantId)
	}
}

// TestNoEmptyPatternStaysSilentOnRealBindings is the half that catches a rule firing on correct
// code.
//
// The default-value forms are the point. `{ a = {} }` and `{ a = [] }` contain the same two
// characters the rule reports elsewhere, in a position where they are a default rather than a
// pattern, and a rule that matched on the characters instead of the position would flag every one
// of them.
func TestNoEmptyPatternStaysSilentOnRealBindings(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"const { a } = foo;",
		"const [a] = foo;",
		"const { a = 1 } = foo;",
		"const [a = 1] = foo;",
		"const { a = {} } = foo;",
		"const { b = [] } = foo;",
		"function foo({ a }) {}",
		"function bar([a]) {}",
		"function baz({ a = {} }) {}",
		"function qux({ a = [] }) {}",
		"const { a: { b } } = foo;",
		"const { a: [b] } = foo;",
		"const { ...rest } = foo;",
		"const [, second] = foo;",
		"const empty = {};",
		"const emptyArray: string[] = [];",
		"foo({});",
	} {
		result := rule_testing.Run(t, NoEmptyPattern, "clean.ts", source)
		rule_testing.ExpectClean(t, result)
	}
}

// TestNoEmptyPatternWithoutOptionsIsStrict pins the default.
//
// The option defaults off, so a registry that forgets to wire it must produce the strict rule
// rather than a silent one. This is the fixture that would catch that inversion.
func TestNoEmptyPatternWithoutOptionsIsStrict(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoEmptyPattern, "strict.ts", "function foo({}) {}", nil)
	rule_testing.ExpectFindings(t, result, "unexpectedObject")
}

// TestNoEmptyPatternAllowsParameterObjectPatternsWhenConfigured pins what the option turns off.
func TestNoEmptyPatternAllowsParameterObjectPatternsWhenConfigured(t *testing.T) {
	t.Parallel()

	options := NoEmptyPatternOptions{AllowObjectPatternsAsParameters: true}
	for _, source := range []string{
		"function foo({}) {}",
		"const bar = function ({}) {};",
		"const qux = ({}) => {};",
		"function baz({} = {}) {}",
		"const quux = ({} = {}) => {};",
		"class C { method({}) {} }",
	} {
		result := rule_testing.RunWithOptions(t, NoEmptyPattern, "allowed.ts", source, options)
		rule_testing.ExpectClean(t, result)
	}
}

// TestNoEmptyPatternKeepsReportingDespiteOption pins the three cases the option does not reach.
//
// Each is a separate claim. An array pattern parameter still destructures and still throws on a
// non-iterable, so the option does not cover it. A nested pattern is a binding that was taken and
// discarded, not a parameter that takes none. And a non-empty default is evaluated at every call for
// a pattern that binds nothing, which is a bug wearing the option's clothes.
func TestNoEmptyPatternKeepsReportingDespiteOption(t *testing.T) {
	t.Parallel()

	options := NoEmptyPatternOptions{AllowObjectPatternsAsParameters: true}
	for _, testCase := range []struct {
		source string
		wantId string
	}{
		{"function baz([]) {}", "unexpectedArray"},
		{"const f = ([]) => {};", "unexpectedArray"},
		{"function foo({ a: {} }) {}", "unexpectedObject"},
		{"const bar = function ({ a: {} }) {};", "unexpectedObject"},
		{"function foo({} = { bar: 1 }) {}", "unexpectedObject"},
		{"function foo({} = bar) {}", "unexpectedObject"},
		{"const item = ({} = { bar: 1 }) => {};", "unexpectedObject"},
		{"const {} = foo;", "unexpectedObject"},
	} {
		result := rule_testing.RunWithOptions(t, NoEmptyPattern, "still.ts", testCase.source, options)
		rule_testing.ExpectFindings(t, result, testCase.wantId)
	}
}
