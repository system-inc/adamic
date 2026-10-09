package typescript

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messagePreferIncludes is upstream's `preferIncludes`.
func messagePreferIncludes() rule.Message {
	return rule.Message{
		Id: "preferIncludes",
		Description: "This compares an `indexOf` result against a sentinel to ask a yes-or-no " +
			"question, so a reader has to remember which comparison means present and which means " +
			"absent, and an off-by-one in the sentinel reverses the answer silently. `includes` " +
			"asks the question directly and returns a boolean.",
	}
}

// messagePreferStringIncludes is upstream's `preferStringIncludes`.
func messagePreferStringIncludes() rule.Message {
	return rule.Message{
		Id: "preferStringIncludes",
		Description: "This pattern matches one fixed string, so the regular expression buys " +
			"nothing over a substring test and costs a reader the work of confirming it holds no " +
			"metacharacters. `String#includes` says the same thing in a form that cannot hide one.",
	}
}

// PreferIncludes requires `includes` over an `indexOf` comparison, and over a fixed-string regex.
//
//	valid:   a.indexOf(b) + 0;                  not a presence comparison
//	valid:   /ba[rz]/.test(a);                  the pattern is not one fixed string
//	valid:   /foo|bar/.test(a);                 two alternatives
//	valid:   type U = { indexOf(x: any): number }; declare const u: U; u.indexOf(b) !== -1;
//	invalid: a.indexOf(b) !== -1;               becomes a.includes(b)
//	invalid: a.indexOf(b) === -1;               becomes !a.includes(b)
//	invalid: /bar/.test(a);                     becomes a.includes('bar')
//
// Ported from `@typescript-eslint/prefer-includes`, reading the clone at
// `packages/eslint-plugin/src/rules/prefer-includes.ts` and measuring every verdict and every
// repair against the installed 8.67.0 build driven through the ESLint 10.8.1 Linter API over a real
// TypeScript program, since the rule is type-aware and the plain Linter path cannot reach it.
//
// # Two rules in one file, and only the second has tree exposure here
//
// The `indexOf` half judges a comparison; the regex half judges `/lit/.test(x)`. Measured on this
// tree with a control: five `indexOf` presence comparisons against 260 `.test(` sites, of which
// about eighteen are patterns upstream would rewrite. So the regex half is where the value is, and
// it is also the half with no substrate on the shelf. See `preferIncludesSimplePatternText`.
//
// # `includes` has to exist AND take the same parameters
//
// Upstream walks every declaration of the `indexOf` symbol, takes the type that declares it, and
// requires an `includes` on that type whose parameter list is textually identical to some
// `indexOf` overload. Three of upstream's clean cases turn on the parameter comparison alone: a
// user type declaring `indexOf(x, fromIndex?)` beside `includes(x)` is clean, and so is one
// declaring `includes: boolean`.
//
// The comparison is on the parameter's SOURCE TEXT, not on its type. That is upstream's own
// shortcut and it is reproduced rather than improved: comparing types would make `includes(x: any)`
// pair with `indexOf(x: unknown)`, which upstream leaves alone.
//
// # The sentinel is evaluated, not matched
//
// `isNumber(node.right, -1)` runs `getStaticValue`, so the right-hand side has to EVALUATE to the
// sentinel. A negative literal is a unary expression rather than a numeric literal in every parser
// involved, which is why `-1` needs its own arm rather than reading a literal's text.
//
// # An optional chain reports without a repair
//
// `a?.indexOf(b) !== -1` reports and offers nothing, because the rewrite would have to decide what
// `a?.includes(b)` means when `a` is nullish, and `undefined !== -1` is true while
// `undefined.includes` throws. Upstream passes `allowFixing: false` on that listener and both of
// its optional-chain cases carry `output: null`. Reproduced as a report with no fix.
var PreferIncludes = rule.Rule{
	Name:             "@typescript-eslint/prefer-includes",
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				reportPreferIncludesComparison(ctx, node)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				reportPreferIncludesRegexTest(ctx, node)
			},
		}
	},
}

