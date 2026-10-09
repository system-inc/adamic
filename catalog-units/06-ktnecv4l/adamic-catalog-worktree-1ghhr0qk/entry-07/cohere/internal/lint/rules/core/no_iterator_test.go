package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// iteratorFile is where the fixtures pretend to live.
const iteratorFile = "/repository/source/Iterator.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_iterator.rs`:
// 4 pass, 5 fail, and the snapshot records 5 diagnostics from those 5 inputs, so one finding per
// input is right here rather than assumed. Copied because a fixture a porter invents encodes the
// same belief as the port, and the case that catches a bug is the one nobody would think to write.
func TestNoIteratorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a dotted access", "var a = test.__iterator__;"},
		{"an assignment to the prototype", "Foo.prototype.__iterator__ = function() {};"},
		{"a string-literal subscript", "var a = test['__iterator__'];"},
		{"a template-literal subscript", "var a = test[`__iterator__`];"},
		{"an assignment through a template subscript", "test[`__iterator__`] = function () {};"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoIterator, iteratorFile, testCase.sourceText), "noIterator")
		})
	}
}

// The clean cases are the whole discrimination, and each fails a different way.
//
// The first reads a *variable* named `__iterator__`, so the property accessed is whatever that
// variable holds and is not this one. The second is a declaration rather than an access. The third
// is a different key that merely starts the same. The fourth is the sharpest: its text does contain
// `__iterator__`, followed by a newline, so any rule matching on source text reports it.
func TestNoIteratorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a computed access through a variable", "var a = test[__iterator__];"},
		{"a declaration binding the name", "var __iterator__ = null;"},
		{"a template naming a different key", "foo[`__iterator`] = null;"},
		{"a template whose text merely contains it", "foo[`__iterator__\n`] = null;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoIterator, iteratorFile, testCase.sourceText))
		})
	}
}

// A case written because somebody read our code rather than upstream's.
//
// The imported corpus is a floor and not a ceiling: it exercises no substituting template, and a
// port reading a template's cooked text without checking for substitutions would report this while
// upstream does not. `staticPropertyName` declines it by kind, and this pins that.
func TestNoIteratorDeclinesASubstitutingTemplate(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoIterator, iteratorFile,
		"declare const part: string;\nexport const a = test[`__iterator${part}__`];\n"))
}

// The suggestion's span, which the message-id fixtures above cannot see.
//
// A mutation replacing the whole node rather than the tail from the object's end compiled and
// changed no fixture, because every assertion here checks which message fired and none checked what
// the repair would do. That is the gap a reviewer named without reproducing, and it was real.
//
// Asserted by applying the suggestion rather than by comparing offsets: offsets are the thing most
// likely to be wrong in the same direction as the code that produced them.
func TestNoIteratorSuggestsTheRightSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		// The dot goes with the property, which is what makes the tail-from-object span right rather
		// than merely convenient.
		{"a dotted access", "var a = test.__iterator__;", "var a = test[Symbol.iterator];"},
		// The brackets go too, and the same span covers it with no second arm.
		{"a string subscript", "var a = test['__iterator__'];", "var a = test[Symbol.iterator];"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoIterator, iteratorFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
				t.Fatalf("wanted one suggestion carrying one fix, got %d suggestions", len(suggestions))
			}

			fix := suggestions[0].Fixes[0]
			rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
				testCase.sourceText[fix.Range.End():]
			if rewritten != testCase.wantSource {
				t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, testCase.wantSource)
			}
		})
	}
}
