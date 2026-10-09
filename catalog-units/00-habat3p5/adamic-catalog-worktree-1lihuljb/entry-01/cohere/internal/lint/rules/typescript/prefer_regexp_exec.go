package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messagePreferRegexpExecOverStringMatch is the rule's single message.
var messagePreferRegexpExecOverStringMatch = rule.Message{
	Id: "regExpExecOverStringMatch",
	Description: "Use `RegExp#exec()` instead of `String#match()` here. Without a global flag the " +
		"two return the same match object, but `exec` says what it does at the call site and is " +
		"the faster of the two. `match` is the one to reach for when you want every match at once, " +
		"which is what the `g` flag turns it into, and this call has no `g`.",
}

// PreferRegexpExec reports `String#match` used where `RegExp#exec` would do.
//
//	valid:   'str'.match(/pattern/g);          a global flag makes match the right call
//	valid:   'str'.match(/a/, extra);          not a one-argument call
//	valid:   notAString.match(/a/);            the receiver has to be a string
//	invalid: 'str'.match(/pattern/);
//	invalid: 'str'.match('pattern');
//	invalid: str.match(regexpVariable);        when the variable's regex has no `g`
//
// # Why this rule needs the checker, and what it asks it
//
// The receiver must be a string, and that is a type question rather than a syntactic one: `s.match`
// where `s` is `'a' | 'b'`, or `string & { __brand: void }`, or a type parameter constrained to a
// string union, all count. `type_checking.GetTypeName` is upstream's `getTypeName` line for line,
// including the union, intersection and type-parameter-constraint arms that decide exactly those
// three shapes, so this rule calls it rather than re-deriving the judgment.
//
// The argument's type is asked separately and drives which of two repairs applies: a RegExp argument
// becomes `argument.exec(receiver)`, a string argument becomes `RegExp(argument).exec(receiver)`, and
// anything else is not reported at all.
//
// # The `g` flag decides the verdict, and reading it needs a value rather than a type
//
// A global regex must stay a `match`, because `exec` returns one match where a global `match`
// returns all of them -- so reporting it would change what the code does. Upstream answers "does
// this definitely lack a `g`" with `getStaticValue`, a general constant folder from eslint-utils
// with its own scope analyser: 476 lines of operations table plus identifier resolution.
//
// There is no such evaluator here, and porting one would be a substrate rather than a rule. What
// this rule actually needs from it is narrower and the checker answers it directly. Measured with a
// probe: `ctx.TypeChecker.GetSymbolAtLocation` on a reference to `const r = /x/g` resolves to the
// declaration, whose initializer is the regex literal with its flags intact. So the question
// "what regex is this identifier bound to" is answered by symbol resolution rather than by scope
// tracking, and the two agree on every shape the corpus and the tree contain.
//
// Where the answer is genuinely unknowable -- a regex built from a runtime value, a `let` reassigned
// elsewhere -- upstream's rule is to stay SILENT, and that direction matters: the guard is
// `definitelyDoesNotContainGlobalFlag`, so an unproven absence of `g` is treated as a possible `g`
// and nothing is reported. A port that inverted this would rewrite global matches and change
// behaviour, which is why the helper below is named for what it proves rather than for what it
// checks.
//
// # The fixer's parenthesisation is upstream's `getWrappingFixer`, and its predicate already ships
//
// `checking.IsStrongPrecedenceNode` is that helper's core test, already ported and already used by
// three rules here. An inner node that is not strongly precedent gets wrapped, which is why
// a template-literal receiver comes out parenthesised, matching upstream: a template literal is
// absent from upstream's list too, so upstream wraps it as well. Measured against the installed
// build rather than inferred, because the redundant-looking parentheses read like a defect.
//
// The outer half is the one that changes meaning if dropped: when the call sits under a weaker
// operator, the whole replacement is wrapped, so `a.match(/x/) || b` becomes `(/x/.exec(a)) || b`.
// Upstream's corpus contains no such case -- every one of its twelve reporting cases has a bare
// identifier or literal receiver in statement position -- so this is measured from the installed
// build directly and pinned in a fixture of its own.
var PreferRegexpExec = rule.Rule{
	Name:             "@typescript-eslint/prefer-regexp-exec",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return rule.Listeners{}
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				checkPreferRegexpExec(ctx, node)
			},
		}
	},
}

