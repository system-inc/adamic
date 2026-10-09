package typescript

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferNullishCoalescingPrimitives says which primitive types are exempt.
//
// Upstream's `ignorePrimitives` is a `oneOf`: this object, or the bare literal `true`. `IgnoreAll`
// carries the second spelling rather than expanding it into four trues, because the two are NOT
// interchangeable at the config surface even though they behave identically once decoded, and
// collapsing them would lose the ability to refuse what upstream refuses. See the decoder.
type PreferNullishCoalescingPrimitives struct {
	IgnoreAll bool
	Bigint    bool
	Boolean   bool
	Number    bool
	String    bool
}

// PreferNullishCoalescingOptions is the decoded option surface.
//
// # Every default here was PROVED by a moving verdict, not read off upstream's defaultOptions block
//
// A default read from a source block is a claim about what upstream wrote; a default proved by
// driving the installed rule with the key absent, explicitly false, and explicitly true is a fact
// about what runs. Measured on the installed 8.67.0 build, one source per option chosen so the
// option is the only thing that can change the verdict:
//
//	option                          absent   explicit false   explicit true   default
//	ignoreConditionalTests          silent   reports          silent          TRUE
//	ignoreTernaryTests              reports  reports          silent          false
//	ignoreMixedLogicalExpressions   reports  reports          silent          false
//	ignoreBooleanCoercion           reports  reports          silent          false
//	ignoreIfStatements              reports  reports          silent          false
//
// `ignoreConditionalTests` defaulting TRUE is the one that matters, and it is the second non-zero
// default found in two batches. A port assuming Go's zero value reports on every `if (a || b)` in
// the tree, which is plausible enough that nobody would question the count.
type PreferNullishCoalescingOptions struct {
	// AllowWithoutStrictNullChecks suppresses the whole-file `noStrictNullCheck` complaint.
	AllowWithoutStrictNullChecks bool
	// IgnoreBooleanCoercion exempts arguments to the `Boolean` constructor.
	IgnoreBooleanCoercion bool
	// IgnoreConditionalTests exempts a `||` inside a conditional test. Defaults TRUE.
	IgnoreConditionalTests bool
	// IgnoreIfStatements exempts an `if` that could become `??=`.
	IgnoreIfStatements bool
	// IgnoreMixedLogicalExpressions exempts a `||` that shares an expression with `&&`.
	IgnoreMixedLogicalExpressions bool
	// IgnorePrimitives exempts nullable unions of the named primitives.
	IgnorePrimitives PreferNullishCoalescingPrimitives
	// IgnoreTernaryTests exempts a ternary that could become `??`.
	IgnoreTernaryTests bool
	// ReactCompiler is ours, not upstream's, and defaults to TRUE: the if-statement check is silent
	// inside a function React Compiler compiles, because the `??=` it suggests is a shorthand the
	// compiler refuses. It is `logical-assignment-operators`' option of the same name, for the same
	// reason; see preferNullishCoalescingJudgeIf.
	ReactCompiler bool
}

// DefaultPreferNullishCoalescingOptions is upstream's defaultOptions, with the one non-false entry,
// and the React Compiler assumed on.
func DefaultPreferNullishCoalescingOptions() PreferNullishCoalescingOptions {
	return PreferNullishCoalescingOptions{IgnoreConditionalTests: true, ReactCompiler: true}
}

type preferNullishCoalescingWire struct {
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
	IgnoreBooleanCoercion                                  *bool `json:"ignoreBooleanCoercion"`
	IgnoreConditionalTests                                 *bool `json:"ignoreConditionalTests"`
	IgnoreIfStatements                                     *bool `json:"ignoreIfStatements"`
	IgnoreMixedLogicalExpressions                          *bool `json:"ignoreMixedLogicalExpressions"`
	IgnoreTernaryTests                                     *bool `json:"ignoreTernaryTests"`
	ReactCompiler                                          *bool `json:"reactCompiler"`
	// Polymorphic, so raw. No struct tag can express `oneOf(object, enum:[true])`, and this is the
	// same shape `ban_ts_comment` keeps raw for the same reason.
	IgnorePrimitives json.RawMessage `json:"ignorePrimitives"`
}

