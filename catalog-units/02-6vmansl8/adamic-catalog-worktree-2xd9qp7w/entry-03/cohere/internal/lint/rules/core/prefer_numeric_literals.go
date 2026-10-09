package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messagePreferNumericLiteralsId is the id. The message is built per finding because it names both
// the number system and the exact text of the callee, and the second is read from the source.
const messagePreferNumericLiteralsId = "useLiteral"

// preferNumericLiteralsMessage renders the finding.
//
// Upstream's template is `Use {{system}} literals instead of {{functionName}}().` The function name
// is the callee's own source text rather than a fixed string, so `Number?.parseInt` and
// `(Number.parseInt)` each name themselves.
func preferNumericLiteralsMessage(system string, functionName string) rule.Message {
	return rule.Message{
		Id: messagePreferNumericLiteralsId,
		Description: fmt.Sprintf("Use %s literals instead of %s(). "+
			"A literal says the value in the base it was written in, with no call to read and no "+
			"radix argument to get wrong, and the engine reads it at parse time rather than at "+
			"every evaluation.", system, functionName),
	}
}

// preferNumericLiteralsRadix maps the three radixes that have a literal spelling.
//
// Only 2, 8 and 16 are here. Base 10 needs no rewrite and every other base has no literal form, so
// upstream's map has exactly these three and a radix outside it is a passing case.
var preferNumericLiteralsRadix = map[int]struct {
	system string
	prefix string
}{
	2:  {"binary", "0b"},
	8:  {"octal", "0o"},
	16: {"hexadecimal", "0x"},
}

// PreferNumericLiterals flags `parseInt` or `Number.parseInt` on a string literal in base 2, 8 or
// 16, and rewrites it as a numeric literal.
//
//	valid:   parseInt(1, 3);
//	valid:   parseInt('11', '2');
//	valid:   parseInt(foo, 2);
//	valid:   function foo(parseInt) { parseInt("111110111", 2); }
//	invalid: parseInt("111110111", 2) === 503;
//	invalid: Number.parseInt("1F7", 16) === 255;
//
// # What has to line up before this reports
//
// Two arguments exactly. The first a string literal or an untagged template with no substitutions,
// whose value the syntax settles. The second a plain numeric literal, not a string and not a
// BigInt, whose value is one of the three radixes with a literal spelling. And the callee `parseInt`
// or `Number.parseInt`, resolving to the global rather than to something shadowing it.
//
// Each of those is a passing case in upstream's corpus, and the shadow ones are the reason this
// rule reads the checker: `function foo(parseInt) { parseInt("111110111", 2); }` is clean, and
// nothing about the syntax says so.
//
// # The fixer, and the three families of decline
//
// Twenty seven of upstream's sixty two invalid cases carry `output: null`, so the declines are
// nearly half the port rather than an edge. They fall into three families and all three are
// reproduced:
//
//	the value would change      `parseInt('7999', 8)` is 7, while `0o7999` is not a number at
//	                            all, and `parseInt('1234.5', 8)` stops at the dot. Upstream tests
//	                            this by evaluating both sides and comparing, and so does this.
//	a comment anywhere inside   `Number/**/.parseInt('11', 2)` and eleven more. The replacement
//	                            spans the whole call, so any comment in it would be deleted.
//	a numeric separator         `parseInt('1_0', 2)` is 1, because parseInt stops at the
//	                            underscore, while `0b1_0` is 2. Upstream reaches this through the
//	                            same value comparison, since unary plus on a separated literal is
//	                            not a number.
//
// # Where the repair needs a space
//
// The replacement can end up beside a token it would merge with, in either direction. `yield` and
// `0b11` written adjacently become one identifier, and `0b11` beside `in` becomes `0b11in`. Both
// are in the corpus and both get a space here.
var PreferNumericLiterals = rule.Rule{
	Name:             "prefer-numeric-literals",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) != 2 {
					return
				}

				stringValue, hasStringValue := preferNumericLiteralsStringValue(call.Arguments.Nodes[0])
				if !hasStringValue {
					return
				}
				radix, hasRadix := preferNumericLiteralsRadixValue(call.Arguments.Nodes[1])
				if !hasRadix {
					return
				}
				spelling, isSpellable := preferNumericLiteralsRadix[radix]
				if !isSpellable {
					return
				}
				if !preferNumericLiteralsIsParseInt(ctx, call.Expression) {
					return
				}

				calleeRange := rule.TokenRange(ctx.SourceFile, call.Expression)
				functionName := ctx.SourceFile.Text()[calleeRange.Pos():calleeRange.End()]
				message := preferNumericLiteralsMessage(spelling.system, functionName)

				if fix, canFix := preferNumericLiteralsFix(ctx, node, spelling.prefix, stringValue,
					radix); canFix {
					ctx.ReportNodeWithFixes(node, message, fix)
					return
				}
				ctx.ReportNode(node, message)
			},
		}
	},
}

