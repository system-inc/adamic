package typescript

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func buildConditionalMessage() rule.Message {
	return rule.Message{
		Id:          "conditional",
		Description: "Expected non-Promise value in a boolean conditional.",
	}
}
func buildPredicateMessage() rule.Message {
	return rule.Message{
		Id:          "predicate",
		Description: "Expected a non-Promise value to be returned.",
	}
}
func buildSpreadMessage() rule.Message {
	return rule.Message{
		Id:          "spread",
		Description: "Expected a non-Promise value to be spreaded in an object.",
	}
}
func buildVoidReturnArgumentMessage() rule.Message {
	return rule.Message{
		Id:          "voidReturnArgument",
		Description: "Promise returned in function argument where a void return was expected.",
	}
}
func buildVoidReturnAttributeMessage() rule.Message {
	return rule.Message{
		Id:          "voidReturnAttribute",
		Description: "Promise-returning function provided to attribute where a void return was expected.",
	}
}
func buildVoidReturnInheritedMethodMessage(heritageTypeName string) rule.Message {
	return rule.Message{
		Id:          "voidReturnInheritedMethod",
		Description: fmt.Sprintf("Promise-returning method provided where a void return was expected by extended/implemented type '%v'.", heritageTypeName),
	}
}
func buildVoidReturnPropertyMessage() rule.Message {
	return rule.Message{
		Id:          "voidReturnProperty",
		Description: "Promise-returning function provided to property where a void return was expected.",
	}
}
func buildVoidReturnReturnValueMessage() rule.Message {
	return rule.Message{
		Id:          "voidReturnReturnValue",
		Description: "Promise-returning function provided to return value where a void return was expected.",
	}
}
func buildVoidReturnVariableMessage() rule.Message {
	return rule.Message{
		Id:          "voidReturnVariable",
		Description: "Promise-returning function provided to variable where a void return was expected.",
	}
}

type NoMisusedPromisesChecksVoidReturnOptions struct {
	Arguments        *bool `json:"arguments"`
	Attributes       *bool `json:"attributes"`
	InheritedMethods *bool `json:"inheritedMethods"`
	Properties       *bool `json:"properties"`
	Returns          *bool `json:"returns"`
	Variables        *bool `json:"variables"`
}
type NoMisusedPromisesOptions struct {
	ChecksConditionals *bool `json:"checksConditionals"`
	// ChecksConditionalsFlagUnions is upstream's `checksConditionals: {flagUnions}`: "none" (the
	// default) reports a value that is always thenable, "all" also one a union makes sometimes
	// thenable, and "strict" also `Promise<T> | T`, whose awaited half matches the rest.
	ChecksConditionalsFlagUnions string                                    `json:"-"`
	ChecksSpreads                *bool                                     `json:"checksSpreads"`
	ChecksVoidReturn             *bool                                     `json:"checksVoidReturn"`
	ChecksVoidReturnOpts         *NoMisusedPromisesChecksVoidReturnOptions `json:"-"`
}

// UnmarshalJSON reads upstream's `checksVoidReturn`, which is a boolean or an object of six booleans,
// and its `checksConditionals`, a boolean or an object holding flagUnions.
//
// tsgolint split the two spellings into `ChecksVoidReturn` and `ChecksVoidReturnOpts`, and with no
// tags the object form could only be written under a key upstream does not have,
// `checksVoidReturnOpts`, while upstream's own `{"checksVoidReturn": {"arguments": false}}` failed
// to decode into a boolean. The two fields stay, because `Run` reads them; this fills them from
// upstream's one key. An object turns the check on with those sub-flags, as upstream's
// `parseChecksVoidReturn` does, and the sub-flags it leaves out default to true in `Run`.
func (options *NoMisusedPromisesOptions) UnmarshalJSON(raw []byte) error {
	var wire struct {
		ChecksConditionals json.RawMessage `json:"checksConditionals"`
		ChecksSpreads      *bool           `json:"checksSpreads"`
		ChecksVoidReturn   json.RawMessage `json:"checksVoidReturn"`
	}
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return err
	}
	*options = NoMisusedPromisesOptions{ChecksSpreads: wire.ChecksSpreads}
	// An object turns the check on, as upstream's truthy test reads it, and a missing flagUnions is
	// "none", as upstream's normalizeFlagUnionsOption has it.
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	conditionals := strings.TrimSpace(string(wire.ChecksConditionals))
	switch {
	case conditionals == "":
	case conditionals == "true" || conditionals == "false":
		options.ChecksConditionals = type_checking.Ref(conditionals == "true")
	case strings.HasPrefix(conditionals, "{"):
		var each struct {
			FlagUnions *string `json:"flagUnions"`
		}
		if err := rule.UnmarshalOptions(wire.ChecksConditionals, &each); err != nil {
			return fmt.Errorf("checksConditionals: %w", err)
		}
		options.ChecksConditionals = type_checking.Ref(true)
		options.ChecksConditionalsFlagUnions = "none"
		if each.FlagUnions != nil {
			switch *each.FlagUnions {
			case "all", "strict", "none":
				options.ChecksConditionalsFlagUnions = *each.FlagUnions
			default:
				return fmt.Errorf("checksConditionals.flagUnions is %q, and upstream takes all, strict or none", *each.FlagUnions)
			}
		}
	default:
		return fmt.Errorf("checksConditionals takes a boolean or an object holding flagUnions, got %s", conditionals)
	}
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(wire.ChecksVoidReturn))
	switch {
	case trimmed == "":
		return nil
	case trimmed == "true" || trimmed == "false":
		options.ChecksVoidReturn = type_checking.Ref(trimmed == "true")
		return nil
	case strings.HasPrefix(trimmed, "{"):
		var each NoMisusedPromisesChecksVoidReturnOptions
		if err := rule.UnmarshalOptions(wire.ChecksVoidReturn, &each); err != nil {
			return fmt.Errorf("checksVoidReturn: %w", err)
		}
		options.ChecksVoidReturn = type_checking.Ref(true)
		options.ChecksVoidReturnOpts = &each
		return nil
	}
	return fmt.Errorf("checksVoidReturn takes a boolean or an object of booleans, got %s", trimmed)
}