// DecodePreferNullishCoalescingOptions reads the option object off the config.
//
// # `ignorePrimitives` is a oneOf, and BOTH spellings were pinned before this was written
//
// The schema is an object of four booleans OR the bare literal `true`. Driving the installed rule
// across every spelling, on one nullable union per primitive:
//
//	spelling            string  number  boolean  bigint
//	absent                 1       1       1        1
//	{}                     1       1       1        1     an empty object exempts NOTHING
//	true                   0       0       0        0     the bare literal exempts everything
//	false                REFUSED REFUSED REFUSED REFUSED  `enum: [true]`, so false is not legal
//	{string: true}         0       1       1        1     per-primitive, as written
//	{string: false}        1       1       1        1
//	all four true          0       0       0        0
//
// Three of those rows are invisible to the corpus. `{}` and `all four true` bracket the object form
// from both ends; `false` is REFUSED BY THE SCHEMA rather than treated as "ignore nothing", which is
// the row a decoder would most naturally get wrong by accepting it and doing something reasonable.
// Reproduced as a refusal, because silently accepting a spelling upstream rejects means a config
// that loads here and fails there.
//
// The corpus splits 110 object-form against 25 bare-`true`, so a decoder handling only the object
// form still passes 110 cases. That is why this was measured first rather than tested afterwards.
func DecodePreferNullishCoalescingOptions(raw []byte) (any, error) {
	options := DefaultPreferNullishCoalescingOptions()
	if len(raw) == 0 {
		return options, nil
	}

	var wire preferNullishCoalescingWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}

	if wire.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing != nil {
		options.AllowWithoutStrictNullChecks = *wire.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing
	}
	if wire.IgnoreBooleanCoercion != nil {
		options.IgnoreBooleanCoercion = *wire.IgnoreBooleanCoercion
	}
	if wire.IgnoreConditionalTests != nil {
		options.IgnoreConditionalTests = *wire.IgnoreConditionalTests
	}
	if wire.IgnoreIfStatements != nil {
		options.IgnoreIfStatements = *wire.IgnoreIfStatements
	}
	if wire.IgnoreMixedLogicalExpressions != nil {
		options.IgnoreMixedLogicalExpressions = *wire.IgnoreMixedLogicalExpressions
	}
	if wire.IgnoreTernaryTests != nil {
		options.IgnoreTernaryTests = *wire.IgnoreTernaryTests
	}
	if wire.ReactCompiler != nil {
		options.ReactCompiler = *wire.ReactCompiler
	}

	primitives, err := decodePreferNullishCoalescingPrimitives(wire.IgnorePrimitives)
	if err != nil {
		return options, err
	}
	options.IgnorePrimitives = primitives
	return options, nil
}

// errPreferNullishCoalescingPrimitivesFalse names the one boolean upstream's schema refuses.
//
// `ignorePrimitives` is `oneOf(object, {type: boolean, enum: [true]})`, so `true` is legal and
// `false` is not. Measured: ESLint rejects `{"ignorePrimitives": false}` at config load with
// "Value false should match exactly one schema in oneOf".
var errPreferNullishCoalescingPrimitivesFalse = errors.New(
	"prefer-nullish-coalescing: ignorePrimitives accepts an object or the literal true; " +
		"false is not a legal value, and upstream refuses it at config load")

// decodePreferNullishCoalescingPrimitives reads the polymorphic key, in upstream's own order.
func decodePreferNullishCoalescingPrimitives(raw json.RawMessage) (
	PreferNullishCoalescingPrimitives, error) {

	var primitives PreferNullishCoalescingPrimitives
	if len(raw) == 0 {
		return primitives, nil
	}

	// The boolean arm first, because `enum: [true]` makes exactly one boolean legal. `false` is
	// refused rather than read as "ignore nothing": upstream's schema rejects it at config load, and
	// accepting it here would let a config load in this tree and fail in ESLint.
	var asBool bool
	if json.Unmarshal(raw, &asBool) == nil {
		if !asBool {
			return primitives, errPreferNullishCoalescingPrimitivesFalse
		}
		primitives.IgnoreAll = true
		return primitives, nil
	}

	// An unknown key is refused, matching `additionalProperties: false`, so a project writing
	// `{strings: true}` for `{string: true}` is told rather than silently exempting nothing.
	var asObject struct {
		Bigint  *bool `json:"bigint"`
		Boolean *bool `json:"boolean"`
		Number  *bool `json:"number"`
		String  *bool `json:"string"`
	}
	if err := rule.UnmarshalOptions(raw, &asObject); err != nil {
		return primitives, err
	}
	primitives.Bigint = asObject.Bigint != nil && *asObject.Bigint
	primitives.Boolean = asObject.Boolean != nil && *asObject.Boolean
	primitives.Number = asObject.Number != nil && *asObject.Number
	primitives.String = asObject.String != nil && *asObject.String
	return primitives, nil
}