// preferNumericLiteralsStringValue reads the cooked value of a string literal or of an untagged
// template with no substitutions, which are the two shapes upstream accepts.
//
// A template WITH a substitution is declined, and so is a tagged one: `parseInt(`+"`11${foo}`"+`, 2)`
// is a passing case because its value is not settled by the syntax.
func preferNumericLiteralsStringValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.Text(), true
	case ast.KindNoSubstitutionTemplateLiteral:
		// A tagged template is a TaggedTemplateExpression rather than an argument, so reaching this
		// kind already means the literal stands alone.
		return node.Text(), true
	}
	return "", false
}

// preferNumericLiteralsRadixValue reads the radix, which upstream requires to be a plain numeric
// literal.
//
// A string radix and a BigInt radix are both passing cases upstream, with six between them, so the
// kind test is the whole discrimination and a value-shaped read would report all six.
func preferNumericLiteralsRadixValue(node *ast.Node) (int, bool) {
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return 0, false
	}
	// The node's own Text is the canonical rendering of the double, so `0x10` reads as 16 here and
	// the radix written in another base still resolves.
	value, err := strconv.ParseFloat(node.Text(), 64)
	if err != nil || value != math.Trunc(value) {
		return 0, false
	}
	return int(value), true
}

// preferNumericLiteralsIsParseInt answers upstream's `isParseInt`, which accepts two callee shapes
// and requires the name at the root of each to be the global.
func preferNumericLiteralsIsParseInt(ctx rule.Context, callee *ast.Node) bool {
	unwrapped := preferNumericLiteralsUnwrap(callee)
	if unwrapped == nil {
		return false
	}

	if unwrapped.Kind == ast.KindIdentifier && unwrapped.Text() == "parseInt" {
		return preferNumericLiteralsResolvesToAGlobal(ctx, unwrapped)
	}

	// `Number.parseInt`, in either spelling and with optional chaining anywhere in it.
	if unwrapped.Kind != ast.KindPropertyAccessExpression &&
		unwrapped.Kind != ast.KindElementAccessExpression {
		return false
	}
	if name, isStatic := property.AccessedName(unwrapped, property.Static); !isStatic ||
		name != "parseInt" {
		return false
	}
	var accessedObject *ast.Node
	switch unwrapped.Kind {
	case ast.KindPropertyAccessExpression:
		accessedObject = unwrapped.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		accessedObject = unwrapped.AsElementAccessExpression().Expression
	}
	accessedObject = preferNumericLiteralsUnwrap(accessedObject)
	if accessedObject == nil || accessedObject.Kind != ast.KindIdentifier ||
		accessedObject.Text() != "Number" {
		return false
	}
	return preferNumericLiteralsResolvesToAGlobal(ctx, accessedObject)
}

// preferNumericLiteralsUnwrap strips parentheses in a loop.
//
// Upstream's parser folds them away, so `(parseInt)('A', 16)` and `(Number?.parseInt)("1F7", 16)`
// reach its check already bare. Five of its invalid cases write that shape, so without this the
// port goes silent on inputs the corpus asserts.
func preferNumericLiteralsUnwrap(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// preferNumericLiteralsResolvesToAGlobal answers upstream's `sourceCode.isGlobalReference`.
//
// A declaration in a declaration file is the global. Five of upstream's passing cases are shadows,
// covering a parameter and a local for each of the two names, and nothing about the syntax
// separates them from the real thing.
func preferNumericLiteralsResolvesToAGlobal(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		declaringFile := ast.GetSourceFileOfNode(declaration)
		if declaringFile == nil || !declaringFile.IsDeclarationFile {
			return false
		}
	}
	return true
}

// preferNumericLiteralsFix builds the repair, or declines.
func preferNumericLiteralsFix(ctx rule.Context, node *ast.Node, prefix string, stringValue string,
	radix int) (rule.Fix, bool) {
	callRange := rule.TokenRange(ctx.SourceFile, node)

	// Any comment inside the call would be deleted by a whole-call replacement. Twelve of
	// upstream's declines are this and nothing else.
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= callRange.Pos() && comment.Range.End() <= callRange.End() {
			return rule.Fix{}, false
		}
	}

	// The literal has to mean what the call meant. Upstream writes this as `+replacement !==
	// parseInt(str, radix)`, evaluating both sides in JavaScript; here both sides are computed
	// explicitly, and the comparison is the same one.
	replacement := prefix + stringValue
	literalValue, literalIsValid := preferNumericLiteralsLiteralValue(replacement)
	if !literalIsValid || literalValue != preferNumericLiteralsParseInt(stringValue, radix) {
		return rule.Fix{}, false
	}

	text := ctx.SourceFile.Text()
	if callRange.Pos() > 0 && scanner.IsIdentifierPart(rune(text[callRange.Pos()-1])) {
		// `yield` beside `0b11` would read as one identifier.
		replacement = " " + replacement
	}
	if callRange.End() < len(text) && scanner.IsIdentifierPart(rune(text[callRange.End()])) {
		// `0b11` beside `in` would read as `0b11in`.
		replacement = replacement + " "
	}
	return rule.ReplaceRange(core.NewTextRange(callRange.Pos(), callRange.End()), replacement), true
}