// checkPreferRegexpExec is upstream's `CallExpression[arguments.length=1] > MemberExpression`
// listener.
//
// Upstream keys on the member expression and reaches the call through its parent; here the listener
// is on the call and the member is read off it, which is the same pair reached from the other end
// and makes the one-argument test a direct read rather than a selector clause.
func checkPreferRegexpExec(ctx rule.Context, node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return
	}
	// An optional call (`s?.match(x)`) is a different node upstream, wrapped in a ChainExpression.
	// Here optionality is a token on this node, and upstream's selector reaches the member either
	// way, so no arm is needed.
	if call.Expression == nil || call.Expression.Kind != ast.KindPropertyAccessExpression {
		return
	}
	member := call.Expression.AsPropertyAccessExpression()
	if member == nil || member.Name() == nil {
		return
	}
	// Upstream's `isStaticMemberAccessOfValue(memberNode, context, 'match')`. A computed access
	// spelled `s['match'](x)` reaches this as an element access rather than a property access, so
	// it is not reported here -- and it is not in upstream's corpus either.
	if member.Name().Kind != ast.KindIdentifier || member.Name().Text() != "match" {
		return
	}

	receiver := member.Expression
	argument := call.Arguments.Nodes[0]
	if receiver == nil || argument == nil {
		return
	}

	// The receiver has to be a string, which is the type question this rule exists to ask.
	receiverType := ctx.TypeChecker.GetTypeAtLocation(receiver)
	if receiverType == nil || type_checking.GetTypeName(ctx.TypeChecker, receiverType) != "string" {
		return
	}

	// A regex that may carry `g` must stay a `match`. The direction is upstream's: only a PROVEN
	// absence of the flag allows a report.
	if !preferRegexpExecDefinitelyHasNoGlobalFlag(ctx, argument) {
		return
	}

	// A string literal argument is turned into a regex literal in the repair, which is why upstream
	// builds one and bails when it will not compile.
	if argument.Kind == ast.KindStringLiteral {
		pattern, ok := preferRegexpExecRegexLiteralFor(argument.Text())
		if !ok {
			return
		}
		ctx.ReportNodeWithFixes(member.Name(), messagePreferRegexpExecOverStringMatch,
			preferRegexpExecWrappingFix(ctx, node, pattern+".exec(",
				[]*ast.Node{receiver}, ")"))
		return
	}

	// Otherwise the argument's own type picks the repair.
	argumentType := ctx.TypeChecker.GetTypeAtLocation(argument)
	if argumentType == nil {
		return
	}
	switch preferRegexpExecArgumentKind(ctx, argumentType) {
	case preferRegexpExecArgumentRegExp:
		ctx.ReportNodeWithFixes(member.Name(), messagePreferRegexpExecOverStringMatch,
			preferRegexpExecWrappingFixTwo(ctx, node, "", argument, ".exec(", receiver, ")"))
	case preferRegexpExecArgumentString:
		ctx.ReportNodeWithFixes(member.Name(), messagePreferRegexpExecOverStringMatch,
			preferRegexpExecWrappingFixTwo(ctx, node, "RegExp(", argument, ").exec(", receiver, ")"))
	}
}

// preferRegexpExecArgumentClass is upstream's `ArgumentType` bitset, reduced to the three answers
// the switch below distinguishes.
type preferRegexpExecArgumentClass int