// reportPreferIncludesComparison is upstream's `checkArrayIndexOf`, reached from the comparison.
//
// Upstream anchors on the member access through two selectors that name the comparison as an
// ancestor; anchoring on the comparison instead reaches the same set and makes the optional-chain
// distinction a property of the node in hand rather than of which selector fired.
func reportPreferIncludesComparison(ctx rule.Context, node *ast.Node) {
	comparison := node.AsBinaryExpression()
	if comparison.OperatorToken == nil {
		return
	}

	negative, positive := preferIncludesCheckDirection(ctx, comparison)
	if !negative && !positive {
		return
	}

	// The left side is the call, possibly wrapped in an optional chain. Our parser marks the
	// optionality on the access itself rather than wrapping it, so the call is reached directly and
	// the chain is read off the access below.
	callNode := comparison.Left
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return
	}
	callee := callNode.AsCallExpression().Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	propertyName := access.Name()
	if propertyName == nil || propertyName.Text() != "indexOf" {
		return
	}

	if !preferIncludesTypeHasMatchingIncludes(ctx, propertyName) {
		return
	}

	// An optional chain anywhere between the receiver and the call means the repair would have to
	// decide what a nullish receiver means, so upstream reports without one.
	fixable := access.QuestionDotToken == nil && callNode.AsCallExpression().QuestionDotToken == nil

	if !fixable {
		ctx.ReportNode(node, messagePreferIncludes())
		return
	}

	// Three edits, upstream's, in upstream's order: a leading `!` for the negative direction, the
	// method name, and the removal of everything from the call's end to the comparison's end.
	//
	// The removal is by RANGE rather than by node, and it deliberately spans the operator and the
	// sentinel together. Nothing type-bearing lives inside it: the left operand ends at the call and
	// the right operand is the numeric sentinel this rule already evaluated, so the span holds only
	// the comparison itself.
	callRange := rule.TokenRange(ctx.SourceFile, callNode)
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	propertyRange := rule.TokenRange(ctx.SourceFile, propertyName)

	fixes := make([]rule.Fix, 0, 3)
	if negative {
		fixes = append(fixes, rule.ReplaceRange(
			shimcore.NewTextRange(callRange.Pos(), callRange.Pos()), "!"))
	}
	fixes = append(fixes,
		rule.ReplaceRange(propertyRange, "includes"),
		rule.RemoveRange(shimcore.NewTextRange(callRange.End(), nodeRange.End())))

	ctx.ReportNodeWithFixes(node, messagePreferIncludes(), fixes...)
}

// preferIncludesCheckDirection is upstream's `isNegativeCheck` and `isPositiveCheck` together.
//
// Returning both answers from one function because the caller needs the direction and upstream's
// two predicates partition the same switch.
func preferIncludesCheckDirection(ctx rule.Context, comparison *ast.BinaryExpression) (negative bool, positive bool) {
	right := comparison.Right
	if right == nil {
		return false, false
	}

	switch comparison.OperatorToken.Kind {
	case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken,
		ast.KindGreaterThanToken:
		return false, preferIncludesEvaluatesTo(ctx, right, -1)
	case ast.KindGreaterThanEqualsToken:
		return false, preferIncludesEvaluatesTo(ctx, right, 0)
	case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken,
		ast.KindLessThanEqualsToken:
		return preferIncludesEvaluatesTo(ctx, right, -1), false
	case ast.KindLessThanToken:
		return preferIncludesEvaluatesTo(ctx, right, 0), false
	}
	return false, false
}