// messagePreferNullishOverOr is the one message id this rule reports, and it renders TWO ways.
//
// Upstream's text interpolates `{{description}}` and `{{equals}}`:
//
//	a || b     "...(`??`) instead of a logical or (`||`)..."
//	a ||= b    "...(`??=`) instead of a logical assignment (`||=`)..."
//
// One id, two renderings, and a fixture asserting only the id cannot tell them apart. That is
// section 3c exactly, and it is sharpest here because BOTH arms also propose different repair text
// (`??` against `??=`), so a rule that rendered one arm's message while writing the other arm's fix
// would pass every message-id assertion in the corpus.
//
// So the description is built per arm rather than being one constant, and the fixtures assert the
// whole message rather than a prefix.
var messagePreferNullishOverOr = rule.Message{Id: "preferNullishOverOr"}

var messagePreferNullishSuggest = rule.Message{Id: "suggestNullish"}

// preferNullishCoalescingDescribe renders the finding for one arm.
//
// `operator` is the source spelling being replaced (`||` or `||=`) and `replacement` is what would
// replace it. Upstream's `equals` is the empty string for `||` and `=` for `||=`, which is the same
// distinction expressed as a suffix; carried here as whole spellings because that is what the fix
// writes and keeping one source for both removes the chance of them disagreeing.
func preferNullishCoalescingDescribe(operator string, replacement string, kind string) string {
	return "Prefer using nullish coalescing operator (`" + replacement + "`) instead of a logical " +
		kind + " (`" + operator + "`), as it is a safer operator. `" + operator + "` falls through " +
		"on every falsy value, so an empty string, a zero and a false all take the right-hand " +
		"branch alongside null and undefined. Where the left side is nullable, `" + replacement +
		"` says what was meant and leaves the valid falsy values alone."
}

// preferNullishCoalescingSuggestion renders the suggestion text, which also varies per arm.
func preferNullishCoalescingSuggestion(replacement string) string {
	return "Fix to nullish coalescing operator (`" + replacement + "`). Only null and undefined " +
		"will take the right-hand branch."
}

