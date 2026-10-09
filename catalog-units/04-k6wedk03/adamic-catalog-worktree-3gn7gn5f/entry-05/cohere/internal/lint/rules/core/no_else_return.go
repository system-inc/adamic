package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoElseReturn = rule.Message{
	Id: "unexpected",
	Description: "This `else` follows a branch that always returns, so it is doing nothing. The " +
		"code after it only runs when the `if` did not return, which is what an `else` says, and " +
		"the extra nesting costs the reader a level of indentation to learn nothing.",
}

// NoElseReturnSettings is the decoded option surface.
//
// `allowElseIf` defaults to TRUE, which is the trap in this rule's configuration. The generic
// decoder would hand back a zero-value struct with it false, and false is not a no-op here: it
// switches the rule from checking an `if` chain's whole consequent list to checking a single
// consequent, which reports `else if` chains upstream deliberately leaves alone.
type NoElseReturnSettings struct {
	// AllowElseIf permits an `else if` after a returning branch.
	AllowElseIf bool
}

// DefaultNoElseReturnSettings is upstream's `allowElseIf: true`.
func DefaultNoElseReturnSettings() NoElseReturnSettings {
	return NoElseReturnSettings{AllowElseIf: true}
}

type noElseReturnWireShape struct {
	AllowElseIf *bool `json:"allowElseIf"`
}

// DecodeNoElseReturnOptions reads the option object off the config.
//
// Hand-rolled with a pointer wire field so an absent key and an explicit `false` stay
// distinguishable. That is not decoration: the default is true, so a decoder falling back to the
// zero value inverts the rule rather than disabling it, and every fixture routed through a struct
// built by hand would pass anyway.
func DecodeNoElseReturnOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultNoElseReturnSettings(), nil
	}

	var wire noElseReturnWireShape
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return DefaultNoElseReturnSettings(), err
	}

	settings := DefaultNoElseReturnSettings()
	if wire.AllowElseIf != nil {
		settings.AllowElseIf = *wire.AllowElseIf
	}
	return settings, nil
}

// NoElseReturn flags an `else` that follows a branch which always returns.
//
//	valid:   function f() { if (a) { return 1; } return 2; }
//	valid:   function f() { if (a) { b(); } else { return 1; } }
//	invalid: function f() { if (a) { return 1; } else { return 2; } }
//	invalid: function f() { if (a) { return 1; } else { b(); } }
//
// # The judgment is small and the repair is most of the rule
//
// Reporting is a few lines: an `if` whose consequent always returns, whose parent admits a statement
// list, and which has an `else`. The repair unwraps that `else` into the enclosing scope, and
// unwrapping can change what the code means in three separate ways. Upstream declines in all three
// and its corpus asserts every one with `output: null`: 46 of its 78 failing cases are declines
// against 32 fix vectors, so refusing correctly IS this port rather than a detail of it.
//
// # Decline one, a name collision
//
// An `else` block's `let`, `const` and `class` declarations are block scoped, and unwrapping lifts
// them into the enclosing scope where a name may already be taken. `function foo(a) { if (bar) {
// return true; } else { let a; } }` would become a redeclaration of the parameter, which is a syntax
// error rather than a behaviour change.
//
// Upstream asks eslint-scope five ways: declared variables in the scope, a catch clause's binding,
// implicit variables that were referenced, names reaching through from an upper scope, and a `var`
// hoisted out of a nested block. We have no scope table, so this asks resolution the equivalent
// question at the `if` statement itself, and what it can answer was measured before anything was
// built on it. `GetSymbolsInScope` at the `if` reports a collision for a parameter, a defaulted
// parameter, a rest parameter, a catch binding, a `var`, a `let`, a `class`, and a `var` hoisted out
// of a sibling block, and reports none where there is none. Eight of upstream's nine collision
// shapes, and the ninth is not a shape: `let arguments` is a strict-mode parse error, so the two
// corpus cases using it are sloppy-mode only and upstream reports nothing on them here either.
//
// # Decline two, automatic semicolon insertion at the front
//
// When the `if` has no braces and does not end in a semicolon, removing `else` leaves two statements
// that rejoin: `if (foo) return bar \nelse { [1,2,3].map(foo) }` becomes `return bar\n[1,2,3]...`,
// which parses as an index into `bar`. The test is upstream's and deliberately textual.
//
// # Decline three, automatic semicolon insertion at the back
//
// The same hazard at the other end: when the unwrapped block's last token is not a semicolon and
// what follows starts with one of the continuation characters, or simply sits on the same line, the
// two rejoin. `function foo() { if (foo) return bar; else { baz() } qaz() }` is upstream's case.
//
// # Where the fix is deliberately NOT ported
//
// Upstream wraps its replacement in a `FixTracker` that extends the edited range to the whole
// enclosing function, to avoid conflicting with `no-useless-return` and with a second `else` block
// in the same function. We have no such tracker, and widening a repair to a whole function by hand
// is the kind of range this project has already been burned by. The narrow replacement is what ships
// and the conflict case is left to the edit engine, which sees every proposal at once and is the
// only component that can reason about two overlapping ones.
//
// # A TypeScript hazard upstream's corpus cannot express
//
// The repair replaces a span rather than deleting one, so everything living inside that span has to
// survive. Upstream's corpus is JavaScript, and two rules in this tree destroyed type information
// this month while passing every one of its cases. The span here runs from the `else` keyword to the
// end of the else statement and the replacement is that block's own inner text, so a type annotation
// inside the block is carried through unchanged rather than reconstructed. Fixtures below assert
// exactly that on annotated declarations, a generic, and a satisfies expression, and the guard was
// mutated to confirm they fail without it.
var NoElseReturn = rule.Rule{
	Name: "no-else-return",

	// The collision check asks what names are already in scope at the `if`, which only resolution
	// can answer. The rule reports without it, and declines to repair.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoElseReturnSettings](options)
		if !ok {
			settings = DefaultNoElseReturnSettings()
		}

		return rule.Listeners{
			ast.KindIfStatement: func(node *ast.Node) {
				checkNoElseReturn(ctx, node, settings.AllowElseIf)
			},
		}
	},
}

