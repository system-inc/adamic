package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only independent whole-type body proofs retire checks. A tag-only summary
// does not prove an open interface's other fields.
func (l *lowering) checkedPredicateBodyProven(annotation *ast.Node) bool {
	if annotation == nil || annotation.Kind != ast.KindTypePredicate {
		return false
	}
	if l.proveFlowPredicate(annotation) == nil {
		return true
	}
	proof, err := l.provePredicate(annotation)
	return err == nil && !proof.TaggedView
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
		case ast.KindSwitchStatement:
			// Under switch (true), every case clause's false branch reaches the clauses after it.
			statement := node.AsSwitchStatement()
			if ast.SkipParentheses(statement.Expression).Kind == ast.KindTrueKeyword {
				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
					if clause.Kind == ast.KindCaseClause && conditionNarrows(clause.Expression(), false) {
						directions |= predicateFalse
					}
				}
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

// checkedPredicateUnclassifiedUse names where the call's result goes when the direction walk
// can't say what the checker narrows there, "" when it can. Climbing through what the checker's
// narrowType reads through (parentheses, !, && and ||, a comparison with a boolean literal, an
// assignment's or a comma's right side), it stops at the first other parent. A condition the walk
// reads, a const alias, or a discarded value is read. Any other parent is a value the walk reads
// no direction from, so it must not stand in a condition or an alias, and must be one the checker
// narrows nothing by; anything else is named, so an unread narrowing stops as NotYet instead of
// going unchecked.
func (l *lowering) checkedPredicateUnclassifiedUse(call *ast.CallExpression) string {
	// A function whose returns the checker infers a predicate from narrows its callers by this
	// result, and a call of an inferred predicate is never checked.
	if function := ast.GetContainingFunction(call.AsNode()); function != nil && function.Type() == nil {
		if signature := l.checker.GetSignatureFromDeclaration(function); signature != nil && predicateOfSignature(l.checker, signature) != nil {
			return "returned from a function the checker infers a type predicate for"
		}
	}
	node, parent := call.AsNode(), call.AsNode().Parent
	for ; parent != nil && checkedPredicateTransparent(parent, node); node, parent = parent, parent.Parent {
	}
	if parent == nil || checkedPredicateCondition(parent, node) || checkedPredicateAlias(parent, node) || parent.Kind == ast.KindExpressionStatement || parent.Kind == ast.KindVoidExpression {
		return ""
	}
	for child, holder := parent, parent.Parent; holder != nil && !ast.IsFunctionLike(holder); child, holder = holder, holder.Parent {
		if checkedPredicateCondition(holder, child) || checkedPredicateAlias(holder, child) {
			return "read by " + describe(parent) + " inside a condition"
		}
	}
	switch parent.Kind {
	case ast.KindBinaryExpression, ast.KindPrefixUnaryExpression, ast.KindConditionalExpression, ast.KindSwitchStatement, ast.KindCaseClause, ast.KindVariableDeclaration,
		ast.KindTypeOfExpression, ast.KindAwaitExpression, ast.KindAsExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindTemplateSpan, ast.KindArrayLiteralExpression,
		ast.KindPropertyAssignment, ast.KindSpreadElement, ast.KindReturnStatement, ast.KindArrowFunction, ast.KindThrowStatement, ast.KindPropertyDeclaration, ast.KindParameter,
		ast.KindElementAccessExpression, ast.KindForStatement:
		return ""
	}
	return "read by " + describe(parent)
}

// checkedPredicateTransparent is whether the checker's narrowType reads through parent to child.
func checkedPredicateTransparent(parent, child *ast.Node) bool {
	switch parent.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			return true
		case ast.KindEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindCommaToken:
			return child == binary.Right
		case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			other := binary.Left
			if child == binary.Left {
				other = binary.Right
			}
			return ast.IsBooleanLiteral(ast.SkipParentheses(other)) && !ast.IsAccessExpression(ast.SkipParentheses(child))
		}
	}
	return false
}

