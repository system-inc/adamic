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
	// The census overload proof refuses overload_union_impl too today ("cannot be served by
	// implementation result"), so its fixture goes red through that refusal without this rule: the two
	// guard one claim independently, and this one stays if the census proof is relaxed (Compiler, Oct 30).
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
	if resolved == nil {
		return call, nil
	}
	claim := predicateOfSignature(l.checker, resolved)
	// tsc makes one signature for a union or an intersection of callables and gives it one member's
	// declaration, so neither a check nor a vouch read from that declaration speaks for the others. The
	// call stands only when every member's narrowing signature is vouched for on its own, as two proven
	// predicates are (no_escape.ts's c41), and none is a claim this head would check.
	if callee := l.checker.GetTypeAtLocation(node.AsCallExpression().Expression); claim != nil && callee.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		if !l.predicateCalleeMembersVouched(callee) {
			return nil, l.notYet(node, "a call narrowing through a union or intersection of callables")
		}
		return call, nil
	}
	if resolved.Declaration() == nil || !l.checkedFunctionTypeClaim(resolved.Declaration().Type()) {
		// Admitting these claims means a call can reach lowering through any signature that narrows: one
		// tsc makes for a union or an intersection, or one with no declaration. Only a claim something
		// stands behind goes unchecked (Lower's must 2).
		if claim != nil && !l.predicateCallVouched(resolved.Declaration(), claim) {
			return nil, l.notYet(node, "a call through a predicate signature nothing checks or proves")
		}
		return call, nil
	}
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
		// JavaScript reads a.m before it evaluates the arguments, and the check reads m off the receiver
		// after them, so an argument that runs code (a write to a.m, a getter) would call a different
		// function from Node's. Only arguments that run nothing keep the order (Lower's must 1).
		if !predicateArgumentsRunNothing(call.Arguments) {
			return nil, l.notYet(node, "a checked method claim whose arguments can run code before the method is read")
		}
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

// predicateCallVouched is whether a call that narrows by claim, through declaration, has something behind
// it besides a check at the call: a library's or the prelude's claim; one predicateParameter trusts
// because every producer is proven; a checked closure, which checks itself; a checked declaration,
// whose value is its checked forwarder; or a body that proves it, the annotation's or, for a claim tsc
// inferred, the proof predicateArguments asks of an inferred argument. No declaration vouches for nothing.
func (l *lowering) predicateCallVouched(declaration *ast.Node, claim *checker.TypePredicate) bool {
	if declaration == nil {
		return false
	}
	file := ast.GetSourceFileOfNode(declaration)
	if load.IsLibrary(file) || load.IsPrelude(file) {
		return true
	}
	annotation := declaration.Type()
	if annotation == nil || annotation.Kind != ast.KindTypePredicate {
		return claim.Type() != nil && ast.IsFunctionLike(declaration) && declaration.Body() != nil && l.proveInferredPredicate(declaration, claim.Type())
	}
	if l.predicateParameter(annotation) != nil || l.checkedPredicateClosure(declaration) || l.checkedPredicateDeclaration(declaration) {
		return true
	}
	if _, err := l.provePredicate(annotation); err == nil {
		return true
	}
	return l.proveFlowPredicate(annotation) == nil
}

// predicateCalleeMembersVouched is whether every member of a union or an intersection of callables
// that narrows is vouched for by predicateCallVouched, with no checked function-type claim among them.
func (l *lowering) predicateCalleeMembersVouched(callee *checker.Type) bool {
	for _, member := range callee.Types() {
		for _, signature := range l.checker.GetSignaturesOfType(member, checker.SignatureKindCall) {
			claim := predicateOfSignature(l.checker, signature)
			if claim == nil {
				continue
			}
			declaration := signature.Declaration()
			if declaration == nil || l.checkedFunctionTypeClaim(declaration.Type()) || !l.predicateCallVouched(declaration, claim) {
				return false
			}
		}
	}
	return true
}

// predicateArgumentsRunNothing is whether evaluating arguments can run no code of the program's, so a
// read before them and one after them see the same function: names, constants, operators over them and
// the closures written in place. A field read isn't among them: through a structural type it can reach
// a class's getter at run time (getter_order.ts, Lower's read of f8c25636). Anything else, a call, an
// allocation that runs a constructor or a library operation, is taken to run code.
func predicateArgumentsRunNothing(arguments []ir.Expression) bool {
	quiet := true
	for _, argument := range arguments {
		walk(argument, func(node any) bool {
			switch node.(type) {
			case ir.Read, ir.NumberConstant, ir.StringConstant, ir.BooleanConstant, ir.Undefined, ir.IsUndefined, ir.Unary, ir.Binary, ir.Unwrap, ir.Defined, ir.MaybeOf, ir.Box, ir.Narrow, ir.TypeOf, ir.MakeClosure:
			default:
				quiet = false
			}
			return quiet
		})
	}
	return quiet
}