// checkNoElseReturn decides one if statement.
func checkNoElseReturn(ctx rule.Context, node *ast.Node, allowElseIf bool) {
	// Fixing this would mean splitting one statement into two, so an `if` somewhere only a single
	// statement is legal is left alone. Upstream words this as a check against its statement-list
	// parent set; asking whether the parent holds a statement list is the same question.
	if !holdsAStatementList(node.Parent) {
		return
	}

	statement := node.AsIfStatement()
	if statement.ElseStatement == nil {
		return
	}

	if allowElseIf {
		// The whole `else if` chain is one judgment: every consequent must return before the final
		// alternate is reported, and an `if` with no final `else` is left alone. That is what
		// ALLOWING `else if` means, which reads backwards until you see it: permitting `else if`
		// is permitting the chain, so the rule waits for the chain to end before judging it.
		//
		// Upstream spells this as `allowElseIf ? checkIfWithoutElse : checkIfWithElse`, and this
		// port had the two branches swapped until the corpus caught it. Nothing but the option
		// fixtures could: `function foo19() { if (true) { return x; } else if (false) { return y; } }`
		// appears in upstream's VALID list under allowElseIf true and in its INVALID list under
		// false, the same bytes with opposite verdicts.
		alternate := statement.ElseStatement
		for current := node; current.Kind == ast.KindIfStatement; {
			currentStatement := current.AsIfStatement()
			if currentStatement.ElseStatement == nil {
				return
			}
			if !alwaysReturns(currentStatement.ThenStatement) {
				return
			}
			alternate = currentStatement.ElseStatement
			current = currentStatement.ElseStatement
		}
		reportElse(ctx, node, alternate)
		return
	}

	// With `else if` disallowed, only this one consequent is judged, so an `else if` after a
	// returning branch reports immediately rather than waiting for the chain to end.
	if alwaysReturns(statement.ThenStatement) {
		reportElse(ctx, node, statement.ElseStatement)
	}
}

// holdsAStatementList says whether a node's body is a list of statements.
//
// This is upstream's STATEMENT_LIST_PARENTS. An `if` whose parent is another `if`'s branch, or a
// loop body written without braces, sits somewhere only one statement is legal, and unwrapping an
// else there would need the statement split in two.
func holdsAStatementList(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
		ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}

// alwaysReturns says whether a branch returns on every path.
//
// Upstream's naive check, reproduced including its naivety: a block returns when ANY of its
// statements is a return or a fully-returning if, rather than when control provably cannot fall
// through. A more careful analysis would report cases upstream does not, which is a divergence in
// the reporting direction and the one this rule cannot afford.
func alwaysReturns(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindBlock {
		for _, statement := range node.AsBlock().Statements.Nodes {
			if returnsOrFullyReturningIf(statement) {
				return true
			}
		}
		return false
	}
	return returnsOrFullyReturningIf(node)
}

