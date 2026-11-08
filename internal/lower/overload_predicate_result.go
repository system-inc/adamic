package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A predicate overload promises a narrower result than its implementation produces, as tsc's
// core.ts find does: U | undefined from a body typed T | undefined. The narrowing is the callback's
// claim, so it holds in .ts on two halves (#2qvx0r2, Adamic's and Compiler's ruling). The body
// returns only what already serves the promise or the very value its callback just said yes to
// (predicateOverloadResultProven). Every call passes a callback whose claim is proven where it's
// passed: predicateArguments demands a named function or arrow proving both branches for a predicate
// parameter nothing checks, which an overload's is when its implementation takes a plain boolean
// callback (checkedFunctionTypeClaim). In .a a predicate value is refused, so the overload is too.

// predicateOverloadCallback is a parameter of an overload typed as a function claiming `x is U`:
// its index, and the index of the callback's own parameter the claim names.
type predicateOverloadCallback struct {
	parameter int
	claimed   int
}

// predicateOverloadCallbacks is each of overload's parameters typed with an identifier type
// predicate whose U serves promised, and whose implementation parameter is a plain name the body
// never assigns, so a call through that name in the body calls the callback the caller passed. A
// default would replace a missing callback with one no caller's road checked, and a rest would move
// the callback's position, so neither counts.
func (l *lowering) predicateOverloadCallbacks(implementation, overload *ast.Node, promised *checker.Type) []predicateOverloadCallback {
	callbacks := []predicateOverloadCallback{}
	served := implementation.Parameters()
	for index, parameter := range overload.Parameters() {
		if index >= len(served) || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			break
		}
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(parameter), checker.SignatureKindCall)
		if len(signatures) != 1 {
			continue
		}
		claim := predicateOfSignature(l.checker, signatures[0])
		if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || claim.Type() == nil || claim.ParameterIndex() < 0 {
			continue
		}
		if !l.censusRelated(claim.Type(), promised) {
			continue
		}
		actual := served[index]
		if !ast.IsIdentifier(actual.Name()) || actual.AsParameterDeclaration().Initializer != nil || actual.AsParameterDeclaration().DotDotDotToken != nil {
			continue
		}
		if l.predicateOverloadAssigned(implementation, l.symbol(actual.Name())) {
			continue
		}
		callbacks = append(callbacks, predicateOverloadCallback{parameter: index, claimed: int(claim.ParameterIndex())})
	}
	return callbacks
}

// predicateOverloadAssigned is whether anything in implementation's body, nested functions
// included, assigns symbol: a closure that rebinds the callback rebinds it for the body too.
func (l *lowering) predicateOverloadAssigned(implementation *ast.Node, symbol *ast.Symbol) bool {
	assigned := false
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(node) && l.symbol(node) == symbol {
			assigned = true
		}
		return assigned || node.ForEachChild(writes)
	}
	implementation.Body().ForEachChild(writes)
	return assigned
}

// predicateOverloadResultProven is whether implementation serves overload's narrower result in .ts
// because each return in its body, nested functions aside, either already serves the promise (a
// return undefined) or returns the value a callback's claim was just checked on. That second kind
// is Compiler's four conditions: return v, where v is a const declared in the body, so nothing writes
// it between the call and the return; the return opens the then-branch of an if whose condition is
// the call itself, so the claim's yes is what reaches it; the call is through the callback's own
// parameter with v as the argument the claim names; and the overload is the predicate one, whose
// claim's U serves the promise. A body whose end can be reached returns undefined there, so undefined
// must serve the promise unless the body ends in a return or a throw. It also gives the first return
// that fails, for the refusal to name, or nil when what failed isn't a return.
func (l *lowering) predicateOverloadResultProven(implementation, overload *ast.Node) (bool, *ast.Node) {
	if !l.checkedAssertionSource(implementation) || implementation.Body() == nil {
		return false, nil
	}
	// An async or generator body's return statements don't produce the call's result.
	// generatorDeclaration reads a method's asterisk too: an overloaded method's implementation reaches
	// here, and it isn't a function declaration.
	if ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) || generatorDeclaration(implementation) {
		return false, nil
	}
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	callbacks := l.predicateOverloadCallbacks(implementation, overload, promised)
	if len(callbacks) == 0 {
		return false, nil
	}
	var failed *ast.Node
	var returns ast.Visitor
	returns = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false // A nested function's return is its own result, not this call's.
		}
		if node.Kind == ast.KindReturnStatement && !l.predicateOverloadReturn(implementation, node, callbacks, promised) {
			failed = node
			return true
		}
		return node.ForEachChild(returns)
	}
	if implementation.Body().ForEachChild(returns) {
		return false, failed
	}
	if !l.censusRelated(l.checker.GetUndefinedType(), promised) {
		statements := implementation.Body().AsBlock().Statements.Nodes
		if len(statements) == 0 {
			return false, nil
		}
		last := statements[len(statements)-1]
		if last.Kind != ast.KindReturnStatement && last.Kind != ast.KindThrowStatement {
			return false, nil
		}
	}
	return true, nil
}

