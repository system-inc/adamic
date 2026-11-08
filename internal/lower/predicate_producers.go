package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A producer is a function that returns a predicate it builds from the predicates it's handed, as
// core.ts's orPredicate2 returns `p => f1(p) || f2(p)`. tsc passes one where a predicate is trusted rather
// than checked, `findAncestor(node, orPredicate2(isImportDeclaration, isExportDeclaration))` at
// checker.ts:4767 and six sites like it, and lib's `.find` at 10911: findAncestor's implementation calls
// its callback through a plain boolean signature (utilitiesPublic.ts:786-788), and the library calls its
// own, so nothing checks the claim there and predicateArguments wants it proven or checked. A call of a
// producer stands there when the producer's every return is a checked closure under the producer's own
// written claim (checkedPredicateClaimNode), accepted as a checked closure passed directly is, and every
// predicate the call hands it is accepted as it would be passed directly, judged at that call's instance,
// a fail-closed extra (Adamic's ruling on #fx9waqy, Nov 7, road (B)).
//
// Inside a producer, a call through one of its parameters keeps its check whoever calls the producer
// (checkedFunctionTypeClaim): its claim is never taken as proven by collecting the predicates passed at
// its call sites, since one argument carried in a value breaks that collection (Language's condition).

// predicateProducerProven is whether call, an argument for a parameter whose claim is target at this
// call, is a call of a proven producer: a direct call, by its own name, of a function the program declares
// once in .ts with a body, never assigned, whose parameters are never assigned; whose every return is a
// checked closure under its written claim, checked per instance at each return; whose result at this
// call claims exactly target;
// and whose every claimed parameter is handed an argument that proves its claim at this call
// (predicateArgumentProof, a nested producer call included). Anything else fails closed.
func (l *lowering) predicateProducerProven(call *ast.Node, target *checker.Type) bool {
	expression := call.AsCallExpression()
	callee := ast.SkipParentheses(expression.Expression)
	if callee.Kind != ast.KindIdentifier || expression.QuestionDotToken != nil {
		return false
	}
	symbol := l.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || l.assignedSymbol(symbol) {
		return false
	}
	producer := symbol.Declarations[0]
	returned := l.predicateProducerReturns(producer)
	if returned == nil {
		return false
	}
	for _, function := range returned {
		if !l.checkedPredicateClosure(function) {
			return false
		}
	}
	resolved := l.checker.GetResolvedSignature(call)
	if resolved == nil {
		return false
	}
	results := l.checker.GetSignaturesOfType(l.checker.GetReturnTypeOfSignature(resolved), checker.SignatureKindCall)
	if len(results) != 1 {
		return false
	}
	claim := predicateOfSignature(l.checker, results[0])
	if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || claim.Type() == nil || !checker.Checker_isTypeIdenticalTo(l.checker, claim.Type(), target) {
		return false
	}
	arguments := expression.Arguments.Nodes
	for _, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			return false
		}
	}
	for index, parameter := range producer.Parameters() {
		contract := predicateProducerContract(parameter)
		if contract == nil {
			continue
		}
		if index >= len(arguments) || index >= len(resolved.Parameters()) {
			return false
		}
		if refused := l.predicateArgumentProof(resolved.Parameters()[index], contract, arguments[index]); refused != nil {
			return false
		}
	}
	return true
}

// predicateProducerReturns is every function producer returns, when producer is a producer: a function
// declaration the program writes in .ts, with a body and a result written as a function type whose claim
// names a parameter and a type, no parameter that's a rest, has a default or is ever assigned, at least one
// return, and every return of its own, a nested function's aside, an arrow or a function expression, in
// parentheses or not. Anything else is nil.
func (l *lowering) predicateProducerReturns(producer *ast.Node) []*ast.Node {
	if producer.Kind != ast.KindFunctionDeclaration || producer.Body() == nil || producer.Type() == nil {
		return nil
	}
	file := ast.GetSourceFileOfNode(producer)
	if load.IsLibrary(file) || load.IsPrelude(file) || !l.checkedAssertionSource(producer) {
		return nil
	}
	written := ast.SkipTypeParentheses(producer.Type())
	if written.Kind != ast.KindFunctionType || written.Type() == nil || written.Type().Kind != ast.KindTypePredicate {
		return nil
	}
	if claim := written.Type().AsTypePredicateNode(); claim.AssertsModifier != nil || claim.Type == nil || claim.ParameterName == nil || claim.ParameterName.Kind != ast.KindIdentifier {
		return nil
	}
	for _, parameter := range producer.Parameters() {
		declaration := parameter.AsParameterDeclaration()
		if declaration.DotDotDotToken != nil || declaration.Initializer != nil || parameter.Name().Kind != ast.KindIdentifier || l.assignedSymbol(l.symbol(parameter.Name())) {
			return nil
		}
	}
	return producerReturnedFunctions(producer)
}