// PreferNullishCoalescing asks for `??` where `||` is guarding against null or undefined.
//
//	valid:   declare const a: string; const x = a || 'b';        not nullable, so `||` is a real choice
//	valid:   declare const a: string | null; if (a || 'b') {}    ignoreConditionalTests defaults TRUE
//	invalid: declare const a: string | null; const x = a || 'b';
//
//	invalid: declare const a: string | null; const x = a !== null ? a : 'b';
//	invalid: declare let a: string | null; if (!a) a = 'b';
//
// # Three reporting paths, all ported
//
// Upstream reports from three listeners: `||` and `||=` (its `preferNullishOverOr`, this file), a
// ternary that tests for null or undefined and returns what it tested (`preferNullishOverTernary`),
// and an `if` with no `else` that assigns only when its target is nullish (`preferNullishOverAssignment`).
// The second and third live in prefer_nullish_coalescing_ternary.go, because they share one reading of
// a test that the `||` arm never needs.
//
// For a month the rule shipped with only the first path while its option decoder parsed
// `ignoreTernaryTests` and `ignoreIfStatements` and the set enabled it with both on, so a config could
// not tell. With the rule on in both engines, ESLint found 64 sites in ahra that cohere never reported
// (55 ternaries, 9 if-assignments) (#10kqgs6). The options now gate real checks, and the corpus test
// counts every path so a missing one fails by name.
//
// The `noStrictNullCheck` whole-file complaint is still absent: it needs a program compiled without
// strictNullChecks, which no tree we lint has.
//
// # The type gate is upstream's `isNullableType`, and the shelf's IsTypeFlagSet is the WRONG helper
//
// This is the same trap as `consistent-return`'s, in a far worse form. Upstream imports two
// different functions both spelled `isTypeFlagSet`:
//
//	isNullableType(type)              typescript-eslint's, which decomposes a union FIRST
//	tsutils.isTypeFlagSet(type, ...)  ts-api-utils', which reads the type's OWN flags
//
// `checking.IsTypeFlagSet` is the second reading. It is correct for the two `ignorePrimitives` call
// sites and wrong for this gate, so which spelling a call site wants is a PER-CALL-SITE question
// rather than a per-rule one. Reading the import once and applying the answer throughout is the
// natural shortcut and it is exactly wrong here.
//
// The cost is not a few cases, it is the whole rule. `isNullableType` is the FIRST gate, and every
// shape this rule fires on is a nullable UNION, which carries the Union flag rather than its
// constituents'. Measured across eight shapes, then scored as a consequence with a control:
//
//	string | null          shelf false   upstream true    DIVERGE
//	string | undefined     shelf false   upstream true    DIVERGE
//	number | null          shelf false   upstream true    DIVERGE
//	{ a?: string }.a       shelf false   upstream true    DIVERGE
//	string                 shelf false   upstream false   agree
//	null                   shelf true    upstream true    agree
//	any                    shelf true    upstream true    agree
//
//	on the rule's own target shapes:  upstream 4/4 nullable, shelf 0/4
//
// A port built on the shelf helper reports zero on every real file AND is green, because an author
// who misread the helper writes fixtures matching what their code does. No mutation sweep and no
// span assertion can vote on that. Hence the local helper below, with a fixture on each of the four.
//
// # The span is the OPERATOR token, which is also why optional chaining costs nothing here
//
// Upstream reports `barBarOperator`, the `||` token itself, rather than the expression. Measured
// against the installed 8.67.0 build: `a || b` reports columns 13 to 15 for a left operand ending at
// 12.
//
// That anchor is also why ESTree's ChainExpression wrapper never gates this rule: `a?.b || c`
// reports exactly like `a || c`, because the rule looks at the operator and not at the member
// access. Predicted as a divergence, measured as a non-divergence, and stated here so the next
// reader does not re-derive the worry.
var PreferNullishCoalescing = rule.Rule{
	Name:             "@typescript-eslint/prefer-nullish-coalescing",
	NeedsTypeChecker: true,
	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[PreferNullishCoalescingOptions](options)
		if !ok {
			settings = DefaultPreferNullishCoalescingOptions()
		}

		// One listener over the whole file, rather than one per binary expression, so findings can be
		// emitted in SOURCE order.
		//
		// `||` is left-associative, so `a || b || c` parses with the OUTER node holding the LATER
		// operator, and a pre-order walk reaches that outer node first. Reported as visited, the
		// findings come out right-to-left.
		//
		// No fixture in the imported corpus can see this: the corpus asserts message ids and counts,
		// and a reordered list of identical ids is the same list. It surfaced only once the spans and
		// the per-finding suggestion texts were pinned, where consecutive findings carry DIFFERENT
		// repairs and swapping them is visible. `consistent-return` needed the same treatment for
		// the same reason.
		return rule.Listeners{
			ast.KindSourceFile: func(file *ast.Node) {
				var pending []preferNullishCoalescingFinding

				var visit func(*ast.Node) bool
				visit = func(current *ast.Node) bool {
					var finding preferNullishCoalescingFinding
					reports := false
					switch current.Kind {
					case ast.KindBinaryExpression:
						finding, reports = preferNullishCoalescingJudge(ctx, current, settings)
					case ast.KindConditionalExpression:
						finding, reports = preferNullishCoalescingJudgeTernary(ctx, current, settings)
					case ast.KindIfStatement:
						finding, reports = preferNullishCoalescingJudgeIf(ctx, current, settings)
					}
					if reports {
						pending = append(pending, finding)
					}
					current.ForEachChild(visit)
					return false
				}
				file.ForEachChild(visit)

				sort.SliceStable(pending, func(left, right int) bool {
					return pending[left].reportRange.Pos() < pending[right].reportRange.Pos()
				})
				for _, finding := range pending {
					ctx.ReportRangeWithSuggestions(finding.reportRange, finding.message,
						rule.Suggestion{Message: finding.suggestion, Fixes: finding.fixes})
				}
			},
		}
	},
}

// preferNullishCoalescingFinding is one pending report, held until the file is walked.
//
// `reportRange` is the `||` token for the first path and the whole ternary or `if` for the other two,
// which is where upstream anchors each.
type preferNullishCoalescingFinding struct {
	reportRange core.TextRange
	message     rule.Message
	suggestion  rule.Message
	fixes       []rule.Fix
}

