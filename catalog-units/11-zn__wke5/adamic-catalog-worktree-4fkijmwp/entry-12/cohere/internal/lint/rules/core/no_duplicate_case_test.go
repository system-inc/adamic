package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoDuplicateCaseReportsRepeatedTest(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{"numeric literal", `switch (a) { case 1: break; case 1: break; }`, 1},
		{"string literal", `switch (a) { case "a": break; case "a": break; }`, 1},
		{"identifier", `switch (a) { case a: break; case a: break; }`, 1},
		{"separated by an unrelated arm", `switch (a) { case 1: break; case 2: break; case 1: break; }`, 1},
		{"three arms, two duplicates", `switch (a) { case 1: break; case 1: break; case 1: break; }`, 2},
		{"comment between case and test", `switch (a) { case /*a*/ 1: break; case 1: break; }`, 1},
		{"whitespace differs", `switch (a) { case 1  +  2: break; case 1 + 2: break; }`, 1},
		{"regex with a slash inside a character class", `switch (a) { case /[/*]/: break; case /[/*]/: break; }`, 1},
		{"member expression", `switch (a) { case Status.Ready: break; case Status.Ready: break; }`, 1},
		{"call expression, argument spacing differs", `switch (a) { case f(1, 2): break; case f(1,2): break; }`, 1},
		{"negative number", `switch (a) { case -1: break; case -1: break; }`, 1},
		{"template literal", "switch (a) { case `t${x}`: break; case `t${x}`: break; }", 1},
		{"duplicate alongside a default clause", `switch (a) { case 1: break; case 1: break; default: break; }`, 1},
		{"nested switch has its own duplicate", `switch (a) { case 1: switch (b) { case 2: break; case 2: break; } break; }`, 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDuplicateCase, "file.ts", testCase.source)
			wantIds := make([]string, testCase.want)
			for index := range wantIds {
				wantIds[index] = "unexpected"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// The finding must land on the test expression itself. A range that started at the node's Pos would
// begin inside the preceding trivia, which puts the finding on the comment above it and out of reach
// of a disable comment on the offending line.
func TestNoDuplicateCaseReportsTheTestExpressionWithoutTrivia(t *testing.T) {
	t.Parallel()

	const source = `switch (a) {
  case 1:
    break;
  // the copied arm
  case 1:
    break;
}`
	result := rule_testing.Run(t, NoDuplicateCase, "file.ts", source)
	rule_testing.ExpectFindings(t, result, "unexpected")

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "1" {
		t.Fatalf("reported range = %q, want %q", reported, "1")
	}
}

func TestNoDuplicateCaseAcceptsDistinctTests(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"distinct numbers", `switch (a) { case 1: break; case 2: break; }`},
		{"number and its string", `switch (a) { case 1: break; case "1": break; }`},
		{"one case and a default", `switch (a) { case 1: break; default: break; }`},
		{"distinct strings", `switch (a) { case "a": break; case "b": break; }`},
		{"distinct identifiers", `switch (a) { case a: break; case b: break; }`},
		{"identifier and the number it holds", `switch (a) { case one: break; case 1: break; }`},
		{"url inside a string is not a comment", `switch (a) { case "http://example.com": break; case "other": break; }`},
		{"comment syntax inside a string is content", `switch (a) { case "a /* b */": break; case "a": break; }`},
		{"strings differing only in whitespace", `switch (a) { case "hello  world": break; case "hello world": break; }`},
		{"distinct regexes", `switch (a) { case /foo/: break; case /bar/: break; }`},
		{"same regex, different flags", `switch (a) { case /foo/i: break; case /foo/g: break; }`},
		{"opposite signs", `switch (a) { case -1: break; case +1: break; }`},
		{"optional chain differs from a plain member access", `switch (a) { case x.y: break; case x?.y: break; }`},
		{"distinct member expressions", `switch (a) { case Status.Ready: break; case Status.Failed: break; }`},
		{"same callee, different argument", `switch (a) { case f(x): break; case f(y): break; }`},
		{"conditionals differing in one branch", `switch (a) { case x ? 1 : 2: break; case x ? 1 : 3: break; }`},
		{"typeof on different operands", `switch (a) { case typeof x: break; case typeof y: break; }`},
		{"empty switch", `switch (a) {}`},
		{"single case", `switch (a) { case 1: break; }`},
		{"only a default", `switch (a) { default: break; }`},
		{"same test in two separate switches", `switch (a) { case 1: break; } switch (b) { case 1: break; }`},
		{"same test in a nested switch", `switch (a) { case 1: switch (b) { case 1: break; } break; }`},
		{"template literals with different substitutions", "switch (a) { case `t${x}`: break; case `t${y}`: break; }"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDuplicateCase, "file.ts", testCase.source))
		})
	}
}