// producerReturnedFunctions is every function producer returns when each of its own returns, a nested
// function's aside, is an arrow or a function expression written in place, through parentheses only, and
// there's at least one: an arrow's expression body is its one return. Anything else, a return of a name
// that holds the function or of any other value, is nil, since the function could then be held, stored or
// passed before it's returned (Adamic's ruling on #fx9waqy, Nov 7).
func producerReturnedFunctions(producer *ast.Node) []*ast.Node {
	body := producer.Body()
	if body == nil {
		return nil
	}
	written := func(value *ast.Node) *ast.Node {
		if value == nil {
			return nil
		}
		value = ast.SkipParentheses(value)
		if value.Kind != ast.KindArrowFunction && value.Kind != ast.KindFunctionExpression {
			return nil
		}
		return value
	}
	if body.Kind != ast.KindBlock {
		if value := written(body); value != nil {
			return []*ast.Node{value}
		}
		return nil
	}
	returned := []*ast.Node{}
	whole := true
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if !whole {
			return true
		}
		if ast.IsFunctionLike(n) {
			return false
		}
		if n.Kind == ast.KindReturnStatement {
			value := written(n.AsReturnStatement().Expression)
			if value == nil {
				whole = false
				return true
			}
			returned = append(returned, value)
			return false
		}
		return n.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	if !whole || len(returned) == 0 {
		return nil
	}
	return returned
}

// checkedPredicateClaimNode is the claim a checked closure checks at each return: its own annotation when
// it has one, else the claim its producer's written result carries when function is a value that producer
// returns in place (Adamic's road (B) on #fx9waqy, Nov 7): an unnamed arrow or function expression,
// through parentheses only, from a producer whose every return is one (producerReturnedFunctions) and
// whose own written result, never an overload's, is a function type with a claim checkedFunctionTypeClaim
// admits, as orPredicate2's `(p: P) => p is R1 | R2`. The arrow is then what tsc would have written by
// hand, `(p): p is R1 | R2 => f1(p) || f2(p)`: a checked closure, instantiated per instance, so a trusted
// site accepts the producer's result as it accepts a checked closure passed directly, and nothing new is
// proven. A named function expression stays out, since its name is a binding its own body can store it
// through (`function self(p) { cache.push(self); ... }`). Anything else is nil.
func (l *lowering) checkedPredicateClaimNode(function *ast.Node) *ast.Node {
	if annotation := function.Type(); annotation != nil {
		if annotation.Kind != ast.KindTypePredicate {
			return nil
		}
		return annotation
	}
	if function.Kind != ast.KindArrowFunction && function.Kind != ast.KindFunctionExpression || function.Name() != nil {
		return nil
	}
	value := function
	for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
		value = value.Parent
	}
	producer := brandedChainReturner(value)
	if producer == nil || producer.Type() == nil {
		return nil
	}
	written := ast.SkipTypeParentheses(producer.Type())
	if written.Kind != ast.KindFunctionType || !l.checkedFunctionTypeClaim(written.Type()) {
		return nil
	}
	for _, returned := range producerReturnedFunctions(producer) {
		if returned == function {
			return written.Type()
		}
	}
	return nil
}

// predicateProducerContract is the claim a producer's parameter is written with, `f1: (p1: P) => p1 is
// R1`, or nil for a parameter written with none.
func predicateProducerContract(parameter *ast.Node) *ast.Node {
	written := parameter.Type()
	if written == nil {
		return nil
	}
	written = ast.SkipTypeParentheses(written)
	if written.Kind != ast.KindFunctionType || written.Type() == nil || written.Type().Kind != ast.KindTypePredicate {
		return nil
	}
	return written.Type()
}

// predicateProducerParameter is whether annotation is the claim of a producer's parameter, whose calls
// inside the producer keep their check unconditionally (Language's condition on #fx9waqy).
func (l *lowering) predicateProducerParameter(annotation *ast.Node) bool {
	signature := annotation.Parent
	if signature == nil || signature.Kind != ast.KindFunctionType || signature.Parent == nil || signature.Parent.Kind != ast.KindParameter {
		return false
	}
	return l.predicateProducerReturns(signature.Parent.Parent) != nil
}