// preferIncludesEvaluatesTo is upstream's `isNumber`, narrowed to the sentinels this rule compares.
//
// Upstream runs the general `getStaticValue`, which evaluates arbitrary constant expressions against
// the global scope. Only two values can ever be asked for here, so the narrowing is to the shapes
// that can produce them: a numeric literal, and a unary minus over one.
//
// # What this declines, stated rather than left silent
//
// A sentinel written as `const NOT_FOUND = -1` and compared as `a.indexOf(b) !== NOT_FOUND` is
// evaluated by upstream and declined here. That costs a FINDING rather than adding one, so the
// failure direction is silence. Measured on this tree with a control: five `indexOf` presence
// comparisons in total and none of them names a constant, so the narrowing has no exposure here.
// Closing it means a constant evaluator with scope, which is substrate rather than a line in this
// rule, and the same gap is recorded on `class-literal-property-style`.
func preferIncludesEvaluatesTo(ctx rule.Context, node *ast.Node, want int) bool {
	if node == nil {
		return false
	}

	if want < 0 {
		if node.Kind != ast.KindPrefixUnaryExpression {
			return false
		}
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindMinusToken || unary.Operand == nil {
			return false
		}
		return preferIncludesNumericLiteralIs(unary.Operand, -want)
	}
	return preferIncludesNumericLiteralIs(node, want)
}

// preferIncludesNumericLiteralIs answers whether a node is a numeric literal of this magnitude.
//
// A `NumericLiteral`'s `Text()` already holds the CANONICAL rendering of the double rather than the
// source spelling, which is what lets a text comparison stand in for evaluating the literal. Probed
// against the installed build on the three spellings upstream accepts and this reads identically:
//
//	a.indexOf(b) !== -1      Text() is "1"
//	a.indexOf(b) !== -1.0    Text() is "1", so the fraction costs nothing
//	a.indexOf(b) !== -0x1    Text() is "1", so the radix costs nothing
//
// All three report upstream and all three report here. Reading the source spelling instead would
// have declined the last two, and neither appears in upstream's corpus.
func preferIncludesNumericLiteralIs(node *ast.Node, want int) bool {
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return false
	}
	switch want {
	case 0:
		return node.Text() == "0"
	case 1:
		return node.Text() == "1"
	}
	return false
}