const (
	// preferRegexpExecArgumentOther is upstream's `Other` and its `Both`: neither is reported,
	// because a union of a string and a regex has no single correct repair.
	preferRegexpExecArgumentOther preferRegexpExecArgumentClass = iota
	preferRegexpExecArgumentString
	preferRegexpExecArgumentRegExp
)

// preferRegexpExecArgumentKind classifies the argument's type across its union constituents.
//
// Upstream accumulates a bitset over the constituents and reports only for exactly-String or
// exactly-RegExp; a union containing both, or containing anything else, falls through to `Both` or
// `Other` and is silent. That collapse is reproduced here rather than simplified, because a union
// of the two is exactly the case where either repair would be wrong.
func preferRegexpExecArgumentKind(
	ctx rule.Context,
	argumentType *checker.Type,
) preferRegexpExecArgumentClass {
	sawString, sawRegExp := false, false
	for constituent := range type_checking.UnionTypePartsSeq(argumentType) {
		switch type_checking.GetTypeName(ctx.TypeChecker, constituent) {
		case "string":
			sawString = true
		case "RegExp":
			sawRegExp = true
		}
	}
	switch {
	case sawString && sawRegExp:
		return preferRegexpExecArgumentOther
	case sawRegExp:
		return preferRegexpExecArgumentRegExp
	case sawString:
		return preferRegexpExecArgumentString
	}
	return preferRegexpExecArgumentOther
}

// preferRegexpExecDefinitelyHasNoGlobalFlag proves the argument cannot carry a `g` flag.
//
// The name states the direction, which is the load-bearing part: an argument this cannot prove
// anything about answers FALSE and the rule stays silent. Reporting an unproven case would rewrite
// a global match into an `exec`, which returns one match where the original returned all of them.
//
// Upstream reaches the same answer through `getStaticValue`, a general constant folder with its own
// scope analyser. The three shapes it actually has to settle are handled directly here:
//
//	/x/ and /x/g          a literal, flags read off the source
//	RegExp(p) etc.        a construction, flags read from the second argument
//	r, where const r=/x/g a reference, resolved through the checker to its declaration
//
// The third is the one that needed a substrate check before this rule could be ported at all.
// `GetSymbolAtLocation` resolves the reference to its declaration and the initializer is the literal
// with its flags intact, measured with a probe, so the scope analyser upstream needs has an
// equivalent here rather than an absence.
func preferRegexpExecDefinitelyHasNoGlobalFlag(ctx rule.Context, argument *ast.Node) bool {
	switch argument.Kind {
	case ast.KindRegularExpressionLiteral:
		return !preferRegexpExecLiteralHasGlobalFlag(argument.Text())

	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		// A string argument becomes `RegExp(...)`, which has no flags at all.
		return true

	case ast.KindCallExpression, ast.KindNewExpression:
		return preferRegexpExecConstructionHasNoGlobalFlag(ctx, argument)

	case ast.KindIdentifier:
		// A reference is settled by evaluating it to an actual value, NOT by recursing into
		// whatever it was initialised with. The distinction is upstream's and it is observable:
		//
		//	const r = /x/;                    a.match(r)   REPORTS, the value is a known RegExp
		//	const r = new RegExp(`^${t}$`);   a.match(r)   CLEAN,  the value cannot be computed
		//	a.match(new RegExp(`^${t}$`))                  REPORTS, an INLINE construction is
		//	                                               judged syntactically instead
		//
		// Upstream reaches that split through two separate guards. `getStaticValue` evaluates the
		// argument, and a construction from a runtime value evaluates to nothing; only when that
		// fails does it fall back to `definitelyDoesNotContainGlobalFlag`, which handles a call or
		// a `new` expression and NOTHING ELSE -- an identifier reaching it answers false.
		//
		// So a bound identifier is reportable only when it evaluates to a real regex, and a
		// construction is reportable only when written inline. A first draft here recursed from the
		// binding into its initializer, which is the intuitive reading and is wrong: it reported 8
		// sites in this repository that upstream leaves alone, every one of them
		// `const r = new RegExp(templateWithASubstitution)`. Caught by a differential over 87 real
		// files, where the corpus -- which has no such shape -- agreed all 37 rows either way.
		return preferRegexpExecEvaluatesToNonGlobalRegex(ctx, argument)
	}
	return false
}

