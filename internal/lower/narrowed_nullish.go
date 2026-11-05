package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A narrowing to undefined outlives a call too, narrowed.go's mirror. The checker narrows a binding to
// undefined at an assignment or a check, and keeps it across a call that writes the binding again:
//
//	function fill(): void { later = 4; }
//	let later: number | undefined = undefined;
//	fill();
//	console.log(String(later));
//
// Node prints 4. A consumer that asks the checker for the read's type (String here, a template, ??, ===,
// typeof) sees only undefined and folds the read to it, keeping the read for its effects alone, so both
// backends printed undefined (#r0vzaaa). Reading the storage at its declared type means routing every
// such consumer, so until it's routed a read narrowed to a nullish unit (undefined, null, or the two),
// of a binding declared to hold more, that a function other than the one declaring it writes, is not
// yet. A binding only its own function writes can't change behind a call: the checker sees each of those
// writes in flow, and a call back into that function makes a binding of its own.
func (l *lowering) narrowedNullishReadGap(node *ast.Node) error {
	shorthand := node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment
	if node.Kind != ast.KindIdentifier || !l.isExpression(node) || (ast.IsDeclarationName(node) && !shorthand) || (ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	symbol := l.bindingNamed(node)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 || !l.writtenAcrossFunctions(symbol) {
		return nil
	}
	absent, nullish := l.nullishUnit(l.checker.GetTypeAtLocation(node))
	if !nullish {
		return nil
	}
	if _, declaredNullish := l.nullishUnit(l.checker.GetTypeOfSymbol(symbol)); declaredNullish {
		return nil // Declared to hold nothing else, the narrowed value is the only one it can hold.
	}
	if l.assignedJustBefore(node, symbol) || l.passedAtDeclaredType(node, symbol) {
		return nil
	}
	return l.notYet(node, "a read narrowed to "+absent+" of a binding a call may write")
}

func init() {
	refusalRules.register(3330, "narrowedNullishReadGap", (*lowering).narrowedNullishReadGap)
}

// passedAtDeclaredType reports whether read is the whole argument of a direct call to a function the
// program declares, with one body, whose parameter there, as this call instantiates it, holds every
// member of the binding's declared type:
//
//	console.log(`${keep(build(2))} ${count(kept)}`);
//
// The argument lowers through fit from the read, which localRead takes from storage at the declared
// type, into the parameter's slot, and never asks the checker for the read's narrowed type, so the
// callee sees what a call put back, at its own parameter's type. A parameter typed only undefined,
// which tsc accepts the narrowed read for, or a generic one inferred from it (id(kept), T undefined),
// would fold inside the callee, so it isn't let through. Nor is a library call (String(kept)), a
// function value, a method, which an override may serve with a narrower parameter, an overload set,
// a spread or a rest parameter.
func (l *lowering) passedAtDeclaredType(read *ast.Node, symbol *ast.Symbol) bool {
	argument := read
	for argument.Parent != nil && argument.Parent.Kind == ast.KindParenthesizedExpression {
		argument = argument.Parent
	}
	call := argument.Parent
	// A slot a caller set (slotSet) is where the argument is written, not the parameter the resolved
	// signature names, so nothing here proves the read lowers at its declared type: it stops.
	if call == nil || call.Kind != ast.KindCallExpression || l.slotSet(argument) {
		return false
	}
	index := slices.Index(call.AsCallExpression().Arguments.Nodes, argument)
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	signature := l.checker.GetResolvedSignature(call)
	if index < 0 || callee.Kind != ast.KindIdentifier || signature == nil || signature.Declaration() == nil || index >= len(signature.Parameters()) || (signature.HasRestParameter() && index >= len(signature.Parameters())-1) {
		return false
	}
	function := l.symbol(callee)
	declaration := signature.Declaration()
	if function == nil || len(function.Declarations) != 1 || function.Declarations[0] != declaration || declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil || l.declaredByLibrary(function) {
		return false
	}
	parameter := l.checker.GetTypeOfSymbol(signature.Parameters()[index])
	declared := l.checker.GetTypeOfSymbol(symbol)
	members := []*checker.Type{declared}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		members = declared.Types()
	}
	for _, member := range members {
		if !l.checker.IsTypeAssignableTo(member, parameter) {
			return false
		}
	}
	return true
}