// preferIncludesTypeHasMatchingIncludes is upstream's declaration walk.
//
// Every declaration of the `indexOf` symbol names a type; that type must carry an `includes` whose
// parameter text matches some `indexOf` overload. A single declaration failing the test declines the
// whole finding, which is why this returns on the first failure rather than accepting on the first
// success.
func preferIncludesTypeHasMatchingIncludes(ctx rule.Context, propertyName *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(propertyName)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}

	for _, indexOfDeclaration := range symbol.Declarations {
		parent := indexOfDeclaration.Parent
		if parent == nil {
			return false
		}
		declaringType := ctx.TypeChecker.GetTypeAtLocation(parent)
		if declaringType == nil {
			return false
		}
		includes := shimchecker.Checker_getPropertyOfType(ctx.TypeChecker, declaringType, "includes")
		if includes == nil || len(includes.Declarations) == 0 {
			return false
		}

		matched := false
		for _, includesDeclaration := range includes.Declarations {
			if preferIncludesHasSameParameters(ctx, includesDeclaration, indexOfDeclaration) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// preferIncludesHasSameParameters is upstream's `hasSameParameters`.
//
// Both declarations must be function-like with the same parameter count, and each parameter's
// SOURCE TEXT must match. Comparing text rather than type is upstream's own shortcut and is
// reproduced: it is what makes a `includes: boolean` property decline, since a property declaration
// is not function-like and has no parameters to compare.
func preferIncludesHasSameParameters(ctx rule.Context, first *ast.Node, second *ast.Node) bool {
	firstParameters, firstIsFunction := preferIncludesParametersOf(first)
	secondParameters, secondIsFunction := preferIncludesParametersOf(second)
	if !firstIsFunction || !secondIsFunction {
		return false
	}
	if len(firstParameters) != len(secondParameters) {
		return false
	}

	for index := range firstParameters {
		firstText, firstReadable := preferIncludesNodeText(ctx, firstParameters[index])
		secondText, secondReadable := preferIncludesNodeText(ctx, secondParameters[index])
		if !firstReadable || !secondReadable || firstText != secondText {
			return false
		}
	}
	return true
}

// preferIncludesParametersOf answers a declaration's parameter list, and whether it has one.
//
// `Parameters()` reaches `FunctionLikeData()` and dereferences it, so it panics on a node that is
// not function-like, and a property declaration reaching here is exactly that shape. The kinds are
// named rather than tried, matching upstream's `ts.isFunctionLike` guard. The walk recovers per
// FILE rather than per rule, so one panic would cost every rule that file.
func preferIncludesParametersOf(declaration *ast.Node) ([]*ast.Node, bool) {
	if declaration == nil {
		return nil, false
	}
	switch declaration.Kind {
	case ast.KindMethodSignature, ast.KindMethodDeclaration,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindConstructor, ast.KindConstructSignature, ast.KindCallSignature,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindFunctionType, ast.KindConstructorType:
		return declaration.Parameters(), true
	}
	return nil, false
}

// reportPreferIncludesRegexTest is upstream's `/lit/.test(x)` listener.
func reportPreferIncludesRegexTest(ctx rule.Context, node *ast.Node) {
	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return
	}
	callee := call.Expression
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	propertyName := access.Name()
	if propertyName == nil || propertyName.Text() != "test" {
		return
	}
	// Upstream's selector carries `[computed=false]`, so `re['test'](a)` is not this rule's
	// business. A property access is the non-computed form in this parser, so the kind test above
	// already carries it, and an element access reaches a different node kind entirely.

	text, simple := preferIncludesPatternTextOf(ctx, access.Expression)
	if !simple {
		return
	}

	argument := call.Arguments.Nodes[0]
	if !preferIncludesArgumentHasIncludes(ctx, argument) {
		return
	}

	replacement, buildable := preferIncludesRegexReplacement(ctx, node, access, argument, text)
	if !buildable {
		ctx.ReportNode(node, messagePreferStringIncludes())
		return
	}

	ctx.ReportNodeWithFixes(node, messagePreferStringIncludes(),
		rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), replacement))
}

// preferIncludesRegexReplacement builds the whole rewritten call.
//
// Upstream emits four edits that together delete everything outside the argument, optionally wrap
// the argument in parentheses, and append the `includes` call. One replacement over the same span
// writes the identical text and cannot leave a half-applied pair behind, which matters because the
// fix engine flattens a diagnostic's fixes into independent proposals and may admit one and refuse
// another.
//
// Nothing type-bearing lives in the replaced span. The span is the whole call expression, and every
// byte of it is either the regex literal, the `.test(` punctuation, the argument, or the closing
// parenthesis. The argument is carried through as its own source text rather than re-rendered, so a
// type assertion or a generic call inside it survives.
func preferIncludesRegexReplacement(
	ctx rule.Context,
	node *ast.Node,
	access *ast.PropertyAccessExpression,
	argument *ast.Node,
	text string,
) (string, bool) {
	// Our parser keeps `KindParenthesizedExpression` and TSESTree folds it away, so upstream is
	// handed the INNER expression with a range that excludes the parentheses while we are handed the
	// wrapper. Both the shape test below and the text slice have to see what upstream sees.
	//
	// This is not cosmetic. Upstream's own `/bar/.test((1 + 1, a))` case is a sequence expression,
	// which needs parentheses in receiver position, and upstream adds them because its range never
	// carried any. Without the unwrap the wrapper reads as a shape needing parentheses AND already
	// carries them, so the repair writes `((1 + 1, a)).includes(...)`. That is what this fixture
	// caught, and it is the only upstream case that can see it.
	//
	// A loop rather than a single step, because `((x))` nests, and written out rather than reaching
	// for `ast.SkipParentheses`, which dereferences its argument and is how this project lost 167
	// files to a nil panic.
	for argument.Kind == ast.KindParenthesizedExpression {
		inner := argument.AsParenthesizedExpression().Expression
		if inner == nil {
			return "", false
		}
		argument = inner
	}

	argumentText, readable := preferIncludesNodeText(ctx, argument)
	if !readable {
		return "", false
	}

	// Upstream parenthesizes an argument that is not one of five shapes, because the rewrite puts
	// the argument in receiver position where a comma or a conditional would bind wrongly.
	needsParentheses := true
	switch argument.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression, ast.KindIdentifier,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindCallExpression:
		needsParentheses = false
	}

	var builder strings.Builder
	if needsParentheses {
		builder.WriteString("(")
	}
	builder.WriteString(argumentText)
	if needsParentheses {
		builder.WriteString(")")
	}
	if access.QuestionDotToken != nil {
		builder.WriteString("?.")
	} else {
		builder.WriteString(".")
	}
	builder.WriteString("includes('")
	builder.WriteString(preferIncludesEscapeString(text))
	builder.WriteString("')")
	return builder.String(), true
}