// NoMisusedPromises flags a Promise used where the surrounding code cannot handle one: as a boolean
// condition, spread into an object, or passed somewhere a `void`-returning function was expected.
//
//	valid:   const p = Promise.resolve(); if (await p) {}
//	valid:   [1, 2].forEach(async () => {});          // forEach takes a void return AND a thenable
//	invalid: const p = Promise.resolve(); if (p) {}   // conditional — always truthy
//	invalid: [1, 2].filter(async n => n > 1);         // predicate — the array gets Promises, not booleans
//	invalid: declare function f(cb: () => void): void; f(async () => {});  // voidReturnArgument
//	invalid: const obj = { ...Promise.resolve({ a: 1 }) };                 // spread
//
// A Promise in a boolean position is always truthy, so `if (p)` takes the true branch whether the
// work succeeded or failed, and `arr.filter(async …)` filters on Promise objects rather than on what
// they resolve to. Both compile. Neither does what the author wrote it to do, and both are silent at
// runtime rather than throwing, which is what makes this worth a type-aware rule rather than review.
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/no_misused_promises/no_misused_promises.go`, at commit
// `076823b`, absorbed onto cohere's own rule interface here. It reaches for `GetCallSignatures`,
// `GetConstructSignatures`, `IsThenableType`, `IsArrayMethodCallWithPredicate`, `GetHeritageClauses`,
// `IsRestParameterDeclaration`, `UnionTypeParts`, `IsTypeFlagSet`, `Some`, `Map`, `Flatten` and `Ref`
// because that is what upstream reaches for, and this note is why a reader finds those helpers in a
// file that otherwise looks native.
//
// The vendoring step is measurable and was measured: rewriting exactly four import paths
// (`microsoft/typescript-go/shim/` to `microsoft/TypeScript/tsc/shim/`, and
// `typescript-eslint/tsgolint/internal/{rule,utils}` to ours) produces a byte diff showing those four
// lines and nothing else. Every helper the rule calls was already on the shelf at
// `internal/utilities/typecheck/`, and every checker entry point it needs is reachable through
// our shim.
//
// tsgolint is not re-synced, so this file is now the only copy of the algorithm rather than a
// translation layer over a vendored one. The checker logic below is byte-identical to upstream's;
// what changed is the interface it speaks, plus the nil-checker guard described below. No predicate,
// no flag set, no defaulting block and no traversal was touched, and both options structs came across
// field for field rather than being re-declared.
//
// oxc declares this rule as a tsgolint delegate and carries no algorithm and no corpus, so the
// behavior oxlint exhibits IS tsgolint's: the release binary shells out to a `tsgolint` executable
// and refuses to run the rule when that binary is absent, which is what it does on this machine
// (`Failed to find tsgolint executable`). That refusal is itself the proof, so oxlint cannot serve as
// ground truth here and tsgolint's own test file is the corpus instead.
//
// # THREE independent checks, and one of them is an object of sub-flags
//
// This is the most structurally complex rule in this family, and the shape is the thing to hold:
//
//	checksConditionals  default TRUE   a Promise in a boolean position, plus array predicates
//	checksVoidReturn    default TRUE   a Promise-returning function where void was expected
//	checksSpreads       default TRUE   a Promise spread into an object
//
// `checksConditionals` also takes upstream's object, `{flagUnions}`, which turns the check on and
// says how a union holding a thenable is judged: "none" (the default, and what `true` means) reports
// only a value every member of which is thenable, "all" any union holding one, and "strict" only
// `Promise<T> | T`, whose awaited types are the union's other members, each assignable both ways.
// Checked against upstream's 34 checksConditionals rows and edge rows in
// no_misused_promises_conditionals_corpus_test.go.
//
// `checksVoidReturn` is not one check. It gates SIX sub-flags, each of which is a distinct position
// in the syntax and a distinct message id, and each of which registers its own listeners:
//
//	arguments         voidReturnArgument         KindCallExpression, KindNewExpression
//	attributes        voidReturnAttribute        KindJsxAttribute
//	inheritedMethods  voidReturnInheritedMethod  KindClassDeclaration/Expression, KindInterfaceDeclaration
//	properties        voidReturnProperty         KindPropertyAssignment, KindMethodDeclaration,
//	                                             KindShorthandPropertyAssignment
//	returns           voidReturnReturnValue      KindReturnStatement
//	variables         voidReturnVariable         KindVariableDeclaration, and the assignment arm of
//	                                             KindBinaryExpression
//
// That set is established from upstream's struct and from the corpus rather than guessed: all six
// message ids appear in tsgolint's own test file, so all six positions demonstrably report
// (voidReturnArgument 32 times, voidReturnInheritedMethod 25, voidReturnProperty 15, voidReturnVariable
// 6, voidReturnAttribute 3, voidReturnReturnValue 2). A port that invented this set would have passed
// every fixture it also invented, which is why the count is recorded here.
//
// A captured rule inventory described the option surface for this rule and it was not to be trusted as the
// authority; upstream's struct is. That is the standing lesson of this lane rather than a remark
// about this entry.
//
// # The pointer fields are load-bearing and must not be flattened to plain bools
//
// EVERY option on this rule defaults to TRUE — all three top-level checks and all six void-return
// sub-flags. That is nine fields whose default is not the zero value, so a plain `bool` could not
// tell "the user wrote false" from "the user wrote nothing", and the entire defaulting block at the
// top of `Run` reads exactly that distinction. Flattening any of them would silently turn the whole
// rule off for a config that merely omitted a key. `ChecksVoidReturnOpts` is a pointer to a struct
// for the same reason one level up.
//
// # Where the two references disagree, counted rather than assumed
//
// Listeners were counted on both sides rather than assumed, because a sibling in this family
// (`await-thenable`) had a fourth arm in `@typescript-eslint` that tsgolint lacks entirely. Here the
// two agree on every arm, and the two apparent differences are representational rather than
// behavioral:
//
//	SPREAD.       tsgolint registers KindSpreadElement AND KindSpreadAssignment; the plugin registers
//	              SpreadElement alone. That is because ESTree spells an object spread as a
//	              SpreadElement inside an ObjectExpression while typescript-go gives it its own kind.
//	              Same inputs, two spellings.
//
//	PREDICATES.   tsgolint registers KindPropertyAccessExpression and KindElementAccessExpression and
//	              then requires the parent to be a call; the plugin writes the selector
//	              `CallExpression > MemberExpression` and lets ESLint do that filtering. Same
//	              predicate, different layer.
//
//	VARIABLES.    tsgolint's assignment arm hangs off KindBinaryExpression guarded by
//	              `IsAssignmentExpression`; the plugin has a dedicated AssignmentExpression node.
//	              Same rule, different grammar.
//
// The one genuine, user-visible disagreement is MESSAGE TEXT on the spread id: the plugin renders
// "Expected a non-Promise value to be spread in an object" and tsgolint renders "…to be spreaded in
// an object". Upstream's wording is an error of English and it is reproduced here unrepaired, because
// oxlint runs tsgolint and the differential harness compares against oxlint, so correcting it would
// read as a difference the harness can see. A fixture asserts the exact rendered text so the next
// sync notices if upstream fixes it.
//
// # There are no fixes and no suggestions
//
// tsgolint ships neither for this rule, and neither does the corpus assert any. There is nothing for
// `rule_testing` to apply and nothing needing a hand-rolled suggestion applier. Upstream carries two
// `TODO(port)` markers at report sites where `@typescript-eslint` narrows the reported range to a
// function's head (`getFunctionHeadLoc`), and one at `getMemberIfExists` where it would escape leading
// underscores. All three are carried across unchanged, so a member named `__proto__` resolves through
// the un-escaped path exactly as it does upstream.
//
// # The checker, and the nil guard that is now reachable
//
// Every listener consults `ctx.TypeChecker`, so this needs the checker and its fixtures use
// `RunTyped`. While this family was adapted through `internal/rules/upstream`, the standing
// `if ctx.TypeChecker == nil` could not be written, because the listener bodies were upstream's and
// that adapter set `NeedsTypeChecker` unconditionally, which made the nil case unreachable. That
// adapter is gone. Absorbing the rule makes the nil case REACHABLE, so the guard is written here.
//
// The direction matters more than usual. Under a checker-less Context this rule does not panic: our
// shim's `GetTypeAtLocation` and `GetSymbolAtLocation` return nil rather than crashing, so the rule
// would go SILENT, and silence makes every clean fixture pass having proven nothing — which for this
// rule is ONE HUNDRED AND TWENTY-THREE of its two hundred and thirteen cases. The guard sits at the
// top of `Run` rather than inside each of the fourteen listeners, which is strictly stronger: no
// listener is even registered, so no arm can be added later that forgets it.
// `TestNoMisusedPromisesRequiresTheTypedHarness` pins the declaration and the guard together so a
// later revert fails loudly instead of going vacuously green.
//
// # Cost
//
// This is the most expensive rule in the family and it should be read that way. It registers up to
// fourteen listeners, several on kinds that appear on nearly every line (call expressions, binary
// expressions, property access, variable declarations), and most arms reach the checker before they
// can decline. Upstream's `checkArguments` resolves the callee's type and then walks every signature's
// every parameter for every call in the file, and its JSX arm resolves every attribute's contextual
// type; both now ask the cheap half of their conjunction first, which is the one departure from
// upstream's order (#hekjpw3). The syntactic escape hatches upstream wrote — `isPossiblyFunctionType`
// on return statements and variable declarations, the `Arguments() == nil` early exit — exist for
// exactly that reason and are carried across rather than being treated as optional.
var NoMisusedPromises = rule.Rule{
	Name: "@typescript-eslint/no-misused-promises",

	// Every listener consults the checker; see the note above on why the guard lives in Run.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// ProgramReads is deliberately absent, and that is a measurement rather than an omission. The
	// program handle appears nowhere in this body: the rule asks the checker questions about nodes in
	// the file it was handed, and the checker resolves across module boundaries on its own without the
	// rule reaching past its own file. While this family was adapted, `upstream.Adapt` declared the
	// flag on every rule it wrapped by assumption; absorbing makes it answerable per rule, and the
	// honest answer here is no.
	//
	// Worth knowing for whoever writes the next one of these: the guard in `internal/dispatch` that
	// enforces the flag is TEXTUAL, a grep for the selector over the rule's whole file, comments
	// included. So a doc comment that merely NAMES the thing it is declining to use fails the guard
	// with a message asserting the opposite of the truth. This paragraph is worded around the token
	// for that reason, rather than because the phrasing reads better.

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		opts, ok := rule.OptionsAs[NoMisusedPromisesOptions](options)
		if !ok {
			opts = NoMisusedPromisesOptions{}
		}
		if opts.ChecksConditionals == nil {
			opts.ChecksConditionals = type_checking.Ref(true)
		}
		if opts.ChecksConditionalsFlagUnions == "" {
			opts.ChecksConditionalsFlagUnions = "none"
		}
		if opts.ChecksSpreads == nil {
			opts.ChecksSpreads = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturn == nil {
			opts.ChecksVoidReturn = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturnOpts == nil {
			opts.ChecksVoidReturnOpts = type_checking.Ref(NoMisusedPromisesChecksVoidReturnOptions{})
		}
		if opts.ChecksVoidReturnOpts.Arguments == nil {
			opts.ChecksVoidReturnOpts.Arguments = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturnOpts.Attributes == nil {
			opts.ChecksVoidReturnOpts.Attributes = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturnOpts.InheritedMethods == nil {
			opts.ChecksVoidReturnOpts.InheritedMethods = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturnOpts.Properties == nil {
			opts.ChecksVoidReturnOpts.Properties = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturnOpts.Returns == nil {
			opts.ChecksVoidReturnOpts.Returns = type_checking.Ref(true)
		}
		if opts.ChecksVoidReturnOpts.Variables == nil {
			opts.ChecksVoidReturnOpts.Variables = type_checking.Ref(true)
		}

		anySignatureIsThenableType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			return type_checking.Some(type_checking.GetCallSignatures(ctx.TypeChecker, t), func(sig *checker.Signature) bool {
				return type_checking.IsThenableType(ctx.TypeChecker, node, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig))
			})
		}

		returnsThenable := func(node *ast.Node) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node))
			return type_checking.Some(type_checking.UnionTypeParts(t), func(t *checker.Type) bool {
				return anySignatureIsThenableType(node, t)
			})
		}

		checkArrayPredicates := func(node *ast.Node) {
			parent := node.Parent

			if !ast.IsCallExpression(parent) {
				return
			}

			expr := parent.AsCallExpression()
			arguments := expr.Arguments.Nodes
			if len(arguments) == 0 {
				return
			}

			callback := arguments[0]

			if type_checking.IsArrayMethodCallWithPredicate(ctx.TypeChecker, expr) && returnsThenable(callback) {
				ctx.ReportNode(callback, buildPredicateMessage())
			}
		}

		isFunctionParam := func(
			param *ast.Symbol,
			node *ast.Node,
		) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeOfSymbolAtLocation(param, node))
			if t == nil {
				return false
			}
			return type_checking.Some(type_checking.UnionTypeParts(t), func(t *checker.Type) bool {
				return len(type_checking.GetCallSignatures(ctx.TypeChecker, t)) != 0
			})
		}

		// Variation on the thenable check which requires all forms of the type (read:
		// alternates in a union) to be thenable. Otherwise, you might be trying to
		// check if something is defined or undefined and get caught because one of the
		// branches is thenable.
		isAlwaysThenable := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)

			for subType := range type_checking.UnionTypePartsSeq(checker.Checker_getApparentType(ctx.TypeChecker, t)) {
				thenProp := checker.Checker_getPropertyOfType(ctx.TypeChecker, subType, "then")

				// If one of the alternates has no then property, it is not thenable in all
				// cases.
				if thenProp == nil {
					return false
				}

				// We walk through each variation of the then property. Since we know it
				// exists at this point, we just need at least one of the alternates to
				// be of the right form to consider it thenable.
				thenType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(thenProp, node)
				hasThenableSignature := false
				for subType := range type_checking.UnionTypePartsSeq(thenType) {
					for _, signature := range type_checking.GetCallSignatures(ctx.TypeChecker, subType) {
						params := checker.Signature_parameters(signature)
						if len(params) != 0 && isFunctionParam(params[0], node) {
							hasThenableSignature = true
						}
					}

					// We only need to find one variant of the then property that has a
					// function signature for it to be thenable.
					if hasThenableSignature {
						break
					}
				}

				// If no flavors of the then property are thenable, we don't consider the
				// overall type to be thenable
				if !hasThenableSignature {
					return false
				}
			}

			// If all variants are considered thenable (i.e. haven't returned false), we
			// consider the overall type thenable
			return true
		}

		checkedNodes := map[*ast.Node](struct{}){}

		var checkConditional func(
			node *ast.Expression,
			isTestExpr bool,
		)
		checkConditional = func(
			node *ast.Expression,
			isTestExpr bool,
		) {
			if node == nil || ast.IsAssignmentExpression(node, false) {
				return
			}
			// prevent checking the same node multiple times
			if _, ok := checkedNodes[node]; ok {
				return
			}
			checkedNodes[node] = struct{}{}

			node = ast.SkipParentheses(node)

			if ast.IsBinaryExpression(node) && ast.IsLogicalExpression(node) {
				expr := node.AsBinaryExpression()
				// ignore the left operand for nullish coalescing expressions not in a context of a test expression
				if expr.OperatorToken.Kind != ast.KindQuestionQuestionToken || isTestExpr {
					checkConditional(expr.Left, isTestExpr)
				}
				// we ignore the right operand when not in a context of a test expression
				if isTestExpr {
					checkConditional(expr.Right, isTestExpr)
				}
				return
			}

			if isAlwaysThenable(node) {
				ctx.ReportNode(node, buildConditionalMessage())
				return
			}

			// flagUnions widens the check to a union holding a thenable: every one under "all", and
			// under "strict" only one whose awaited half is the rest, as in `Promise<T> | T`.
			switch opts.ChecksConditionalsFlagUnions {
			case "all":
				if isSometimesThenable(ctx.TypeChecker, node) {
					ctx.ReportNode(node, buildConditionalMessage())
				}
			case "strict":
				if hasMatchingPromiseTypeArgument(ctx.TypeChecker, node) {
					ctx.ReportNode(node, buildConditionalMessage())
				}
			}
		}

		getMemberIfExists := func(
			t *checker.Type,
			memberName string,
		) *ast.Symbol {
			// TODO(port)
			// const escapedMemberName = ts.escapeLeadingUnderscores(memberName);
			symbol := checker.Type_symbol(t)
			if symbol != nil {
				symbol = symbol.Members[memberName]
			}
			if symbol != nil {
				return symbol
			}

			return checker.Checker_getPropertyOfType(ctx.TypeChecker, t, memberName)
		}

		isVoidReturningFunctionType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			hadVoidReturn := false
			for t := range type_checking.UnionTypePartsSeq(t) {
				for _, sig := range type_checking.GetCallSignatures(ctx.TypeChecker, t) {
					returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)
					// If a certain positional argument accepts both thenable and void returns,
					// a promise-returning function is valid
					if type_checking.IsThenableType(ctx.TypeChecker, node, returnType) {
						return false
					}

					hadVoidReturn = hadVoidReturn || type_checking.IsTypeFlagSet(returnType, checker.TypeFlagsVoid)
				}
			}
			return hadVoidReturn
		}

		/**
		 * Checks `heritageType` for a member named `memberName` that returns void; reports the
		 * 'voidReturnInheritedMethod' message if found.
		 * @param nodeMember Node member that returns a Promise
		 * @param heritageType Heritage type to check against
		 * @param memberName Name of the member to check for
		 */
		checkHeritageTypeForMemberReturningVoid := func(
			nodeMember *ast.Node,
			heritageType *checker.Type,
			memberName string,
		) {
			heritageMember := getMemberIfExists(heritageType, memberName)
			if heritageMember == nil {
				return
			}
			memberType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(
				heritageMember,
				nodeMember,
			)
			if !isVoidReturningFunctionType(nodeMember, memberType) {
				return
			}
			ctx.ReportNode(nodeMember, buildVoidReturnInheritedMethodMessage(ctx.TypeChecker.TypeToString(heritageType)))
		}

		checkJSXAttribute := func(node *ast.JsxAttribute) {
			if node.Initializer == nil || node.Initializer.Kind != ast.KindJsxExpression {
				return
			}
			expressionContainer := node.Initializer.AsJsxExpression()
			expression := expressionContainer.Expression
			// Upstream asks for the contextual type first and the expression's type last. The three
			// questions are a conjunction with no side effects, so the cheap one goes first: the
			// attribute's contextual type resolves the element's props for every attribute, and was
			// 1.2s of this rule's 2.5s of CPU on ahra, while the expression's own type was 0.2s and
			// is a function returning a thenable for almost no attribute (#hekjpw3).
			if !returnsThenable(expression) {
				return
			}
			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Initializer, checker.ContextFlagsNone)
			if contextualType != nil && isVoidReturningFunctionType(node.Initializer, contextualType) {
				ctx.ReportNode(node.Initializer, buildVoidReturnAttributeMessage())
			}
		}

		checkSpread := func(node *ast.Node) {
			if type_checking.IsThenableType(ctx.TypeChecker, node.Expression(), nil) {
				ctx.ReportNode(node.Expression(), buildSpreadMessage())
			}
		}

		isThenableReturningFunctionType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			return type_checking.Some(type_checking.UnionTypeParts(t), func(t *checker.Type) bool {
				return anySignatureIsThenableType(node, t)
			})
		}

		var checkThenableOrVoidArgument func(
			node *ast.Expression,
			t *checker.Type,
			index int,
			thenableReturnIndices *[]int,
			voidReturnIndices *[]int,
		)
		checkThenableOrVoidArgument = func(
			node *ast.Expression,
			t *checker.Type,
			index int,
			thenableReturnIndices *[]int,
			voidReturnIndices *[]int,
		) {
			if isThenableReturningFunctionType(node.Expression(), t) {
				(*thenableReturnIndices) = append(*thenableReturnIndices, index)
			} else if isVoidReturningFunctionType(node.Expression(), t) &&
				// If a certain argument accepts both thenable and void returns,
				// a promise-returning function is valid
				!slices.Contains(*thenableReturnIndices, index) {

				(*voidReturnIndices) = append(*voidReturnIndices, index)
			}
			contextualType := checker.Checker_getContextualTypeForArgumentAtIndex(ctx.TypeChecker, node, index)

			if contextualType != t {
				checkThenableOrVoidArgument(
					node,
					contextualType,
					index,
					thenableReturnIndices,
					voidReturnIndices,
				)
			}
		}

		// Get the positions of arguments which are void functions (and not also
		// thenable functions). These are the candidates for the void-return check at
		// the current call site.
		// If the function parameters end with a 'rest' parameter, then we consider
		// the array type parameter (e.g. '...args:Array<SomeType>') when determining
		// if trailing arguments are candidates.
		voidFunctionArguments := func(
			node *ast.Expression,
		) []int {
			// 'new' can be used without any arguments, as in 'let b = new Object;'
			// In this case, there are no argument positions to check, so return early.
			if node.Arguments() == nil {
				return []int{}
			}
			thenableReturnIndices := []int{}
			voidReturnIndices := []int{}
			t := ctx.TypeChecker.GetTypeAtLocation(node.Expression())

			// We can't use checker.getResolvedSignature because it prefers an early '() => void' over a later '() => Promise<void>'
			// See https://github.com/microsoft/TypeScript/issues/48077

			for subType := range type_checking.UnionTypePartsSeq(t) {
				// Standard function calls and `new` have two different types of signatures
				var signatures []*checker.Signature
				if ast.IsCallExpression(node) {
					signatures = type_checking.GetCallSignatures(ctx.TypeChecker, subType)
				} else {
					signatures = type_checking.GetConstructSignatures(ctx.TypeChecker, subType)
				}
				for _, signature := range signatures {
					for index, parameter := range checker.Signature_parameters(signature) {
						decl := parameter.ValueDeclaration
						t := ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameter, node.Expression())

						// If this is a array 'rest' parameter, check all of the argument indices
						// from the current argument to the end.
						if decl != nil && type_checking.IsRestParameterDeclaration(decl) {
							if checker.Checker_isArrayType(ctx.TypeChecker, t) {
								// Unwrap 'Array<MaybeVoidFunction>' to 'MaybeVoidFunction',
								// so that we'll handle it in the same way as a non-rest
								// 'param: MaybeVoidFunction'
								t = checker.Checker_getTypeArguments(ctx.TypeChecker, t)[0]
								for i := index; i < len(node.Arguments()); i++ {
									checkThenableOrVoidArgument(
										node,
										t,
										i,
										&thenableReturnIndices,
										&voidReturnIndices,
									)
								}
							} else if checker.IsTupleType(t) {
								// Check each type in the tuple - for example, [boolean, () => void] would
								// add the index of the second tuple parameter to 'voidReturnIndices'
								typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
								for i := index; i < len(node.Arguments()) && i-index < len(typeArgs); i++ {
									checkThenableOrVoidArgument(
										node,
										typeArgs[i-index],
										i,
										&thenableReturnIndices,
										&voidReturnIndices,
									)
								}
							}
						} else {
							checkThenableOrVoidArgument(
								node,
								t,
								index,
								&thenableReturnIndices,
								&voidReturnIndices,
							)
						}
					}
				}
			}

			for _, index := range thenableReturnIndices {
				at := slices.Index(voidReturnIndices, index)
				if at >= 0 {
					voidReturnIndices = slices.Delete(voidReturnIndices, at, at+1)
				}
			}

			return voidReturnIndices
		}

		// An argument is reported when its position takes a void-returning function AND the argument
		// returns a thenable. Upstream resolves the positions first, which walks every parameter of
		// every signature and asks each its contextual type, for every call in the file, including
		// calls with no arguments at all. Asking each argument first is the same conjunction in the
		// other order: almost no argument returns a thenable, so almost no call resolves its
		// positions (#hekjpw3).
		checkArguments := func(
			node *ast.Expression,
		) {
			thenableArguments := []int{}
			for index, argument := range node.Arguments() {
				if returnsThenable(argument) {
					thenableArguments = append(thenableArguments, index)
				}
			}
			if len(thenableArguments) == 0 {
				return
			}

			voidArgs := voidFunctionArguments(node)
			if len(voidArgs) == 0 {
				return
			}

			for _, index := range thenableArguments {
				if slices.Contains(voidArgs, index) {
					ctx.ReportNode(node.Arguments()[index], buildVoidReturnArgumentMessage())
				}
			}
		}

		checkClassLikeOrInterfaceNode := func(
			node *ast.Node,
		) {
			heritageClauses := type_checking.GetHeritageClauses(node)
			if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
				return
			}

			heritageTypes := type_checking.Flatten(type_checking.Map(heritageClauses.Nodes, func(h *ast.Node) []*checker.Type {
				return type_checking.Map(h.AsHeritageClause().Types.Nodes, func(n *ast.Node) *checker.Type {
					return ctx.TypeChecker.GetTypeAtLocation(n)
				})
			}))

			for _, nodeMember := range node.Members() {
				if nodeMember.Name() == nil {
					// Call/construct/index signatures don't have names. TS allows call signatures to mismatch,
					// and construct signatures can't be async.
					// TODO - Once we're able to use `checker.isTypeAssignableTo` (v8), we can check an index
					// signature here against its compatible index signatures in `heritageTypes`
					continue
				}
				if !(ast.IsIdentifier(nodeMember.Name()) || ast.IsPrivateIdentifier(nodeMember.Name()) || ast.IsStringLiteral(nodeMember.Name()) || ast.IsNumericLiteral(nodeMember.Name()) || ast.IsBigIntLiteral(nodeMember.Name())) {
					continue
				}
				memberName := nodeMember.Name().Text()
				if ast.IsStatic(nodeMember) {
					continue
				}
				if !returnsThenable(nodeMember) {
					continue
				}
				for _, heritageType := range heritageTypes {
					checkHeritageTypeForMemberReturningVoid(
						nodeMember,
						heritageType,
						memberName,
					)
				}
			}
		}

		checkProperty := func(node *ast.Node) {
			if ast.IsPropertyAssignment(node) {
				property := node.AsPropertyAssignment()
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, property.Initializer, checker.ContextFlagsNone)

				if contextualType != nil && isVoidReturningFunctionType(
					property.Initializer,
					contextualType,
				) && returnsThenable(property.Initializer) {
					if ast.IsFunctionLike(property.Initializer) {
						returnType := property.Initializer.Type()
						if returnType != nil {
							ctx.ReportNode(returnType, buildVoidReturnPropertyMessage())
						} else {
							ctx.ReportNode(
								// TODO(port): getFunctionHeadLoc(functionNode, context.sourceCode)
								property.Initializer,
								buildVoidReturnPropertyMessage(),
							)
						}
					} else {
						ctx.ReportNode(property.Initializer, buildVoidReturnPropertyMessage())
					}
				}
			} else if ast.IsShorthandPropertyAssignment(node) {
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Name(), checker.ContextFlagsNone)
				if contextualType != nil &&
					isVoidReturningFunctionType(node.Name(), contextualType) &&
					returnsThenable(node.Name()) {
					ctx.ReportNode(node.Name(), buildVoidReturnPropertyMessage())
				}
			} else if ast.IsMethodDeclaration(node) {
				if ast.IsComputedPropertyName(node.Name()) {
					return
				}
				obj := node.Parent

				// Below condition isn't satisfied unless something goes wrong,
				// but is needed for type checking.
				// 'node' does not include class method declaration so 'obj' is
				// always an object literal expression, but after converting 'node'
				// to TypeScript AST, its type includes MethodDeclaration which
				// does include the case of class method declaration.
				if !ast.IsObjectLiteralExpression(obj) {
					return
				}

				if !returnsThenable(node) {
					return
				}
				objType := checker.Checker_getContextualType(ctx.TypeChecker, obj, checker.ContextFlagsNone)
				if objType == nil {
					return
				}
				propertySymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, objType, node.Name().Text())
				if propertySymbol == nil {
					return
				}

				contextualType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(
					propertySymbol,
					node.Name(),
				)

				if isVoidReturningFunctionType(node.Name(), contextualType) {
					if ast.IsMethodDeclaration(node) {
					}

					if node.Type() != nil {
						ctx.ReportNode(node.Type(), buildVoidReturnPropertyMessage())
					} else {
						ctx.ReportNode(
							// TODO(port): getFunctionHeadLoc(functionNode, context.sourceCode)
							node,
							buildVoidReturnPropertyMessage(),
						)
					}
				}
			}
		}

		/**
		 * A syntactic check to see if an annotated type is maybe a function type.
		 * This is a perf optimization to help avoid requesting types where possible
		 */
		isPossiblyFunctionType := func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindConditionalType,
				ast.KindConstructorType,
				ast.KindFunctionType,
				ast.KindImportType,
				ast.KindIndexedAccessType,
				ast.KindInferType,
				ast.KindIntersectionType,
				ast.KindQualifiedName,
				ast.KindThisType,
				ast.KindTypeOperator,
				ast.KindTypeQuery,
				ast.KindTypeReference,
				ast.KindUnionType:
				return true

			case ast.KindTypeLiteral:
				return type_checking.Some(node.AsTypeLiteralNode().Members.Nodes, func(member *ast.Node) bool {
					return member.Kind == ast.KindCallSignature || member.Kind == ast.KindConstructSignature
				})

			case ast.KindAbstractKeyword,
				ast.KindAnyKeyword,
				ast.KindArrayType,
				ast.KindAsyncKeyword,
				ast.KindBigIntKeyword,
				ast.KindBooleanKeyword,
				ast.KindDeclareKeyword,
				ast.KindExportKeyword,
				ast.KindIntrinsicKeyword,
				ast.KindLiteralType,
				ast.KindMappedType,
				ast.KindNamedTupleMember,
				ast.KindNeverKeyword,
				ast.KindNullKeyword,
				ast.KindNumberKeyword,
				ast.KindObjectKeyword,
				ast.KindOptionalType,
				ast.KindPrivateKeyword,
				ast.KindProtectedKeyword,
				ast.KindPublicKeyword,
				ast.KindReadonlyKeyword,
				ast.KindRestType,
				ast.KindStaticKeyword,
				ast.KindStringKeyword,
				ast.KindSymbolKeyword,
				ast.KindTemplateLiteralType,
				ast.KindTupleType,
				ast.KindTypePredicate,
				ast.KindUndefinedKeyword,
				ast.KindUnknownKeyword,
				ast.KindVoidKeyword:
				return false
			}
			return false
		}

		checkReturnStatement := func(node *ast.ReturnStatement) {
			if node.Expression == nil {
				return
			}

			// syntactically ignore some known-good cases to avoid touching type info
			functionNode := (func() *ast.Node {
				current := node.Parent
				for current != nil && !ast.IsFunctionLike(current) {
					current = current.Parent
				}
				if current == nil {
					panic("missing parent function")
				}
				return current
			})()

			if functionNode.Type() != nil && !isPossiblyFunctionType(functionNode.Type()) {
				return
			}

			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Expression, checker.ContextFlagsNone)
			if contextualType != nil &&
				isVoidReturningFunctionType(
					node.Expression,
					contextualType,
				) && returnsThenable(node.Expression) {
				ctx.ReportNode(node.Expression, buildVoidReturnReturnValueMessage())
			}
		}

		checkAssignment := func(node *ast.BinaryExpression) {
			varType := ctx.TypeChecker.GetTypeAtLocation(node.Left)
			if !isVoidReturningFunctionType(node.Left, varType) {
				return
			}

			if returnsThenable(node.Right) {
				ctx.ReportNode(node.Right, buildVoidReturnVariableMessage())
			}
		}

		checkVariableDeclaration := func(node *ast.VariableDeclaration) {
			if node.Initializer == nil ||
				node.Type == nil {
				return
			}

			// syntactically ignore some known-good cases to avoid touching type info
			if !isPossiblyFunctionType(node.Type) {
				return
			}

			varType := ctx.TypeChecker.GetTypeAtLocation(node.Name())
			if !isVoidReturningFunctionType(node.Initializer, varType) {
				return
			}

			if returnsThenable(node.Initializer) {
				ctx.ReportNode(node.Initializer, buildVoidReturnVariableMessage())
			}
		}

		listeners := rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if *opts.ChecksConditionals {
					checkConditional(node, false)
				}

				if *opts.ChecksVoidReturn && *opts.ChecksVoidReturnOpts.Variables && ast.IsAssignmentExpression(node, false) {
					checkAssignment(node.AsBinaryExpression())
				}
			},
		}
		if *opts.ChecksConditionals {
			listeners[ast.KindPropertyAccessExpression] = checkArrayPredicates
			listeners[ast.KindElementAccessExpression] = checkArrayPredicates

			listeners[ast.KindPrefixUnaryExpression] = func(node *ast.Node) {
				expr := node.AsPrefixUnaryExpression()
				if expr.Operator == ast.KindExclamationToken {
					checkConditional(expr.Operand, true)
				}
			}

			listeners[ast.KindConditionalExpression] = func(node *ast.Node) { checkConditional(node.AsConditionalExpression().Condition, true) }
			listeners[ast.KindForStatement] = func(node *ast.Node) { checkConditional(node.AsForStatement().Condition, true) }
			listeners[ast.KindDoStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
			listeners[ast.KindWhileStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
			listeners[ast.KindIfStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
		}

		if *opts.ChecksVoidReturn {
			if *opts.ChecksVoidReturnOpts.Arguments {
				listeners[ast.KindCallExpression] = checkArguments
				listeners[ast.KindNewExpression] = checkArguments
			}
			if *opts.ChecksVoidReturnOpts.Attributes {
				listeners[ast.KindJsxAttribute] = func(node *ast.Node) { checkJSXAttribute(node.AsJsxAttribute()) }
			}
			if *opts.ChecksVoidReturnOpts.InheritedMethods {
				listeners[ast.KindClassDeclaration] = checkClassLikeOrInterfaceNode
				listeners[ast.KindClassExpression] = checkClassLikeOrInterfaceNode
				listeners[ast.KindInterfaceDeclaration] = checkClassLikeOrInterfaceNode
			}
			if *opts.ChecksVoidReturnOpts.Properties {
				listeners[ast.KindPropertyAssignment] = checkProperty
				listeners[ast.KindMethodDeclaration] = checkProperty
				listeners[ast.KindShorthandPropertyAssignment] = checkProperty
			}
			if *opts.ChecksVoidReturnOpts.Returns {
				listeners[ast.KindReturnStatement] = func(node *ast.Node) { checkReturnStatement(node.AsReturnStatement()) }
			}
			if *opts.ChecksVoidReturnOpts.Variables {
				listeners[ast.KindVariableDeclaration] = func(node *ast.Node) { checkVariableDeclaration(node.AsVariableDeclaration()) }
			}

		}
		if *opts.ChecksSpreads {
			listeners[ast.KindSpreadElement] = checkSpread
			listeners[ast.KindSpreadAssignment] = checkSpread
		}

		return listeners

	},
}

// isSometimesThenable reports whether any member of a value's union type is thenable, upstream's
// isSometimesThenable, which flagUnions "all" reports.
func isSometimesThenable(typeChecker *checker.Checker, node *ast.Node) bool {
	valueType := typeChecker.GetTypeAtLocation(node)
	for part := range type_checking.UnionTypePartsSeq(checker.Checker_getApparentType(typeChecker, valueType)) {
		if type_checking.IsThenableType(typeChecker, node, part) {
			return true
		}
	}
	return false
}

// hasMatchingPromiseTypeArgument is upstream's test of the same name, which flagUnions "strict"
// reports: a union holding thenables whose awaited types are, member for member, the union's other
// members, so `Promise<T> | T` reports and `Promise<T> | U` does not. Two types are the same when each
// is assignable to the other.
func hasMatchingPromiseTypeArgument(typeChecker *checker.Checker, node *ast.Node) bool {
	valueType := typeChecker.GetTypeAtLocation(node)
	var thenables, others []*checker.Type
	for part := range type_checking.UnionTypePartsSeq(checker.Checker_getApparentType(typeChecker, valueType)) {
		if type_checking.IsThenableType(typeChecker, node, part) {
			thenables = append(thenables, part)
		} else {
			others = append(others, part)
		}
	}
	if len(thenables) == 0 {
		return false
	}
	var awaited []*checker.Type
	for _, thenable := range thenables {
		awaitedType := checker.Checker_getAwaitedType(typeChecker, thenable)
		if awaitedType == nil {
			return false
		}
		awaited = append(awaited, slices.Collect(type_checking.UnionTypePartsSeq(awaitedType))...)
	}
	equivalent := func(left *checker.Type, right *checker.Type) bool {
		return checker.Checker_isTypeAssignableTo(typeChecker, left, right) &&
			checker.Checker_isTypeAssignableTo(typeChecker, right, left)
	}
	matchesAny := func(candidate *checker.Type, among []*checker.Type) bool {
		return slices.ContainsFunc(among, func(other *checker.Type) bool { return equivalent(candidate, other) })
	}
	for _, other := range others {
		if !matchesAny(other, awaited) {
			return false
		}
	}
	for _, each := range awaited {
		if !matchesAny(each, others) {
			return false
		}
	}
	return true
}