// assignedJustBefore reports whether read sits in the statement right after `symbol = value;`, in the
// same list and the same code, outside any loop, with nothing evaluated before it in its statement that
// can run the program's own code. Then no function can have written the binding since the assignment,
// and the read sees the value assigned, the one the checker narrowed it to:
//
//	best = undefined;
//	console.log(`${best === undefined}`);
//
// Anything else fails closed, as a call may have run between the narrowing and the read.
func (l *lowering) assignedJustBefore(read *ast.Node, symbol *ast.Symbol) bool {
	statement := read
	for statement.Parent != nil && statement.Parent.Kind != ast.KindBlock && statement.Parent.Kind != ast.KindSourceFile {
		if ast.IsIterationStatement(statement.Parent, false) {
			return false // A loop runs what follows the read before the read again.
		}
		statement = statement.Parent
	}
	if statement.Parent == nil || callBoundary(statement) != callBoundary(read) {
		return false
	}
	var statements []*ast.Node
	if statement.Parent.Kind == ast.KindBlock {
		statements = statement.Parent.AsBlock().Statements.Nodes
	} else {
		statements = statement.Parent.AsSourceFile().Statements.Nodes
	}
	index := slices.Index(statements, statement)
	if index < 1 || statements[index-1].Kind != ast.KindExpressionStatement {
		return false
	}
	assignment := ast.SkipParentheses(statements[index-1].AsExpressionStatement().Expression)
	if assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}
	target := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
	return target.Kind == ast.KindIdentifier && l.symbol(target) == symbol && l.runsNothingBefore(statement, read)
}

// runsNothingBefore reports whether nothing node evaluates before read can run the program's own code.
// JavaScript evaluates an expression's parts left to right, and what an expression does with them
// (a call, a conversion, a getter) after, so what runs before read is what ends before it in the text,
// outside the expressions around it.
func (l *lowering) runsNothingBefore(node *ast.Node, read *ast.Node) bool {
	switch {
	case node == read || node.Pos() >= read.End():
		return true
	case node.End() <= read.Pos():
		return l.runsNothing(node)
	}
	safe := true
	node.ForEachChild(func(child *ast.Node) bool {
		safe = l.runsNothingBefore(child, read)
		return !safe
	})
	return safe
}

// runsNothing reports whether evaluating node can't run the program's own code: a literal, a name, a
// function made but not called, ===, !==, &&, ||, ?? and the comma, !, typeof and void over such parts, a
// template whose substitutions are primitives, or a member the library declares of a global it declares
// (console.log), whose getter is the library's. Anything else may call into the program.
func (l *lowering) runsNothing(node *ast.Node) bool {
	every := func(node *ast.Node) bool {
		safe := true
		node.ForEachChild(func(child *ast.Node) bool {
			safe = l.runsNothing(child)
			return !safe
		})
		return safe
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindArrowFunction, ast.KindFunctionExpression:
		return true
	case ast.KindParenthesizedExpression, ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindTemplateExpression:
		return every(node)
	case ast.KindTemplateSpan:
		return l.primitiveTyped(node.AsTemplateSpan().Expression) && every(node)
	case ast.KindPrefixUnaryExpression:
		return node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken && every(node)
	case ast.KindBinaryExpression:
		switch node.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken, ast.KindCommaToken:
			return every(node)
		}
		return false
	case ast.KindPropertyAccessExpression:
		// Only off a global the library declares: a program's own object may be a Proxy, whose get
		// trap runs for a member the library declares too.
		receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
		if receiver.Kind == ast.KindIdentifier {
			return l.declaredByLibrary(l.checker.GetSymbolAtLocation(node.Name())) && l.declaredByLibrary(l.checker.GetSymbolAtLocation(receiver))
		}
		return receiver.Kind == ast.KindPropertyAccessExpression && l.declaredByLibrary(l.checker.GetSymbolAtLocation(node.Name())) && l.runsNothing(receiver)
	}
	return node.Kind < ast.KindIdentifier // A literal, a template's text, or a punctuation token.
}

