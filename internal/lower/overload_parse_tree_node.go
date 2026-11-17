package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The node-test proof (Adamic's ruling on #0kkpw2e set 3): an overload promising a node parameter's own
// type, T | undefined, from an implementation typed wider, served per call. tsc's utilitiesPublic.ts
// getParseTreeNode is the shape:
//
//	function getParseTreeNode<T extends Node>(node: T | undefined, nodeTest?: (node: Node) => node is T): T | undefined;
//	function getParseTreeNode(node: Node | undefined, nodeTest?: (node: Node) => boolean): Node | undefined {
//	    if (node === undefined || isParseTreeNode(node)) {
//	        return node;
//	    }
//	    node = node.original;
//	    while (node) {
//	        if (isParseTreeNode(node)) {
//	            return !nodeTest || nodeTest(node) ? node : undefined;
//	        }
//	        node = node.original;
//	    }
//	}
//
// Every return is the node argument before anything writes it, a value the node test just said yes to,
// or undefined. That's T | undefined only where the call passes its test: with the test absent, the walk
// hands back an untested node typed T, which is a lie whenever T is narrower than the constraint. So the
// census accepts the declaration on the body, and overloadNodeTestCall judges each direct call:
//
//  1. a call passing the test, its argument's type holding no undefined, is erased, since that argument
//     goes through predicateArguments, which proves it or, in .ts, checks it as a closure;
//  2. a call whose instance of the result the implementation's serves by relation, T at its constraint,
//     is erased, since nothing narrows;
//  3. any other call is an unchecked claim and gets a cast's answer: in .ts, where the instance can be
//     checked at the call (overloadResultCheckable), its result is checked on overloadResultCheck's road,
//     and otherwise it's Refused, naming the fix.
//
// Any other entry is NotYet at certificate A's gate (overloadCertificateCall), which this proof shares.

// overloadNodeTestRoad is how overloadNodeTestCall serves one direct call.
type overloadNodeTestRoad int

const (
	overloadNodeTestErased overloadNodeTestRoad = iota
	overloadNodeTestChecked
	overloadNodeTestRefused
)

// overloadNodeTest is overload's node test: its one parameter typed, past undefined, as a function
// claiming `x is U` of one of its own parameters, where U serves overload's result, and whose
// implementation parameter is a plain name the body never assigns, so a call through that name calls
// what the caller passed. Two such parameters would leave the per-call judgment unsure which one a call
// must pass, so neither counts.
func (l *lowering) overloadNodeTest(implementation, overload *ast.Node) (predicateOverloadCallback, bool) {
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	// The overload's own types, whoever asks: a call lowered inside a generic instance has its mapper
	// installed, and the judgment must be the declaration's.
	l.typeMapper = nil
	test, found := predicateOverloadCallback{}, false
	if implementation == nil || overload == nil {
		return test, false
	}
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	served := implementation.Parameters()
	for index, parameter := range overload.Parameters() {
		if index >= len(served) || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			break // A rest takes any count of arguments, so no position past it says which is the test.
		}
		if !ast.IsIdentifier(parameter.Name()) {
			continue
		}
		// An optional test's type is the function or undefined, which has no call signature of its own.
		signatures := l.checker.GetSignaturesOfType(l.withoutUndefined(l.censusCallableParameterType(l.symbol(parameter.Name()))), checker.SignatureKindCall)
		if len(signatures) != 1 {
			continue
		}
		claim := predicateOfSignature(l.checker, signatures[0])
		if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || claim.Type() == nil || claim.ParameterIndex() < 0 || !l.censusRelated(claim.Type(), promised) {
			continue
		}
		actual := served[index]
		if !ast.IsIdentifier(actual.Name()) || actual.AsParameterDeclaration().Initializer != nil || actual.AsParameterDeclaration().DotDotDotToken != nil || l.predicateOverloadAssigned(implementation, l.symbol(actual.Name())) {
			continue
		}
		if found {
			return predicateOverloadCallback{}, false
		}
		test, found = predicateOverloadCallback{parameter: index, claimed: int(claim.ParameterIndex())}, true
	}
	return test, found
}

