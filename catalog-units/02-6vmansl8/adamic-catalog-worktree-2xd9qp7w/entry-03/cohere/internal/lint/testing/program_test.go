package rule_testing

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// awaitedTypeIsThenable is a minimal type-aware rule, written to exercise the harness rather than
// to ship. It reports an await whose operand is not a Promise, which cannot be decided by looking
// at the syntax: `await value` is identical either way, and only the checker knows which it is.
var awaitedTypeIsThenable = rule.Rule{
	Name: "probe-await-thenable",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			// A type-aware rule handed no checker must decline rather than guess. This is exactly
			// the path that made a fixture pass vacuously under the syntax-only harness.
			return nil
		}
		return rule.Listeners{
			ast.KindAwaitExpression: func(node *ast.Node) {
				operand := node.AsAwaitExpression().Expression
				operandType := ctx.TypeChecker.GetTypeAtLocation(operand)
				if operandType == nil {
					return
				}
				// The checker narrows `const value = 3` to the literal type `3`, not to `number`.
				// That narrowing is itself evidence the checker is live rather than stubbed, and a
				// syntactic walk could not produce it.
				if isNumeric(ctx.TypeChecker.TypeToString(operandType)) {
					ctx.ReportNode(node, rule.Message{
						Id:          "awaitNonThenable",
						Description: "Awaiting a number does nothing.",
					})
				}
			},
		}
	},
}

// isNumeric accepts both `number` and a narrowed numeric literal type such as `3`.
func isNumeric(typeName string) bool {
	if typeName == "number" {
		return true
	}
	for _, character := range typeName {
		if character < '0' || character > '9' {
			return false
		}
	}
	return typeName != ""
}

// TestTypedHarnessSuppliesALiveChecker is the guard that makes every later type-aware fixture mean
// something.
//
// A harness that handed rules a nil checker would make each of them take its decline path, report
// nothing, and pass. Green, and having proven nothing. So this asserts the positive direction with
// a rule that can only answer by asking the checker.
func TestTypedHarnessSuppliesALiveChecker(t *testing.T) {
	t.Parallel()
	result := RunTyped(t, awaitedTypeIsThenable, "Subject.ts", `
		async function main() {
			const value = 3;
			return await value;
		}
	`)

	ExpectFindings(t, result, "awaitNonThenable")
}

// TestTypedHarnessStaysSilentOnCorrectCode is the other direction, and the half a violation-only
// corpus never has. A harness whose rule fired on everything would pass the test above.
func TestTypedHarnessStaysSilentOnCorrectCode(t *testing.T) {
	t.Parallel()
	result := RunTyped(t, awaitedTypeIsThenable, "Subject.ts", `
		async function main() {
			const value = Promise.resolve(3);
			return await value;
		}
	`)

	ExpectClean(t, result)
}

// TestSyntaxOnlyHarnessCannotProveATypeAwareRule states the defect this file exists to prevent, as
// an executable claim rather than as a comment.
//
// Run gives a rule no checker, so the same rule that fires above reports nothing here. A fixture
// written that way passes while exercising only the nil path, which is why a type-aware rule must
// use RunTyped and why that is worth failing loudly about rather than documenting.
func TestSyntaxOnlyHarnessCannotProveATypeAwareRule(t *testing.T) {
	t.Parallel()
	result := Run(t, awaitedTypeIsThenable, "Subject.ts", `
		async function main() {
			const value = 3;
			return await value;
		}
	`)

	if len(result.Diagnostics) != 0 {
		t.Fatalf("the syntax-only harness unexpectedly produced findings: %v", result.MessageIds())
	}
	// The point is not that this is correct behavior. It is that a green result here proves nothing
	// about the rule, and a reader has to know the difference.
}

// TestTypedHarnessResolvesAcrossFiles covers the questions a single-file program cannot pose:
// whether an imported symbol is what the rule thinks it is.
func TestTypedHarnessResolvesAcrossFiles(t *testing.T) {
	t.Parallel()
	result := RunTypedFiles(t, awaitedTypeIsThenable, map[string]string{
		"Helper.ts": `export const helperValue = 3;`,
		"Subject.ts": `
			import { helperValue } from './Helper';
			async function main() {
				return await helperValue;
			}
		`,
	}, "Subject.ts")

	// helperValue is a number declared in another file. Only a program that resolved the import can
	// answer that, so a finding here proves cross-file resolution rather than a lucky local guess.
	ExpectFindings(t, result, "awaitNonThenable")
}

// TestNarrowedLiteralTypesReachTheRule records what the checker actually reports, because it is a
// trap the next type-aware rule will hit.
//
// `const value = 3` has type `3`, not `number`. My first version of the probe rule above compared
// against "number" and reported nothing, which looked exactly like a harness handing over a nil
// checker. It was the opposite: the checker was live and doing real narrowing, and the rule's
// predicate was wrong.
//
// Worth an executable note rather than a comment, because the two failures are indistinguishable
// from the test output alone. A type-aware rule that reports nothing has either lost its checker or
// asked the wrong question, and knowing which is the difference between debugging the harness and
// debugging the rule.
func TestNarrowedLiteralTypesReachTheRule(t *testing.T) {
	t.Parallel()
	reported := []string{}
	recorder := rule.Rule{
		Name: "probe-type-names",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			if ctx.TypeChecker == nil {
				return nil
			}
			return rule.Listeners{
				ast.KindAwaitExpression: func(node *ast.Node) {
					operandType := ctx.TypeChecker.GetTypeAtLocation(node.AsAwaitExpression().Expression)
					reported = append(reported, ctx.TypeChecker.TypeToString(operandType))
				},
			}
		},
	}

	RunTyped(t, recorder, "Subject.ts", `
		async function main() {
			const literal = 3;
			const widened: number = 3;
			return [await literal, await widened];
		}
	`)

	if len(reported) != 2 {
		t.Fatalf("expected the checker to be asked twice, got %v", reported)
	}
	if reported[0] != "3" {
		t.Fatalf("expected the const to narrow to the literal type 3, got %q", reported[0])
	}
	if reported[1] != "number" {
		t.Fatalf("expected the annotated binding to stay number, got %q", reported[1])
	}
}
