package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoTestOnGlobalRegexId = "globalRegexTest"

// correctnessNoTestOnGlobalRegexText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-test-on-global-regex.json`.
var correctnessNoTestOnGlobalRegexText = policy.MessageOf("nexus/correctness-no-test-on-global-regex", correctnessNoTestOnGlobalRegexId)

func correctnessNoTestOnGlobalRegexMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoTestOnGlobalRegexId,
		Description: correctnessNoTestOnGlobalRegexText.Render(nil),
	}
}

// CorrectnessNoTestOnGlobalRegex reports `.test()` on a `g` regex that the call reaches more than once.
//
//	invalid: const dollarPattern = /\$\d+/g; function hasDollar(text: string) { return dollarPattern.test(text); }
//	invalid: const pattern = /x/g; for(const line of lines) { if(pattern.test(line)) count++; }
//	invalid: const pattern = /x/g; lines.filter((line) => pattern.test(line));
//	valid:   const dollarPattern = /\$\d+/; function hasDollar(text: string) { return dollarPattern.test(text); }
//	valid:   function hasDollar(text: string) { const pattern = /\$\d+/g; return pattern.test(text); }
//	valid:   for(let match = pattern.exec(text); match; match = pattern.exec(text)) { ... }
//	valid:   while(pattern.test(text)) count++;
//	valid:   pattern.lastIndex = position; if(pattern.test(text)) { ... }
//
// # Where it came from
//
// `modules/os/wisdom/AhraOsWisdom.ts` in ahra (since rewritten): the comp-leak guard declared
// `compLeakDollarPattern = /\$\s?\d[\d,]*(?:\.\d{2})?/g` at module scope for a `replace`, and also
// called `compLeakDollarPattern.test(text)` from `textCarriesCompCommitment`. A true left `lastIndex`
// past the dollar figure, so the next example's check started mid-string. It was patched by hand with
// `lastIndex = 0` after each true and a comment explaining why. Found by the JavaScript-catalog pass
// of the new-rules sweep (`#tevhg3f`, after sonarjs `stateful-regex`), built as part of `#j03vwm6`.
//
// # The shape, exactly
//
//  1. A call of `.test()` whose symbol is the default library's `RegExp.test`.
//  2. Its receiver resolves, by symbol, to a declaration in this file whose value is a fresh regex
//     with the `g` flag (`gy` included): a `const` binding, or a `readonly` class property
//     (`this.pattern`, `Class.pattern`). The value is a regex literal or `new RegExp(source, flags)`
//     with literal flags and the default library's `RegExp`. A `let` or a writable property may hold
//     another regex by the time of the call, so it is not read.
//  3. The call is reached more than once for one regex: between the regex's creation and the call
//     there is a function boundary (the function may be called again, with the same regex) or a loop
//     around the call that does not also recreate the regex. A regex made in the same run of the same
//     function as the call, outside any loop around it, is new each time, and its `lastIndex` starts at
//     zero: silent, even though the flag does nothing there.
//
// # What it exempts, and why
//
//   - **`.exec()`**, which is not reported at all. `for(let m = re.exec(s); m; m = re.exec(s))` is the
//     idiom that needs `g`, and the research probe's seven `.exec()` hits were all that loop.
//   - **`.test()` as a loop's condition** (`while(re.test(s)) count++`): the same idiom, using the
//     state on purpose, and it leaves `lastIndex` at zero when it ends.
//   - **A regex whose `lastIndex` is set to anything but the literal `0`** anywhere in the file
//     (`re.lastIndex = position`): searching from a position is the one thing `g` gives `.test()`.
//     Setting it to `0` is the hand reset, which does not exempt, because it is the patch this rule
//     replaces.
//   - **A sticky regex without `g`** (`/\w+/y`) is outside the rule: it matches only at `lastIndex`
//     and is written for a tokenizer that positions it on purpose. With `g` as well and no positioning
//     anywhere, it carries state like any `g` regex, and is reported.
//
// # Exact in what it claims
//
// The finding claims the regex carries `lastIndex` from one `.test()` into the next, which is true at
// every reported site. Whether a given site answers wrong depends on what follows a true: a
// find-first loop that returns on the first match, or a function called once per process, never
// tests again after a true. Those are reported too, deliberately: proving that a function runs once
// or that every true exits is a whole-program question, and the fix (drop `g`) is correct and free at
// every one of them.
//
// # No fix
//
// Dropping `g` is right only when no other use of the same regex needs it, and the regex is often
// shared with a `replace` (the AhraOsWisdom case). Splitting it is the author's edit.
var CorrectnessNoTestOnGlobalRegex = rule.Rule{
	Name: "nexus/correctness-no-test-on-global-regex",

	// Symbol identity ties the receiver to its declaration and `test` to RegExp's.
	NeedsTypeChecker: true,

	// Whether `test` and `RegExp` are the default library's, through type_checking.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				correctnessNoTestOnGlobalRegexCheck(ctx, node)
			},
		}
	},
}