// preferRegexpExecEvaluatesToNonGlobalRegex is the narrow slice of `getStaticValue` this rule needs
// for a reference: does this identifier evaluate to a value that is safe to rewrite.
//
// Two values qualify, and both were measured against the installed build rather than reasoned from
// the source:
//
//	const r = /x/;        REPORTS   a regex literal with no `g`
//	const r = /x/g;       clean     the `g` is what match is for
//	const s = 'thing';    REPORTS   a string, which becomes RegExp(s) and has no flags at all
//	const s = 'a' + 'b';  REPORTS   still a computable string
//	declare const s: string;   clean     no value to compute
//	const r = new RegExp(t);   clean     a construction is not a computable value
//
// The last two are the ones that separate this from the intuitive implementation. Upstream evaluates
// the reference with `getStaticValue` and requires an actual VALUE back; a declaration with no
// initializer and a construction from a runtime pattern both yield nothing, so the rule stays
// silent. Recursing into the initializer instead -- asking "could this construction carry a g" --
// reports both, and over-reported 8 real sites in this repository.
//
// Only `const` is followed, since a `let` can be reassigned anywhere in the program and its
// initializer proves nothing at this call site. Measured: upstream is likewise clean on
// `let r = /x/g; a.match(r)`.
func preferRegexpExecEvaluatesToNonGlobalRegex(ctx rule.Context, reference *ast.Node) bool {
	initializer := preferRegexpExecInitializerOf(ctx, reference)
	if initializer == nil {
		return false
	}
	if initializer.Kind == ast.KindRegularExpressionLiteral {
		return !preferRegexpExecLiteralHasGlobalFlag(initializer.Text())
	}
	// A construction is computable exactly when every argument is, which is the line that separates
	// `new RegExp('test', '')` -- a real value `getStaticValue` folds to `/test/` -- from
	// `new RegExp(`^${t}$`)`, which it cannot fold and therefore leaves alone. Both are in evidence:
	// the first is a corpus case that reports, the second is the shape that over-reported 8 sites
	// in this repository.
	if initializer.Kind == ast.KindNewExpression || initializer.Kind == ast.KindCallExpression {
		return preferRegexpExecIsComputableConstruction(ctx, initializer)
	}

	// A computable STRING is safe: the repair wraps it in `RegExp(...)`, which carries no flags.
	// Anything else -- a call to something else, a value from outside the file -- is not computable
	// and therefore not reportable.
	return preferRegexpExecIsComputableString(initializer)
}

// preferRegexpExecIsComputableConstruction answers whether a `RegExp(...)` construction folds to a
// value, and whether that value lacks a `g` flag.
//
// `getStaticValue` can evaluate `new RegExp('test', ”)` because both arguments are literals, and
// cannot evaluate `new RegExp(`^${t}$`)` because the pattern depends on a runtime value. So this
// requires EVERY argument to be a computable string, and then reads the flags from the second.
func preferRegexpExecIsComputableConstruction(ctx rule.Context, node *ast.Node) bool {
	var callee *ast.Node
	var arguments *ast.NodeList
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		if call == nil {
			return false
		}
		callee, arguments = call.Expression, call.Arguments
	} else {
		construction := node.AsNewExpression()
		if construction == nil {
			return false
		}
		callee, arguments = construction.Expression, construction.Arguments
	}
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "RegExp" {
		return false
	}
	if arguments == nil || len(arguments.Nodes) == 0 {
		return false
	}
	for _, argument := range arguments.Nodes {
		if !preferRegexpExecIsComputableString(argument) {
			return false
		}
	}
	if len(arguments.Nodes) < 2 {
		return true
	}
	return !strings.Contains(arguments.Nodes[1].Text(), "g")
}