// preferIncludesArgumentHasIncludes answers whether the argument's type carries an `includes`.
//
// This is what makes `/bar/.test(a)` clean when `a` is untyped and reporting when it is a string.
// Upstream reads the CONSTRAINED type, so a type parameter is judged by its constraint.
func preferIncludesArgumentHasIncludes(ctx rule.Context, argument *ast.Node) bool {
	argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
	if argumentType == nil {
		return false
	}
	constrained := shimchecker.Checker_getBaseConstraintOfType(ctx.TypeChecker, argumentType)
	if constrained != nil {
		argumentType = constrained
	}
	includes := shimchecker.Checker_getPropertyOfType(ctx.TypeChecker, argumentType, "includes")
	return includes != nil && len(includes.Declarations) > 0
}

// preferIncludesPatternTextOf answers the fixed string a regex receiver matches, if it matches one.
//
// Two receivers reach here, both of which upstream resolves through `getStaticValue`: a regex
// literal written in place, and an identifier bound to one. The identifier arm covers upstream's
// `const pattern = /bar/` and `const pattern = new RegExp('bar')` cases.
func preferIncludesPatternTextOf(ctx rule.Context, receiver *ast.Node) (string, bool) {
	if receiver == nil {
		return "", false
	}

	switch receiver.Kind {
	case ast.KindRegularExpressionLiteral:
		return preferIncludesSimplePatternFromLiteral(receiver.Text())
	case ast.KindIdentifier:
		return preferIncludesResolveRegexVariable(ctx, receiver.Text())
	}
	return "", false
}

// preferIncludesResolveRegexVariable resolves an identifier to a regex declared in the same file.
//
// Upstream's `getStaticValue` walks scope and evaluates, which reaches a `new RegExp('bar')` as well
// as a literal. Both shapes are covered here because both are in upstream's corpus, and both are
// resolved only through a file-scope variable whose initializer is one of them.
//
// The LAST matching declaration wins, matching how a second declaration of the same name shadows the
// first at the point the call is written.
func preferIncludesResolveRegexVariable(ctx rule.Context, name string) (string, bool) {
	text := ""
	found := false
	for _, statement := range ctx.SourceFile.AsNode().Statements() {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		list := statement.AsVariableStatement().DeclarationList
		if list == nil {
			continue
		}
		for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
			declarationName := declaration.Name()
			if declarationName == nil || declarationName.Kind != ast.KindIdentifier {
				continue
			}
			if declarationName.Text() != name {
				continue
			}
			initializer := declaration.AsVariableDeclaration().Initializer
			if candidate, ok := preferIncludesRegexInitializerText(initializer); ok {
				text, found = candidate, true
			}
		}
	}
	return text, found
}