// preferNullishCoalescingJudge decides one binary expression.
//
// Upstream registers two listeners, `LogicalExpression[operator="||"]` and
// `AssignmentExpression[operator="||="]`, both calling the same check. Our parser makes both a
// KindBinaryExpression, so they are separated here by the operator token. The `||=` half is 32 of
// the 123 preferNullishOverOr cases, so omitting it is not a rounding error.
func preferNullishCoalescingJudge(ctx rule.Context, node *ast.Node,
	settings PreferNullishCoalescingOptions) (preferNullishCoalescingFinding, bool) {

	expression := node.AsBinaryExpression()
	var operator, replacement, kind string
	switch expression.OperatorToken.Kind {
	case ast.KindBarBarToken:
		operator, replacement, kind = "||", "??", "or"
	case ast.KindBarBarEqualsToken:
		// Upstream calls this arm "assignment" rather than "or assignment", measured against the
		// installed build rather than inferred from the option name.
		operator, replacement, kind = "||=", "??=", "assignment"
	default:
		return preferNullishCoalescingFinding{}, false
	}

	if !preferNullishCoalescingEligible(ctx, node, expression.Left, settings) {
		return preferNullishCoalescingFinding{}, false
	}
	if settings.IgnoreMixedLogicalExpressions && preferNullishCoalescingIsMixed(node) {
		return preferNullishCoalescingFinding{}, false
	}

	operatorRange := rule.TokenRange(ctx.SourceFile, expression.OperatorToken)
	return preferNullishCoalescingFinding{
		reportRange: operatorRange,
		message: rule.Message{
			Id:          messagePreferNullishOverOr.Id,
			Description: preferNullishCoalescingDescribe(operator, replacement, kind),
		},
		suggestion: rule.Message{
			Id:          messagePreferNullishSuggest.Id,
			Description: preferNullishCoalescingSuggestion(replacement),
		},
		fixes: preferNullishCoalescingRepair(ctx, node, operatorRange, replacement),
	}, true
}

// preferNullishCoalescingRepair builds the suggested edit, parenthesising where `??` needs it.
//
// # A chained `||` cannot simply become `??`, and the corpus never says so
//
// `??` may not sit beside `||` or `&&` without parentheses -- `a ?? b || c` is a syntax error -- so
// upstream does not replace the operator when the expression has a logical parent. It replaces the
// whole LEFT-HAND expression with a parenthesised nullish form. Measured against the installed
// 8.67.0 build, reading the suggestion text rather than the message:
//
//	a || b                 one finding,   suggests `??`
//	a || b || c            two findings,  the first suggests `(a ?? b)`, the second `??`
//	a || (b || c)          two findings,  the first suggests `??`,       the second `(b ?? c)`
//	a || b || c || d       three findings, `(a ?? b)`, `(b ?? c)`, then `??`
//	(a || b) && d          one finding,   suggests `??`
//
// The pattern is per-finding rather than per-expression: a `||` whose PARENT is also a `||` gets the
// parenthesised form, and the outermost one gets the plain operator swap. `(a || b) && d` is the row
// that shows the test is on the parent being a logical OR specifically, not on being any logical
// operator, since that one is silenced only by ignoreMixedLogicalExpressions and still suggests the
// plain swap.
//
// **No fixture in the imported corpus can see this**, because a suggestion's text is not a message
// id and this rule has exactly one id for both arms. A port writing `??` everywhere passes all 398
// cases and proposes a syntax error on every chained expression it touches. Found by reading the
// suggestion text out of the installed rule while measuring spans, not by the corpus.
func preferNullishCoalescingRepair(ctx rule.Context, node *ast.Node, operatorRange core.TextRange,
	replacement string) []rule.Fix {

	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	parentIsLogicalOr := parent != nil && parent.Kind == ast.KindBinaryExpression &&
		(parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken ||
			parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarEqualsToken)

	if !parentIsLogicalOr {
		return []rule.Fix{rule.ReplaceRange(operatorRange, replacement)}
	}

	// The expression becomes `(left ?? right)`, so the parenthesis carries the precedence the bare
	// swap would have broken.
	//
	// For a LEFT operand that is itself a `||` chain, upstream renders only its immediate RIGHT
	// operand rather than the whole subtree. Measured on `a || b || c || d`, whose middle finding
	// suggests `(b ?? c)` and not `(a || b ?? c)`: each finding in a chain is an independent repair
	// describing what that one finding would write if applied alone, rather than a step in a shared
	// rewrite.
	expression := node.AsBinaryExpression()
	leftNode := expression.Left
	for {
		unwrapped := preferNullishCoalescingUnwrap(leftNode)
		if unwrapped == nil || unwrapped.Kind != ast.KindBinaryExpression {
			break
		}
		operator := unwrapped.AsBinaryExpression().OperatorToken.Kind
		if operator != ast.KindBarBarToken && operator != ast.KindBarBarEqualsToken {
			break
		}
		leftNode = unwrapped.AsBinaryExpression().Right
	}

	text := ctx.SourceFile.Text()
	leftRange := rule.TokenRange(ctx.SourceFile, leftNode)
	rightRange := rule.TokenRange(ctx.SourceFile, expression.Right)
	return []rule.Fix{rule.ReplaceRange(
		core.NewTextRange(leftRange.Pos(), rightRange.End()),
		"("+text[leftRange.Pos():leftRange.End()]+" "+replacement+" "+
			text[rightRange.Pos():rightRange.End()]+")")}
}