// declaredByLibrary reports whether every declaration of symbol is the library's or the prelude's.
func (l *lowering) declaredByLibrary(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if file := ast.GetSourceFileOfNode(declaration); !load.IsLibrary(file) && !load.IsPrelude(file) {
			return false
		}
	}
	return true
}

// pureWrite reports whether an assignment target is only written there: the target of a plain =, a
// destructuring pattern's included, or a for-in or for-of head. A compound or logical assignment, ++
// and -- read the binding first, so `s += 'a'` reads the value a call put back.
func pureWrite(target *ast.Node) bool {
	assignment := ast.GetAssignmentTarget(target)
	if assignment == nil {
		return false
	}
	switch assignment.Kind {
	case ast.KindBinaryExpression:
		return assignment.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken
	case ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

// bindingNamed is the binding an identifier reads or writes. In { later } and ({ later } = source) the
// name is the literal's property too, and asked for its symbol the checker gives the property.
func (l *lowering) bindingNamed(identifier *ast.Node) *ast.Symbol {
	if identifier.Parent == nil || identifier.Parent.Kind != ast.KindShorthandPropertyAssignment {
		return l.symbol(identifier)
	}
	symbol := l.checker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	if symbol == nil {
		return nil
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	return l.checker.GetExportSymbolOfSymbol(symbol)
}

// nullishUnit reports whether every member of t is undefined or null, and names what it holds: undefined,
// null, or undefined or null.
func (l *lowering) nullishUnit(t *checker.Type) (string, bool) {
	members := []*checker.Type{t}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		members = t.Types()
	}
	undefined, null := false, false
	for _, member := range members {
		switch {
		case member.Flags()&checker.TypeFlagsUndefined != 0:
			undefined = true
		case member.Flags()&checker.TypeFlagsNull != 0:
			null = true
		default:
			return "", false
		}
	}
	switch {
	case undefined && null:
		return "undefined or null", true
	case null:
		return "null", true
	}
	return "undefined", undefined
}

// writtenAcrossFunctions reports whether any write to symbol sits in a function, a class field's
// initializer or a static block other than the nearest one around its declaration: code a call can run.
// A write is any form ast.IsAssignmentTarget knows, as assignedSymbol's walk (predicates.go) counts
// them: =, a compound or logical assignment, ++ or -- either side, a destructuring target and a for-in
// or for-of head. One walk over every module of the program collects the answer for every symbol the
// first time it's asked; the answer is the program's alone, so nothing invalidates it.
func (l *lowering) writtenAcrossFunctions(symbol *ast.Symbol) bool {
	if l.crossFunctionWrites == nil {
		l.crossFunctionWrites = map[*ast.Symbol]bool{}
		var writes ast.Visitor
		writes = func(node *ast.Node) bool {
			if node.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(node) {
				if written := l.bindingNamed(node); written != nil && written.ValueDeclaration != nil && callBoundary(node) != callBoundary(written.ValueDeclaration) {
					l.crossFunctionWrites[written] = true
				}
			}
			node.ForEachChild(writes)
			return false
		}
		seen := map[*ast.SourceFile]bool{}
		for _, root := range l.program.Files() {
			// Every module the program imports, not only the files it was given: a module's own
			// function can write the binding another module reads (ESMModuleOrder has no effects).
			order, _ := ESMModuleOrder(l.checker, root)
			for _, module := range order {
				if !seen[module] {
					seen[module] = true
					module.AsNode().ForEachChild(writes)
				}
			}
		}
	}
	return l.crossFunctionWrites[symbol]
}

// callBoundary is the nearest code around node that runs when something calls or makes it, rather than
// in the flow around it: a function, a class field's initializer, which runs as its class is made, or a
// static block. Its source file where there is none.
func callBoundary(node *ast.Node) *ast.Node {
	for at := node.Parent; at != nil; at = at.Parent {
		if ast.IsFunctionLike(at) || at.Kind == ast.KindClassStaticBlockDeclaration || at.Kind == ast.KindPropertyDeclaration || at.Kind == ast.KindSourceFile {
			return at
		}
	}
	return nil
}
