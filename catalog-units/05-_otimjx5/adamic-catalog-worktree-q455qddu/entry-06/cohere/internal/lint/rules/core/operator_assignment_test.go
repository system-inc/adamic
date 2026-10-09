package core

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// operatorAssignmentFile is where the fixtures pretend to live.
const operatorAssignmentFile = "/repository/source/OperatorAssignment.ts"

// decodedOperatorAssignmentOptions routes a fixture's options through the rule's own decoder rather
// than building the struct directly.
//
// The default is `always`, so a struct built by hand would carry a non-nil setting and every case
// would pass while the live config's bare "error" still handed the rule a nil it read as silence.
// An empty string here means that bare configuration.
func decodedOperatorAssignmentOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeOperatorAssignmentOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/operator-assignment.js`, and it was not
// transcribed. Upstream's tester file was LOADED with a stub RuleTester that captured the case
// objects, so every string here is the cooked value upstream's own tester would have used and no
// escape could be changed on the way in. 50 clean cases and 69 reporting ones, which is the whole
// file.
//
// The extraction was then replayed against the installed eslint at 10.8.1: all 119 cases agree on
// verdict, on message id, and on whether the case is fixed or declined.
//
// 24 of the 69 carry `output: null`, meaning upstream reports and deliberately withholds the
// repair. Those are asserted as declines rather than dropped, because a fixer that repairs a case
// upstream refuses to touch is a defect no message-id fixture can see.
func TestOperatorAssignmentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		options      string
		wantIds      []string
		wantOperator []string
		wantFixed    string
		fixable      bool
	}{
		{"x = x + y", "", []string{"replaced"}, []string{"+="}, "x += y", true},
		{"x = x - y", "", []string{"replaced"}, []string{"-="}, "x -= y", true},
		{"x = x * y", "", []string{"replaced"}, []string{"*="}, "x *= y", true},
		{"x = y * x", "", []string{"replaced"}, []string{"*="}, "", false},
		{"x = (y * z) * x", "", []string{"replaced"}, []string{"*="}, "", false},
		{"x = x / y", "", []string{"replaced"}, []string{"/="}, "x /= y", true},
		{"x = x % y", "", []string{"replaced"}, []string{"%="}, "x %= y", true},
		{"x = x << y", "", []string{"replaced"}, []string{"<<="}, "x <<= y", true},
		{"x = x >> y", "", []string{"replaced"}, []string{">>="}, "x >>= y", true},
		{"x = x >>> y", "", []string{"replaced"}, []string{">>>="}, "x >>>= y", true},
		{"x = x & y", "", []string{"replaced"}, []string{"&="}, "x &= y", true},
		{"x = x ^ y", "", []string{"replaced"}, []string{"^="}, "x ^= y", true},
		{"x = x | y", "", []string{"replaced"}, []string{"|="}, "x |= y", true},
		{"x[0] = x[0] - y", "", []string{"replaced"}, []string{"-="}, "x[0] -= y", true},
		{"x.y[z['a']][0].b = x.y[z['a']][0].b * 2", "", []string{"replaced"}, []string{"*="}, "", false},
		{"x = x + y", "\"always\"", []string{"replaced"}, []string{"+="}, "x += y", true},
		{"x = (x + y)", "\"always\"", []string{"replaced"}, []string{"+="}, "x += y", true},
		{"x = x + (y)", "\"always\"", []string{"replaced"}, []string{"+="}, "x += (y)", true},
		{"x += (y)", "\"never\"", []string{"unexpected"}, []string{"+="}, "x = x + (y)", true},
		{"x += y", "\"never\"", []string{"unexpected"}, []string{"+="}, "x = x + y", true},
		{"foo.bar = foo.bar + baz", "", []string{"replaced"}, []string{"+="}, "foo.bar += baz", true},
		{"foo.bar += baz", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo.bar = foo.bar + baz", true},
		{"this.foo = this.foo + bar", "", []string{"replaced"}, []string{"+="}, "this.foo += bar", true},
		{"this.foo += bar", "\"never\"", []string{"unexpected"}, []string{"+="}, "this.foo = this.foo + bar", true},
		{"foo.bar.baz = foo.bar.baz + qux", "", []string{"replaced"}, []string{"+="}, "", false},
		{"foo.bar.baz += qux", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"this.foo.bar = this.foo.bar + baz", "", []string{"replaced"}, []string{"+="}, "", false},
		{"this.foo.bar += baz", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"foo[bar] = foo[bar] + baz", "", []string{"replaced"}, []string{"+="}, "", false},
		{"this[foo] = this[foo] + bar", "", []string{"replaced"}, []string{"+="}, "", false},
		{"foo[bar] >>>= baz", "\"never\"", []string{"unexpected"}, []string{">>>="}, "", false},
		{"this[foo] >>>= bar", "\"never\"", []string{"unexpected"}, []string{">>>="}, "", false},
		{"foo[5] = foo[5] / baz", "", []string{"replaced"}, []string{"/="}, "foo[5] /= baz", true},
		{"this[5] = this[5] / foo", "", []string{"replaced"}, []string{"/="}, "this[5] /= foo", true},
		{"/*1*/x/*2*/./*3*/y/*4*/= x.y +/*5*/z/*6*/./*7*/w/*8*/;", "\"always\"", []string{"replaced"}, []string{"+="}, "/*1*/x/*2*/./*3*/y/*4*/+=/*5*/z/*6*/./*7*/w/*8*/;", true},
		{"x // 1\n . // 2\n y // 3\n = x.y + //4\n z //5\n . //6\n w;", "\"always\"", []string{"replaced"}, []string{"+="}, "x // 1\n . // 2\n y // 3\n += //4\n z //5\n . //6\n w;", true},
		{"x = /*1*/ x + y", "\"always\"", []string{"replaced"}, []string{"+="}, "", false},
		{"x = //1\n x + y", "\"always\"", []string{"replaced"}, []string{"+="}, "", false},
		{"x.y = x/*1*/.y + z", "\"always\"", []string{"replaced"}, []string{"+="}, "", false},
		{"x.y = x. //1\n y + z", "\"always\"", []string{"replaced"}, []string{"+="}, "", false},
		{"x = x /*1*/ + y", "\"always\"", []string{"replaced"}, []string{"+="}, "", false},
		{"x = x //1\n + y", "\"always\"", []string{"replaced"}, []string{"+="}, "", false},
		{"/*1*/x +=/*2*/y/*3*/;", "\"never\"", []string{"unexpected"}, []string{"+="}, "/*1*/x = x +/*2*/y/*3*/;", true},
		{"x +=//1\n y", "\"never\"", []string{"unexpected"}, []string{"+="}, "x = x +//1\n y", true},
		{"(/*1*/x += y)", "\"never\"", []string{"unexpected"}, []string{"+="}, "(/*1*/x = x + y)", true},
		{"x/*1*/+=  y", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"x //1\n +=  y", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"(/*1*/x) +=  y", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"x/*1*/.y +=  z", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"x.//1\n y +=  z", "\"never\"", []string{"unexpected"}, []string{"+="}, "", false},
		{"(foo.bar) ^= ((((((((((((((((baz))))))))))))))))", "\"never\"", []string{"unexpected"}, []string{"^="}, "(foo.bar) = (foo.bar) ^ ((((((((((((((((baz))))))))))))))))", true},
		{"foo = foo ** bar", "", []string{"replaced"}, []string{"**="}, "foo **= bar", true},
		{"foo **= bar", "\"never\"", []string{"unexpected"}, []string{"**="}, "foo = foo ** bar", true},
		{"foo *= bar + 1", "\"never\"", []string{"unexpected"}, []string{"*="}, "foo = foo * (bar + 1)", true},
		{"foo -= bar - baz", "\"never\"", []string{"unexpected"}, []string{"-="}, "foo = foo - (bar - baz)", true},
		{"foo += bar + baz", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo = foo + (bar + baz)", true},
		{"foo += bar = 1", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo = foo + (bar = 1)", true},
		{"foo *= (bar + 1)", "\"never\"", []string{"unexpected"}, []string{"*="}, "foo = foo * (bar + 1)", true},
		{"foo+=-bar", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo= foo+-bar", true},
		{"foo/=bar", "\"never\"", []string{"unexpected"}, []string{"/="}, "foo= foo/bar", true},
		{"foo/=/**/bar", "\"never\"", []string{"unexpected"}, []string{"/="}, "foo= foo/ /**/bar", true},
		{"foo/=//\nbar", "\"never\"", []string{"unexpected"}, []string{"/="}, "foo= foo/ //\nbar", true},
		{"foo/=/^bar$/", "\"never\"", []string{"unexpected"}, []string{"/="}, "foo= foo/ /^bar$/", true},
		{"foo+=+bar", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo= foo+ +bar", true},
		{"foo+= +bar", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo= foo+ +bar", true},
		{"foo+=/**/+bar", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo= foo+/**/+bar", true},
		{"foo+=+bar===baz", "\"never\"", []string{"unexpected"}, []string{"+="}, "foo= foo+(+bar===baz)", true},
		{"(obj?.a).b = (obj?.a).b + y", "", []string{"replaced"}, []string{"+="}, "", false},
		{"obj.a = obj?.a + b", "", []string{"replaced"}, []string{"+="}, "", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name+" ["+testCase.options+"]", func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OperatorAssignment, operatorAssignmentFile,
				testCase.name, decodedOperatorAssignmentOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			// The rendered text, not only the id. Both messages interpolate the operator, and a
			// rule naming the wrong one renders a sentence that reads perfectly.
			for index, diagnostic := range result.Diagnostics {
				var want string
				switch testCase.wantIds[index] {
				case "replaced":
					want = operatorAssignmentReplacedMessage(testCase.wantOperator[index]).Description
				default:
					want = operatorAssignmentUnexpectedMessage(testCase.wantOperator[index]).Description
				}
				if got := diagnostic.Message.Description; got != want {
					t.Fatalf("finding %d message:\n got %q\nwant %q", index, got, want)
				}
			}

			fixCount := 0
			for _, diagnostic := range result.Diagnostics {
				fixCount += len(diagnostic.Fixes)
			}

			if !testCase.fixable {
				if fixCount != 0 {
					t.Fatalf("upstream declines to fix this case, but the rule proposed %d fix(es)", fixCount)
				}
				return
			}
			if fixCount == 0 {
				t.Fatalf("upstream fixes this case and the rule proposed no repair")
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

func TestOperatorAssignmentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options string
	}{
		{"x = y", ""},
		{"x = y + x", ""},
		{"x += x + y", ""},
		{"x = (x + y) - z", ""},
		{"x -= y", ""},
		{"x = y - x", ""},
		{"x *= x", ""},
		{"x = y * z", ""},
		{"x = (x * y) * z", ""},
		{"x = y / x", ""},
		{"x /= y", ""},
		{"x %= y", ""},
		{"x <<= y", ""},
		{"x >>= x >> y", ""},
		{"x >>>= y", ""},
		{"x &= y", ""},
		{"x **= y", ""},
		{"x ^= y ^ z", ""},
		{"x |= x | y", ""},
		{"x = x && y", ""},
		{"x = x || y", ""},
		{"x = x < y", ""},
		{"x = x > y", ""},
		{"x = x <= y", ""},
		{"x = x >= y", ""},
		{"x = x instanceof y", ""},
		{"x = x in y", ""},
		{"x = x == y", ""},
		{"x = x != y", ""},
		{"x = x === y", ""},
		{"x = x !== y", ""},
		{"x[y] = x['y'] + z", ""},
		{"x.y = x['y'] / z", ""},
		{"x.y = z + x.y", ""},
		{"x[fn()] = x[fn()] + y", ""},
		{"x += x + y", "\"always\""},
		{"x = x + y", "\"never\""},
		{"x = x ** y", "\"never\""},
		{"x = y ** x", ""},
		{"x = x * y + z", ""},
		{"this.x = this.y + z", "\"always\""},
		{"this.x = foo.x + y", "\"always\""},
		{"this.x = foo.this.x + y", "\"always\""},
		{"const foo = 0; class C { foo = foo + 1; }", ""},
		{"x = x && y", "\"always\""},
		{"x = x || y", "\"always\""},
		{"x = x ?? y", "\"always\""},
		{"x &&= y", "\"never\""},
		{"x ||= y", "\"never\""},
		{"x ??= y", "\"never\""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name+" ["+testCase.options+"]", func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OperatorAssignment, operatorAssignmentFile,
				testCase.name, decodedOperatorAssignmentOptions(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestOperatorAssignmentTypeScriptShapes covers what upstream's corpus structurally cannot.
//
// Upstream's tests are JavaScript, so no imported case carries a type assertion, a type argument or
// a non-null operator. This project has now lost type information twice to fixers that were correct
// about JavaScript, so these shapes are enumerated rather than assumed safe.
//
// Both fixers copy BYTES out of the source rather than re-rendering the construct, which is what
// makes most of these safe by construction: a generic argument or an as-expression inside a copied
// slice survives because nothing looked at it. That is worth pinning, because the natural way to
// write either fixer is to rebuild the expression from its parts, and that is how the two earlier
// incidents happened.
//
// The `as` and `satisfies` rows are a defect this found rather than a property it confirms. Both
// bind LOOSER than every arithmetic operator, so before the precedence table carried a row for them
// they were ranked tightest and the expansion silently re-associated: `x += 1 as number` became
// `x = x + 1 as number`, which asserts the type of the SUM rather than of the addend. The two
// parses were compared directly to establish that, rather than read off the grammar.
func TestOperatorAssignmentTypeScriptShapes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		options    string
		wantId     string
		wantFixed  string
		fixable    bool
	}{
		{"an as-expression inside the right operand", "x = x + (1 as number)", "", "replaced", "x += (1 as number)", true},
		{"a satisfies expression inside the right operand", "x = x + (1 satisfies number)", "", "replaced", "x += (1 satisfies number)", true},
		{"a generic call on the right", "x = x + f<number>(1)", "", "replaced", "x += f<number>(1)", true},
		{"a generic call expanded", "x += f<number>(1)", "\"never\"", "unexpected", "x = x + f<number>(1)", true},
		{"an as-expression needs parentheses when expanded", "x += 1 as number", "\"never\"", "unexpected", "x = x + (1 as number)", true},
		{"a satisfies expression needs them too", "x += 1 satisfies number", "\"never\"", "unexpected", "x = x + (1 satisfies number)", true},
		{"a non-null assertion on the target reports without a repair", "x! += 1", "\"never\"", "unexpected", "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, OperatorAssignment, operatorAssignmentFile,
				testCase.sourceText, decodedOperatorAssignmentOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.wantId)

			fixCount := 0
			for _, diagnostic := range result.Diagnostics {
				fixCount += len(diagnostic.Fixes)
			}
			if !testCase.fixable {
				if fixCount != 0 {
					t.Fatalf("expected no repair, got %d", fixCount)
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}