// preferNullishCoalescingEligible is upstream's `isTruthinessCheckEligibleForPreferNullish`.
func preferNullishCoalescingEligible(ctx rule.Context, node *ast.Node, testNode *ast.Node,
	settings PreferNullishCoalescingOptions) bool {

	if settings.IgnoreConditionalTests && preferNullishCoalescingIsConditionalTest(node) {
		return false
	}
	// A ternary passed straight to `Boolean(...)` is still reported: upstream exempts the coercion
	// context everywhere except there, since `Boolean(a ? a : b)` and `Boolean(a ?? b)` can differ
	// when `a` is a falsy non-nullish value. Measured, and pinned by upstream's own corpus.
	if settings.IgnoreBooleanCoercion && preferNullishCoalescingInBooleanCall(ctx, node) &&
		!(node.Kind == ast.KindConditionalExpression &&
			preferNullishCoalescingParentSkippingParentheses(node) != nil &&
			preferNullishCoalescingParentSkippingParentheses(node).Kind == ast.KindCallExpression) {
		return false
	}
	if ctx.TypeChecker == nil || testNode == nil {
		return false
	}
	return preferNullishCoalescingTypeEligible(ctx,
		ctx.TypeChecker.GetTypeAtLocation(testNode), settings)
}

// preferNullishCoalescingTypeEligible is upstream's `isTypeEligibleForPreferNullish`.
//
// The order is upstream's and it matters: nullable first, then the `any`/`unknown` bail, then the
// per-primitive exemptions. A type that is not nullable is never reported, whatever else is true.
func preferNullishCoalescingTypeEligible(ctx rule.Context, subject *checker.Type,
	settings PreferNullishCoalescingOptions) bool {

	if !preferNullishCoalescingIsNullable(subject) {
		return false
	}

	ignorable := preferNullishCoalescingIgnorableFlags(settings.IgnorePrimitives)
	if ignorable == 0 {
		return true
	}

	// `any` and `unknown` could be any primitive at runtime even though the flags are not set, so
	// upstream declines them once ANY primitive is being ignored. This one reads the type's own
	// flags, which is ts-api-utils' spelling and is what `checking.IsTypeFlagSet` provides.
	if type_checking.IsTypeFlagSet(subject, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return false
	}

	for part := range type_checking.UnionTypePartsSeq(subject) {
		for _, intersected := range type_checking.IntersectionTypeParts(part) {
			if type_checking.IsTypeFlagSet(intersected, ignorable) {
				return false
			}
		}
	}
	return true
}

// preferNullishCoalescingIsNullable is upstream's `isNullableType`, which decomposes the union.
//
// NOT `checking.IsTypeFlagSet`. See the rule's doc comment for the measurement: the shelf helper
// answers false for every nullable union, which is every shape this rule reports on.
//
// Written local rather than added to the shelf, because `checking.IsTypeFlagSet` is not wrong -- it
// is the other reading, and two of this rule's own call sites want it. A second shelf function whose
// name differs by a word is how two callers end up asking different questions believing they asked
// the same one.
func preferNullishCoalescingIsNullable(subject *checker.Type) bool {
	const nullish = checker.TypeFlagsAny | checker.TypeFlagsUnknown |
		checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid
	return preferNullishCoalescingUnionFlagSet(subject, nullish)
}