// predicateOverloadReturn is whether statement, a return in implementation's body, returns what
// serves promised: a value whose type already does, or the binding its callback's claim just said yes to
// (predicateOverloadResultProven, predicateOverloadResultTested).
func (l *lowering) predicateOverloadReturn(implementation, statement *ast.Node, callbacks []predicateOverloadCallback, promised *checker.Type) bool {
	expression := statement.AsReturnStatement().Expression
	if expression == nil {
		return false
	}
	if l.predicateOverloadUndefinedServes(expression, promised) || l.censusRelated(l.checker.GetTypeAtLocation(expression), promised) {
		return true
	}
	returned := ast.SkipParentheses(expression)
	if !ast.IsIdentifier(returned) {
		return false
	}
	local := l.symbol(returned)
	if local == nil || len(local.Declarations) != 1 {
		return false
	}
	// The return is the then-branch itself, or the first statement of the block that is.
	then := statement
	if statement.Parent.Kind == ast.KindBlock {
		if block := statement.Parent.AsBlock(); len(block.Statements.Nodes) == 0 || block.Statements.Nodes[0] != statement {
			return false
		}
		then = statement.Parent
	}
	if then.Parent == nil || then.Parent.Kind != ast.KindIfStatement || then.Parent.AsIfStatement().ThenStatement != then {
		return false
	}
	branch := then.Parent
	condition := ast.SkipParentheses(branch.AsIfStatement().Expression)
	switch {
	case condition.Kind == ast.KindCallExpression:
		// The call is the condition, so its yes is what reaches the return. A const declared in the body
		// can't be written between the call and the return.
		declaration := local.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || ast.GetContainingFunction(declaration) != implementation {
			return false
		}
		return l.predicateOverloadCallTests(implementation, condition, callbacks, local)
	case ast.IsIdentifier(condition):
		return false
	}
	return false
}

// predicateOverloadCallTests is whether call is a call through one of callbacks' own parameters whose
// claimed argument is local, with no spread before it, so the claim's yes is about local.
func (l *lowering) predicateOverloadCallTests(implementation, call *ast.Node, callbacks []predicateOverloadCallback, local *ast.Symbol) bool {
	test := call.AsCallExpression()
	callee := ast.SkipParentheses(test.Expression)
	if !ast.IsIdentifier(callee) {
		return false
	}
	served := implementation.Parameters()
	for _, callback := range callbacks {
		if l.symbol(callee) != l.symbol(served[callback.parameter].Name()) || callback.claimed >= len(test.Arguments.Nodes) {
			continue
		}
		// A spread before the claimed argument moves which value the claim names.
		shifted := false
		for _, argument := range test.Arguments.Nodes[:callback.claimed+1] {
			shifted = shifted || argument.Kind == ast.KindSpreadElement
		}
		argument := ast.SkipParentheses(test.Arguments.Nodes[callback.claimed])
		if !shifted && ast.IsIdentifier(argument) && l.symbol(argument) == local {
			return true
		}
	}
	return false
}