// overloadNodeTestProof is whether implementation serves overload's result under the node-test premise,
// that a call passes its test: every return in its body, nested functions aside, is one of
//
//	(a) a parameter itself, typed by the overload as its result, at a return in a top-level statement of
//	    the body that no earlier or same statement writes it in, and outside any loop, so it's the argument;
//	(b) C ? x : undefined, where C is test(x) or !test || test(x) through the node test, and x a name
//	    nothing the test could run writes, so the value is x only where the test said yes;
//	(c) undefined, as is falling off the end, which the result must hold.
//
// Anything else fails, and the overload stays refused at its declaration. A default would replace an
// argument the caller passed, so any parameter with one fails, as does a body reading arguments, through
// which sloppy code writes a parameter by its position.
func (l *lowering) overloadNodeTestProof(implementation, overload *ast.Node) bool {
	if implementation == nil || implementation.Kind != ast.KindFunctionDeclaration || implementation.Body() == nil || len(implementation.TypeParameters()) != 0 || ast.HasSyntacticModifier(implementation, ast.ModifierFlagsAsync) || generatorDeclaration(implementation) {
		return false
	}
	test, found := l.overloadNodeTest(implementation, overload)
	if !found {
		return false
	}
	outerMapper := l.typeMapper
	defer func() { l.typeMapper = outerMapper }()
	l.typeMapper = nil
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	if !l.censusRelated(l.checker.GetUndefinedType(), promised) {
		return false
	}
	parameters, declared := implementation.Parameters(), overload.Parameters()
	for _, parameter := range parameters {
		declaration := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || declaration.Initializer != nil || declaration.DotDotDotToken != nil {
			return false
		}
	}
	parameterOf := func(node *ast.Node) int {
		node = ast.SkipParentheses(node)
		if !ast.IsIdentifier(node) {
			return -1
		}
		for index, parameter := range parameters {
			if l.symbol(node) == l.symbol(parameter.Name()) {
				return index
			}
		}
		return -1
	}
	undefined := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return ast.IsIdentifier(node) && l.symbol(node) == l.checker.GetUndefinedSymbol()
	}

	// writtenAt is, for each parameter, the first top-level statement that writes it anywhere inside, or
	// one past the last when none does. A write in a nested function counts before the first statement,
	// since a hoisted declaration can be called before its own statement runs.
	statements := implementation.Body().AsBlock().Statements.Nodes
	writtenAt := make([]int, len(parameters))
	for index := range writtenAt {
		writtenAt[index] = len(statements)
	}
	sloppy := false
	var writes func(node *ast.Node, position int, nested bool)
	writes = func(node *ast.Node, position int, nested bool) {
		if ast.IsIdentifier(node) && node.Text() == "arguments" {
			sloppy = true
		}
		if ast.IsIdentifier(node) && ast.IsAssignmentTarget(node) {
			if index := parameterOf(node); index >= 0 {
				at := position
				if nested {
					at = -1
				}
				writtenAt[index] = min(writtenAt[index], at)
			}
		}
		inner := nested || ast.IsFunctionLike(node)
		node.ForEachChild(func(child *ast.Node) bool {
			writes(child, position, inner)
			return false
		})
	}
	for position, statement := range statements {
		writes(statement, position, false)
	}
	if sloppy {
		return false
	}

	returned := func(statement *ast.Node, position int, looped bool) bool {
		expression := statement.AsReturnStatement().Expression
		if expression == nil || undefined(expression) {
			return true // (c)
		}
		expression = ast.SkipParentheses(expression)
		if index := parameterOf(expression); index >= 0 {
			if looped {
				return false // A loop's return runs after the loop's own later statements, a write among them.
			}
			if writtenAt[index] <= position {
				return false // Written first, it's no longer the argument the overload typed.
			}
			given, _, rest, err := l.censusOverloadParameter(declared, index)
			return err == nil && !rest && given != nil && l.censusRelated(given, promised) // (a)
		}
		if expression.Kind != ast.KindConditionalExpression {
			return false
		}
		conditional := expression.AsConditionalExpression()
		value := ast.SkipParentheses(conditional.WhenTrue)
		if !ast.IsIdentifier(value) || !undefined(conditional.WhenFalse) {
			return false
		}
		// The test runs program code between reading x and the conditional's value, so x must be a
		// parameter or local no nested function writes (predicateOverloadBinding).
		local := l.symbol(value)
		if local == nil || len(local.Declarations) != 1 || !l.predicateOverloadBinding(implementation, local) {
			return false
		}
		condition := ast.SkipParentheses(conditional.Condition)
		if condition.Kind == ast.KindBinaryExpression && condition.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken {
			// !test || test(x): under the premise the test is a function, so !test is false.
			absent := ast.SkipParentheses(condition.AsBinaryExpression().Left)
			if absent.Kind != ast.KindPrefixUnaryExpression || absent.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken || parameterOf(absent.AsPrefixUnaryExpression().Operand) != test.parameter {
				return false
			}
			condition = ast.SkipParentheses(condition.AsBinaryExpression().Right)
		}
		return condition.Kind == ast.KindCallExpression && l.predicateOverloadCallTests(implementation, condition, []predicateOverloadCallback{test}, local) // (b)
	}
	proven := true
	var returns func(node *ast.Node, position int, looped bool)
	returns = func(node *ast.Node, position int, looped bool) {
		if !proven || ast.IsFunctionLike(node) {
			return // A nested function's return is its own result, not this call's.
		}
		if node.Kind == ast.KindReturnStatement && !returned(node, position, looped) {
			proven = false
			return
		}
		inner := looped || ast.IsIterationStatement(node, false)
		node.ForEachChild(func(child *ast.Node) bool {
			returns(child, position, inner)
			return false
		})
	}
	for position, statement := range statements {
		returns(statement, position, false)
	}
	return proven
}