// returnsOrFullyReturningIf says whether one statement is a return, or an if returning both ways.
func returnsOrFullyReturningIf(node *ast.Node) bool {
	if node.Kind == ast.KindReturnStatement {
		return true
	}
	if node.Kind != ast.KindIfStatement {
		return false
	}
	statement := node.AsIfStatement()
	if statement.ElseStatement == nil || statement.ThenStatement == nil {
		return false
	}
	return naiveHasReturn(statement.ElseStatement) && naiveHasReturn(statement.ThenStatement)
}

// naiveHasReturn is upstream's last-statement-only check.
//
// Deliberately only the LAST statement of a block, where alwaysReturns scans all of them. The
// asymmetry is upstream's and reproducing it matters: a block whose return is not last counts for
// one question and not the other.
func naiveHasReturn(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindBlock {
		statements := node.AsBlock().Statements.Nodes
		if len(statements) == 0 {
			return false
		}
		return statements[len(statements)-1].Kind == ast.KindReturnStatement
	}
	return node.Kind == ast.KindReturnStatement
}

// reportElse reports the else and offers the repair when it can be shown safe.
func reportElse(ctx rule.Context, ifStatement *ast.Node, alternate *ast.Node) {
	if repair, ok := elseUnwrapFix(ctx, ifStatement, alternate); ok {
		ctx.ReportNodeWithFixes(alternate, messageNoElseReturn, repair)
		return
	}
	ctx.ReportNode(alternate, messageNoElseReturn)
}

// elseUnwrapFix builds the repair that removes an else, or declines to.
//
// Three independent declines, each reproducing one of upstream's. See the rule's doc comment for
// why each exists; the order here is upstream's and is cheapest-first only by accident.
func elseUnwrapFix(ctx rule.Context, ifStatement *ast.Node, alternate *ast.Node) (rule.Fix, bool) {
	if !isSafeFromNameCollisions(ctx, ifStatement, alternate) {
		return rule.Fix{}, false
	}

	source := ctx.SourceFile.Text()
	elseKeyword := elseKeywordOffset(source, ifStatement, alternate)
	if elseKeyword < 0 {
		return rule.Fix{}, false
	}

	// The text the unwrap keeps: the block's inner text, or the whole statement when there are no
	// braces. Taken as SOURCE TEXT rather than reconstructed, which is what carries a type
	// annotation, a generic argument or a satisfies expression through untouched.
	// TokenRange rather than Pos(): Pos() reaches back over the whitespace between `else` and the
	// statement, and the removal already covers that span, so keeping it too writes a double space.
	// `if (true) return x; else return y;` came out as `... return x;  return y;` until this line.
	keptStart, keptEnd := rule.TokenRange(ctx.SourceFile, alternate).Pos(), alternate.End()
	if alternate.Kind == ast.KindBlock {
		keptStart, keptEnd = alternate.Pos()+1, alternate.End()-1
		if openBrace := indexOfByteFrom(source, '{', alternate.Pos(), alternate.End()); openBrace >= 0 {
			keptStart = openBrace + 1
		}
		if closeBrace := lastIndexOfByteBefore(source, '}', alternate.End(), keptStart); closeBrace >= 0 {
			keptEnd = closeBrace
		}
	}
	if keptStart < 0 || keptEnd > len(source) || keptStart > keptEnd {
		return rule.Fix{}, false
	}

	if rejoinsAtTheFront(ctx, ifStatement, source, keptStart, keptEnd) {
		return rule.Fix{}, false
	}
	if rejoinsAtTheBack(source, alternate, keptStart, keptEnd) {
		return rule.Fix{}, false
	}

	return rule.ReplaceRange(core.NewTextRange(elseKeyword, alternate.End()),
		source[keptStart:keptEnd]), true
}

