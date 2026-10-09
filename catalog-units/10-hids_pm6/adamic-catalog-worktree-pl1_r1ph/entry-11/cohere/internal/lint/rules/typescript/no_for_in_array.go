package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoForInArray flags a `for...in` loop whose subject is array-like.
//
//	valid:   for (const x of [3, 4, 5]) {}
//	valid:   for (const x in { a: 1, b: 2, c: 3 }) {}
//	valid:   declare const obj: { [key: number]: number }; for (const key in obj) {}
//	invalid: for (const x in [3, 4, 5]) {}
//	invalid: declare const array: [number, string]; for (const key in array) {}
//	invalid: function f() { for (const a in arguments) {} }
//
// A for-in loop over an array skips holes, hands back indices as strings rather than numbers, and
// walks the prototype chain along with any other enumerable property, so it is almost never what
// the author meant.
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/no_for_in_array/no_for_in_array.go`, vendored at commit
// `05b7fbc` and absorbed onto cohere's own rule interface here. It reaches for
// `GetConstrainedTypeAtLocation`, `TypeRecurser`, `GetNumberIndexType` and `GetForStatementHeadLoc`
// because that is what upstream reaches for, and this note is why a reader finds those helpers in a
// file that otherwise looks native.
//
// tsgolint is not re-synced, so this file is now the only copy of the algorithm rather than a
// translation layer over a vendored one. The checker logic below is byte-identical to upstream's;
// what changed is the interface it speaks: `rule.RuleContext` became `rule.Context`,
// `RuleListeners` became `Listeners`, and `rule.RuleMessage` became `rule.Message`. No predicate,
// no flag set and no traversal was touched.
//
// oxc has no implementation. `no_for_in_array.rs` declares `NoForInArray(tsgolint)` and its body is
// `impl Rule for NoForInArray {}`, sixty four lines of documentation around an empty block. There
// is no oxc corpus either, so the extractor has nothing to extract. oxlint's behavior for this rule
// IS tsgolint's, by delegation, and on this machine oxlint refuses to run it at all with `Failed to
// find tsgolint executable` rather than falling back to something of its own.
//
// # The one correction to upstream: it misspells its own rule name
//
// Upstream declares `Name: "no-for-in-array-rule"`. Every other rule in tsgolint spells its name as
// its directory with underscores turned to dashes, and this one appends "-rule".
//
// That was measured rather than eyeballed: all 38 rules under tsgolint's `internal/rules/` were
// fetched and each Name compared against its directory, and this is the only disagreement in the
// family. It is live upstream rather than absorbed downstream, because `cmd/tsgolint/main.go`
// builds its `ConfiguredRule` with `Name: r.Name` and no mapping table.
//
// The suffix cannot ship here. cohere keys the catalog, the config file, and suppression comments
// on this exact string, so `no-for-in-array-rule` would register a rule that
// `typescript/no-for-in-array` in `CohereSettings.json` never enables, that the inventory entry
// never matches, and that no suppression comment an author would plausibly write could silence. It
// is corrected below, at the line, with the measurement recorded beside it.
//
// # Where the two references disagree: they do not
//
// This is the opposite of what `await-thenable`'s port found, and it was established by running
// both rather than by reading them. `@typescript-eslint` 8.67.0 was driven through its Linter API
// on a real program over all 24 of tsgolint's own cases, and it agreed on every one: same verdicts,
// same finding counts, and the same line and column spans on all 20 diagnostics. Zero
// disagreements.
//
// Listener counts match too, which is the check `await-thenable` earned. Both sides register
// exactly ONE listener on the for-in statement. `@typescript-eslint` declares `schema: []`, so
// there is no option surface on either side, and the message id and its full text are identical
// strings. There is no fourth arm here of the kind `await-thenable` had.
//
// # What decides a finding
//
// The subject's type is resolved through its base constraint first, so a type parameter reports
// exactly when what it is constrained to would: `<T extends any[]>(arr: T)` fires. The predicate
// then recurses through unions and intersections and asks two things of each part, both of which
// must hold: it has a NUMBER index signature, and it has a `length` property whose type is
// number-like.
//
// Both halves are load-bearing and the corpus pins each one from the other side. `{ [key: number]:
// number }` has the index signature and no length, and it is a PASSING case; adding `length: 1` to
// that same object is a FAILING case. And because the recursion is a `some` rather than an `every`,
// a union reports when any part is array-like, which is why `string[] | null` and `boolean[] | { a:
// 1 }` both fire.
//
// So the set of things that report is wider than "an array", and it was read off the corpus rather
// than off the rule's name: arrays, tuples, `RegExpExecArray`, `HTMLCollection`, `NodeList`, an
// `arguments` object, a union of an array with a plain object, a type parameter constrained to an
// array, and any hand-written object carrying both a number index and a numeric length. A string is
// NOT among them and neither is a `Set`; upstream writes no case for either, and neither has a
// number index signature.
//
// # It reports and offers nothing
//
// No fix and no suggestion, on either side. `@typescript-eslint` reports with a bare `loc` and
// `messageId`, and the vendored rule calls `ctx.ReportRange` rather than any of the fix-carrying
// forms. Rewriting a for-in into a for-of changes what the loop binds, from a string key to a
// value, so there is no repair that preserves meaning and nothing to apply unattended.
//
// # The span is the head of the loop, not the statement
//
// Upstream reports `GetForStatementHeadLoc`, which runs from the `for` keyword to the start of the
// loop BODY, deliberately excluding the body itself. It is also the reason the node-report trivia
// question does not arise for this rule: this is a `ReportRange` of a range the helper computed, so
// the `TokenRange` trimming that `ReportNode` applies is not in the path at all, and the trimming
// that does happen is `TrimNodeTextRange` inside `GetForStatementHeadLoc`. Absorbing the rule
// therefore cannot move this span: neither the adapter nor our own node helpers ever touched it.
//
// That span is asserted rather than assumed. All 18 of upstream's single-file diagnostics are
// replayed in the test file against their exact line and column, including the two cases where the
// head runs across ten lines through nested parentheses and trailing comments.
//
// # The checker, and the nil guard that now lives here
//
// The listener reads `ctx.TypeChecker` unconditionally, so this needs the checker and its fixtures
// use `RunTyped`. While this rule was adapted, the standing `if ctx.TypeChecker == nil { return }`
// could not be written, because the listener was upstream's and editing it would have turned a
// re-sync into a merge. `upstream.Adapt` set `NeedsTypeChecker` on every rule it wrapped, so the
// nil case was unreachable through it. Absorbing the rule removes that constraint AND makes the nil
// case reachable, so the guard is now written where the advice always wanted it, and it is the one
// addition to the body.
//
// It matters more than a crash-avoidance would suggest. `GetSymbolAtLocation` and its neighbours on
// this shim return nil rather than panicking under a nil checker, so a missing guard here does not
// crash — it buys a VACUOUS GREEN, where every clean fixture passes having proven nothing and every
// reporting one fails. A test in this package pins the declaration and the guard together so a
// later revert fails loudly.
var NoForInArray = rule.Rule{
	// Upstream spells this "no-for-in-array-rule". That trailing "-rule" is a typo, and correcting
	// it is the one correction this absorption carries.
	//
	// It is not cosmetic. cohere keys the catalog, the config, and suppression comments on this
	// string, so shipping upstream's spelling would register a rule no `CohereSettings.json` entry
	// enables, that the inventory's `typescript/no-for-in-array` never matches, and that no
	// `cohere-disable` comment an author would actually write could silence.
	//
	// Measured rather than assumed before editing: all 38 rules in tsgolint's `internal/rules/` were
	// fetched and their Name compared against their directory, and this is the ONLY one that
	// disagrees. tsgolint's own `cmd/tsgolint/main.go` passes `r.Name` straight through to its
	// reporter with no mapping table, so upstream really does emit the suffixed name; the typo is
	// live there rather than absorbed somewhere downstream.
	//
	// Both other references spell it without the suffix: oxc declares `NoForInArray(tsgolint)` and
	// `@typescript-eslint` 8.67.0 declares `name: 'no-for-in-array'`.
	Name: "@typescript-eslint/no-for-in-array",

	// The listener resolves the subject's type on every for-in statement, so the checker is required
	// rather than opportunistic.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		hasArrayishLength := func(t *checker.Type) bool {
			lengthProperty := checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "length")
			if lengthProperty == nil {
				return false
			}

			return type_checking.IsTypeFlagSet(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, lengthProperty), checker.TypeFlagsNumberLike)
		}
		isArrayLike := func(t *checker.Type) bool {
			return type_checking.TypeRecurser(t, func(t *checker.Type) bool {
				return type_checking.GetNumberIndexType(ctx.TypeChecker, t) != nil && hasArrayishLength(t)
			})
		}

		return rule.Listeners{
			ast.KindForInStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				t := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, node.AsForInOrOfStatement().Expression)

				if isArrayLike(t) {
					ctx.ReportRange(
						type_checking.GetForStatementHeadLoc(ctx.SourceFile, node),
						buildForInViolationMessage(),
					)
				}
			},
		}
	},
}

// buildForInViolationMessage is upstream's message, text unchanged.
func buildForInViolationMessage() rule.Message {
	return rule.Message{
		Id:          "forInViolation",
		Description: "For-in loops over arrays skips holes, returns indices as strings, and may visit the prototype chain or other enumerable properties. Use a more robust iteration method such as for-of or array.forEach instead.",
	}
}