// preferNullishCoalescingUnionFlagSet is typescript-eslint's `getTypeFlags`: OR every union
// constituent's flags before masking.
func preferNullishCoalescingUnionFlagSet(subject *checker.Type, flags checker.TypeFlags) bool {
	if subject == nil {
		return false
	}
	var combined checker.TypeFlags
	for part := range type_checking.UnionTypePartsSeq(subject) {
		if part != nil {
			combined |= checker.Type_flags(part)
		}
	}
	return combined&flags != 0
}

// preferNullishCoalescingIgnorableFlags turns the decoded primitives into one mask.
//
// The `IgnoreAll` arm is upstream's `ignorePrimitives === true`, which sets every flag rather than
// short-circuiting, so the `any`/`unknown` bail below still applies. Reproduced that way rather than
// as an early return, because the two differ on an `any` left operand.
func preferNullishCoalescingIgnorableFlags(
	primitives PreferNullishCoalescingPrimitives) checker.TypeFlags {

	var ignorable checker.TypeFlags
	if primitives.IgnoreAll || primitives.Bigint {
		ignorable |= checker.TypeFlagsBigIntLike
	}
	if primitives.IgnoreAll || primitives.Boolean {
		ignorable |= checker.TypeFlagsBooleanLike
	}
	if primitives.IgnoreAll || primitives.Number {
		ignorable |= checker.TypeFlagsNumberLike
	}
	if primitives.IgnoreAll || primitives.String {
		ignorable |= checker.TypeFlagsStringLike
	}
	return ignorable
}

// preferNullishCoalescingInBooleanCall is upstream's `isBooleanConstructorContext`.
//
// A value flowing into `Boolean(...)` is being coerced to a boolean, so `||` and `??` cannot differ
// in the result and the rule has nothing to say. The walk follows the same three shapes as the
// conditional-test walk, since a value reaches the call through them: a logical expression, either
// branch of a ternary, and the last element of a sequence.
//
// # The callee must be the GLOBAL Boolean, and the checker answers that better than upstream can
//
// Upstream resolves the callee through eslint-scope and accepts it when `variable == null ||
// variable.defs.length === 0`, which is "unresolvable, or declared nowhere in source" -- what a
// global looks like from a scope manager that cannot see `lib.es5.d.ts`.
//
// A first draft skipped this check and documented the gap as narrow and one-directional. Running the
// corpus disproved that: upstream's `invalid[297]` shadows `Boolean` with a local arrow and REPORTS,
// and the draft was silent on it. The gap was real and the note was wrong about it being acceptable.
//
// The checker answers the question directly, and better: it resolves the symbol and says where it
// was declared. Probed across three shapes:
//
//	Boolean(a || b)                            declared in bundled:///libs/lib.es5.d.ts
//	const Boolean = ...; inside a function     declared in the source file
//	const Boolean = ...; at module scope       declared in the source file
//
// `checking.IsSymbolFromDefaultLibrary` is the shelf's existing question for exactly this, so this
// is a caller rather than a reimplementation, and it needs no scope resolver at all.
func preferNullishCoalescingInBooleanCall(ctx rule.Context, node *ast.Node) bool {
	for current := node; current != nil; {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			if call.Expression == nil || call.Expression.Kind != ast.KindIdentifier ||
				call.Expression.Text() != "Boolean" {
				return false
			}
			if ctx.TypeChecker == nil || ctx.Program == nil {
				return false
			}
			if !type_checking.IsSymbolFromDefaultLibrary(ctx.Program,
				ctx.TypeChecker.GetSymbolAtLocation(call.Expression)) {
				return false
			}
			// Only the FIRST argument is coerced; `Boolean(x, y)` reads y for its side effects.
			//
			// The argument is unwrapped because upstream's parser folds parentheses away, so
			// `Boolean(((a = b), b || c))` presents its sequence directly. Without the unwrap the
			// argument is a KindParenthesizedExpression and never equals the node walked up to.
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
				return false
			}
			// BOTH sides are unwrapped, and that took a probe to get right. The walk arrives here
			// holding whatever node it last stepped to, which for `Boolean(((a = b), b || c))` is
			// the KindParenthesizedExpression wrapping the sequence, while the argument list holds
			// that same parenthesized node. Unwrapping only the argument compares a bare sequence
			// against a parenthesized one and answers false.
			//
			// Upstream needs neither unwrap, because its parser folded both away before the walk.
			return preferNullishCoalescingUnwrap(call.Arguments.Nodes[0]) ==
				preferNullishCoalescingUnwrap(current)
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional.WhenTrue != current && conditional.WhenFalse != current {
				return false
			}
			current = parent
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			// ESTree's LogicalExpression covers `??` alongside `||` and `&&`, so upstream's
			// recursion reaches through all three. Ours has to name `??` explicitly, and
			// upstream's corpus pins it with `Boolean(a ?? (b || c))`.
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken,
				ast.KindQuestionQuestionToken:
				current = parent
			case ast.KindCommaToken:
				if binary.Right != current {
					return false
				}
				current = parent
			default:
				return false
			}
		default:
			return false
		}
	}
	return false
}