func correctnessNoTestOnGlobalRegexCheck(ctx rule.Context, node *ast.Node) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name()
	if name.Kind != ast.KindIdentifier || name.Text() != "test" {
		return
	}
	if !correctnessNoTestOnGlobalRegexIsRegExpMember(ctx, ctx.TypeChecker.GetSymbolAtLocation(name)) {
		return
	}
	receiver := ast.SkipParentheses(access.Expression)
	symbol, creation := correctnessNoTestOnGlobalRegexRegex(ctx, receiver)
	if creation == nil {
		return
	}
	if correctnessNoTestOnGlobalRegexIsLoopCondition(node) {
		return
	}
	if !correctnessNoTestOnGlobalRegexRepeats(node, creation) {
		return
	}
	if correctnessNoTestOnGlobalRegexIsPositioned(ctx, symbol) {
		return
	}
	ctx.ReportNode(node, correctnessNoTestOnGlobalRegexMessage())
}

// correctnessNoTestOnGlobalRegexIsRegExpMember says whether a symbol is a member of the default
// library's RegExp interface, with every declaration there.
func correctnessNoTestOnGlobalRegexIsRegExpMember(ctx rule.Context, symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil || owner.Name().Text() != "RegExp" {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !type_checking.IsSourceFileDefaultLibrary(ctx.Program, file) {
			return false
		}
	}
	return true
}

// correctnessNoTestOnGlobalRegexRegex resolves a `.test()` receiver to the symbol it names and the
// expression that created its regex, when that is a `g` regex held by a `const` or a `readonly`
// property declared in this file. The creation is nil otherwise.
func correctnessNoTestOnGlobalRegexRegex(ctx rule.Context, receiver *ast.Node) (*ast.Symbol, *ast.Node) {
	var name *ast.Node
	switch receiver.Kind {
	case ast.KindIdentifier:
		name = receiver
	case ast.KindPropertyAccessExpression:
		name = receiver.AsPropertyAccessExpression().Name()
	default:
		return nil, nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil
	}
	declaration := symbol.Declarations[0]
	if ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
		return nil, nil
	}
	var initializer *ast.Node
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		list := declaration.Parent
		if receiver.Kind != ast.KindIdentifier || list == nil || list.Kind != ast.KindVariableDeclarationList ||
			list.Flags&ast.NodeFlagsConst == 0 {
			return nil, nil
		}
		initializer = declaration.AsVariableDeclaration().Initializer
	case ast.KindPropertyDeclaration:
		if receiver.Kind != ast.KindPropertyAccessExpression || !ast.HasSyntacticModifier(declaration, ast.ModifierFlagsReadonly) {
			return nil, nil
		}
		initializer = declaration.AsPropertyDeclaration().Initializer
	default:
		return nil, nil
	}
	if initializer == nil {
		return nil, nil
	}
	creation := ast.SkipParentheses(initializer)
	flags, known := correctnessNoTestOnGlobalRegexFlags(ctx, creation)
	if !known || !strings.Contains(flags, "g") {
		return nil, nil
	}
	return symbol, creation
}

