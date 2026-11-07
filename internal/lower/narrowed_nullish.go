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
func (lowering *lowering) narrowedNullishReadGap(node *ast.Node) error {
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		return lowering.narrowedNullishPropertyReadGap(node)
	}
	shorthand := node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment
	if node.Kind != ast.KindIdentifier ||
		!lowering.isExpression(node) ||
		(ast.IsDeclarationName(node) && !shorthand) ||
		(ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	symbol := lowering.bindingNamed(node)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 || !lowering.writtenAcrossFunctions(symbol) {
		return nil
	}
	absent, nullish := lowering.nullishUnit(lowering.checker.GetTypeAtLocation(node))
	if !nullish {
		return nil
	}
	if _, declaredNullish := lowering.nullishUnit(lowering.checker.GetTypeOfSymbol(symbol)); declaredNullish {
		return nil // Declared to hold nothing else, the narrowed value is the only one it can hold.
	}
	if lowering.assignedJustBefore(node, symbol) || lowering.passedAtDeclaredType(node, symbol) {
		return nil
	}
	return lowering.notYet(node, "a read narrowed to "+absent+" of a binding a call may write")
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
func (lowering *lowering) passedAtDeclaredType(read *ast.Node, symbol *ast.Symbol) bool {
	argument := read
	for argument.Parent != nil && argument.Parent.Kind == ast.KindParenthesizedExpression {
		argument = argument.Parent
	}
	call := argument.Parent
	// A slot a caller set (slotSet) is where the argument is written, not the parameter the resolved
	// signature names, so nothing here proves the read lowers at its declared type: it stops.
	if call == nil || call.Kind != ast.KindCallExpression || lowering.slotSet(argument) {
		return false
	}
	index := slices.Index(call.AsCallExpression().Arguments.Nodes, argument)
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	signature := lowering.checker.GetResolvedSignature(call)
	if index < 0 ||
		callee.Kind != ast.KindIdentifier ||
		signature == nil ||
		signature.Declaration() == nil ||
		index >= len(signature.Parameters()) ||
		(signature.HasRestParameter() && index >= len(signature.Parameters())-1) {
		return false
	}
	function := lowering.symbol(callee)
	declaration := signature.Declaration()
	if function == nil ||
		len(function.Declarations) != 1 ||
		function.Declarations[0] != declaration ||
		declaration.Kind != ast.KindFunctionDeclaration ||
		declaration.Body() == nil ||
		lowering.declaredByLibrary(function) {
		return false
	}
	parameter := lowering.checker.GetTypeOfSymbol(signature.Parameters()[index])
	declared := lowering.checker.GetTypeOfSymbol(symbol)
	members := []*checker.Type{declared}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		members = declared.Types()
	}
	for _, member := range members {
		if !lowering.checker.IsTypeAssignableTo(member, parameter) {
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
func (lowering *lowering) assignedJustBefore(read *ast.Node, symbol *ast.Symbol) bool {
	statement := read
	for statement.Parent != nil &&
		statement.Parent.Kind != ast.KindBlock &&
		statement.Parent.Kind != ast.KindSourceFile {
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
	if assignment.Kind != ast.KindBinaryExpression ||
		assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}
	target := ast.SkipParentheses(assignment.AsBinaryExpression().Left)
	return target.Kind == ast.KindIdentifier &&
		lowering.symbol(target) == symbol &&
		lowering.runsNothingBefore(statement, read)
}

// runsNothingBefore reports whether nothing node evaluates before read can run the program's own code.
// JavaScript evaluates an expression's parts left to right, and what an expression does with them
// (a call, a conversion, a getter) after, so what runs before read is what ends before it in the text,
// outside the expressions around it.
func (lowering *lowering) runsNothingBefore(node *ast.Node, read *ast.Node) bool {
	switch {
	case node == read || node.Pos() >= read.End():
		return true
	case node.End() <= read.Pos():
		return lowering.runsNothing(node)
	}
	safe := true
	node.ForEachChild(func(child *ast.Node) bool {
		safe = lowering.runsNothingBefore(child, read)
		return !safe
	})
	return safe
}

// runsNothing reports whether evaluating node can't run the program's own code: a literal, a name, a
// function made but not called, ===, !==, &&, ||, ?? and the comma, !, typeof and void over such parts, a
// template whose substitutions are primitives, or a member the library declares of a global it declares
// (console.log), whose getter is the library's. Anything else may call into the program.
func (lowering *lowering) runsNothing(node *ast.Node) bool {
	every := func(node *ast.Node) bool {
		safe := true
		node.ForEachChild(func(child *ast.Node) bool {
			safe = lowering.runsNothing(child)
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
		return lowering.primitiveTyped(node.AsTemplateSpan().Expression) && every(node)
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
			return lowering.declaredByLibrary(lowering.checker.GetSymbolAtLocation(node.Name())) &&
				lowering.declaredByLibrary(lowering.checker.GetSymbolAtLocation(receiver))
		}
		return receiver.Kind == ast.KindPropertyAccessExpression &&
			lowering.declaredByLibrary(lowering.checker.GetSymbolAtLocation(node.Name())) &&
			lowering.runsNothing(receiver)
	}
	return node.Kind < ast.KindIdentifier // A literal, a template's text, or a punctuation token.
}

// declaredByLibrary reports whether every declaration of symbol is the library's or the prelude's.
func (lowering *lowering) declaredByLibrary(symbol *ast.Symbol) bool {
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
// and String(b.x) printed undefined. A property has no declaration that keeps its writes in one function:
// any alias of its object can write it, in any function, the read's own included (o2.x = 4 with o2 the
// same object as o, or a recursive call back into it). A getter of the name, or a program that could
// install one (propertyWrites.accessors), stops every such read first, since a getter answers each read
// itself. Then the read lowers when the place it was last seen holding the narrowed value (an
// assignment, a delete or a check, nullishObservationOf) is followed by nothing that can run program
// code before the read (codeBetween). Otherwise it stops when a number spells its name (t[0], o['1']),
// whose elements an array's own methods write behind a call, when some write could reach any name, or
// when a write of its name could run between: one in the read's own code between the two, or one a
// call between can run (writeMayRunBetween). A read through a key that isn't a literal is never narrowed.
func (lowering *lowering) narrowedNullishPropertyReadGap(node *ast.Node) error {
	if !lowering.isExpression(node) || (ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	var key *ast.Node
	if node.Kind == ast.KindPropertyAccessExpression {
		key = node.Name()
	} else {
		key = ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind != ast.KindStringLiteral &&
			key.Kind != ast.KindNoSubstitutionTemplateLiteral &&
			key.Kind != ast.KindNumericLiteral {
			return nil
		}
	}
	absent, nullish := lowering.nullishUnit(lowering.checker.GetTypeAtLocation(node))
	if !nullish {
		return nil
	}
	if property := lowering.checker.GetSymbolAtLocation(key); property != nil {
		if _, declaredNullish := lowering.nullishUnit(lowering.checker.GetTypeOfSymbol(property)); declaredNullish {
			return nil // Declared to hold nothing else, as for a binding.
		}
	}
	stop := lowering.notYet(node, "a read narrowed to "+absent+" of a property a call may write")
	// A getter answers each read itself, so neither an assignment just before nor a missing write
	// proves what it gives: get x() { return n > 0 ? 4 : undefined; } changes with n alone. One the
	// program writes by name (accessorName), or one a descriptor or a prototype could install
	// (accessors), stops the read whatever the flow says.
	writes := lowering.propertyWritesOf()
	if writes.accessors || lowering.accessorName(key.Text()) {
		return stop
	}
	_, notNumeric := strconv.ParseFloat(key.Text(), 64) // A name a number spells parses, NaN and Infinity among them.
	numeric := notNumeric == nil
	observation := lowering.nullishObservationOf(node, absent)
	if observation == nil {
		if numeric || writes.everyName || len(writes.named[key.Text()]) > 0 {
			return stop
		}
		return nil
	}
	if !lowering.codeBetween(observation.from, node, lowering.appendsBetween(observation, node, numeric)) {
		return nil
	}
	if numeric || writes.everyName {
		return stop
	}
	reach := lowering.reachBetween(observation.from, node)
	for _, target := range writes.named[key.Text()] {
		if target != observation.target && lowering.writeMayRunBetween(target, observation.from, node, reach) {
			return stop
		}
	}
	return nil
}

// nullishObservation is where a read narrowed to a nullish unit was last seen holding the value the
// checker narrowed it to, so that the read gives that value unless something writes it after: target
// is the assignment's or the delete's target, or the reference a check reads, and from is where what
// runs after it starts. assigned is a plain =, after which the object has the name as its own, so an
// array has that index.
type nullishObservation struct {
	target   *ast.Node
	from     int
	assigned bool
}

// nullishObservationDepth bounds the joins nullishObservationOf walks through; past it, there's no
// observation, and the read is judged as one with none.
const nullishObservationDepth = 32

// nullishObservationOf walks the read's flow straight back to the place it was last seen holding the
// narrowed value: a plain assignment of it, a delete of it, or a check that proves it holds exactly that
// nullish unit (=== null, !== undefined on its false side, typeof x === 'undefined'). A condition that doesn't mention
// it, a call, which codeBetween judges, and a join whose every way in reaches the same observation (an optional
// chain's) are walked through. A loop, a label of any other kind, a write of what holds the read, another test of it or
// the start of a function ends the walk with none, since the value there isn't the one seen.
func (lowering *lowering) nullishObservationOf(read *ast.Node, absent string) *nullishObservation {
	data := read.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return nil
	}
	return lowering.nullishObservationAlong(read, absent, data.FlowNode, 0, map[*ast.FlowNode]*nullishObservation{})
}

func (lowering *lowering) nullishObservationAlong(
	read *ast.Node,
	absent string,
	flow *ast.FlowNode,
	joins int,
	seen map[*ast.FlowNode]*nullishObservation,
) *nullishObservation {
	for flow != nil {
		switch {
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(lowering.checker, read, flow.Node) {
				parent := flow.Node.Parent
				switch {
				case parent != nil &&
					parent.Kind == ast.KindBinaryExpression &&
					parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken &&
					parent.AsBinaryExpression().Left == flow.Node:
					return &nullishObservation{target: flow.Node, from: parent.End(), assigned: true}
				case parent != nil && parent.Kind == ast.KindDeleteExpression:
					return &nullishObservation{target: flow.Node, from: parent.End()}
				}
				return nil
			}
			if enumTagContainsReference(lowering.checker, read, flow.Node) {
				return nil
			}
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if checked := lowering.nullishCheck(read, absent, flow); checked != nil {
				return &nullishObservation{target: checked, from: checked.End()}
			}
			if lowering.mentionsReference(read, flow.Node) {
				return nil
			}
		case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
		case flow.Flags&ast.FlowFlagsBranchLabel != 0:
			if found, walked := seen[flow]; walked {
				return found
			}
			seen[flow] = nil // A cycle can't arise without a loop label, but if one did it ends here.
			if joins >= nullishObservationDepth {
				return nil
			}
			var common *nullishObservation
			for list := flow.Antecedents; list != nil; list = list.Next {
				if list.Flow.Flags&ast.FlowFlagsUnreachable != 0 {
					continue
				}
				found := lowering.nullishObservationAlong(read, absent, list.Flow, joins+1, seen)
				if found == nil || common != nil && found.target != common.target {
					return nil
				}
				common = found
			}
			seen[flow] = common
			return common
		default:
			return nil
		}
		flow = flow.Antecedent
	}
	return nil
}

// nullishCheck is the reference a condition reads when, on the side flow takes, it proves the
// reference holds exactly absent: x === null or x === undefined when true, x !== null or x !== undefined
// when false, and typeof x === 'undefined' the same way. Nil for any other condition (== is refused).
func (lowering *lowering) nullishCheck(read *ast.Node, absent string, flow *ast.FlowNode) *ast.Node {
	condition := ast.SkipParentheses(flow.Node)
	if condition.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := condition.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken:
		if flow.Flags&ast.FlowFlagsTrueCondition == 0 {
			return nil
		}
	case ast.KindExclamationEqualsEqualsToken:
		if flow.Flags&ast.FlowFlagsTrueCondition != 0 {
			return nil
		}
	default:
		return nil
	}
	for _, sides := range [][2]*ast.Node{{binary.Left, binary.Right}, {binary.Right, binary.Left}} {
		reference, other := ast.SkipParentheses(sides[0]), ast.SkipParentheses(sides[1])
		var proves string
		switch {
		case reference.Kind == ast.KindTypeOfExpression &&
			(other.Kind == ast.KindStringLiteral || other.Kind == ast.KindNoSubstitutionTemplateLiteral) &&
			other.Text() == "undefined":
			reference, proves = ast.SkipParentheses(reference.AsTypeOfExpression().Expression), "undefined"
		case other.Kind == ast.KindNullKeyword:
			proves = "null"
		case other.Kind == ast.KindIdentifier &&
			other.Text() == "undefined" &&
			lowering.checker.GetTypeAtLocation(other).Flags() == checker.TypeFlagsUndefined:
			proves = "undefined"
		default:
			continue
		}
		if proves == absent && predicateMatchingReference(lowering.checker, read, reference) {
			return reference
		}
	}
	return nil
}

// mentionsReference is whether node reads or writes what read reads, anywhere inside it.
func (lowering *lowering) mentionsReference(read *ast.Node, node *ast.Node) bool {
	if node == nil {
		return false
	}
	if predicateMatchingReference(lowering.checker, read, node) {
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		found = lowering.mentionsReference(read, child)
		return found
	})
	return found
}

// appendsBetween is the calls between an observation and its read that codeBetween needn't count: a
// library push onto an array, whose arguments run nothing, runs no program code and writes only an index
// past the array's end and its length. So it can't change a named property, nor an index that a plain
// assignment just gave the array, which is below its end. After a delete or a check the index may be
// the end, which push writes, so a number's name exempts none then.
func (lowering *lowering) appendsBetween(
	observation *nullishObservation,
	read *ast.Node,
	numeric bool,
) map[*ast.Node]bool {
	if numeric && !observation.assigned {
		return nil
	}
	var exempt map[*ast.Node]bool
	lowering.visitBetween(observation.from, read, func(node *ast.Node) {
		if node.Kind != ast.KindCallExpression {
			return
		}
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression ||
			callee.Name().Text() != "push" ||
			!lowering.declaredByLibrary(lowering.checker.GetSymbolAtLocation(callee.Name())) ||
			!lowering.isArrayType(lowering.checker.GetNonNullableType(
				lowering.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression),
			)) {
			return
		}
		for _, argument := range call.Arguments.Nodes {
			if !lowering.primitiveTyped(argument) {
				return
			}
		}
		if exempt == nil {
			exempt = map[*ast.Node]bool{}
		}
		exempt[node] = true
	})
	return exempt
}

// visitBetween visits every node in the read's own code (codeBetween's walk) that completes after from
// and before the read starts, nested functions included as nodes but not entered.
func (lowering *lowering) visitBetween(from int, read *ast.Node, visit func(*ast.Node)) {
	container := read.Parent
	for container != nil && !ast.IsFunctionLike(container) && container.Kind != ast.KindSourceFile {
		container = container.Parent
	}
	if container == nil {
		return
	}
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		if node.Pos() >= read.Pos() {
			return false
		}
		if node.End() > from && node.End() <= read.Pos() {
			visit(node)
		}
		if !ast.IsFunctionLike(node) {
			node.ForEachChild(walk)
		}
		return false
	}
	container.ForEachChild(walk)
}

// codeReach is the program code that may run between an observation and its read: every function a
// call there can run, transitively, or everything, when some code there runs what can't be named.
type codeReach struct {
	everything bool
	functions  map[*ast.Node]bool
}

// reachBetween is what may run between from and read. A direct call of a function the program declares
// once, with a body, runs that body, and what it calls in turn. A library call runs no program code but
// the callbacks passed to it, when its receiver and every other argument are primitives or a library
// global, so nothing of the program's is converted or iterated. A function made there may be called.
// A write runs nothing, since a setter anywhere makes every name written (everyName). Anything else that
// can run code (a method call, a function value's call, new, a conversion, an iterator, a getter, an
// element read, which a computed getter could serve) reaches everything, and so fails closed.
func (lowering *lowering) reachBetween(from int, read *ast.Node) *codeReach {
	reach := &codeReach{functions: map[*ast.Node]bool{}}
	lowering.visitBetween(from, read, func(node *ast.Node) {
		if !reach.everything {
			lowering.reachFrom(node, reach)
		}
	})
	return reach
}

// reachFrom adds what evaluating node itself may run, beyond its parts, to reach.
func (lowering *lowering) reachFrom(node *ast.Node, reach *codeReach) {
	if ast.IsFunctionLike(node) {
		lowering.reachFunction(node, reach)
		return
	}
	if !lowering.mayRunCode(node) {
		return
	}
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindIdentifier {
			if symbol := lowering.symbol(callee); symbol != nil &&
				len(symbol.Declarations) == 1 &&
				symbol.Declarations[0].Kind == ast.KindFunctionDeclaration &&
				symbol.Declarations[0].Body() != nil &&
				!lowering.declaredByLibrary(symbol) {
				lowering.reachFunction(symbol.Declarations[0], reach)
				return
			}
		}
		if lowering.libraryCallRunsOnlyCallbacks(callee, call.Arguments.Nodes, reach) {
			return
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) &&
			writesStorage(binary.Left) &&
			(binary.OperatorToken.Kind == ast.KindEqualsToken ||
				lowering.primitiveTyped(binary.Left) && lowering.primitiveTyped(binary.Right)) {
			return // A write, judged by name: = converts nothing, and a compound one here only primitives.
		}
	case ast.KindPrefixUnaryExpression:
		if lowering.primitiveTyped(node.AsPrefixUnaryExpression().Operand) {
			return // ++, --, +, - and ~ on a primitive convert nothing of the program's.
		}
	case ast.KindPostfixUnaryExpression:
		if lowering.primitiveTyped(node.AsPostfixUnaryExpression().Operand) {
			return
		}
	case ast.KindDeleteExpression:
		return // A delete writes, judged by name, and runs nothing.
	case ast.KindElementAccessExpression:
		if len(lowering.programAccessorNames()) == 0 {
			return // With no getter anywhere and none installed, an element read runs nothing.
		}
	}
	reach.everything = true
}