// isSafeFromNameCollisions says whether unwrapping the else would collide with a name already there.
//
// A block-scoped declaration inside the else is lifted into the enclosing scope by the unwrap, and a
// name already taken there makes the result a redeclaration error or silently rebinds a reference.
// Only `let`, `const` and `class` matter: a `var` or a function declaration is already function
// scoped, so unwrapping moves nothing.
//
// Upstream reads eslint-scope; this asks resolution what names are in scope AT the if statement and
// whether any of them is declared in source rather than in the standard library. Measured before
// being built on, across eight of upstream's nine collision shapes: a parameter, a defaulted
// parameter, a rest parameter, a catch binding, a var, a let, a class, and a var hoisted out of a
// sibling block all answer collision, and a name nothing else declares answers none. The ninth,
// `let arguments`, is a strict-mode parse error rather than a shape.
//
// A conditional function declaration is refused outright, matching upstream: its scoping and
// hoisting differ between engines, so no analysis here would be worth trusting.
func isSafeFromNameCollisions(ctx rule.Context, ifStatement *ast.Node, alternate *ast.Node) bool {
	if alternate.Kind == ast.KindFunctionDeclaration {
		return false
	}
	if alternate.Kind != ast.KindBlock {
		return true
	}
	if ctx.TypeChecker == nil {
		// No checker means the collision question cannot be asked, so the repair is withheld rather
		// than guessed at. The finding still fires.
		return false
	}

	names := blockScopedNamesDeclaredIn(alternate)
	if len(names) == 0 {
		return true
	}

	// Everything visible at the `if`, which is where the unwrapped declarations would land.
	//
	// One narrow divergence lives here, and it runs toward declining rather than toward repairing.
	// `GetSymbolsInScope` reports a HOISTED function declaration as visible at the `if` even when
	// that function belongs to a scope one block further out, so an else block nested inside another
	// block reads as colliding with it. Upstream distinguishes the two levels through eslint-scope
	// and repairs that case.
	//
	// Measured against the installed rule, the divergence is exactly one shape wide:
	//
	//	function foo() { if (bar) { if (baz) { return true; } else { let a; } } function a(){} }
	//	                                                            upstream fixes, this declines
	//	function foo() { if (bar) { return true; } else { let a; } function a(){} }
	//	                                                            both decline
	//	function foo() { function a(){} if (bar) { return true; } else { let a; } }
	//	                                                            both decline
	//
	// Left as is deliberately. A withheld repair costs a manual edit; a wrong one lifts a `let` into
	// a scope where the name is taken and changes what the program means, and the edit engine's
	// parse check cannot refuse that because the result parses. The finding still fires either way,
	// so nothing is missed, only the automatic repair on one shape.
	for _, symbol := range ctx.TypeChecker.GetSymbolsInScope(ifStatement, ^ast.SymbolFlags(0)) {
		if !names[symbol.Name] {
			continue
		}
		for _, declaration := range symbol.Declarations {
			// A declaration in the standard library is not a collision: shadowing a global is legal
			// and is what most of this tree's code does. Only a name declared in SOURCE collides.
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil || file.IsDeclarationFile {
				continue
			}
			// A declaration inside the else block itself is the one being moved, not a collision
			// with something already outside it.
			if declaration.Pos() >= alternate.Pos() && declaration.End() <= alternate.End() {
				continue
			}
			return false
		}
	}

	// A name REFERENCED after the if statement, but declared nowhere the checker can see, would be
	// captured by the lifted declaration and stop resolving to whatever it resolved to before.
	// Upstream asks eslint-scope for `scope.through`, the names reaching out of this scope; asking
	// whether the enclosing function mentions the name outside the else block is the same question
	// asked of the tree.
	//
	// Measured against the installed rule: `function foo() { if (bar) { return true; } else { let a;
	// } a; }` is declined and the same input without the trailing `a;` is fixed.
	if mentionsAnyNameOutside(ctx, ifStatement, alternate, names) {
		return false
	}

	return true
}

// mentionsAnyNameOutside says whether an enclosing scope reads one of the lifted names elsewhere.
//
// Only identifiers outside the else block count, since a mention inside it already refers to the
// declaration being moved. The search is bounded by the nearest enclosing function or the file,
// because a name mentioned beyond that cannot be captured by a declaration landing here.
func mentionsAnyNameOutside(ctx rule.Context, ifStatement *ast.Node, alternate *ast.Node,
	names map[string]bool) bool {
	// The block the unwrapped declarations LAND in, which is the if statement's own parent, and
	// nothing wider. A name mentioned further out belongs to a scope the lifted declaration never
	// reaches, so it is not captured.
	//
	// Measured against the installed rule, and the boundary is exact:
	//
	//	function foo() { if (bar) { if (baz) { return true; } else { let a; } a; } }   declined
	//	function foo() { if (bar) { if (baz) { return true; } else { let a; } } a; }   fixed
	//	function foo() { if (bar) { return true; } else { let a; } a; }               declined
	//
	// Walking to the enclosing function instead declined the middle one, which upstream repairs.
	boundary := ifStatement.Parent
	if boundary == nil {
		return false
	}

	mentioned := false
	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		if mentioned {
			return true
		}
		// Inside the else block is the declaration's own home, not a capture.
		if current.Pos() >= alternate.Pos() && current.End() <= alternate.End() {
			return false
		}
		if current.Kind == ast.KindIdentifier && names[current.Text()] && isAReadOfTheName(current) {
			mentioned = true
			return true
		}
		current.ForEachChild(visit)
		return false
	}
	boundary.ForEachChild(visit)
	return mentioned
}