// checkedPredicateCondition is whether child is a condition the binder gives flow conditions to:
// a statement's or a conditional's test, an operand of && or ||, or a switch (true) case.
func checkedPredicateCondition(holder, child *ast.Node) bool {
	switch holder.Kind {
	case ast.KindIfStatement:
		return holder.AsIfStatement().Expression == child
	case ast.KindWhileStatement:
		return holder.AsWhileStatement().Expression == child
	case ast.KindDoStatement:
		return holder.AsDoStatement().Expression == child
	case ast.KindForStatement:
		return holder.AsForStatement().Condition == child
	case ast.KindConditionalExpression:
		return holder.AsConditionalExpression().Condition == child
	case ast.KindBinaryExpression:
		operator := holder.AsBinaryExpression().OperatorToken.Kind
		return operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken
	case ast.KindCaseClause:
		statement := holder.Parent.Parent
		return holder.Expression() == child && statement.Kind == ast.KindSwitchStatement && ast.SkipParentheses(statement.AsSwitchStatement().Expression).Kind == ast.KindTrueKeyword
	}
	return false
}

// checkedPredicateAlias is whether child initializes a const, which the checker narrows through
// wherever the const stands in a condition.
func checkedPredicateAlias(holder, child *ast.Node) bool {
	return holder.Kind == ast.KindVariableDeclaration && holder.AsVariableDeclaration().Initializer == child && holder.Parent.Flags&ast.NodeFlagsConst != 0
}

// A tag is a full membership test only when the admitted source has a
// complete discriminant partition. Open structural domains need a negative
// view-membership operation, which this train does not provide.
func (l *lowering) checkedPredicateTagDomain(source, target *checker.Type) bool {
	return l.checkedPredicateTagDomainSeen(source, target, false, map[[2]*checker.Type]bool{})
}

// The true check proves the layout only when no source member outside the
// target can hold the wanted tag (#drwm5pj): two members of one enum sharing a
// value are one type to the checker, and a field an open numeric enum types,
// like tsc's kind: SyntaxKind, can hold any of its values. Untagged members and
// a plain kind: number stay the existing reification's gap.
func (l *lowering) checkedPredicateTagProvesLayout(source, target *checker.Type) bool {
	return l.checkedPredicateTagDomainSeen(source, target, true, map[[2]*checker.Type]bool{})
}