// overloadNodeTestCertified is whether censusOverload accepted overload by the node-test proof, judged
// once per overload as overloadArrayCertified judges certificate A: an overload not yet judged, when its
// body passes the proof, is judged by running its set's census, which records the acceptance.
func (l *lowering) overloadNodeTestCertified(overload *ast.Node) (bool, error) {
	if overload == nil || overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil || overload.Name() == nil {
		return false, nil
	}
	if certified, judged := l.overloadNodeTests[overload]; judged {
		return certified, nil
	}
	certified := false
	if implementation := l.censusImplementation(overload); implementation != nil && l.overloadNodeTestProof(implementation, overload) {
		if err := l.censusOverloads(implementation); err != nil {
			return false, err
		}
		certified = l.overloadNodeTests[overload]
	}
	l.noteOverloadNodeTest(overload, certified)
	return certified, nil
}

// noteOverloadNodeTest records whether census accepted overload by the node-test proof.
func (l *lowering) noteOverloadNodeTest(overload *ast.Node, certified bool) {
	if l.overloadNodeTests == nil {
		l.overloadNodeTests = map[*ast.Node]bool{}
	}
	l.overloadNodeTests[overload] = certified
}

// overloadNodeTestRoadOf is the road a direct call, node, resolved to an overload the node-test proof
// accepted, takes by the ruling's three cases.
func (l *lowering) overloadNodeTestRoadOf(node *ast.Node, resolved *checker.Signature) overloadNodeTestRoad {
	overload := resolved.Declaration()
	implementation := l.censusImplementation(overload)
	test, found := l.overloadNodeTest(implementation, overload)
	if !found {
		return overloadNodeTestRefused
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if test.parameter < len(arguments) && arguments[test.parameter].Kind != ast.KindSpreadElement && !l.censusHasUndefined(l.concrete(l.checker.GetTypeAtLocation(arguments[test.parameter]))) {
		return overloadNodeTestErased // 1: the test is passed, and predicateArguments holds it to its claim.
	}
	instance := l.concrete(l.checker.GetReturnTypeOfSignature(resolved))
	produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
	if l.censusRelated(produced, instance) {
		return overloadNodeTestErased // 2: the instance narrows nothing the implementation's result doesn't hold.
	}
	// 3: a .a file never holds a hidden check (rulings.md:127), so there it's Refused.
	if l.checkedAssertionSource(implementation) && l.checkedAssertionSource(overload) && l.overloadResultCheckable(instance, map[*checker.Type]bool{}) {
		return overloadNodeTestChecked
	}
	return overloadNodeTestRefused
}

// overloadNodeTestCall is nil unless node, a direct call resolved to an overload the node-test proof
// accepted, is an unchecked claim whose instance can't be checked at the call, which is Refused there.
// overloadCertificateCall asks it once the call's entry is direct, before the callee or an argument is
// lowered.
func (l *lowering) overloadNodeTestCall(node *ast.Node, resolved *checker.Signature) error {
	// The proof is judged by now, so a checked call is noted for checkOverloadResults even where the road
	// noted it before the census ran.
	l.noteOverloadCall(node)
	if l.overloadNodeTestRoadOf(node, resolved) != overloadNodeTestRefused {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "a call to an overload whose result is honest only for a proven node test, made without one", Fix: "pass the node test that proves the result's type"}
}

// overloadNodeTestChecked is whether node, a call resolved to resolved, is one whose result the node-test
// proof leaves to overloadResultCheck (case 3, checkable). It reads only what census recorded, since
// narrowOverloadCall asks it after lowering too; overloadCertificateCall records it before any road
// lowers the call.
func (l *lowering) overloadNodeTestChecked(node *ast.Node, resolved *checker.Signature) bool {
	overload := resolved.Declaration()
	if overload == nil || !l.overloadNodeTests[overload] || !l.overloadCertificateEntry(node, overload) {
		return false
	}
	return l.overloadNodeTestRoadOf(node, resolved) == overloadNodeTestChecked
}