// preferRegexpExecIsComputableString answers whether an expression is a string `getStaticValue`
// could fold to a literal.
//
// A literal is one, and so is a concatenation of them, which upstream's operations table folds and
// its corpus exercises through `const s = 'a' + 'b'`. A template with a substitution is NOT, unless
// every substitution is itself computable -- and this stops at the shapes measured rather than
// reimplementing a folder, because anything it declines simply goes unreported, which is the safe
// direction.
func preferRegexpExecIsComputableString(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken.Kind != ast.KindPlusToken {
			return false
		}
		return preferRegexpExecIsComputableString(binary.Left) &&
			preferRegexpExecIsComputableString(binary.Right)
	}
	return false
}

// preferRegexpExecInitializerOf resolves an identifier to the expression it was initialised with,
// when that initializer is the binding's only value.
//
// # The test is "written once", not "declared const", and that distinction was measured
//
// A first draft required a `const` declaration, which reads as the safe conservative choice and is
// simply a different rule: upstream reports `let r = /x/; a.match(r)` and this was silent on it.
// What `getStaticValue` actually does is fold a variable only when its scope shows exactly ONE
// write, so the keyword is irrelevant and reassignment is everything. Measured against the installed
// 8.67.0 build:
//
//	let r = /x/;              a.match(r)   REPORTS   one write, foldable
//	var r = /x/;              a.match(r)   REPORTS   likewise
//	let r = /x/; r = /y/g;    a.match(r)   clean     two writes, not foldable
//	let r = /x/; a.match(r); r = /y/g;     clean     the later write still counts
//
// The third and fourth are why a declaration-keyword test cannot substitute: `const` would report
// the first two correctly and would have to be wrong about something, and a `let` test would be
// wrong about all four. The last one also rules out any ordering shortcut -- a write AFTER the call
// site disqualifies the fold just as much as one before it.
//
// So the binding's other references are examined, and any one of them writing disqualifies the
// fold. `reference.WritesToBinding` is the shelf helper for exactly this question, lifted after four
// rules were each found hand-rolling it differently.
func preferRegexpExecInitializerOf(ctx rule.Context, reference *ast.Node) *ast.Node {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(reference)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return nil
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return nil
	}
	if preferRegexpExecBindingIsReassigned(ctx, declaration, symbol) {
		return nil
	}
	return initializer
}

// preferRegexpExecBindingIsReassigned answers whether anything in the file writes to this binding
// after its declaration.
//
// The whole source file is walked rather than a scope, because a write can sit anywhere -- inside a
// nested function, after the call site, in a branch never taken. That breadth is the point: the
// question is whether the initializer is the ONLY value the binding ever holds, and a narrower
// search answers a weaker question while looking like the same one.
func preferRegexpExecBindingIsReassigned(
	ctx rule.Context,
	declaration *ast.Node,
	symbol *ast.Symbol,
) bool {
	if ctx.SourceFile == nil {
		return false
	}
	declarationName := declaration.Name()
	if declarationName == nil || declarationName.Kind != ast.KindIdentifier {
		// A destructured binding has no single initializer to fold anyway.
		return true
	}

	reassigned := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || reassigned {
			return
		}
		if node.Kind == ast.KindIdentifier && node != declarationName &&
			node.Text() == declarationName.Text() &&
			reference.WritesToBinding(node) {
			// Same spelling and a write. Confirm it names the same binding before believing it, so
			// a shadowing declaration elsewhere cannot disqualify this one.
			if ctx.TypeChecker.GetSymbolAtLocation(node) == symbol {
				reassigned = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(ctx.SourceFile.AsNode())
	return reassigned
}

// preferRegexpExecConstructionHasNoGlobalFlag is upstream's
// `definitelyDoesNotContainGlobalFlag` for `RegExp(...)` and `new RegExp(...)`.
//
// A construction with no second argument has no flags. With one, the flags have to be a literal this
// can read: a runtime string is unproven, so the rule stays silent. `undefined` counts as absent,
// which upstream reaches through `getStaticValue` returning the value `undefined` rather than no
// value at all -- a corpus case turns on exactly that.
func preferRegexpExecConstructionHasNoGlobalFlag(ctx rule.Context, node *ast.Node) bool {
	var callee *ast.Node
	var arguments *ast.NodeList
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		if call == nil {
			return false
		}
		callee, arguments = call.Expression, call.Arguments
	} else {
		construction := node.AsNewExpression()
		if construction == nil {
			return false
		}
		callee, arguments = construction.Expression, construction.Arguments
	}
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "RegExp" {
		return false
	}
	if arguments == nil || len(arguments.Nodes) < 2 {
		// No flags argument at all, so there is no `g`.
		return true
	}

	flags := arguments.Nodes[1]
	switch flags.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return !strings.Contains(flags.Text(), "g")
	case ast.KindIdentifier:
		// `undefined` means no flags. Any other identifier is a runtime value and unproven.
		return flags.Text() == "undefined"
	}
	return false
}

