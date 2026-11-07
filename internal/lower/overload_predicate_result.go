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
func (lowering *lowering) predicateOverloadCallbacks(
	implementation, overload *ast.Node,
	promised *checker.Type,
) []predicateOverloadCallback {
	callbacks := []predicateOverloadCallback{}
	served := implementation.Parameters()
	for index, parameter := range overload.Parameters() {
		if index >= len(served) || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			break
		}
		signatures := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeAtLocation(parameter),
			checker.SignatureKindCall,
		)
		if len(signatures) != 1 {
			continue
		}
		claim := predicateOfSignature(lowering.checker, signatures[0])
		if claim == nil ||
			claim.Kind() != checker.TypePredicateKindIdentifier ||
			claim.Type() == nil ||
			claim.ParameterIndex() < 0 {
			continue
		}
		if !lowering.censusRelated(claim.Type(), promised) {
			continue
		}
		actual := served[index]
		if !ast.IsIdentifier(actual.Name()) ||
			actual.AsParameterDeclaration().Initializer != nil ||
			actual.AsParameterDeclaration().DotDotDotToken != nil {
			continue
		}
		if lowering.predicateOverloadAssigned(implementation, lowering.symbol(actual.Name())) {
			continue
		}
		callbacks = append(callbacks, predicateOverloadCallback{parameter: index, claimed: int(claim.ParameterIndex())})
	}
	return callbacks
}

// predicateOverloadAssigned is whether anything in implementation's body, nested functions
// included, assigns symbol: a closure that rebinds the callback rebinds it for the body too.
func (lowering *lowering) predicateOverloadAssigned(implementation *ast.Node, symbol *ast.Symbol) bool {
	assigned := false
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(node) && lowering.symbol(node) == symbol {
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
func (lowering *lowering) predicateOverloadResultProven(implementation, overload *ast.Node) (bool, *ast.Node) {
	if !lowering.checkedAssertionSource(implementation) || implementation.Body() == nil {
		return false, nil
	}
	// An async or generator body's return statements don't produce the call's result.
	// generatorDeclaration reads a method's asterisk too: an overloaded method's implementation reaches
	// here, and it isn't a function declaration.
	if ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) || generatorDeclaration(implementation) {
		return false, nil
	}
	promised := lowering.checker.GetReturnTypeOfSignature(lowering.checker.GetSignatureFromDeclaration(overload))
	callbacks := lowering.predicateOverloadCallbacks(implementation, overload, promised)
	if len(callbacks) == 0 {
		return false, nil
	}
	var failed *ast.Node
	var returns ast.Visitor
	returns = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false // A nested function's return is its own result, not this call's.
		}
		if node.Kind == ast.KindReturnStatement &&
			!lowering.predicateOverloadReturn(implementation, node, callbacks, promised) {
			failed = node
			return true
		}
		return node.ForEachChild(returns)
	}
	if implementation.Body().ForEachChild(returns) {
		return false, failed
	}
	if !lowering.censusRelated(lowering.checker.GetUndefinedType(), promised) {
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
// serves promised: a value whose type already does, or the const its callback's claim just said yes
// to (predicateOverloadResultProven).
func (lowering *lowering) predicateOverloadReturn(
	implementation, statement *ast.Node,
	callbacks []predicateOverloadCallback,
	promised *checker.Type,
) bool {
	expression := statement.AsReturnStatement().Expression
	if expression == nil {
		return false
	}
	if lowering.censusRelated(lowering.checker.GetTypeAtLocation(expression), promised) {
		return true
	}
	returned := ast.SkipParentheses(expression)
	if !ast.IsIdentifier(returned) {
		return false
	}
	local := lowering.symbol(returned)
	if local == nil || len(local.Declarations) != 1 {
		return false
	}
	declaration := local.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration ||
		declaration.Parent == nil ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 ||
		ast.GetContainingFunction(declaration) != implementation {
		return false
	}
	// The return is the then-branch itself, or the first statement of the block that is.
	then := statement
	if statement.Parent.Kind == ast.KindBlock {
		if block := statement.Parent.AsBlock(); len(block.Statements.Nodes) == 0 ||
			block.Statements.Nodes[0] != statement {
			return false
		}
		then = statement.Parent
	}
	if then.Parent == nil ||
		then.Parent.Kind != ast.KindIfStatement ||
		then.Parent.AsIfStatement().ThenStatement != then {
		return false
	}
	condition := ast.SkipParentheses(then.Parent.AsIfStatement().Expression)
	if condition.Kind != ast.KindCallExpression {
		return false
	}
	test := condition.AsCallExpression()
	callee := ast.SkipParentheses(test.Expression)
	if !ast.IsIdentifier(callee) {
		return false
	}
	served := implementation.Parameters()
	for _, callback := range callbacks {
		if lowering.symbol(callee) != lowering.symbol(served[callback.parameter].Name()) ||
			callback.claimed >= len(test.Arguments.Nodes) {
			continue
		}
		// A spread before the claimed argument moves which value the claim names.
		shifted := false
		for _, argument := range test.Arguments.Nodes[:callback.claimed+1] {
			shifted = shifted || argument.Kind == ast.KindSpreadElement
		}
		argument := ast.SkipParentheses(test.Arguments.Nodes[callback.claimed])
		if !shifted && ast.IsIdentifier(argument) && lowering.symbol(argument) == local {
			return true
		}
	}
	return false
}