// predicateOverloadResultTested is the road tsc's findAncestor (utilitiesPublic.ts:786-799) takes, where
// the callback's answer is held and tested before the return:
//
//	const result = callback(node);
//	if (result === "quit") { return undefined; }
//	else if (result) { return node; }
//	node = node.parent;
//
// #2qvx0r2's rule applied through flow (Adamic's ruling on #fx9waqy, Nov 8, with Language's conditions on
// #45abpv9). The return of local stands only when the one definition of local reaching it is the
// argument of the callback call whose result is tested, which the structure below makes so:
//
//   - The tested value is the callback call's own result, its only definition: a const declared once in
//     the body, with that call as its initializer (Language's condition 1). A let fails closed.
//   - That call's claimed argument is local by symbol, never by name, through callbacks' own parameter
//     (Language's condition 2). A copy, `const n = node; return n`, fails closed, since copies aren't
//     followed.
//   - Nothing between the call and the return can run program code or write local: the const's statement
//     is the one right before the if chain, every condition before the tested one compares locals and
//     literals by === or !== alone, every argument of the call is a name or a literal, and no function
//     nested in the body assigns local, since the callback could call it. So a reassignment of local,
//     findAncestor's `node = node.parent`, is only ever on the path after the callback's no.
//   - Every answer but the claim's yes is excluded first: the checker's own flow type of the tested name
//     at the test is boolean, so `"quit"` was ruled out before it (a body testing `if (result)` with
//     `"quit"` still possible keeps the refusal).
//   - local is a parameter or a local of the implementation, declared once: unknown reach fails closed.
//
// The claim is the callback's, checked where it's made (rulings.md:139, :140): predicateArguments makes
// every value reaching the predicate overload's parameter prove or check its own claim, which is what
// this return stands on, since the body calls it through the implementation's plain type and checks
// nothing there.
func (l *lowering) predicateOverloadResultTested(implementation, branch, condition *ast.Node, callbacks []predicateOverloadCallback, local *ast.Symbol) bool {
	if !l.predicateOverloadBinding(implementation, local) {
		return false
	}
	result := l.symbol(condition)
	if result == nil || len(result.Declarations) != 1 {
		return false
	}
	declaration := result.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || ast.GetContainingFunction(declaration) != implementation {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil || ast.SkipParentheses(initializer).Kind != ast.KindCallExpression {
		return false
	}
	call := ast.SkipParentheses(initializer)
	list := declaration.Parent
	statement := list.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement || len(list.AsVariableDeclarationList().Declarations.Nodes) != 1 {
		return false
	}
	top := branch
	for top.Parent != nil && top.Parent.Kind == ast.KindIfStatement && top.Parent.AsIfStatement().ElseStatement == top {
		if !predicateOverloadQuiet(top.Parent.AsIfStatement().Expression) {
			return false
		}
		top = top.Parent
	}
	if top.Parent == nil || top.Parent.Kind != ast.KindBlock {
		return false
	}
	adjacent := false
	statements := top.Parent.AsBlock().Statements.Nodes
	for position, candidate := range statements {
		if candidate == top {
			adjacent = position > 0 && statements[position-1] == statement
		}
	}
	if !adjacent {
		return false
	}
	answer := l.checker.GetTypeAtLocation(condition)
	answers := []*checker.Type{answer}
	if answer.Flags()&checker.TypeFlagsUnion != 0 {
		answers = answer.Types()
	}
	for _, member := range answers {
		if member.Flags()&checker.TypeFlagsBooleanLike == 0 {
			return false
		}
	}
	for _, argument := range call.AsCallExpression().Arguments.Nodes {
		// A name or a literal, never an operator over them: an argument is a value handed on, not a test.
		if kind := ast.SkipParentheses(argument).Kind; !predicateOverloadQuiet(argument) || kind == ast.KindBinaryExpression || kind == ast.KindPrefixUnaryExpression {
			return false
		}
	}
	return l.predicateOverloadCallTests(implementation, call, callbacks, local)
}

// predicateOverloadBinding is whether local is one of implementation's parameters, a plain name with no
// default and no rest, or a local declared once in its body, and no function nested in the body assigns
// it, so a callback that calls one can't write it between a call and a return.
func (l *lowering) predicateOverloadBinding(implementation *ast.Node, local *ast.Symbol) bool {
	declaration := local.Declarations[0]
	switch declaration.Kind {
	case ast.KindParameter:
		parameter := declaration.AsParameterDeclaration()
		if declaration.Parent != implementation || !ast.IsIdentifier(declaration.Name()) || parameter.Initializer != nil || parameter.DotDotDotToken != nil {
			return false
		}
	case ast.KindVariableDeclaration:
		if ast.GetContainingFunction(declaration) != implementation || !ast.IsIdentifier(declaration.Name()) {
			return false
		}
	default:
		return false
	}
	written := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if written {
			return true
		}
		if ast.IsFunctionLike(node) {
			var writes ast.Visitor
			writes = func(inner *ast.Node) bool {
				if inner.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(inner) && l.symbol(inner) == local {
					written = true
				}
				return written || inner.ForEachChild(writes)
			}
			node.ForEachChild(writes)
			return written
		}
		return node.ForEachChild(visit)
	}
	implementation.Body().ForEachChild(visit)
	return !written
}

// predicateOverloadQuiet is whether expression runs no program code and writes nothing: a name, a
// literal, or a strict comparison or negation of those, through parentheses. A loose == can call an
// object's valueOf, and a property read can run a getter, so neither is quiet.
func predicateOverloadQuiet(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindNoSubstitutionTemplateLiteral:
		return true
	case ast.KindPrefixUnaryExpression:
		unary := expression.AsPrefixUnaryExpression()
		return unary.Operator == ast.KindExclamationToken && predicateOverloadQuiet(unary.Operand)
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		return (operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) && predicateOverloadQuiet(binary.Left) && predicateOverloadQuiet(binary.Right)
	}
	return false
}

// predicateOverloadUndefinedServes is whether expression is the value undefined, written as the name
// `undefined` the checker resolves to its own symbol or as `void` of a literal, and promised, as the
// overload writes it, holds undefined as one of its members: findAncestor's `return undefined` in its
// "quit" arm serving `T | undefined` (#448a51k). That needs no claim and no representation of T, so it
// holds whatever T is mapped to where the overload is judged: censusRelated concretes the union, and in
// the census's own unit for findAncestor's overload it refused this return (Correctness's census of
// 5a86e3b3, utilitiesPublic.ts:792).
func (l *lowering) predicateOverloadUndefinedServes(expression *ast.Node, promised *checker.Type) bool {
	value := ast.SkipParentheses(expression)
	switch {
	case value.Kind == ast.KindIdentifier:
		if value.Text() != "undefined" || l.checker.GetSymbolAtLocation(value) != l.checker.GetUndefinedSymbol() {
			return false
		}
	case value.Kind == ast.KindVoidExpression:
		operand := ast.SkipParentheses(value.AsVoidExpression().Expression)
		if operand.Kind != ast.KindNumericLiteral && operand.Kind != ast.KindStringLiteral {
			return false
		}
	default:
		return false
	}
	members := []*checker.Type{promised}
	if promised.Flags()&checker.TypeFlagsUnion != 0 {
		members = promised.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			return true
		}
	}
	return false
}