// preferRegexpExecLiteralHasGlobalFlag answers whether a regex literal's flags include `g`.
//
// The flags are the run after the final unescaped `/`, which is found by scanning from the end
// rather than the start: a pattern may contain escaped slashes and a character class may contain
// bare ones, so the first `/` after the opening one is not reliably the closing delimiter.
func preferRegexpExecLiteralHasGlobalFlag(literal string) bool {
	closing := strings.LastIndexByte(literal, '/')
	if closing < 0 {
		return false
	}
	return strings.Contains(literal[closing+1:], "g")
}

// preferRegexpExecRegexLiteralFor turns a string literal's value into the regex literal upstream
// writes into the repair, and says whether that is possible.
//
// Upstream builds `RegExp(value)` and bails if the constructor throws, then writes
// `regExp.toString()`. `toString` escapes any unescaped forward slash, which is why
// `'^[a-z]+thing/?$'` repairs to `/^[a-z]+thing\/?$/` -- a corpus case turns on exactly that, and a
// port that simply wrapped the value in slashes would write a literal that does not parse.
//
// An empty pattern is `(?:)` in JavaScript, because `//` opens a comment.
//
// The pattern is COMPILED first, and a pattern that will not compile is declined. That is upstream's
// try/catch and it is not decorative: `str.match('[a-z')` is a valid string and an invalid regular
// expression, so upstream reports nothing and a port that only escaped the text would report it and
// then write `/[a-z/.exec(str)`, which does not parse. Upstream's corpus carries that exact case and
// a first draft here failed it -- the only row of thirty-seven it got wrong.
func preferRegexpExecRegexLiteralFor(value string) (string, bool) {
	// `esregexp` is this tree's JavaScript-semantics regular-expression engine, already used by
	// two other rules to answer "would `new RegExp(source)` throw".
	if _, err := esregexp.Compile(value, ""); err != nil {
		return "", false
	}
	if value == "" {
		return "/(?:)/", true
	}
	var escaped strings.Builder
	escaped.WriteByte('/')
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character == '\\' && index+1 < len(value) {
			escaped.WriteByte(character)
			index++
			escaped.WriteByte(value[index])
			continue
		}
		if character == '/' {
			escaped.WriteString(`\/`)
			continue
		}
		if character == '\n' || character == '\r' {
			// A literal newline cannot appear inside a regex literal, and upstream's `toString`
			// would produce one here. Declining is the honest answer; upstream's own guard is the
			// try/catch around the constructor.
			return "", false
		}
		escaped.WriteByte(character)
	}
	escaped.WriteByte('/')
	return escaped.String(), true
}

