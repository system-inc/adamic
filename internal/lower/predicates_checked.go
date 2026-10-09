package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Independent body proofs retire predicate checks. Tag summaries still hand
// the complete target to the shared checked-view builder for remaining fields.
func (l *lowering) checkedPredicateBodyProven(annotation *ast.Node) bool {
	if annotation == nil || annotation.Kind != ast.KindTypePredicate {
		return false
	}
	if l.proveFlowPredicate(annotation) == nil {
		return true
	}
	proof, err := l.provePredicate(annotation)
	if err != nil {
		return false
	}
	if proof.TaggedView {
		return l.predicateRefusal(annotation) == nil
	}
	return true
}

// Callback producers may retain the shared field views required by a tag proof.
func (l *lowering) checkedPredicateArgumentProven(annotation *ast.Node) bool {
	if l.proveFlowPredicate(annotation) == nil {
		return true
	}
	if _, err := l.provePredicate(annotation); err != nil {
		return false
	}
	return l.predicateRefusal(annotation) == nil
}

func (l *lowering) checkedPredicateDirections(call *ast.CallExpression, claim *checker.TypePredicate) predicateTruth {
	if claim.Kind() == checker.TypePredicateKindAssertsIdentifier {
		return predicateTrue
	}
	directions := predicateTrue
	index := int(claim.ParameterIndex())
	if index < 0 || index >= len(call.Arguments.Nodes) {
		return directions
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[index])
	if argument.Kind != ast.KindIdentifier && argument.Kind != ast.KindPropertyAccessExpression && argument.Kind != ast.KindElementAccessExpression && argument.Kind != ast.KindThisKeyword {
		return directions
	}
	source := l.checker.GetTypeAtLocation(argument)
	// Ask the pinned checker whether this call's false condition excludes a part
	// of the incoming type. The annotation supplies a claim, never body evidence.
	flow := &ast.FlowNode{Flags: ast.FlowFlagsFalseCondition, Node: call.AsNode(), Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
	narrowed := predicateFlowType(l.checker, argument, source, source, nil, flow)
	if checker.Checker_isTypeIdenticalTo(l.checker, source, narrowed) {
		return directions
	}
	// Empty branches still count. Read-based flow discovery also covers boolean
	// aliases, captures, and continuation paths after an early return.
	if l.predicateUseDirections(call)&predicateFalse != 0 {
		return predicateEither
	}
	conditionNarrows := func(condition *ast.Node, truth bool) bool {
		if condition == nil || l.predicateConditionDirection(condition, call.AsNode(), truth, 0)&predicateFalse == 0 {
			return false
		}
		flags := ast.FlowFlagsFalseCondition
		if truth {
			flags = ast.FlowFlagsTrueCondition
		}
		branch := &ast.FlowNode{Flags: flags, Node: condition, Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
		narrowed := predicateFlowType(l.checker, argument, source, source, nil, branch)
		return !checker.Checker_isTypeIdenticalTo(l.checker, source, narrowed)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var condition *ast.Node
		switch node.Kind {
		case ast.KindIfStatement:
			condition = node.AsIfStatement().Expression
		case ast.KindConditionalExpression:
			condition = node.AsConditionalExpression().Condition
		case ast.KindWhileStatement:
			condition = node.AsWhileStatement().Expression
		case ast.KindDoStatement:
			condition = node.AsDoStatement().Expression
		case ast.KindForStatement:
			condition = node.AsForStatement().Condition
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindBarBarToken && conditionNarrows(binary.Left, false) {
				directions |= predicateFalse
			}
		}
		if conditionNarrows(condition, true) || conditionNarrows(condition, false) {
			directions |= predicateFalse
		}
		node.ForEachChild(visit)
		return false
	}
	ast.GetSourceFileOfNode(call.AsNode()).AsNode().ForEachChild(visit)
	return directions
}

// A tag is a full membership test only when the admitted source has a
// complete discriminant partition. Open structural domains need a negative
// view-membership operation, which this train does not provide.
func (l *lowering) checkedPredicateTagDomain(source, target *checker.Type) bool {
	return l.checkedPredicateTagDomainSeen(source, target, map[[2]*checker.Type]bool{})
}
func (l *lowering) checkedPredicateTagDomainSeen(source, target *checker.Type, seen map[[2]*checker.Type]bool) bool {
	source, target = l.concrete(source), l.concrete(target)
	key := [2]*checker.Type{source, target}
	if seen[key] {
		return false
	}
	seen[key] = true
	defer delete(seen, key)
	if l.checker.IsArrayType(target) {
		members := []*checker.Type{source}
		if source.Flags()&checker.TypeFlagsUnion != 0 {
			members = source.Types()
		}
		for _, member := range members {
			if !l.checker.IsArrayType(member) {
				continue
			}
			from, to := l.checker.GetElementTypeOfArrayType(member), l.checker.GetElementTypeOfArrayType(target)
			if !checker.Checker_isTypeIdenticalTo(l.checker, from, to) && !l.checkedPredicateTagDomainSeen(from, to, seen) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !l.checkedPredicateTagDomainSeen(source, member, seen) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) || l.checker.IsArrayType(target) {
		return true
	}
	tag := l.fieldLiteral(target, "kind")
	if tag == nil {
		return true
	} // Existing reification names untagged gaps.
	wanted, ok := predicateLiteral(tag)
	if !ok {
		return false
	}
	members := []*checker.Type{source}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		members = source.Types()
	}
	for _, member := range members {
		held := l.fieldLiteral(member, "kind")
		if held == nil {
			return false
		}
		key, ok := predicateLiteral(held)
		if !ok {
			return false
		}
		if key == wanted && !l.checker.IsTypeAssignableTo(member, target) {
			return false
		}
	}
	return true
}

func (l *lowering) checkedPredicateResult(call *ast.CallExpression, value ir.Expression, implementation, declaration *ast.Node) (ir.Expression, error) {
	resolved := l.checker.GetResolvedSignature(call.AsNode())
	if resolved == nil {
		return value, nil
	}
	if declaration == nil {
		declaration = resolved.Declaration()
	}
	// Inferred predicates still use their existing body proof admission path.
	if declaration == nil || declaration.Type() == nil || declaration.Type().Kind != ast.KindTypePredicate {
		return value, nil
	}
	if declaration.Kind != ast.KindFunctionDeclaration {
		if l.checkedPredicateBodyProven(declaration.Type()) || l.predicateParameter(declaration.Type()) != nil {
			return value, nil
		}
		return nil, l.notYet(call.AsNode(), "a checked predicate without a named implementation")
	}
	if implementation == nil {
		implementation = declaration
	}
	if !l.checkedAssertionSource(declaration) {
		return value, nil
	}
	claim := predicateOfSignature(l.checker, resolved)
	if claim == nil || claim.Type() == nil {
		return nil, l.notYet(call.AsNode(), "a checked predicate without a reifiable target")
	}
	assertion := claim.Kind() == checker.TypePredicateKindAssertsIdentifier
	index := int(claim.ParameterIndex())
	directions := l.checkedPredicateDirections(call, claim)
	proven := l.checkedPredicateBodyProven(declaration.Type())
	if declaration.Body() == nil {
		proven = l.predicateOverloadProven(implementation, declaration)
	}
	name := implementation.Name().Text()
	site := ir.PredicateCallCheck{Where: l.program.Where(call.AsNode()), Function: name}
	if declaration.Body() == nil {
		site.Overload = l.predicateOverloadOrdinal(implementation, declaration)
	}
	sourceName := "missing argument"
	if index >= 0 && index < len(call.Arguments.Nodes) {
		sourceName = l.checker.TypeToString(l.checker.GetTypeAtLocation(call.Arguments.Nodes[index]))
	}
	targetName := l.checker.TypeToString(claim.Type())
	for _, direction := range []predicateTruth{predicateTrue, predicateFalse} {
		if assertion && direction == predicateFalse {
			continue
		}
		branch := "true"
		if direction == predicateFalse {
			branch = "false"
		}
		if assertion {
			branch = "asserts"
		}
		report := ir.PredicateDirectionCheck{Direction: branch, Reason: sourceName + " -> " + targetName}
		switch {
		case proven:
			report.Status = "proven"
			l.result.PredicateChecks.Proven++
		case directions&direction != 0:
			report.Status = "checked"
			l.result.PredicateChecks.Checked++
		default:
			report.Status = "unobservable"
			l.result.PredicateChecks.Unobservable++
		}
		site.Directions = append(site.Directions, report)
	}
	l.result.PredicateChecks.Sites = append(l.result.PredicateChecks.Sites, site)
	if proven {
		return value, nil
	}
	invoked, ok := value.(ir.Call)
	if !ok {
		return nil, l.notYet(call.AsNode(), "an indirect call of a checked predicate")
	}
	if index < 0 || index >= len(invoked.Arguments) {
		return nil, l.notYet(call.AsNode(), "a checked predicate on a missing argument")
	}
	if (!assertion && invoked.Type() != ir.Boolean) || (assertion && invoked.Type() != 0) {
		return nil, l.notYet(call.AsNode(), "a checked predicate with a non-boolean result")
	}
	wrapper := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: fmt.Sprintf("%s_predicate_check_%d", name, wrapper), Returns: invoked.Type()})
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: wrapper})
		return index
	}
	arguments := []ir.Expression{}
	for i, argument := range invoked.Arguments {
		if undefined, ok := argument.(ir.Undefined); ok && i < len(l.result.Functions[l.result.CallTargets(invoked)[0]].Parameters) {
			parameter := l.result.Functions[l.result.CallTargets(invoked)[0]].Parameters[i]
			undefined.Of = l.result.Locals[parameter].Type
			argument = undefined
			invoked.Arguments[i] = argument
		}
		parameter := local(fmt.Sprintf("argument_%d", i), argument.Type())
		l.result.Functions[wrapper].Parameters = append(l.result.Functions[wrapper].Parameters, parameter)
		arguments = append(arguments, ir.Read{Local: parameter, Of: argument.Type()})
	}
	checked := invoked
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
	if directions&predicateFalse != 0 && !l.checkedPredicateTagDomain(source, claim.Type()) {
		return nil, l.notYet(call.AsNode(), "checked predicate membership over an open tag domain")
	}
	setup, membership, err := l.predicateMembership(declaration, arguments[index], source, claim.Type(), local, 0)
	if err != nil {
		return nil, err
	}
	failure := func(branch string, condition ir.Expression) ir.Statement {
		message := fmt.Sprintf("predicate %s at %s %s branch: source %s, target %s", name, site.Where, branch, sourceName, targetName)
		return ir.If{Condition: condition, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}}
	}
	positive := append([]ir.Statement{}, setup...)
	branch := "true"
	if assertion {
		branch = "asserts"
	}
	positive = append(positive, failure(branch, ir.Unary{Operator: ir.Not, Operand: membership}))
	if assertion {
		body = append(body, positive...)
	} else {
		check := ir.If{Condition: returned, Then: positive}
		if directions&predicateFalse != 0 {
			check.Else = append(append([]ir.Statement{}, setup...), failure("false", membership))
		}
		body = append(body, check)
	}
	body = append(body, ir.Return{Value: returned})
	l.result.Functions[wrapper].Body = body
	return ir.Call{Function: wrapper, Arguments: invoked.Arguments, Returns: invoked.Type()}, nil
}