// preferNumericLiteralsLiteralValue evaluates the literal the fix would write, the way unary plus
// evaluates it.
//
// Returns false when the text is not a valid literal at all, which is `0b1234` and its siblings, and
// also when it carries a numeric separator, because unary plus does not accept one. Both are
// declines upstream reaches through the same comparison.
func preferNumericLiteralsLiteralValue(literal string) (float64, bool) {
	// Subsumed today by the ParseUint below and kept deliberately, with the measurement recorded
	// rather than the guard deleted.
	//
	// A mutant removing this survived the whole corpus, and the reason is that ParseUint with an
	// EXPLICIT base rejects a separator: measured, `ParseUint("1_0", 2)` is a syntax error while
	// `ParseUint("0b1_0", 0)` is 2. So the four separator cases decline either way, through the
	// invalid-literal path instead of through this one.
	//
	// The subsumption is a property of how the call below is spelled rather than of the language,
	// and base 0 is the spelling someone would reach for if they wanted the prefix parsed here too.
	// That change would be correct-looking, would keep every fixture green, and would silently
	// start fixing `parseInt('1_0', 2)` to `0b1_0`, which is 1 becoming 2. One line is cheap
	// insurance against a change nothing else would catch.
	if strings.ContainsRune(literal, '_') {
		return 0, false
	}
	if len(literal) < 3 {
		// The prefix alone, which `parseInt('', 8)` would produce and which is not a number.
		return 0, false
	}
	base := 0
	switch literal[1] {
	case 'b':
		base = 2
	case 'o':
		base = 8
	case 'x':
		base = 16
	default:
		return 0, false
	}
	value, err := strconv.ParseUint(literal[2:], base, 64)
	if err != nil {
		return 0, false
	}
	return float64(value), true
}

// preferNumericLiteralsParseInt reproduces what `parseInt(string, radix)` returns, for the three
// radixes this rule can reach.
//
// Not a general implementation and it does not need to be. The string reaching here came from a
// literal and the radix is 2, 8 or 16, so the parts of parseInt that matter are: leading whitespace
// is skipped, an optional sign is read, digits are consumed while they are valid in the radix, and
// everything from the first invalid character onward is discarded. An empty digit run is NaN, which
// is returned as a value no literal can equal so the caller declines.
//
// The discard rule is the whole reason the declines exist. `parseInt('7999', 8)` reads `7`, stops at
// the `9`, and returns 7 while `0o7999` is not a literal; `parseInt('1_0', 2)` reads `1`, stops at
// the underscore, and returns 1 while `0b1_0` is 2.
func preferNumericLiteralsParseInt(text string, radix int) float64 {
	// The whitespace set is written as escapes rather than as literal characters. A non-breaking
	// space and a byte order mark typed into Go source are invisible, and the second is not
	// even legal there: the first draft of this line carried both and the compiler rejected it.
	trimmed := strings.TrimLeft(text, " \t\n\r\v\f\u00a0\ufeff\u2028\u2029")
	negative := false
	if strings.HasPrefix(trimmed, "+") || strings.HasPrefix(trimmed, "-") {
		negative = trimmed[0] == '-'
		trimmed = trimmed[1:]
	}

	// The scan stops at the first character that is not a digit in this radix, which is what makes
	// `parseInt('7999', 8)` equal 7 rather than an error.
	//
	// A mutant reading the whole string instead survived the corpus, and an exhaustive check says
	// why: over 75,777 strings of up to three characters drawn from digits, letters, a dot, an
	// underscore, spaces and signs, crossed with all three radixes, the two versions never once
	// disagree about whether a FIX IS OFFERED. They disagree about the value often, on six of the
	// twelve shapes probed by hand, but only where the discarded suffix is also what makes the
	// literal invalid, so both routes reach the same decline.
	//
	// So it is equivalent at the level the rule decides, and the stopping scan is kept because it
	// is what the sentence above claims the function computes. A version that agreed by accident
	// would be a worse thing to leave behind than one that agrees by construction.
	digits := 0
	for digits < len(trimmed) && preferNumericLiteralsDigitValue(trimmed[digits]) < radix &&
		preferNumericLiteralsDigitValue(trimmed[digits]) >= 0 {
		digits++
	}
	if digits == 0 {
		// NaN. Returned as a sentinel the caller cannot match, which makes the case a decline.
		return math.NaN()
	}

	value := 0.0
	for index := 0; index < digits; index++ {
		value = value*float64(radix) + float64(preferNumericLiteralsDigitValue(trimmed[index]))
	}
	if negative {
		return -value
	}
	return value
}

// preferNumericLiteralsDigitValue is the value of one character as a digit, or -1.
func preferNumericLiteralsDigitValue(character byte) int {
	switch {
	case character >= '0' && character <= '9':
		return int(character - '0')
	case character >= 'a' && character <= 'z':
		return int(character-'a') + 10
	case character >= 'A' && character <= 'Z':
		return int(character-'A') + 10
	}
	return -1
}
