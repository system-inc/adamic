package lower

import (
	"slices"
	"strconv"

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
// writes in flow, and a call back into that function makes a binding of its own. A property is the same,
// proved by name (narrowedNullishPropertyReadGap).
func (l *lowering) narrowedNullishReadGap(node *ast.Node) error {
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		return l.narrowedNullishPropertyReadGap(node)
	}
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
	if l.assignedJustBefore(node, symbol) {
		return nil
	}
	return l.notYet(node, "a read narrowed to "+absent+" of a binding a call may write")
}

func init() {
	refusalRules.register(3330, "narrowedNullishReadGap", (*lowering).narrowedNullishReadGap)
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

// narrowedNullishPropertyReadGap is the rule for a property read, b.x, this.x, a.b.x or b['x'], narrowed
// the same way: after b.x = undefined, a call that writes b.x = 4 leaves the checker's undefined in place,
// and String(b.x) printed undefined. A property has no declaration that keeps its writes in one function,
// since any alias of its object can write it, so the proof is by name over the whole program: the read
// stops when a write of its name sits in code other than the read's own (callBoundary), or when the
// program has a write that could reach any name (propertyWritesOf). A write of the name in the read's own
// code, through the same reference, is one the checker sees in flow. A read of a name a number spells,
// t[0], o['1'] or o.NaN, stops whenever it's narrowed: an array's own methods write its elements behind
// a call, and so does o[i] = 4 anywhere. A read through a key that isn't a literal is never narrowed.
func (l *lowering) narrowedNullishPropertyReadGap(node *ast.Node) error {
	if !l.isExpression(node) || (ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	var key *ast.Node
	if node.Kind == ast.KindPropertyAccessExpression {
		key = node.Name()
	} else {
		key = ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNoSubstitutionTemplateLiteral && key.Kind != ast.KindNumericLiteral {
			return nil
		}
	}
	absent, nullish := l.nullishUnit(l.checker.GetTypeAtLocation(node))
	if !nullish {
		return nil
	}
	if property := l.checker.GetSymbolAtLocation(key); property != nil {
		if _, declaredNullish := l.nullishUnit(l.checker.GetTypeOfSymbol(property)); declaredNullish {
			return nil // Declared to hold nothing else, as for a binding.
		}
	}
	writes := l.propertyWritesOf()
	_, notNumeric := strconv.ParseFloat(key.Text(), 64) // A name a number spells parses, NaN and Infinity among them.
	stop := writes.everyName || notNumeric == nil
	for _, boundary := range writes.named[key.Text()] {
		stop = stop || boundary != callBoundary(node)
	}
	if !stop {
		return nil
	}
	return l.notYet(node, "a read narrowed to "+absent+" of a property a call may write")
}

// propertyWrites is every write of a property in the program, by what it can reach (propertyWritesOf).
type propertyWrites struct {
	// named is the code each write of a name sits in (callBoundary), nil for one through a literal key,
	// o['x'] = 4, which the rule counts as written elsewhere whatever code it sits in.
	named map[string][]*ast.Node
	// everyName is a write that could reach any name: through a key that isn't a literal, a number or a
	// symbol, through an any or a string index signature, a write of __proto__, a setter anywhere, or a
	// library member that writes any name (writingLibraryMembers).
	everyName bool
}

// writingLibraryMembers are the library's members that can write any name of an object the program
// already has, the road a spread takes into one: a prototype that gains f changes what an absent f
// reads too. Any other library or host call writes no name the program spells, and a callback one
// calls is the program's own code, whose writes are counted by name. Object or Reflect passed on whole
// counts as one of these, since what's called through it can't be told.
var writingLibraryMembers = map[string]map[string]bool{
	"Object":  {"assign": true, "defineProperty": true, "defineProperties": true, "setPrototypeOf": true},
	"Reflect": {"set": true, "defineProperty": true, "setPrototypeOf": true},
}

// propertyWritesOf collects every property write in every module the program imports, the first time
// it's asked, as writtenAcrossFunctions does for bindings. A write is any form ast.IsAssignmentTarget
// knows, on a property access or an element access: =, a compound or logical assignment, ++ or --, a
// destructuring target and a for-in or for-of head. A delete counts as one too: after o.x = null, a
// delete in a callee leaves undefined where the checker still says null.
func (l *lowering) propertyWritesOf() *propertyWrites {
	if l.propertyWrites != nil {
		return l.propertyWrites
	}
	l.propertyWrites = &propertyWrites{named: map[string][]*ast.Node{}}
	l.walkProgramModules(func(node *ast.Node) {
		switch {
		case (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression) && ast.IsAssignmentTarget(node):
			l.notePropertyWrite(node)
		case node.Kind == ast.KindDeleteExpression:
			if operand := ast.SkipParentheses(node.AsDeleteExpression().Expression); operand.Kind == ast.KindPropertyAccessExpression || operand.Kind == ast.KindElementAccessExpression {
				l.notePropertyWrite(operand)
			}
		case node.Kind == ast.KindSetAccessor:
			l.propertyWrites.everyName = true // o.f = 4 through a setter runs its body, which can write any name.
		case node.Kind == ast.KindIdentifier && writingLibraryMembers[node.Text()] != nil && l.isExpression(node) && l.isLibraryGlobal(node, node.Text()):
			member := node.Parent
			if member.Kind != ast.KindPropertyAccessExpression || member.AsPropertyAccessExpression().Expression != node || writingLibraryMembers[node.Text()][member.Name().Text()] {
				l.propertyWrites.everyName = true
			}
		}
	})
	return l.propertyWrites
}

// notePropertyWrite files one write of a property access or an element access by what it can reach.
func (l *lowering) notePropertyWrite(target *ast.Node) {
	var receiver *ast.Node
	if target.Kind == ast.KindPropertyAccessExpression {
		receiver = target.AsPropertyAccessExpression().Expression
	} else {
		receiver = target.AsElementAccessExpression().Expression
	}
	receiverType := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	if receiverType.Flags()&checker.TypeFlagsAny != 0 {
		l.propertyWrites.everyName = true
	}
	for _, info := range l.checker.GetIndexInfosOfType(receiverType) {
		if info.KeyType().Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsESSymbolLike) == 0 {
			l.propertyWrites.everyName = true
		}
	}
	if target.Kind == ast.KindPropertyAccessExpression {
		name := target.Name().Text()
		l.propertyWrites.everyName = l.propertyWrites.everyName || name == "__proto__"
		l.propertyWrites.named[name] = append(l.propertyWrites.named[name], callBoundary(target))
		return
	}
	key := ast.SkipParentheses(target.AsElementAccessExpression().ArgumentExpression)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral {
		l.propertyWrites.everyName = l.propertyWrites.everyName || key.Text() == "__proto__"
		l.propertyWrites.named[key.Text()] = append(l.propertyWrites.named[key.Text()], nil)
		return
	}
	keyType := l.checker.GetTypeAtLocation(key)
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	for _, member := range members {
		// A number key writes only a name a number spells, and every read of one stops already. A
		// symbol key writes no name a read spells.
		if member.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsESSymbolLike) == 0 {
			l.propertyWrites.everyName = true
		}
	}
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
		l.walkProgramModules(func(node *ast.Node) {
			if node.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(node) {
				if written := l.bindingNamed(node); written != nil && written.ValueDeclaration != nil && callBoundary(node) != callBoundary(written.ValueDeclaration) {
					l.crossFunctionWrites[written] = true
				}
			}
		})
	}
	return l.crossFunctionWrites[symbol]
}

// walkProgramModules visits every node of every module the program imports, each module once, and not
// only the files it was given: a module's own function can write what another module reads
// (ESMModuleOrder has no effects).
func (l *lowering) walkProgramModules(visit func(*ast.Node)) {
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		visit(node)
		node.ForEachChild(walk)
		return false
	}
	seen := map[*ast.SourceFile]bool{}
	for _, root := range l.program.Files() {
		order, _ := ESMModuleOrder(l.checker, root)
		for _, module := range order {
			if !seen[module] {
				seen[module] = true
				module.AsNode().ForEachChild(walk)
			}
		}
	}
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