// programAccessorNames is every getter and setter name the program writes, accessorName's set.
func (lowering *lowering) programAccessorNames() map[string]bool {
	lowering.accessorName("")
	return lowering.writtenAccessorNames
}

// libraryCallRunsOnlyCallbacks reports whether a call is the library's, through a callee every
// declaration of which is the library's or the prelude's, on a primitive or a library global, with
// every argument a primitive or a callback, and adds each callback to reach: the program's own code it
// can run. Nil when the call isn't one, and reach is left as it was.
func (lowering *lowering) libraryCallRunsOnlyCallbacks(callee *ast.Node, arguments []*ast.Node, reach *codeReach) bool {
	switch callee.Kind {
	case ast.KindIdentifier:
		if !lowering.declaredByLibrary(lowering.symbol(callee)) {
			return false
		}
	case ast.KindPropertyAccessExpression:
		receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		if !lowering.declaredByLibrary(lowering.checker.GetSymbolAtLocation(callee.Name())) {
			return false
		}
		if !lowering.primitiveTyped(receiver) &&
			!(receiver.Kind == ast.KindIdentifier && lowering.declaredByLibrary(lowering.symbol(receiver))) {
			return false
		}
	default:
		return false
	}
	var callbacks []*ast.Node
	for _, argument := range arguments {
		argument = ast.SkipParentheses(argument)
		switch {
		case lowering.primitiveTyped(argument):
		case ast.IsFunctionLike(argument):
			callbacks = append(callbacks, argument)
		case argument.Kind == ast.KindIdentifier:
			symbol := lowering.symbol(argument)
			if symbol == nil ||
				len(symbol.Declarations) != 1 ||
				symbol.Declarations[0].Kind != ast.KindFunctionDeclaration ||
				symbol.Declarations[0].Body() == nil {
				return false
			}
			callbacks = append(callbacks, symbol.Declarations[0])
		default:
			return false
		}
	}
	for _, callback := range callbacks {
		lowering.reachFunction(callback, reach)
	}
	return true
}