// preferIncludesRegexInitializerText reads a regex out of a variable's initializer.
func preferIncludesRegexInitializerText(initializer *ast.Node) (string, bool) {
	if initializer == nil {
		return "", false
	}

	switch initializer.Kind {
	case ast.KindRegularExpressionLiteral:
		return preferIncludesSimplePatternFromLiteral(initializer.Text())

	case ast.KindNewExpression:
		// `new RegExp('bar')`. A second argument is the flags, and upstream's flag test rejects
		// `i` and `g`, so a flags argument that is not a plain empty string declines here rather
		// than being parsed.
		newExpression := initializer.AsNewExpression()
		callee := newExpression.Expression
		if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "RegExp" {
			return "", false
		}
		if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) != 1 {
			return "", false
		}
		pattern := newExpression.Arguments.Nodes[0]
		if pattern.Kind != ast.KindStringLiteral && pattern.Kind != ast.KindNoSubstitutionTemplateLiteral {
			return "", false
		}
		return preferIncludesSimplePatternText(pattern.Text(), regexsyntax.ParseRegexFlags(""))
	}
	return "", false
}

// preferIncludesSimplePatternFromLiteral splits a regex literal and answers its fixed string.
func preferIncludesSimplePatternFromLiteral(literalText string) (string, bool) {
	pattern, flags := regexsyntax.PatternAndFlags(literalText)

	// Upstream rejects `i` and `g` outright: a case-insensitive pattern does not match one fixed
	// string, and a global one carries `lastIndex` state that `includes` does not have.
	parsed := regexsyntax.ParseRegexFlags(flags)
	if strings.ContainsAny(flags, "ig") {
		return "", false
	}
	return preferIncludesSimplePatternText(pattern, parsed)
}

// preferIncludesSimplePatternText answers upstream's question: is this pattern a single alternative
// made only of plain characters, and if so what string does it match?
//
// # Why this is written here rather than taken from the shelf
//
// `regexpattern.Walk` is the shelf's pattern walker and it CANNOT answer this, which was measured
// rather than assumed. Walk reports characters with a depth counter and erases exactly the
// distinctions the question turns on. Run over this rule's own patterns it renders `b.r` as `b.r`,
// `^bar` and `bar$` and `(bar)` and `b(?:a)r` all as `bar`, and `foo|bar` as `foobar`, so a port
// built on it would rewrite `/^bar$/.test(a)` to `a.includes('bar')`, which is wrong. Upstream
// declines every one of those, measured against the installed build.
//
// The shelf's own doc says as much: it is "deliberately not a regex AST", and a caller wanting
// alternation "should reach for the layer underneath rather than extend this". That is what this
// does. `regexsyntax.SkipPatternEscape` handles the escape grammar and everything else is a
// metacharacter test, which is the whole of upstream's question and none of a parse tree.
//
// Note `SkipPatternEscape` returns a LENGTH rather than an absolute index. A probe reading it as an
// index panicked on the third pattern it saw, which is recorded here because the name reads like an
// index and the two agree for the first escape in a pattern.
//
// # The verdict on every shape, measured against the installed build
//
//	/bar/       bar      /b\.r/      b.r      /\x41/     A       /a\-b/    a-b
//	/ba[rz]/    decline  /foo|bar/   decline  /b.r/      decline /ba+r/    decline
//	/^bar/      decline  /bar$/      decline  /(bar)/    decline /b(?:a)r/ decline
//	/\d/        decline  /a{2}/      decline  /bar/i     decline /bar/g    decline
func preferIncludesSimplePatternText(pattern string, flags regexsyntax.RegexFlags) (string, bool) {
	var builder strings.Builder

	for index := 0; index < len(pattern); {
		character := pattern[index]

		switch character {
		// Every metacharacter that makes the pattern something other than one literal alternative.
		// `|` is alternation, the brackets and braces open a class, a group or a quantifier, and
		// the rest are anchors, quantifiers, or the any-character dot.
		case '|', '(', ')', '[', ']', '{', '}', '^', '$', '*', '+', '?', '.':
			return "", false

		case '\\':
			length, ok := regexsyntax.SkipPatternEscape(pattern, index, flags)
			if !ok || length <= 0 || index+length > len(pattern) {
				return "", false
			}
			value, decoded := preferIncludesDecodeEscape(pattern[index : index+length])
			if !decoded {
				return "", false
			}
			builder.WriteRune(value)
			index += length

		default:
			value, size := utf8.DecodeRuneInString(pattern[index:])
			if size == 0 {
				return "", false
			}
			builder.WriteRune(value)
			index += size
		}
	}

	return builder.String(), true
}

