package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// RequireArraySortCompareOptions is the rule's option surface.
//
// One key, and it DEFAULTS TO TRUE, which is the whole reason this rule hand-rolls its decoder
// below rather than reaching for the generic helper.
type RequireArraySortCompareOptions struct {
	// IgnoreStringArrays exempts an array whose every element type is a string.
	//
	// Sorting strings with no comparator is the one case where the default sort does what a reader
	// expects, since the default converts to string and compares code units, which for values that
	// are already strings is ordinary lexicographic order.
	IgnoreStringArrays bool
}

// DefaultRequireArraySortCompareSettings is upstream's `defaultOptions`.
func DefaultRequireArraySortCompareSettings() RequireArraySortCompareOptions {
	return RequireArraySortCompareOptions{IgnoreStringArrays: true}
}

// requireArraySortCompareRawOptions is the wire shape, with a pointer so an absent key is
// distinguishable from an explicit false.
//
// This is the trap `rule.DecodeOptionsInto` sets for a default-true option, and it is silent in
// both directions. The generic helper yields a ZERO-VALUE struct when the config names the rule as
// a bare "error", which for this rule means `ignoreStringArrays: false`, the exact inverse of
// upstream's default, turning a rule that deliberately exempts string arrays into one that reports
// every one of them. Nothing would have caught it: a fixture built by handing `RunTypedWithOptions`
// an options STRUCT never routes through the decoder, so the inversion is invisible to the suite
// that is supposed to guard it.
type requireArraySortCompareRawOptions struct {
	IgnoreStringArrays *bool `json:"ignoreStringArrays"`
}

// DecodeRequireArraySortCompareOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple before dispatch, so what
// arrives here is the bare object rather than upstream's one-element array.
//
// An absent key keeps the default rather than zeroing it, which is the entire point of the pointer
// above. On a decode error the default is returned alongside the error, so a malformed option can
// never leave the rule running inverted.
func DecodeRequireArraySortCompareOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[requireArraySortCompareRawOptions]()(raw)
	if err != nil {
		return DefaultRequireArraySortCompareSettings(), err
	}

	wire, _ := decoded.(requireArraySortCompareRawOptions)
	options := DefaultRequireArraySortCompareSettings()
	if wire.IgnoreStringArrays != nil {
		options.IgnoreStringArrays = *wire.IgnoreStringArrays
	}
	return options, nil
}