// reachFunction adds a function's body to reach, and what it may run in turn, nested functions
// included, since its body may call or hand them out.
func (lowering *lowering) reachFunction(function *ast.Node, reach *codeReach) {
	if reach.everything || reach.functions[function] {
		return
	}
	reach.functions[function] = true
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		if reach.everything {
			return true
		}
		if ast.IsFunctionLike(node) {
			lowering.reachFunction(node, reach)
			return false
		}
		lowering.reachFrom(node, reach)
		node.ForEachChild(walk)
		return reach.everything
	}
	function.ForEachChild(walk)
}

// writeMayRunBetween reports whether a write of the read's name may run between from and the read. One in
// the read's own code runs then only if it's there, between the two, or if that code can be entered
// again by what runs between (reentered). One in another function, a class field's initializer or a
// static block runs then only if what runs between reaches it. One at another module's top level runs
// only when that module is loaded, which a dynamic import, reaching everything, can do.
func (lowering *lowering) writeMayRunBetween(target *ast.Node, from int, read *ast.Node, reach *codeReach) bool {
	boundary, own := callBoundary(target), callBoundary(read)
	if boundary == own {
		return target.Pos() >= from && target.End() <= read.Pos() || lowering.reentered(own, reach)
	}
	if boundary == nil || boundary.Kind == ast.KindSourceFile {
		return reach.everything
	}
	return reach.everything || reach.functions[boundary]
}

