package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// RadixMode is upstream's deprecated option, which names a mode the rule no longer has.
//
// Kept as a string rather than dropped at decode time so a config reads back as what was written.
// Nothing in this file branches on it, which is upstream's own behavior: see "The option is
// deprecated and does nothing" on Radix.
type RadixMode string

const (
	// RadixAlways once meant "always require a radix", which is what the rule now always does.
	RadixAlways RadixMode = "always"

	// RadixAsNeeded once meant "report a radix of 10 as redundant". At 10.8.1 it reports exactly
	// what RadixAlways does.
	RadixAsNeeded RadixMode = "as-needed"
)

// DecodeRadixOptions turns the configured value into the rule's options.
//
// Upstream's schema is one optional element, `{enum: ["always", "as-needed"]}`, so a config writes
// `["error", "as-needed"]` and what arrives here is the bare JSON string. Anything the enum refuses
// is refused here too: another string, a number, an object, and `null`, which unmarshals into a Go
// string as the empty one and is caught by the same switch.
//
// A bare severity decodes to the zero mode. The rule ignores the mode either way, so there is no
// default to get wrong.
func DecodeRadixOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return RadixMode(""), nil
	}

	var configured string
	if json.Unmarshal(raw, &configured) == nil {
		switch mode := RadixMode(configured); mode {
		case RadixAlways, RadixAsNeeded:
			return mode, nil
		}
	}
	return RadixMode(""), fmt.Errorf("radix takes \"always\" or \"as-needed\", got %s",
		// Go whitespace: raw JSON bytes of the rule's options, echoed in cohere's own error message.
		strings.TrimSpace(string(raw)))
}

var messageRadixMissingParameters = rule.Message{
	Id: "missingParameters",
	Description: "This calls `parseInt` with no arguments at all, which returns NaN every time. " +
		"It parses nothing, so whatever this line was meant to read is not being read.",
}

var messageRadixMissingRadix = rule.Message{
	Id: "missingRadix",
	Description: "This calls `parseInt` without a radix. The base is then inferred from the text, " +
		"so a string that happens to start with `0x` parses as hexadecimal and a leading zero has " +
		"meant octal in some engines. Pass 10 explicitly when you mean decimal.",
}

var messageRadixInvalidRadix = rule.Message{
	Id: "invalidRadix",
	Description: "The radix here is not an integer between 2 and 36, so `parseInt` ignores it and " +
		"falls back to inferring the base from the text. It reads as a decision that has been made " +
		"and is not one.",
}

var messageRadixAddRadixParameter10 = rule.Message{
	Id:          "addRadixParameter10",
	Description: "Add a radix of 10 to parse this as a decimal number.",
}