// correctnessNoTestOnGlobalRegexFlags reads the flags of a regex literal, or of `new RegExp(source,
// flags)` with the default library's RegExp and a literal flags argument.
func correctnessNoTestOnGlobalRegexFlags(ctx rule.Context, creation *ast.Node) (string, bool) {
	switch creation.Kind {
	case ast.KindRegularExpressionLiteral:
		_, flags := regexsyntax.PatternAndFlags(creation.Text())
		return flags, true
	case ast.KindNewExpression:
		newExpression := creation.AsNewExpression()
		constructor := ast.SkipParentheses(newExpression.Expression)
		if constructor.Kind != ast.KindIdentifier || constructor.Text() != "RegExp" ||
			!type_checking.IsSymbolFromDefaultLibrary(ctx.Program, ctx.TypeChecker.GetSymbolAtLocation(constructor)) {
			return "", false
		}
		if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) != 2 {
			return "", false
		}
		flags := ast.SkipParentheses(newExpression.Arguments.Nodes[1])
		if flags.Kind != ast.KindStringLiteral && flags.Kind != ast.KindNoSubstitutionTemplateLiteral {
			return "", false
		}
		return flags.Text(), true
	}
	return "", false
}

// correctnessNoTestOnGlobalRegexIsLoopCondition says whether a call sits in the condition of a
// `while`, `do`, or `for` loop, where `.test()` walks the matches on purpose.
func correctnessNoTestOnGlobalRegexIsLoopCondition(call *ast.Node) bool {
	child := call
	for current := call.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) || ast.IsStatement(current) && !correctnessNoTestOnGlobalRegexIsLoop(current) {
			return false
		}
		switch current.Kind {
		case ast.KindWhileStatement:
			return current.AsWhileStatement().Expression == child
		case ast.KindDoStatement:
			return current.AsDoStatement().Expression == child
		case ast.KindForStatement:
			return current.AsForStatement().Condition == child
		}
		child = current
	}
	return false
}

func correctnessNoTestOnGlobalRegexIsLoop(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement:
		return true
	}
	return false
}

// correctnessNoTestOnGlobalRegexRepeats says whether the call can run more than once against one regex
// made by the creation: climbing from the call to the nearest node that also holds the creation, it
// passes a function boundary or a part of a loop that runs on every turn.
func correctnessNoTestOnGlobalRegexRepeats(call *ast.Node, creation *ast.Node) bool {
	holdsCreation := map[*ast.Node]bool{}
	for current := creation; current != nil; current = current.Parent {
		holdsCreation[current] = true
	}
	child := call
	for current := call.Parent; current != nil && !holdsCreation[current]; current = current.Parent {
		if ast.IsFunctionLike(current) || current.Kind == ast.KindClassStaticBlockDeclaration {
			return true
		}
		if correctnessNoTestOnGlobalRegexRepeatsInLoop(current, child) {
			return true
		}
		child = current
	}
	return false
}

// correctnessNoTestOnGlobalRegexRepeatsInLoop says whether a child of a loop runs on every turn: the
// body, and a `for`'s condition and update, but not a `for`'s initializer or the iterated expression
// of a `for...of` or `for...in`, which run once.
func correctnessNoTestOnGlobalRegexRepeatsInLoop(loop *ast.Node, child *ast.Node) bool {
	switch loop.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement:
		return true
	case ast.KindForStatement:
		return loop.AsForStatement().Initializer != child
	case ast.KindForOfStatement, ast.KindForInStatement:
		statement := loop.AsForInOrOfStatement()
		return statement.Statement == child
	}
	return false
}

// correctnessNoTestOnGlobalRegexIsPositioned says whether anything in the file sets the regex's
// `lastIndex` to a value other than the literal `0`, which is the one use `g` has for `.test()`.
func correctnessNoTestOnGlobalRegexIsPositioned(ctx rule.Context, symbol *ast.Symbol) bool {
	positioned := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if positioned {
			return true
		}
		if node.Kind == ast.KindBinaryExpression {
			binary := node.AsBinaryExpression()
			target := ast.SkipParentheses(binary.Left)
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && target.Kind == ast.KindPropertyAccessExpression &&
				target.AsPropertyAccessExpression().Name().Text() == "lastIndex" {
				receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
				var name *ast.Node
				switch receiver.Kind {
				case ast.KindIdentifier:
					name = receiver
				case ast.KindPropertyAccessExpression:
					name = receiver.AsPropertyAccessExpression().Name()
				}
				value := ast.SkipParentheses(binary.Right)
				isZero := binary.OperatorToken.Kind == ast.KindEqualsToken && value.Kind == ast.KindNumericLiteral && value.Text() == "0"
				if name != nil && !isZero && ctx.TypeChecker.GetSymbolAtLocation(name) == symbol {
					positioned = true
					return true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)
	return positioned
}