// preferIncludesDecodeEscape answers the character an escape sequence matches, or declines.
//
// A SET escape has no single character to answer and declines, which is what makes `/\d/` clean.
// `\b` and `\B` are assertions outside a class and decline for the same reason.
func preferIncludesDecodeEscape(escaped string) (rune, bool) {
	if len(escaped) < 2 || escaped[0] != '\\' {
		return 0, false
	}

	switch escaped[1] {
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	case 'v':
		return '\v', true
	case 'f':
		return '\f', true
	case '0':
		// `\0` with nothing following is the null character. A digit following makes it a legacy
		// octal escape, which `SkipPatternEscape` does not consume, so the length is the tell.
		if len(escaped) == 2 {
			return 0, true
		}
		return 0, false
	case 'd', 'D', 's', 'S', 'w', 'W', 'b', 'B', 'p', 'P', 'k', 'q':
		return 0, false
	case 'x':
		if len(escaped) == 4 && regexsyntax.AllHexDigits(escaped[2:]) {
			return rune(regexsyntax.ParseHexUint(escaped[2:])), true
		}
		return 0, false
	case 'u':
		if len(escaped) == 6 && regexsyntax.AllHexDigits(escaped[2:]) {
			return rune(regexsyntax.ParseHexUint(escaped[2:])), true
		}
		if len(escaped) > 4 && escaped[2] == '{' && escaped[len(escaped)-1] == '}' &&
			regexsyntax.AllHexDigits(escaped[3:len(escaped)-1]) {
			return rune(regexsyntax.ParseHexUint(escaped[3 : len(escaped)-1])), true
		}
		return 0, false
	case 'c':
		if len(escaped) == 3 {
			letter := escaped[2]
			if (letter >= 'a' && letter <= 'z') || (letter >= 'A' && letter <= 'Z') {
				return rune(letter % 32), true
			}
		}
		return 0, false
	}

	// An identity escape: a backslash before a character that needed no escaping, like `\.` or
	// `\-`. The character itself is what the pattern matches.
	if len(escaped) == 2 {
		value, size := utf8.DecodeRuneInString(escaped[1:])
		if size == 0 {
			return 0, false
		}
		return value, true
	}
	return 0, false
}

// preferIncludesEscapeString is upstream's `escapeString`, rendering a string into single quotes.
//
// Upstream's map deliberately omits `\b`, with the comment that it "cause unexpected replacements",
// because the backspace character and the word-boundary escape share a spelling. That omission is
// reproduced: a backspace is written raw rather than escaped.
func preferIncludesEscapeString(text string) string {
	var builder strings.Builder
	for _, character := range text {
		switch character {
		case 0:
			builder.WriteString("\\0")
		case '\t':
			builder.WriteString("\\t")
		case '\n':
			builder.WriteString("\\n")
		case '\v':
			builder.WriteString("\\v")
		case '\f':
			builder.WriteString("\\f")
		case '\r':
			builder.WriteString("\\r")
		case '\'':
			builder.WriteString("\\'")
		case '\\':
			builder.WriteString("\\\\")
		default:
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

// preferIncludesNodeText slices a node's own source text, trivia excluded.
func preferIncludesNodeText(ctx rule.Context, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil {
		return "", false
	}
	trimmed := type_checking.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(text) || trimmed.Pos() > trimmed.End() {
		return "", false
	}
	return text[trimmed.Pos():trimmed.End()], true
}
