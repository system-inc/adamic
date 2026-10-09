package typescript

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnsafeUnaryMinus flags unary negation applied to a value that is not a number or a bigint.
//
//	valid:   -42
//	valid:   -42n
//	valid:   (a: number | bigint) => -a
//	valid:   (a: any) => -a
//	valid:   (a: never) => -a
//	valid:   <T extends number>(t: T) => -t
//	valid:   (a: string) => +a
//	invalid: (a: string) => -a
//	invalid: (a: {}) => -a
//	invalid: (a: unknown) => -a
//	invalid: <T,>(t: T) => -t
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/no_unsafe_unary_minus/no_unsafe_unary_minus.go`, vendored at
// commit `05b7fbc` and absorbed onto cohere's own rule interface here. It reaches for
// `GetConstrainedTypeAtLocation` and `UnionTypeParts` because that is what upstream reaches for,
// and this note is why a reader finds those helpers in a file that otherwise looks native.
//
// tsgolint is not re-synced, so this file is now the only copy of the algorithm rather than a
// translation layer over a vendored one. The checker logic below is byte-identical to upstream's;
// what changed is the interface it speaks: `rule.RuleContext` became `rule.Context`,
// `RuleListeners` became `Listeners`, and `rule.RuleMessage` became `rule.Message`. No predicate,
// no flag set and no traversal was touched.
//
// oxc has no implementation. `no_unsafe_unary_minus.rs` is sixty lines of documentation ending in
// `impl Rule for NoUnsafeUnaryMinus {}`, an empty body, with the declaration `NoUnsafeUnaryMinus`
// carrying the `(tsgolint)` marker. There is no algorithm and no corpus on that side at all, so the
// behavior oxlint exhibits for this rule IS tsgolint's: the release binary shells out to a
// `tsgolint` executable and refuses with `Failed to find tsgolint executable` when it is absent,
// which is what it does on this machine. That refusal is the proof rather than an obstacle.
//
// # Where the two references disagree: the message text, not the verdict
//
// Both implementations register exactly ONE listener, on prefix unary expressions, and both declare
// `schema: []`, so there is no option surface on either side and nothing for the config to carry.
// That is the first thing checked, because await-thenable's port found `@typescript-eslint`
// carrying a fourth listener tsgolint has no counterpart for. Here the arms match.
//
// The verdicts match too. All fifty four inputs below were driven through `@typescript-eslint`
// 8.67.0 on a real program and through the adapted rule here, and all fifty four agree on whether
// to report and on where the finding points, byte for byte on the reported span.
//
// What differs is the type NAMED in the message, and it differs on five of the twenty four
// reporting inputs. tsgolint reports the offending union PART, `ctx.TypeChecker.TypeToString(t)`
// inside the loop, while `@typescript-eslint` reports the WHOLE argument type,
// `checker.typeToString(argType)` outside it. Measured:
//
//	(a: number | string) => -a          tsgolint: "is string"   ts-eslint: "is string | number"
//	(a: number|string|bigint) => -a     tsgolint: "is string"   ts-eslint: "is string | number | bigint"
//	<T extends number|string>(t: T)     tsgolint: "is string"   ts-eslint: "is string | number"
//	declare const b: boolean; -b        tsgolint: "is false"    ts-eslint: "is boolean"
//
// The `boolean` row is the one worth staring at, because it was not predicted from reading either
// source. `boolean` is internally the union `false | true`, so the union walk splits it and names
// the first part. Nothing about the input looks like a union, and no fixture asserting a message id
// or a count could ever see it. That is why the test file asserts rendered text.
//
// tsgolint wins, and this is reproduced rather than corrected: oxlint is what the differential
// harness compares against, and oxlint runs tsgolint, so naming the whole type would report text
// the gate does not.
//
// # What decides a finding, established by probing flags rather than by reasoning about them
//
// The operand's type goes through `GetConstrainedTypeAtLocation`, which resolves a type parameter
// to its base constraint and falls back to the raw type, and then through `UnionTypeParts`. Each
// part must carry one of `Any | Never | BigIntLike | NumberLike` or the rule reports and stops at
// the first offender, so one expression yields at most one finding.
//
// The union walk is load-bearing rather than a convenience, and there is an input that proves it:
// `number | bigint` has flags `Union` and NOTHING else. Not `NumberLike`, not `BigIntLike`. A
// version asking the union's own flags would report the safe case that upstream's corpus lists as
// valid. Measured with a flags probe rather than reasoned about, because the two spellings look
// interchangeable in source.
//
// The boundary, all measured the same way:
//
//	number | bigint       Union alone on the whole, NumberLike and BigIntLike on the parts   silent
//	number | string       parts split, `string` carries String and reports                   reports
//	any                   Any                                                                silent
//	unknown               Unknown, which is NOT Any                                          reports
//	never                 Never                                                              silent
//	5                     NumberLiteral, inside NumberLike                                   silent
//	enum E { A = 1 }      E.A is NumberLiteral|EnumLiteral, and the whole enum splits into
//	                      its members, each NumberLike                                       silent
//	enum S { A = 'a' }    StringLiteral|EnumLiteral, not a union, not NumberLike             reports
//	Number, BigInt        Object, because a boxed wrapper is not NumberLike                  reports
//	boolean               splits into false | true, both BooleanLiteral                      reports
//	T extends number      constrained to number before the walk                              silent
//	T unconstrained       stays TypeParameter                                                reports
//	T extends any         constrained to UNKNOWN, not to any                                 reports
//	string | any          collapses to `any` at the checker, so no union survives            silent
//
// The `T extends any` row contradicts the natural reading twice over: the constraint is written as
// `any` and resolves to `unknown`, which is the one intrinsic in this list that is not exempt.
//
// # The error type is `any`, so a fixture touching a missing type asserts the opposite of upstream
//
// The harness tsconfig pins `lib: ["ES2022"]` and cannot be raised. A type it does not carry
// resolves to the error type, whose flags are exactly `Any`, and `Any` is in the safe set. So
// `declare const x: Disposable; -x;` and `declare const x: NotDefinedAnywhere; -x;` both go SILENT
// here, and both were measured that way rather than assumed. For this rule that failure runs in the
// direction that hides: a fixture naming a type beyond ES2022 lands in the clean list, passes, and
// asserts nothing. Every fixture in the test file was checked against that hazard, and the corpus
// happens to name no such type, so none needed a second file.
//
// # The checker, and the nil guard that now lives here
//
// The single listener reads `ctx.TypeChecker` unconditionally, so the fixtures use `RunTyped` and
// the rule declares `NeedsTypeChecker`. While this rule was adapted, that declaration was made by
// `upstream.Adapt` on every rule it wrapped and the standing `if ctx.TypeChecker == nil { return }`
// could not be written, because the listener was upstream's and editing it would have turned a
// re-sync into a merge. Absorbing the rule removes that constraint: the listener is ours now, so
// the guard is written where the advice always wanted it, and it is the one addition to the body.
//
// The guard is unreachable through registration, since `NeedsTypeChecker` is declared right above.
// It is here for the harness path, where a Context can be built by hand, and because the dangerous
// direction for this rule is silence rather than a panic: under an untyped harness the rule would
// report nothing and every clean fixture would pass having proven nothing. A test in this package
// pins the declaration so a later revert fails loudly instead of going vacuously green.
//
// # Cost
//
// The anchor is `KindPrefixUnaryExpression` and the operator test is the first line, so every `+`,
// `!`, `~`, `++` and `--` exits before the checker is touched. Only `-` pays for a type query, and
// a negated numeric literal is the overwhelmingly common case in real source.
var NoUnsafeUnaryMinus = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-unary-minus",

	// The listener reads ctx.TypeChecker on every `-` operand, so the checker is required rather
	// than opportunistic.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				expr := node.AsPrefixUnaryExpression()

				if expr.Operator != ast.KindMinusToken {
					return
				}

				argType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, expr.Operand)

				for t := range type_checking.UnionTypePartsSeq(argType) {
					if !type_checking.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsNever|checker.TypeFlagsBigIntLike|checker.TypeFlagsNumberLike) {
						ctx.ReportNode(node, buildUnaryMinusMessage(ctx.TypeChecker.TypeToString(t)))
						break
					}
				}
			},
		}
	},
}

// buildUnaryMinusMessage names the offending union PART rather than the whole argument type, which
// is tsgolint's behavior and differs observably from `@typescript-eslint`. See the note above.
func buildUnaryMinusMessage(t string) rule.Message {
	return rule.Message{
		Id:          "unaryMinus",
		Description: fmt.Sprintf("Argument of unary negation should be assignable to number | bigint but is %v instead.", t),
	}
}