// reentered reports whether the code holding a read may run again from what runs between: never a
// module's top level, a function what runs between reaches, and, when that's everything, any function
// but one only its module's top level calls (calledOnlyFromTopLevel).
func (lowering *lowering) reentered(code *ast.Node, reach *codeReach) bool {
	if code == nil || code.Kind == ast.KindSourceFile {
		return false
	}
	if reach.everything {
		return !lowering.calledOnlyFromTopLevel(code)
	}
	return reach.functions[code]
}

// calledOnlyFromTopLevel reports whether a function declaration is named nowhere in the program but
// as the direct callee of a call at a module's top level, which no call runs again, so nothing that
// runs inside it can enter it again. Judged by the name's text, so an export, an alias, a value or a
// shadowing name fails closed.
func (lowering *lowering) calledOnlyFromTopLevel(function *ast.Node) bool {
	if function.Kind != ast.KindFunctionDeclaration || function.Name() == nil {
		return false
	}
	if answer, asked := lowering.topLevelOnlyFunctions[function]; asked {
		return answer
	}
	name, only := function.Name().Text(), true
	lowering.walkProgramModules(func(node *ast.Node) {
		if !only || node.Kind != ast.KindIdentifier || node.Text() != name || node == function.Name() {
			return
		}
		call := node.Parent
		if call == nil ||
			call.Kind != ast.KindCallExpression ||
			call.AsCallExpression().Expression != node ||
			callBoundary(call) == nil ||
			callBoundary(call).Kind != ast.KindSourceFile ||
			lowering.symbol(node) != lowering.symbol(function.Name()) {
			only = false
		}
	})
	if lowering.topLevelOnlyFunctions == nil {
		lowering.topLevelOnlyFunctions = map[*ast.Node]bool{}
	}
	lowering.topLevelOnlyFunctions[function] = only
	return only
}