// RequireArraySortCompare flags `sort()` and `toSorted()` called on an array with no comparator.
//
//	valid:   function f(a: any[]) { a.sort((a, b) => a - b); }
//	valid:   function f(a: { sort(): void }) { a.sort(); }
//	valid:   function f(a: any) { a.sort(); }
//	valid:   ['foo', 'bar'].sort();                          under the default option
//	invalid: function f(a: number[]) { a.sort(); }
//	invalid: function f(a: string[] | number[]) { a.sort(); }
//
// `Array#sort` with no comparator converts every element to a string and compares code units, so
// `[1, 10, 2].sort()` yields `[1, 10, 2]` rather than `[1, 2, 10]`. That is almost never what the
// author meant for anything but strings, and it is silent: the result is a plausible-looking array
// in the wrong order.
//
// The rule ships no fixer, and upstream has no `meta.fixable`, because there is no comparator the
// rule could write. Whether the intended order is ascending, descending, or by some field is the
// author's judgment, and the corpus carries no `output` on any of its seventeen failing cases.
//
// # What decides a finding
//
// The anchor is a call with EXACTLY ZERO ARGUMENTS whose callee is a member access. Upstream writes
// that as the selector `CallExpression[arguments.length=0] > MemberExpression`, so the argument
// count is part of the match rather than a test inside the body. `a.sort(undefined)` is silent even
// though it behaves identically at run time, which is upstream's decision and is reproduced: it
// appears in the corpus as a PASSING case, so an explicit `undefined` reads as the author having
// considered the question.
//
// The member name must be `sort` or `toSorted`, and then the RECEIVER's type is resolved through
// `GetConstrainedTypeAtLocation`, which walks a type parameter to its base constraint. The finding
// requires that type to be an array or a union of nothing but arrays. So `any` is silent, a
// user-defined `sort()` on an interface is silent, and a `string[] | number[]` reports while a
// `number[] | string` does not.
//
// The string-array exemption is checked BEFORE the array test, and it asks a narrower question than
// the report does: it requires the receiver to be an array or a TUPLE, then that every type argument
// is a string. A tuple's type arguments are its element types, so `['a', 'b'] as const` is exempt
// while `['a', 1]` is not.
//
// # The member name, and a substrate difference stated rather than papered over
//
// Upstream resolves the member name through `getStaticMemberAccessValue`, which reads a plain
// identifier directly but falls back to ESLint's `getStaticValue` over the enclosing SCOPE for a
// computed key. That means `const key = 'sort'; a[key]()` resolves to `sort` upstream and reports.
//
// This port reads a plain property access and a string-literal computed access, and does NOT
// constant-fold a computed key through a variable, because that requires the scope-analysis and
// constant-folding layer upstream gets from ESLint and this tree does not have. The gap is one
// shape wide and it is a FALSE NEGATIVE, which is the safe direction for a rule with no fixer.
//
// Measured rather than assumed, on the installed build, with a control alongside:
//
//	a['sort']()                        REPORT upstream, REPORT here   string-literal key
//	a?.['sort']()                      REPORT upstream, REPORT here   optional computed
//	const key = 'sort'; a[key]()       REPORT upstream, SILENT here   the divergence
//
// Upstream's corpus writes no computed access at all, so nothing in it can see any of these. The
// gap is recorded here so the next reader does not read the narrowness as deliberate and stop
// checking, and a fixture in this package pins the divergence as reporting-nothing rather than
// leaving it undocumented.
//
// # Cost
//
// A call expression is a common anchor, so the body exits on the argument count before touching the
// checker. The type is resolved only for a zero-argument call to something named `sort` or
// `toSorted`, which is rare.
var RequireArraySortCompare = rule.Rule{
	Name: "@typescript-eslint/require-array-sort-compare",

	// The receiver's type decides every finding, so the checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[RequireArraySortCompareOptions](options)
		if !isSettings {
			// A rule configured as a bare "error" is handed nil options, and `options.(T)` on nil
			// yields the zero value, which for this rule is the INVERSE of upstream's default. The
			// fallback is what keeps a plainly-configured rule behaving as documented, and it is
			// pinned by a fixture that bypasses the decoder.
			settings = DefaultRequireArraySortCompareSettings()
		}

		// isStringArrayNode answers upstream's `isStringArrayNode`: an array or tuple whose every
		// type argument is a string.
		//
		// The TUPLE half of that test is inert, in upstream as much as here, and it is kept anyway.
		// A mutation removing it survives every fixture, which reads as a blind spot and is not one:
		// a tuple is not an array type, so `isTypeArrayTypeOrUnionOfArrayTypes` below declines every
		// tuple regardless, and the exemption never gets to matter for one. Measured on the
		// installed build over `[string, string]`, `[string]`, `readonly [string, string]` and
		// `[number, number]`, with the option BOTH ways: all eight runs are clean, so no tuple's
		// verdict moves with the exemption on either side. Our port answers all eight identically.
		//
		// Kept rather than deleted because it is upstream's text and deleting it would be a
		// divergence with no observable benefit. Recorded rather than left silent because the next
		// reader will score the same mutant and needs to know the survival was measured.
		isStringArrayNode := func(receiver *ast.Node) bool {
			receiverType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)
			if !checker.Checker_isArrayType(ctx.TypeChecker, receiverType) &&
				!checker.IsTupleType(receiverType) {
				return false
			}

			typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, receiverType)
			for _, argument := range typeArguments {
				if type_checking.GetTypeName(ctx.TypeChecker, argument) != "string" {
					return false
				}
			}
			// An empty type-argument list satisfies `every` upstream too, so it is exempt here for
			// the same reason rather than by omission.
			return true
		}

		// isArrayOrUnionOfArrays answers upstream's `isTypeArrayTypeOrUnionOfArrayTypes`: every
		// constituent of a union must be an array, and a non-union is its own single constituent.
		isArrayOrUnionOfArrays := func(receiverType *checker.Type) bool {
			for constituent := range type_checking.UnionTypePartsSeq(receiverType) {
				if !checker.Checker_isArrayType(ctx.TypeChecker, constituent) {
					return false
				}
			}
			return true
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				call := node.AsCallExpression()

				// Upstream matches `arguments.length=0` in the selector, so a call with any
				// argument at all never reaches the body. `a.sort(undefined)` is silent.
				if call.Arguments != nil && len(call.Arguments.Nodes) != 0 {
					return
				}

				// Parentheses are skipped on the callee, and this line was written the other way
				// first on an argument that sounded right and was false. estree has no
				// parenthesized-expression node at all, so upstream's `> MemberExpression` child
				// selector sees straight through them and `(a.sort)()` REPORTS. Measured against
				// the installed build rather than reasoned about, after the reasoned version left
				// a false negative that no imported fixture could see, since the corpus writes no
				// parenthesized callee anywhere.
				//
				// SkipParentheses dereferences its argument, so the nil test comes first.
				if call.Expression == nil {
					return
				}
				callee := ast.SkipParentheses(call.Expression)

				receiver, memberName, isMemberAccess := sortMemberAccessOf(callee)
				if !isMemberAccess || (memberName != "sort" && memberName != "toSorted") {
					return
				}

				if settings.IgnoreStringArrays && isStringArrayNode(receiver) {
					return
				}

				if !isArrayOrUnionOfArrays(type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)) {
					return
				}

				// Upstream reports `callee.parent`, which is the whole call expression rather than
				// the member access, so the finding underlines `a.sort()` including the parentheses.
				ctx.ReportNode(node, buildRequireCompareMessage())
			},
		}
	},
}

// sortMemberAccessOf reads a callee as a member access, returning its receiver and the member name.
//
// A property access supplies the name directly. A computed access supplies it only when the key is
// a string literal, which is the part of upstream's `getStaticMemberAccessValue` that needs no scope
// analysis; the variable-folding half is the stated divergence in the rule's doc comment.
//
// The optional-chain forms are accepted rather than excluded. `a?.sort()` is a member access in
// estree too, and upstream's corpus carries `a?.sort()` in BOTH lists: passing where the receiver
// is a user-defined interface, and reporting where it is an array. So the chain is not the
// discriminator, the receiver's type is.
func sortMemberAccessOf(callee *ast.Node) (receiver *ast.Node, memberName string, isMemberAccess bool) {
	if ast.IsPropertyAccessExpression(callee) {
		access := callee.AsPropertyAccessExpression()
		if access.Name() == nil || !ast.IsIdentifier(access.Name()) {
			return nil, "", false
		}
		return access.Expression, access.Name().Text(), true
	}

	if ast.IsElementAccessExpression(callee) {
		access := callee.AsElementAccessExpression()
		key := access.ArgumentExpression
		if key == nil || key.Kind != ast.KindStringLiteral {
			return nil, "", false
		}
		return access.Expression, key.Text(), true
	}

	return nil, "", false
}

func buildRequireCompareMessage() rule.Message {
	return rule.Message{
		Id:          "requireCompare",
		Description: "Require 'compare' argument.",
	}
}