// preferNullishCoalescingIsConditionalTest is upstream's `isConditionalTest`, walking up through the
// shapes that pass a truthiness test along: a logical expression, either branch of a ternary, and
// the last element of a sequence.
func preferNullishCoalescingIsConditionalTest(node *ast.Node) bool {
	for current := node; current != nil; {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindIfStatement:
			return parent.AsIfStatement().Expression == current
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional.Condition == current {
				return true
			}
			// A value flowing out of either branch is still in the enclosing test position.
			if conditional.WhenTrue == current || conditional.WhenFalse == current {
				current = parent
				continue
			}
			return false
		case ast.KindWhileStatement:
			return parent.AsWhileStatement().Expression == current
		case ast.KindDoStatement:
			return parent.AsDoStatement().Expression == current
		case ast.KindForStatement:
			return parent.AsForStatement().Condition == current
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			// `??` is a LogicalExpression in ESTree, so upstream's recursion reaches through it.
			// Pinned by `if ((a || b) ?? c)` and `if (a ?? (b || c))`, both upstream valid cases.
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken,
				ast.KindQuestionQuestionToken:
				current = parent
				continue
			case ast.KindCommaToken:
				// Only the last element of a sequence carries the value onward.
				if binary.Right == current {
					current = parent
					continue
				}
			}
			return false
		case ast.KindPrefixUnaryExpression:
			// `if (!(a || b))` and `if (!!(a || b))` are still conditional tests: the negation is
			// consumed by the test rather than escaping it. Upstream reaches this through its
			// UnaryExpression recursion and its corpus pins both depths.
			if parent.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
				return false
			}
			current = parent
			continue
		case ast.KindParenthesizedExpression:
			// Our parser keeps this node where upstream's folds it away, so walking through it is
			// what makes `if ((a || b))` behave like `if (a || b)`.
			current = parent
			continue
		default:
			return false
		}
	}
	return false
}

// preferNullishCoalescingParentSkippingParentheses is the node's ESTree parent: the first ancestor
// that is not a pair of parentheses.
func preferNullishCoalescingParentSkippingParentheses(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

// preferNullishCoalescingUnwrap strips parentheses, which upstream's parser folds away.
//
// A loop, because `((a))` nests. Not `ast.SkipParentheses`, which dereferences its argument and is
// how this project lost 167 files to a nil panic; every caller here holds an optional node.
func preferNullishCoalescingUnwrap(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// preferNullishCoalescingIsMixed is upstream's `isMixedLogicalExpression`.
//
// A breadth walk over the parent and both operands, following `||` links and answering true on the
// first `&&`. The `seen` set is upstream's and is load bearing rather than defensive: the queue
// pushes parents, so a node reachable both upward and downward would otherwise be revisited forever.
func preferNullishCoalescingIsMixed(node *ast.Node) bool {
	expression := node.AsBinaryExpression()
	seen := map[*ast.Node]bool{}
	queue := []*ast.Node{node.Parent, expression.Left, expression.Right}

	for index := 0; index < len(queue); index++ {
		// Parentheses are stripped on the way in rather than pushed as a node, because upstream's
		// parser folds them away before this walk runs. Without it `a || (b && c)` never reaches the
		// `&&`, and the walk stops at a KindParenthesizedExpression that upstream never sees.
		// Twenty-seven of upstream's own valid cases are exactly that shape, all under
		// ignoreMixedLogicalExpressions, and every one of them reported before this was added.
		current := preferNullishCoalescingUnwrap(queue[index])
		if current == nil || seen[current] {
			continue
		}
		seen[current] = true

		if current.Kind != ast.KindBinaryExpression {
			continue
		}
		switch current.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			return true
		case ast.KindBarBarToken, ast.KindBarBarEqualsToken:
			// Follow the `||` chain, so `a || b || c && d` is caught.
			inner := current.AsBinaryExpression()
			queue = append(queue, current.Parent, inner.Left, inner.Right)
		}
	}
	return false
}
