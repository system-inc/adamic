package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A claim on a function type has no body to prove it: `test: (node: Node) => node is T` on a
// parameter, a field or a return, or `isLateBound(node: Declaration): node is LateBoundDeclaration` on an
// interface. In .ts it's a checked claim (#1w5bve2, agreed Oct 30 inside Adamic's closed-world rule: a
// call checks the claim of the signature it went through). Every call that resolves to the signature is
// a checked call: the claim's true side is checked at the call, against the target the call resolved
// to, so a lying body stops there whichever one runs. The false side is the walk's
// (predicate_false_flow.go), which a program holding such a claim turns on. A target with no runtime
// test is NotYet at the call, never a check that passes vacuously.
//
// A library's claim (lib.d.ts's filter, find and every) is called inside the library, where nothing is
// checked, so it isn't one: an argument for its parameter keeps predicateArguments' proof, a forward of
// a checked parameter included (Lower's must). A claim on this, and one on an abstract method, whose
// calls don't go through a function value, stay refused. .a needs proof, and keeps the refusal.

// checkedFunctionTypeClaim is whether annotation is a checked claim on a function type: in a .ts file of
// the program, on a function type, an interface method or a call signature, naming a parameter, and not
// one predicateParameter trusts because every producer of it is proven.
func (l *lowering) checkedFunctionTypeClaim(annotation *ast.Node) bool {
	if annotation == nil || annotation.Kind != ast.KindTypePredicate || annotation.AsTypePredicateNode().ParameterName.Kind != ast.KindIdentifier {
		return false
	}
	signature := annotation.Parent
	if signature == nil || signature.Type() != annotation {
		return false
	}
	switch signature.Kind {
	case ast.KindFunctionType, ast.KindMethodSignature, ast.KindCallSignature:
	default:
		return false
	}
	file := ast.GetSourceFileOfNode(annotation)
	if load.IsLibrary(file) || load.IsPrelude(file) || !l.checkedAssertionSource(annotation) {
		return false
	}
	if l.predicateParameter(annotation) != nil {
		return false
	}
	// An overload's parameter is called in its implementation, through the implementation's own
	// parameter type, and the overload's result leans on the answers (takeWhile's U[]). So it's checked
	// only when that parameter carries a checked claim too, which checks every call the body makes.
	if parameter := signature.Parent; parameter.Kind == ast.KindParameter && parameter.Parent.Kind == ast.KindFunctionDeclaration && parameter.Parent.Body() == nil {
		implementation := l.censusImplementation(parameter.Parent)
		if implementation == nil {
			return false
		}
		for index, candidate := range parameter.Parent.Parameters() {
			if candidate != parameter {
				continue
			}
			if index >= len(implementation.Parameters()) {
				return false
			}
			implemented := implementation.Parameters()[index].Type()
			return implemented != nil && implemented.Kind == ast.KindFunctionType && l.checkedFunctionTypeClaim(implemented.Type())
		}
		return false
	}
	return true
}