// isAReadOfTheName says whether an identifier reads a binding rather than naming a new one.
//
// The distinction is load bearing and this port had it wrong first. Upstream's `scope.through` is
// the names REACHING OUT of a scope, so a sibling declaration of the same name is not one of them:
// `function foo() { if (bar) { if (baz) { return true; } else { let a; } } function a(){} }` is
// FIXED upstream, and treating that `function a` as a mention withheld the repair.
//
// So a declaration's own name is not a read, and neither is a property name, since `x.a` does not
// resolve `a` as a binding.
func isAReadOfTheName(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return true
	}
	// The name a declaration introduces is not a read of an existing binding.
	if parent.Name() == identifier {
		switch parent.Kind {
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindVariableDeclaration,
			ast.KindParameter, ast.KindBindingElement, ast.KindMethodDeclaration,
			ast.KindPropertyDeclaration, ast.KindPropertyAssignment, ast.KindEnumDeclaration,
			ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindModuleDeclaration:
			return false
		}
	}
	// A property name is not a binding reference.
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == identifier {
		return false
	}
	return true
}

// blockScopedNamesDeclaredIn collects the names an else block would lift into the enclosing scope.
//
// Only the block's own direct statements, because only those are lifted: a declaration nested inside
// a further block stays block scoped after the unwrap. And only the block-scoped forms, since a
// `var` or a function declaration already belongs to the function scope.
func blockScopedNamesDeclaredIn(block *ast.Node) map[string]bool {
	names := map[string]bool{}
	for _, statement := range block.AsBlock().Statements.Nodes {
		switch statement.Kind {
		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList
			if list == nil {
				continue
			}
			declarationList := list.AsVariableDeclarationList()
			if declarationList == nil ||
				declarationList.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst) == 0 {
				continue
			}
			for _, declaration := range declarationList.Declarations.Nodes {
				collectBindingNames(declaration.Name(), names)
			}

		case ast.KindClassDeclaration:
			if name := statement.Name(); name != nil && name.Kind == ast.KindIdentifier {
				names[name.Text()] = true
			}

		case ast.KindFunctionDeclaration:
			// A function declaration inside a block IS block scoped in strict mode, so lifting one
			// can collide exactly as a `let` does. Upstream reaches the same conclusion from the
			// other side, refusing outright when the whole else IS a function declaration; a
			// function declaration among other statements inside a braced else is this case.
			// Measured against the installed rule: `function foo() { let a; if (bar) { return true;
			// } else { function a(){} } }` is declined and the same input without the outer `let a`
			// is fixed.
			if name := statement.Name(); name != nil && name.Kind == ast.KindIdentifier {
				names[name.Text()] = true
			}
		}
	}
	return names
}

// collectBindingNames records every name a binding introduces, walking patterns.
func collectBindingNames(name *ast.Node, into map[string]bool) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		into[name.Text()] = true
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind == ast.KindBindingElement {
				collectBindingNames(element.AsBindingElement().Name(), into)
			}
		}
	}
}

// rejoinsAtTheFront says whether removing the else lets the consequent swallow what follows.
//
// When the `if` has no braces and does not end in a semicolon, the consequent and the unwrapped
// block become adjacent statements with nothing between them, and a block opening with one of the
// continuation characters binds to the consequent instead. `if (foo) return bar \nelse { [1,2,3] }`
// becomes `return bar\n[1,2,3]`, which indexes into `bar`.
//
// Upstream's test is deliberately textual and both halves are needed: a braced consequent cannot
// rejoin, and neither can one already ending in a semicolon.
func rejoinsAtTheFront(ctx rule.Context, ifStatement *ast.Node, source string,
	keptStart int, keptEnd int) bool {
	consequent := ifStatement.AsIfStatement().ThenStatement
	if consequent == nil || consequent.Kind == ast.KindBlock {
		return false
	}
	// The last non-space byte of the consequent. A semicolon there already terminates it.
	for index := consequent.End() - 1; index >= 0; index-- {
		character := source[index]
		if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		if character == ';' {
			return false
		}
		break
	}
	return startsWithAContinuationCharacter(source, keptStart, keptEnd)
}