// Radix flags a call to `parseInt` with no radix, or with one that is not a base it accepts.
//
//	valid:   parseInt("10", 10);
//	valid:   parseInt("10", foo);
//	valid:   var parseInt; parseInt();
//	invalid: parseInt("10");
//	invalid: parseInt("10", 1);
//	invalid: parseInt();
//
// # Three judgments over one call, and they are ordered
//
// No arguments is `missingParameters`, one argument is `missingRadix`, and two or more with a bad
// second is `invalidRadix`. They are exclusive, so a call reports at most once, and upstream's
// corpus states that: no case here carries two findings for one call.
//
// # A valid radix is an integer literal from 2 to 36, and everything else is a question
//
// The rule is deliberately narrow about what it will call invalid. A literal that is not an integer
// in range reports, and so does a bare `undefined`. Anything the rule cannot evaluate is accepted:
// `parseInt("10", foo)` and `parseInt("10", +radix)` and `parseInt("10", ~1)` are all clean, because
// a rule that guessed at them would report correct code. The one arithmetic form it does evaluate is
// a leading `+` or `-` on a numeric literal, which is why `+37` reports and `+10` does not.
//
// Our parser keeps a numeric literal's canonical text, so `0x10`, `1.6e1` and `10.0` all read as 16,
// 16 and 10 without a numeric parser being written here. All three are upstream passing cases.
//
// # A spread before the radix abandons the call
//
// `parseInt(...args)` could be any number of arguments, so nothing can be said about where the radix
// is, and upstream declines. A spread AFTER the first two positions does not hide anything, which is
// why `parseInt("10", 1, ...args)` still reports an invalid radix. The index test is upstream's and
// the boundary is exact.
//
// # Shadowing, and where this port is more precise than upstream
//
// A local `parseInt` is not the global one, so the rule has nothing to say about it. Upstream asks
// eslint-scope for the PROGRAM scope's variable and checks whether anything defines it, which makes
// its answer whole-file: measured against the installed rule, `{ let parseInt; } parseInt("10");`
// is CLEAN upstream even though the block-scoped binding cannot reach the call.
//
// This port resolves each call site instead, through `resolvesToAGlobal`, so that input reports
// here. The divergence is deliberate and it is in the direction of reporting a real defect that
// upstream misses, rather than of reporting correct code. It is pinned by a fixture that states
// both answers.
//
// The mechanism was probed before being built on: `parseInt` at a plain call site resolves to a
// declaration in the standard library, and every shadow form -- a `var`, a parameter, a `let` --
// resolves to a declaration in source. A type-only `interface parseInt {}` resolves to the library,
// which is right, since a type does not shadow a value.
//
// # Suggestions rather than fixes, and the trailing comma is why
//
// Upstream offers a suggestion rather than a fix, so a human chooses it. Adding a radix changes what
// the call returns for any string the reader was relying on being inferred, which is a meaning
// change rather than a spelling one, and that is exactly the line between the two.
//
// The insertion goes before the closing parenthesis and has to read what precedes it: with a
// trailing comma already there, upstream writes `" 10,"` rather than `", 10"`, turning
// `parseInt("10",)` into `parseInt("10", 10,)`. Both spellings are in the corpus.
//
// # The option is deprecated and does nothing
//
// `meta.schema` still accepts `"always"` or `"as-needed"`, marked `// deprecated` in
// eslint/lib/rules/radix.js at 10.8.1, and `create` never reads `context.options`. So the two values
// behave identically there, and identically to no option at all. Upstream's corpus states it row by
// row: `parseInt("10", 10)` is clean under `"as-needed"`, where older releases reported a redundant
// radix, and `parseInt("10")` reports `missingRadix` under `"as-needed"` as it does under
// `"always"`. Both values are accepted here by `DecodeRadixOptions` and ignored the same way, since
// refusing them would stop a whole run over a config ESLint accepts.
var Radix = rule.Rule{
	Name: "radix",

	// The shadow question is which file declares the name, which only resolution can answer.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()
				if call == nil || call.Expression == nil {
					return
				}
				if !isAParseIntCallee(ctx, call.Expression) {
					return
				}
				checkRadixArguments(ctx, node, call)
			},
		}
	},
}

// isAParseIntCallee says whether a callee names the global `parseInt` or `Number.parseInt`.
//
// Both spellings of the member form are accepted, since `Number["parseInt"]` and Number[`parseInt`]
// are the same call and upstream reports all three. A computed access whose key is not a literal
// string is not, which is what keeps `Number[parseInt]()` clean.
func isAParseIntCallee(ctx rule.Context, callee *ast.Node) bool {
	// The optional-call form `parseInt?.()` and the optional-member form `Number?.parseInt()` reach
	// here as the same node kinds, so nothing special is needed for them. A parenthesized callee
	// does need unwrapping: `(Number?.parseInt)("10")` is one of upstream's failing cases, and our
	// parser keeps the parentheses that upstream's folds away.
	for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
	}
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "parseInt" && resolvesToAGlobal(ctx, callee)

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		// A private name is not a property of `Number`, and `Number.#parseInt()` is one of
		// upstream's passing cases.
		//
		// The KIND half of this guard is EQUIVALENT rather than load bearing, and it is kept
		// because it states the intent where a reader looks for it and because it does not depend
		// on how a private name renders. Our parser gives a private identifier a Text() of
		// `#parseInt`, hash included, so the string comparison below already declines every one of
		// them. Measured over seven shapes including a private `#parseInt` called with a radix and
		// a class named to collide with the receiver: byte identical findings with the kind test
		// present and absent.
		if access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return false
		}
		return access.Name().Text() == "parseInt" && isTheGlobalNumber(ctx, access.Expression)

	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		return isTheLiteralText(access.ArgumentExpression, "parseInt") &&
			isTheGlobalNumber(ctx, access.Expression)
	}
	return false
}