// propertyWrites is every write of a property in the program, by what it can reach (propertyWritesOf).
type propertyWrites struct {
	// named is the target of each write of a name, o.x or o['x'], or of each name a key's literal
	// union spells, o[key] with key "a" | "b".
	named map[string][]*ast.Node
	// everyName is a write that could reach any name: through a key that isn't a literal, a literal
	// union, a number or a symbol, a setter anywhere, or a library member that writes any name
	// (writingLibraryMembers).
	everyName bool
	// accessors is a write that could make a name a getter, on an object or its prototype: a
	// descriptor, a prototype set, a write of __proto__, or Object or Reflect handed on whole.
	accessors bool
}

// writingLibraryMembers are the library's members that can write any name of an object the program
// already has, the road a spread takes into one, each marked true when it can install a getter too
// (a descriptor, or a prototype, since one that gains get f changes what f reads). Any other library or
// host call writes no name the program spells, and a callback one calls is the program's own code,
// whose writes are counted by name. Object or Reflect passed on whole counts as an installing one,
// since what's called through it can't be told.
var writingLibraryMembers = map[string]map[string]bool{
	"Object":  {"assign": false, "defineProperty": true, "defineProperties": true, "setPrototypeOf": true},
	"Reflect": {"set": false, "defineProperty": true, "setPrototypeOf": true},
}