func (l *lowering) checkedPredicateTagDomainSeen(source, target *checker.Type, positive bool, seen map[[2]*checker.Type]bool) bool {
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
			if !checker.Checker_isTypeIdenticalTo(l.checker, from, to) && !l.checkedPredicateTagDomainSeen(from, to, positive, seen) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !l.checkedPredicateTagDomainSeen(source, member, positive, seen) {
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
		if held == nil && positive {
			if field := l.checker.GetPropertyOfType(member, "kind"); member.Flags()&checker.TypeFlagsPrimitive == 0 && field != nil && l.openNumericEnumType(l.checker.GetTypeOfSymbol(field)) && l.checker.IsTypeAssignableTo(tag, l.checker.GetTypeOfSymbol(field)) && !l.checker.IsTypeAssignableTo(member, target) {
				return false
			}
			continue
		}
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

// A direct predicate call names the declaration itself, through a bare name or a
// namespace/module binding. A structural value with the same member is not that
// binding; its predicate keeps the check carried by functionValue's forwarder.
func (l *lowering) directPredicateCallee(callee, declaration *ast.Node) bool {
	if declaration == nil || declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil {
		return false
	}
	callee = ast.SkipParentheses(callee)
	switch callee.Kind {
	case ast.KindIdentifier:
	case ast.KindPropertyAccessExpression:
		object := l.symbol(ast.SkipParentheses(callee.Expression()))
		if object == nil || object.Flags&ast.SymbolFlagsModule == 0 || object.Flags&(ast.SymbolFlagsValue & ^ast.SymbolFlagsModule) != 0 {
			return false
		}
	default:
		return false
	}
	// symbol follows import and re-export aliases with GetAliasedSymbol and maps
	// the local declaration to its export symbol, on both sides of the comparison.
	symbol := l.symbol(callee)
	return symbol != nil && symbol == l.symbol(declaration.Name())
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
	if !assertion {
		if use := l.checkedPredicateUnclassifiedUse(call); use != "" {
			return nil, l.notYet(call.AsNode(), "a checked predicate result "+use+", which its direction walk doesn't read")
		}
	}
	invoked, ok := value.(ir.Call)
	if !ok || !l.directPredicateCallee(call.Expression, declaration) {
		return nil, l.notYet(call.AsNode(), "an indirect call of a checked predicate")
	}
	if index < 0 || index >= len(invoked.Arguments) {
		return nil, l.notYet(call.AsNode(), "a checked predicate on a missing argument")
	}
	if (!assertion && invoked.Type() != ir.Boolean) || (assertion && invoked.Type() != 0) {
		return nil, l.notYet(call.AsNode(), "a checked predicate with a non-boolean result")
	}
	wrapper := len(l.result.Functions)
	// Same-named predicates may share this IR label: native keys include the
	// semantic body and resolved callee, so the label must not encode encounter order.
	l.result.Functions = append(l.result.Functions, ir.Function{Name: fmt.Sprintf("%s_predicate_check", name), Returns: invoked.Type()})
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
	check, err := l.checkedPredicateCheck(call.AsNode(), declaration, arguments[index], source, claim.Type(), returned, directions&predicateFalse != 0, local, func(branch string) string {
		return fmt.Sprintf("predicate %s at %s %s branch: source %s, target %s", name, site.Where, branch, sourceName, targetName)
	})
	if err != nil {
		return nil, err
	}
	body = append(body, check...)
	body = append(body, ir.Return{Value: returned})
	l.result.Functions[wrapper].Body = body
	return ir.Call{Function: wrapper, Arguments: invoked.Arguments, Returns: invoked.Type()}, nil
}

// checkedPredicateCheck is the one check a checked predicate's claim gets, wherever it's checked: at
// a direct call (checkedPredicateResult) and in the value a predicate becomes (functionValue). result
// is the predicate's boolean, or nil for an asserts predicate, whose claim holds once it returns.
// argument is the value the claim names. The true side is always checked, and the false side only
// where falseToo says a caller narrows on it.
func (l *lowering) checkedPredicateCheck(node, declaration *ast.Node, argument ir.Expression, source, target *checker.Type, result ir.Expression, falseToo bool, local func(string, ir.Type) int, message func(branch string) string) ([]ir.Statement, error) {
	if falseToo && !l.checkedPredicateTagDomain(source, target) {
		return nil, l.notYet(node, "checked predicate membership over an open tag domain")
	}
	// enum_tag_flow.go's hazards hold here too: a singleton open enum tag reads as a literal. A tag
	// another member can hold proves no layout, so on the true branch the target is a checked view:
	// predicateMembership registers its fields, as a .ts cast does, and a lying tag stops at the first
	// field read (#k1t2wp9). A false branch has no view to read through, so it stays pending. The view
	// is keyed by the target's type across the program, so a read narrowed by a call through a
	// predicate's value reads through it too.
	if (l.enumTagHazardOf(source, "kind") != enumTagNoHazard || !l.checkedPredicateTagProvesLayout(source, target)) && falseToo {
		return nil, l.notYet(node, "checked predicate membership by an enum tag another member can hold")
	}
	setup, membership, err := l.predicateMembership(declaration, argument, source, target, local, 0)
	if err != nil {
		return nil, err
	}
	failure := func(branch string, condition ir.Expression) ir.Statement {
		return ir.If{Condition: condition, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message(branch))}}}}
	}
	if result == nil {
		return append(append([]ir.Statement{}, setup...), failure("asserts", ir.Unary{Operator: ir.Not, Operand: membership})), nil
	}
	check := ir.If{Condition: result, Then: append(append([]ir.Statement{}, setup...), failure("true", ir.Unary{Operator: ir.Not, Operand: membership}))}
	if falseToo {
		check.Else = append(append([]ir.Statement{}, setup...), failure("false", membership))
	}
	return []ir.Statement{check}, nil
}