// rejoinsAtTheBack says whether the unwrapped text runs into whatever follows the if statement.
//
// The mirror of the front hazard. When the kept text's last token is not a semicolon, the statement
// after the `if` can bind to it, either because it opens with a continuation character or simply
// because it sits on the same line. A closing brace is the one same-line neighbour that is safe.
func rejoinsAtTheBack(source string, alternate *ast.Node, keptStart int, keptEnd int) bool {
	lastByte := lastNonSpaceByte(source, keptStart, keptEnd)
	if lastByte == ';' {
		return false
	}

	next := firstNonSpaceOffsetFrom(source, alternate.End())
	if next < 0 {
		return false
	}
	if isAContinuationCharacter(source[next]) {
		return true
	}
	if source[next] == '}' {
		return false
	}
	// Same line as the kept text's last token means the two would run together.
	return !anyNewlineBetween(source, keptEnd, next)
}

// startsWithAContinuationCharacter says whether kept text opens with a byte that binds leftward.
func startsWithAContinuationCharacter(source string, start int, end int) bool {
	for index := start; index < end; index++ {
		character := source[index]
		if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		return isAContinuationCharacter(character)
	}
	return false
}

// isAContinuationCharacter is upstream's `/^[([/+`-]/u`.
func isAContinuationCharacter(character byte) bool {
	switch character {
	case '(', '[', '/', '+', '`', '-':
		return true
	}
	return false
}

// lastNonSpaceByte gives the last non-whitespace byte in a range, or zero.
func lastNonSpaceByte(source string, start int, end int) byte {
	for index := end - 1; index >= start; index-- {
		character := source[index]
		if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		return character
	}
	return 0
}

// firstNonSpaceOffsetFrom gives the offset of the next non-whitespace byte, or -1.
func firstNonSpaceOffsetFrom(source string, start int) int {
	for index := start; index < len(source); index++ {
		character := source[index]
		if character == ' ' || character == '\t' || character == '\n' || character == '\r' {
			continue
		}
		return index
	}
	return -1
}

// anyNewlineBetween says whether a line break separates two offsets.
func anyNewlineBetween(source string, start int, end int) bool {
	for index := start; index < end && index < len(source); index++ {
		if source[index] == '\n' {
			return true
		}
	}
	return false
}

// elseKeywordOffset finds the `else` token between the consequent and the alternate.
//
// Scanned rather than taken from a node, because the keyword is a token the tree does not hand back
// as a child. The search is bounded by the consequent's end and the alternate's start, so it cannot
// wander into either.
func elseKeywordOffset(source string, ifStatement *ast.Node, alternate *ast.Node) int {
	// The `else` immediately before the ALTERNATE being reported, not the first one in the chain.
	//
	// In `if (x) {...} else if (y) {...} else {...}` under allowElseIf, the reported alternate is
	// the final block and only its own `else` is removed. Searching forward from the outermost
	// consequent finds the chain's FIRST `else` and swallows the whole `else if`, which is what this
	// port did until upstream's own output for foo9 said otherwise.
	//
	// Searching backward from the alternate is what pins it to the right one, and the search is
	// floored at the enclosing if so it cannot wander out of the statement.
	floor := ifStatement.Pos()
	for index := alternate.Pos() - 4; index >= floor && index >= 0; index-- {
		if index+4 <= len(source) && source[index:index+4] == "else" {
			return index
		}
	}
	return -1
}

// indexOfByteFrom finds a byte within a range, or -1.
func indexOfByteFrom(source string, target byte, start int, end int) int {
	for index := start; index < end && index < len(source); index++ {
		if source[index] == target {
			return index
		}
	}
	return -1
}

// lastIndexOfByteBefore finds the last occurrence of a byte before an offset, or -1.
func lastIndexOfByteBefore(source string, target byte, end int, floor int) int {
	for index := end - 1; index >= floor && index >= 0; index-- {
		if source[index] == target {
			return index
		}
	}
	return -1
}