// propertyWritesOf collects every property write in every module the program imports, the first time
// it's asked, as writtenAcrossFunctions does for bindings. A write is any form ast.IsAssignmentTarget
// knows, on a property access or an element access: =, a compound or logical assignment, ++ or --, a
// destructuring target and a for-in or for-of head. A delete counts as one too: after o.x = null, a
// delete in a callee leaves undefined where the checker still says null.
func (lowering *lowering) propertyWritesOf() *propertyWrites {
	if lowering.propertyWrites != nil {
		return lowering.propertyWrites
	}
	lowering.propertyWrites = &propertyWrites{named: map[string][]*ast.Node{}}
	lowering.walkProgramModules(func(node *ast.Node) {
		switch {
		case (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression) &&
			ast.IsAssignmentTarget(node):
			lowering.notePropertyWrite(node)
		case node.Kind == ast.KindDeleteExpression:
			if operand := ast.SkipParentheses(
				node.AsDeleteExpression().Expression,
			); operand.Kind == ast.KindPropertyAccessExpression || operand.Kind == ast.KindElementAccessExpression {
				lowering.notePropertyWrite(operand)
			}
		case node.Kind == ast.KindSetAccessor:
			lowering.propertyWrites.everyName = true // o.f = 4 through a setter runs its body, which can write any name.
		case node.Kind == ast.KindIdentifier &&
			writingLibraryMembers[node.Text()] != nil &&
			lowering.isExpression(node) &&
			lowering.isLibraryGlobal(node, node.Text()):
			member := node.Parent
			if member.Kind != ast.KindPropertyAccessExpression ||
				member.AsPropertyAccessExpression().Expression != node {
				lowering.propertyWrites.accessors = true
				return
			}
			if installs, writes := writingLibraryMembers[node.Text()][member.Name().Text()]; writes {
				lowering.propertyWrites.everyName = true
				lowering.propertyWrites.accessors = lowering.propertyWrites.accessors || installs
			}
		}
	})
	return lowering.propertyWrites
}