// isTheGlobalNumber says whether a receiver is the global `Number` rather than something shadowing.
func isTheGlobalNumber(ctx rule.Context, receiver *ast.Node) bool {
	for receiver != nil && receiver.Kind == ast.KindParenthesizedExpression {
		receiver = receiver.AsParenthesizedExpression().Expression
	}
	return receiver != nil && receiver.Kind == ast.KindIdentifier &&
		receiver.Text() == "Number" && resolvesToAGlobal(ctx, receiver)
}

// isTheLiteralText says whether a computed key is a string literal spelling the given name.
//
// A template with no substitutions counts, because Number[`parseInt`] is the same call and upstream
// treats it as one. A template WITH substitutions does not, since its value is not known here.
func isTheLiteralText(key *ast.Node, name string) bool {
	if key == nil {
		return false
	}
	switch key.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return key.Text() == name
	}
	return false
}

// checkRadixArguments applies the three judgments in upstream's own order.
func checkRadixArguments(ctx rule.Context, node *ast.Node, call *ast.CallExpression) {
	arguments := []*ast.Node{}
	if call.Arguments != nil {
		arguments = call.Arguments.Nodes
	}

	// A spread in either of the first two positions hides where the radix is, so nothing can be
	// said. A spread after them hides nothing, which is why `parseInt("10", 1, ...args)` still
	// reports.
	for index, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement && index < 2 {
			return
		}
	}

	switch {
	case len(arguments) == 0:
		ctx.ReportNode(node, messageRadixMissingParameters)

	case len(arguments) == 1:
		ctx.ReportNodeWithSuggestions(node, messageRadixMissingRadix,
			addRadixSuggestion(ctx, node))

	case !isAValidRadix(ctx, arguments[1]):
		ctx.ReportNode(node, messageRadixInvalidRadix)
	}
}

// isAValidRadix says whether an argument is a base `parseInt` would actually use.
//
// Deliberately narrow, matching upstream. Only a literal is judged, plus a numeric literal carrying
// a leading sign, and a bare `undefined` that resolves to the global. Everything else is accepted,
// because a rule guessing at a computed value would report correct code.
func isAValidRadix(ctx rule.Context, radix *ast.Node) bool {
	if radix == nil {
		return true
	}

	// A signed numeric literal is the one arithmetic form upstream evaluates, which is why `+37`
	// reports while `~1` does not.
	if radix.Kind == ast.KindPrefixUnaryExpression {
		unary := radix.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindPlusToken && unary.Operator != ast.KindMinusToken {
			return true
		}
		if unary.Operand == nil || unary.Operand.Kind != ast.KindNumericLiteral {
			return true
		}
		value, ok := integerRadixValue(unary.Operand.Text())
		if !ok {
			return false
		}
		if unary.Operator == ast.KindMinusToken {
			value = -value
		}
		return value >= 2 && value <= 36
	}

	switch radix.Kind {
	case ast.KindNumericLiteral:
		value, ok := integerRadixValue(radix.Text())
		return ok && value >= 2 && value <= 36

	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral:
		// A literal that is not a number is never a base in range.
		return false

	case ast.KindIdentifier:
		// `undefined` is a literal in every way that matters here, and upstream checks that it
		// really is the global one rather than a shadowing parameter. Its own passing case
		// `function foo(undefined) { parseInt("10", undefined); }` is what says so.
		//
		// `resolvesToAGlobal` is deliberately NOT used, and this is the one place in this file
		// where it would be wrong. It answers false when a symbol carries zero declarations, and
		// the real global `undefined` has exactly that shape, so it answers false on every input
		// this arm must report. Probed: a bare `undefined` resolves to a symbol with 0
		// declarations, while the shadowing parameter resolves to a symbol with 1, declared in
		// source.
		//
		// So the predicate is the complement -- ask whether the name is declared in SOURCE, and
		// treat "not declared in source" as the global. Written this way round, a name with no
		// symbol at all also reads as the global, which is the correct answer for `undefined` in a
		// file the checker cannot fully resolve.
		return radix.Text() != "undefined" || isDeclaredInSource(ctx, radix)
	}

	// Anything else is a value this rule cannot evaluate, and upstream accepts every one of them.
	return true
}