// checkedFunctionTypeCall is call, a call through a function value, checked against the claim of the
// signature it resolved to when that's a checked function-type claim. The callee, or a method's
// receiver, and the arguments are evaluated once, at the call site, into the check's parameters, so the
// claim reads the value the call received.
func (l *lowering) checkedFunctionTypeCall(node *ast.Node, call ir.CallClosure) (ir.Expression, error) {
	resolved := l.checker.GetResolvedSignature(node)
	if resolved == nil || resolved.Declaration() == nil || !l.checkedFunctionTypeClaim(resolved.Declaration().Type()) {
		return call, nil
	}
	claim := predicateOfSignature(l.checker, resolved)
	if claim == nil || (claim.Kind() != checker.TypePredicateKindIdentifier && claim.Kind() != checker.TypePredicateKindAssertsIdentifier) {
		return nil, l.notYet(node, "a checked function-type predicate without an identifier claim")
	}
	assertion := claim.Kind() == checker.TypePredicateKindAssertsIdentifier
	if claim.Type() == nil && !assertion {
		return nil, l.notYet(node, "a checked predicate without a reifiable target")
	}
	if len(call.Spread) != 0 {
		return nil, l.notYet(node, "spread arguments to a checked function-type predicate")
	}
	index := int(claim.ParameterIndex())
	if index < 0 || index >= len(call.Arguments) || index >= len(resolved.Parameters()) {
		return nil, l.notYet(node, "a checked predicate on a missing argument")
	}
	if (!assertion && call.Type() != ir.Boolean) || (assertion && call.Type() != 0) {
		return nil, l.notYet(node, "a checked predicate with a non-boolean result")
	}
	expression := node.AsCallExpression()
	directions := l.checkedPredicateDirections(expression, claim)
	if !assertion {
		if use := l.checkedPredicateUnclassifiedUse(expression); use != "" {
			return nil, l.notYet(node, "a checked predicate result "+use+", which its direction walk doesn't read")
		}
	}
	name := sourceExpression(expression.Expression)
	sourceName := l.checker.TypeToString(l.checker.GetTypeAtLocation(expression.Arguments.Nodes[index]))
	targetName := "truthy"
	if claim.Type() != nil {
		targetName = l.checker.TypeToString(claim.Type())
	}
	site := ir.PredicateCallCheck{Where: l.program.Where(node), Function: name}
	l.recordPredicateSite(&site, assertion, false, directions, sourceName+" -> "+targetName)
	wrapper := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "function_type_predicate_check", Returns: call.Type()})
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: wrapper})
		return index
	}
	// A method through an interface keeps its receiver, so the check takes the receiver and reads the
	// method off it; any other callee is the function value itself.
	var callee ir.Expression
	var checked ir.CallClosure = call
	if property, isMethod := call.Closure.(ir.Property); isMethod && property.Method {
		receiver := local("receiver", property.Object.Type())
		callee = property.Object
		property.Object = ir.Read{Local: receiver, Of: property.Object.Type()}
		checked.Closure = property
		l.result.Functions[wrapper].Parameters = append(l.result.Functions[wrapper].Parameters, receiver)
	} else {
		function := local("function", ir.Closure)
		callee = call.Closure
		checked.Closure = ir.Read{Local: function, Of: ir.Closure}
		l.result.Functions[wrapper].Parameters = append(l.result.Functions[wrapper].Parameters, function)
	}
	arguments := []ir.Expression{}
	for position, argument := range call.Arguments {
		if argument.Type() == 0 {
			return nil, l.notYet(node, "a checked function-type predicate passed an argument without a representation")
		}
		parameter := local(fmt.Sprintf("argument_%d", position), argument.Type())
		l.result.Functions[wrapper].Parameters = append(l.result.Functions[wrapper].Parameters, parameter)
		arguments = append(arguments, ir.Read{Local: parameter, Of: argument.Type()})
	}
	checked.Arguments = arguments
	body := []ir.Statement{}
	var returned ir.Expression
	if assertion {
		body = append(body, ir.Evaluate{Value: checked})
	} else {
		result := local("result", ir.Boolean)
		body = append(body, ir.Declare{Local: result, Value: checked})
		returned = ir.Read{Local: result, Of: ir.Boolean}
	}
	source := l.checker.GetTypeOfSymbol(resolved.Parameters()[index])
	check, err := l.checkedPredicateCheck(node, node, arguments[index], source, claim.Type(), returned, directions&predicateFalse != 0, local, func(branch string) string {
		return fmt.Sprintf("predicate %s at %s %s branch: source %s, target %s", name, site.Where, branch, sourceName, targetName)
	})
	if err != nil {
		return nil, err
	}
	body = append(body, check...)
	body = append(body, ir.Return{Value: returned})
	l.result.Functions[wrapper].Body = body
	return ir.Call{Function: wrapper, Arguments: append([]ir.Expression{callee}, call.Arguments...), Returns: call.Type()}, nil
}