// notePropertyWrite files one write of a property access or an element access by what it can reach. A
// name written through a property access or a literal key is that name, whatever the receiver's type:
// an any or a dictionary that aliases the object writes it by that name too.
func (lowering *lowering) notePropertyWrite(target *ast.Node) {
	if target.Kind == ast.KindPropertyAccessExpression {
		name := target.Name().Text()
		lowering.propertyWrites.accessors = lowering.propertyWrites.accessors || name == "__proto__"
		lowering.propertyWrites.named[name] = append(lowering.propertyWrites.named[name], target)
		return
	}
	key := ast.SkipParentheses(target.AsElementAccessExpression().ArgumentExpression)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral {
		lowering.propertyWrites.accessors = lowering.propertyWrites.accessors || key.Text() == "__proto__"
		lowering.propertyWrites.named[key.Text()] = append(lowering.propertyWrites.named[key.Text()], target)
		return
	}
	keyType := lowering.checker.GetTypeAtLocation(key)
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	for _, member := range members {
		switch {
		case member.Flags()&checker.TypeFlagsStringLiteral != 0:
			// A literal union names what it writes: drop(options, "target") deletes target alone.
			name, _ := member.AsLiteralType().Value().(string)
			lowering.propertyWrites.accessors = lowering.propertyWrites.accessors || name == "__proto__"
			lowering.propertyWrites.named[name] = append(lowering.propertyWrites.named[name], target)
		case member.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsESSymbolLike) != 0:
			// A number key writes only a name a number spells, and every read of one stops already. A
			// symbol key writes no name a read spells.
		default:
			lowering.propertyWrites.everyName = true
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
func (lowering *lowering) bindingNamed(identifier *ast.Node) *ast.Symbol {
	if identifier.Parent == nil || identifier.Parent.Kind != ast.KindShorthandPropertyAssignment {
		return lowering.symbol(identifier)
	}
	symbol := lowering.checker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	if symbol == nil {
		return nil
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = lowering.checker.GetAliasedSymbol(symbol)
	}
	return lowering.checker.GetExportSymbolOfSymbol(symbol)
}

// nullishUnit reports whether every member of t is undefined or null, and names what it holds: undefined,
// null, or undefined or null.
func (lowering *lowering) nullishUnit(t *checker.Type) (string, bool) {
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
func (lowering *lowering) writtenAcrossFunctions(symbol *ast.Symbol) bool {
	if lowering.crossFunctionWrites == nil {
		lowering.crossFunctionWrites = map[*ast.Symbol]bool{}
		lowering.walkProgramModules(func(node *ast.Node) {
			if node.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(node) {
				if written := lowering.bindingNamed(node); written != nil &&
					written.ValueDeclaration != nil &&
					callBoundary(node) != callBoundary(written.ValueDeclaration) {
					lowering.crossFunctionWrites[written] = true
				}
			}
		})
	}
	return lowering.crossFunctionWrites[symbol]
}

// walkProgramModules visits every node of every module the program imports, each module once, and not
// only the files it was given: a module's own function can write what another module reads
// (ESMModuleOrder has no effects).
func (lowering *lowering) walkProgramModules(visit func(*ast.Node)) {
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		visit(node)
		node.ForEachChild(walk)
		return false
	}
	seen := map[*ast.SourceFile]bool{}
	for _, root := range lowering.program.Files() {
		order, _ := ESMModuleOrder(lowering.checker, root)
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
		if ast.IsFunctionLike(at) ||
			at.Kind == ast.KindClassStaticBlockDeclaration ||
			at.Kind == ast.KindPropertyDeclaration ||
			at.Kind == ast.KindSourceFile {
			return at
		}
	}
	return nil
}