// integerRadixValue reads a numeric literal's canonical text as an integer.
//
// The parser has already normalised the text, so `0x10` arrives as `16`, `1.6e1` as `16` and `10.0`
// as `10`, which is why no numeric parsing happens here beyond reading digits. A text carrying
// anything but digits is not an integer and cannot be a base: `10.5` arrives as `10.5` and answers
// false, which is upstream's `10.5` failing case.
func integerRadixValue(text string) (int, bool) {
	if text == "" {
		return 0, false
	}
	value := 0
	for index := 0; index < len(text); index++ {
		character := text[index]
		if character < '0' || character > '9' {
			return 0, false
		}
		value = value*10 + int(character-'0')
		// Nothing above 36 is a base, so a long literal can stop early rather than overflowing.
		if value > 36 {
			return value, true
		}
	}
	return value, true
}

// addRadixSuggestion offers to insert a decimal radix, which a human chooses.
//
// A suggestion rather than a fix, matching upstream: adding a radix changes what the call returns
// for any string whose base was being inferred, which is a change of meaning rather than of
// spelling.
//
// The text depends on whether a trailing comma is already there. `parseInt("10",)` becomes
// `parseInt("10", 10,)` rather than `parseInt("10",, 10)`, which is why the comma is read rather
// than assumed. Both spellings are in upstream's corpus.
func addRadixSuggestion(ctx rule.Context, node *ast.Node) rule.Suggestion {
	source := ctx.SourceFile.Text()

	// The closing parenthesis, found by walking back from the call's end over trivia. The node's
	// End() sits just past it.
	closingParenthesis := node.End() - 1
	for closingParenthesis > 0 && source[closingParenthesis] != ')' {
		closingParenthesis--
	}

	// A trailing comma before the closing parenthesis, ignoring whitespace between them.
	insertion := ", 10"
	for index := closingParenthesis - 1; index >= 0; index-- {
		character := source[index]
		if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		if character == ',' {
			insertion = " 10,"
		}
		break
	}

	return rule.Suggestion{
		Message: messageRadixAddRadixParameter10,
		Fixes: []rule.Fix{
			rule.ReplaceRange(
				core.NewTextRange(closingParenthesis, closingParenthesis), insertion),
		},
	}
}

// isDeclaredInSource says whether a name has a declaration written in this program's source.
//
// The complement of `resolvesToAGlobal`, and deliberately not expressed in terms of it. That helper
// answers false when a symbol carries no declarations, which is the shape of the real `undefined`
// global, so it answers false on exactly the inputs a rule about `undefined` must report. Asking the
// question the other way round makes a zero-declaration symbol read as the global, which is right.
func isDeclaredInSource(ctx rule.Context, identifier *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	// Every declaration is checked rather than the first, since a merged symbol can carry a library
	// declaration ahead of a source one: `declare global { interface Number {...} }` was measured
	// producing four library declarations followed by one local.
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file != nil && !file.IsDeclarationFile {
			return true
		}
	}
	return false
}