// preferRegexpExecWrappingFix builds a repair that keeps one inner node.
func preferRegexpExecWrappingFix(
	ctx rule.Context,
	call *ast.Node,
	prefix string,
	inner []*ast.Node,
	suffix string,
) rule.Fix {
	var replacement strings.Builder
	replacement.WriteString(prefix)
	for _, node := range inner {
		replacement.WriteString(preferRegexpExecInnerText(ctx, node))
	}
	replacement.WriteString(suffix)
	return preferRegexpExecOuterFix(ctx, call, replacement.String())
}

// preferRegexpExecWrappingFixTwo builds a repair that keeps two inner nodes, in the order the
// repair writes them.
func preferRegexpExecWrappingFixTwo(
	ctx rule.Context,
	call *ast.Node,
	prefix string,
	first *ast.Node,
	middle string,
	second *ast.Node,
	suffix string,
) rule.Fix {
	replacement := prefix + preferRegexpExecInnerText(ctx, first) + middle +
		preferRegexpExecInnerText(ctx, second) + suffix
	return preferRegexpExecOuterFix(ctx, call, replacement)
}

// preferRegexpExecInnerText renders a preserved node, parenthesised when upstream would parenthesise
// it.
//
// This is the inner half of `getWrappingFixer`: a node that is not strongly precedent is wrapped, so
// that dropping it into a new expression cannot change how it binds.
// `checking.IsStrongPrecedenceNode` is that predicate, already ported and already shared by three
// rules, and it deliberately omits template literals -- which is why a template receiver comes out
// parenthesised here exactly as it does upstream.
func preferRegexpExecInnerText(ctx rule.Context, node *ast.Node) string {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]
	if type_checking.IsStrongPrecedenceNode(node) {
		return text
	}
	return "(" + text + ")"
}

// preferRegexpExecOuterFix wraps the whole replacement when the call sits under a weaker operator.
//
// This is the outer half of `getWrappingFixer`, and it is the half that changes meaning when it is
// missing: `a.match(/x/) || b` has to become `(/x/.exec(a)) || b`, because the replacement is a
// different expression from the one whose precedence the surrounding code was written around.
// Measured against the installed build, since upstream's corpus has no case in this shape.
func preferRegexpExecOuterFix(ctx rule.Context, call *ast.Node, replacement string) rule.Fix {
	callRange := rule.TokenRange(ctx.SourceFile, call)
	if preferRegexpExecHasWeakPrecedenceParent(call) && !preferRegexpExecIsParenthesized(ctx, call) {
		replacement = "(" + replacement + ")"
	}
	return rule.ReplaceRange(core.NewTextRange(callRange.Pos(), callRange.End()), replacement)
}

// preferRegexpExecHasWeakPrecedenceParent is upstream's `isWeakPrecedenceParent`.
//
// The list is upstream's: the parent kinds that bind loosely enough that a replaced child needs its
// own parentheses. A property access is included because `a.match(x).length` would otherwise become
// `RegExp(x).exec(a).length`, which is fine, but `/x/.exec(a).length` after a literal repair is not
// always -- upstream includes the member case and it is reproduced rather than reasoned about.
func preferRegexpExecHasWeakPrecedenceParent(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression, ast.KindConditionalExpression,
		ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindAwaitExpression, ast.KindTypeOfExpression, ast.KindVoidExpression,
		ast.KindSpreadElement, ast.KindTypeAssertionExpression, ast.KindAsExpression:
		return true
	case ast.KindPropertyAccessExpression:
		// Only when the call is the OBJECT being accessed. As the property it cannot be, and as an
		// argument it is already inside parentheses.
		return parent.AsPropertyAccessExpression().Expression == node
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression == node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression == node
	}
	return false
}

// preferRegexpExecIsParenthesized answers whether the call is already wrapped, so the outer fix does
// not add a second pair.
func preferRegexpExecIsParenthesized(ctx rule.Context, node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && parent.Kind == ast.KindParenthesizedExpression
}