// checkedPredicateDeclaration is whether declaration is a predicate R-T80 checks at a direct call: a
// named .ts function claiming a predicate its body doesn't prove, with a body or an implementation.
// The direct call checks it, the value it becomes checks it (functionValue), and a program where it
// escapes as a value walks its false side (predicate_false_flow.go).
func (l *lowering) checkedPredicateDeclaration(declaration *ast.Node) bool {
	return declaration != nil && declaration.Kind == ast.KindFunctionDeclaration && declaration.Type() != nil && declaration.Type().Kind == ast.KindTypePredicate && (declaration.Body() != nil || l.censusImplementation(declaration) != nil) && l.checkedAssertionSource(declaration) && !l.checkedPredicateBodyProven(declaration.Type())
}

// checkedPredicateForwarder is the body of the value a checked predicate becomes (#vdt9f1q, ruled
// Oct 25): functionValue's one forwarder, which every road to the value goes through, so the check
// travels with it into an argument, a field, an alias or a return. It checks the predicate's own
// claim, its true side or an asserts on return, at every call through the value; the false side is
// the walk's (predicate_false_flow.go). forwarder is the forwarder's index, already in the result,
// and call is its call of the predicate. A target that isn't a checked predicate gets no body here.
// It's no call site, so PredicateChecks, which counts the calls R-T80 checks, doesn't count it.
func (l *lowering) checkedPredicateForwarder(node *ast.Node, symbol *ast.Symbol, forwarder int, call ir.Call) ([]ir.Statement, bool, error) {
	var declaration *ast.Node
	if symbol != nil {
		for _, candidate := range symbol.Declarations {
			if candidate.Body() != nil && l.checkedPredicateDeclaration(candidate) {
				declaration = candidate
			}
		}
	}
	if declaration == nil {
		return nil, false, nil
	}
	if len(declaration.TypeParameters()) != 0 {
		return nil, false, l.notYet(node, "a checked generic predicate used as a value")
	}
	signature := l.checker.GetSignatureFromDeclaration(declaration)
	claim := predicateOfSignature(l.checker, signature)
	if claim == nil || claim.Type() == nil {
		return nil, false, l.notYet(node, "a checked predicate without a reifiable target")
	}
	assertion := claim.Kind() == checker.TypePredicateKindAssertsIdentifier
	index := int(claim.ParameterIndex())
	if index < 0 || index >= len(call.Arguments) || index >= len(signature.Parameters()) {
		return nil, false, l.notYet(node, "a checked predicate on a missing argument")
	}
	if (!assertion && call.Returns != ir.Boolean) || (assertion && call.Returns != 0) {
		return nil, false, l.notYet(node, "a checked predicate with a non-boolean result")
	}
	name := declaration.Name().Text()
	where := l.program.Where(declaration)
	source := l.checker.GetTypeOfSymbol(signature.Parameters()[index])
	sourceName, targetName := l.checker.TypeToString(source), l.checker.TypeToString(claim.Type())
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: forwarder})
		return index
	}
	body := []ir.Statement{}
	var result ir.Expression
	if assertion {
		body = append(body, ir.Evaluate{Value: call})
	} else {
		held := local("result", ir.Boolean)
		body = append(body, ir.Declare{Local: held, Value: call})
		result = ir.Read{Local: held, Of: ir.Boolean}
	}
	check, err := l.checkedPredicateCheck(node, declaration, call.Arguments[index], source, claim.Type(), result, false, local, func(branch string) string {
		return fmt.Sprintf("predicate %s used as a value, declared at %s, %s branch: source %s, target %s", name, where, branch, sourceName, targetName)
	})
	if err != nil {
		return nil, false, err
	}
	body = append(body, check...)
	return append(body, ir.Return{Value: result}), true, nil
}
