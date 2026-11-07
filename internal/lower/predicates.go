package lower

import (
	"errors"
	"fmt"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// As in instantiate.go, this calls the pinned checker's own implementation until its shim
// exports it. The annotation is a claimed postcondition, never a flow fact. The flow nodes below contain
// only checks independently trusted by Adamic, not assertion calls or assignments.
//
//go:linkname predicateFlowType github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getFlowTypeOfReferenceEx
func predicateFlowType(
	receiver *checker.Checker,
	reference *ast.Node,
	declared, initial *checker.Type,
	container *ast.Node,
	flow *ast.FlowNode,
) *checker.Type

type predicateFlowProof struct {
	l          *lowering
	function   *ast.Node
	parameter  *ast.Node
	declared   *checker.Type
	admitted   *checker.Type
	target     *checker.Type
	assertion  bool
	trueTypes  []*checker.Type
	lastReturn *ast.Node
	// lyingTag is why check refused a comparison on a tag that can lie, for the refusal to name.
	lyingTag string
}

type predicateFlowPath struct {
	flow    *ast.FlowNode
	checked bool
}

func (proof *predicateFlowProof) refused(node *ast.Node, reason string) *Refused {
	return &Refused{
		Where: proof.l.program.Where(node),
		What:  "a type predicate whose return is not proven (" + reason + ")",
		Fix:   "inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)",
	}
}

func (lowering *lowering) proveFlowPredicate(node *ast.Node) error {
	if lowering.censusPredicateMarkerContract(node) {
		return nil
	}
	annotation := node.AsTypePredicateNode()
	function := node.Parent
	if !ast.IsFunctionLike(function) || function.Body() == nil || annotation.ParameterName.Kind != ast.KindIdentifier {
		return (&predicateFlowProof{l: lowering}).refused(node, "there is no body proving this parameter")
	}
	var parameter *ast.Node
	for _, candidate := range function.Parameters() {
		if ast.IsIdentifier(candidate.Name()) && candidate.Name().Text() == annotation.ParameterName.Text() {
			parameter = candidate
			break
		}
	}
	p := &predicateFlowProof{
		l:         lowering,
		function:  function,
		parameter: parameter,
		assertion: annotation.AssertsModifier != nil,
	}
	if parameter == nil ||
		parameter.AsParameterDeclaration().Initializer != nil ||
		parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return p.refused(node, "the predicate must name an unchanged plain parameter")
	}
	p.declared = lowering.checker.GetTypeAtLocation(parameter.Name())
	if annotation.Type != nil {
		p.target = lowering.checker.GetTypeFromTypeNode(annotation.Type)
	} else {
		// asserts cond is accepted only for a boolean parameter, with a proven true postcondition.
		if p.declared.Flags()&checker.TypeFlagsBooleanLike == 0 {
			return p.refused(node, "asserts cond needs a boolean parameter")
		}
		p.target = predicateTrueType(lowering.checker)
	}
	// A rebinding changes the value being tested, not the caller's argument. Include writes in
	// nested closures, even if the checker happens to retain a narrowing across their calls.
	var changed *ast.Node
	var visit ast.Visitor
	visit = func(current *ast.Node) bool {
		if current.Kind == ast.KindIdentifier &&
			lowering.checker.GetSymbolAtLocation(current) == lowering.checker.GetSymbolAtLocation(parameter.Name()) &&
			ast.IsAssignmentTarget(current) {
			changed = current
		}
		current.ForEachChild(visit)
		return false
	}
	function.Body().ForEachChild(visit)
	if changed != nil {
		return p.refused(changed, "the predicate parameter is assigned")
	}
	return p.proveBody(node)
}

func (proof *predicateFlowProof) proveBody(node *ast.Node) error {
	function := proof.function
	paths := []predicateFlowPath{{flow: &ast.FlowNode{Flags: ast.FlowFlagsStart}}}
	var remaining []predicateFlowPath
	var err error
	if function.Body().Kind == ast.KindBlock {
		remaining, err = proof.statement(function.Body(), paths)
	} else {
		err = proof.returned(function.Body(), function.Body(), paths)
	}
	if err != nil {
		return err
	}
	for _, path := range remaining {
		if !proof.assertion {
			return proof.refused(function.Body(), "implicit return is not a proven boolean")
		}
		if normalReturnError := proof.normalReturn(function.Body(), path); normalReturnError != nil {
			return normalReturnError
		}
	}
	if !proof.assertion &&
		len(proof.trueTypes) > 0 &&
		!checker.Checker_isTypeIdenticalTo(
			proof.l.checker,
			proof.l.checker.GetUnionType(proof.trueTypes),
			proof.target,
		) {
		return proof.refused(
			proof.lastReturn,
			"the body's true narrowing does not match "+proof.l.checker.TypeToString(proof.target),
		)
	}
	return nil
}

func (proof *predicateFlowProof) narrowed(path predicateFlowPath, initial *checker.Type) *checker.Type {
	if path.flow.Node != nil &&
		proof.l.isArrayPredicateCall(path.flow.Node) &&
		proof.reference(
			ast.SkipParentheses(ast.SkipParentheses(path.flow.Node).AsCallExpression().Arguments.Nodes[0]),
		) {
		return proof.l.arrayPredicateMembers(initial, path.flow.Flags&ast.FlowFlagsTrueCondition != 0)
	}
	if proof.admitted != nil && initial == proof.declared {
		initial = proof.admitted
	}
	return proof.l.concrete(
		predicateFlowType(proof.l.checker, proof.parameter.Name(), proof.declared, initial, proof.function, path.flow),
	)
}

func predicateBranch(path predicateFlowPath, expression *ast.Node, truth bool) predicateFlowPath {
	flags := ast.FlowFlagsFalseCondition
	if truth {
		flags = ast.FlowFlagsTrueCondition
	}
	return predicateFlowPath{flow: &ast.FlowNode{Flags: flags, Node: expression, Antecedent: path.flow}, checked: true}
}

// Immutable scalar arguments carry their admitted value type into the body.
// An object view does not certify its current tag without an actual test.
func (proof *predicateFlowProof) admittedValueProof() bool {
	return proof.admitted != nil &&
		proof.target.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined) != 0
}

func (proof *predicateFlowProof) normalReturn(node *ast.Node, path predicateFlowPath) error {
	narrowed := proof.narrowed(path, proof.declared)
	if (!path.checked && !proof.admittedValueProof()) ||
		!checker.Checker_isTypeIdenticalTo(proof.l.checker, narrowed, proof.target) ||
		proof.l.nominalMismatch(narrowed, proof.target, map[[2]*checker.Type]bool{}) != nil {
		return proof.refused(
			node,
			"normal return has not narrowed "+
				proof.parameter.Name().Text()+
				" to "+
				proof.l.checker.TypeToString(proof.target),
		)
	}
	return nil
}

func (proof *predicateFlowProof) returned(node, expression *ast.Node, paths []predicateFlowPath) error {
	proof.lastReturn = node
	if proof.assertion {
		if expression != nil {
			return proof.refused(node, "an assertion must return without a value")
		}
		for _, path := range paths {
			if normalReturnError := proof.normalReturn(node, path); normalReturnError != nil {
				return normalReturnError
			}
		}
		return nil
	}
	if expression == nil {
		return proof.refused(node, "return has no boolean check")
	}
	expression = ast.SkipParentheses(expression)
	literal := expression.Kind == ast.KindTrueKeyword || expression.Kind == ast.KindFalseKeyword
	if !literal && !proof.check(expression) {
		return proof.untrusted(node, "return expression is not a trusted check on "+proof.parameter.Name().Text())
	}
	for _, path := range paths {
		for _, truth := range []bool{true, false} {
			if literal && truth != (expression.Kind == ast.KindTrueKeyword) {
				continue
			}
			branch := path
			if !literal {
				branch = predicateBranch(path, expression, truth)
			}
			// For a false path, an argument initially in the claimed type must be
			// impossible, as for inferred predicates. That proves the caller's else
			// narrowing independently of the true-return comparison.
			if truth {
				narrowed := proof.narrowed(branch, proof.declared)
				if (!branch.checked && !proof.admittedValueProof()) ||
					!proof.l.checker.IsTypeAssignableTo(narrowed, proof.target) ||
					proof.l.nominalMismatch(narrowed, proof.target, map[[2]*checker.Type]bool{}) != nil {
					return proof.refused(node, "true return narrows to "+
						proof.l.checker.TypeToString(narrowed)+
						", not "+
						proof.l.checker.TypeToString(proof.target))
				}
				proof.trueTypes = append(proof.trueTypes, narrowed)
			} else if proof.narrowed(branch, proof.target).Flags()&checker.TypeFlagsNever == 0 {
				return proof.refused(node, "false return can still contain "+proof.l.checker.TypeToString(proof.target))
			}
		}
	}
	return nil
}

func (proof *predicateFlowProof) statement(node *ast.Node, paths []predicateFlowPath) ([]predicateFlowPath, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	if len(paths) > 256 {
		return nil, proof.refused(node, "too many return paths to prove")
	}
	switch node.Kind {
	case ast.KindBlock:
		for _, statement := range node.AsBlock().Statements.Nodes {
			var statementError error
			paths, statementError = proof.statement(statement, paths)
			if statementError != nil {
				return nil, statementError
			}
		}
		return paths, nil
	case ast.KindIfStatement:
		statement := node.AsIfStatement()
		if !proof.check(statement.Expression) {
			return nil, proof.untrusted(node, "branch is not a trusted parameter check")
		}
		var then, otherwise []predicateFlowPath
		for _, path := range paths {
			then = append(then, predicateBranch(path, statement.Expression, true))
			otherwise = append(otherwise, predicateBranch(path, statement.Expression, false))
		}
		then, statementError := proof.statement(statement.ThenStatement, then)
		if statementError != nil {
			return nil, statementError
		}
		if statement.ElseStatement != nil {
			otherwise, statementError = proof.statement(statement.ElseStatement, otherwise)
			if statementError != nil {
				return nil, statementError
			}
		}
		return append(then, otherwise...), nil
	case ast.KindReturnStatement:
		return nil, proof.returned(node, node.AsReturnStatement().Expression, paths)
	case ast.KindThrowStatement:
		return nil, nil
	case ast.KindEmptyStatement:
		return paths, nil
	case ast.KindExpressionStatement, ast.KindVariableStatement:
		if node.Kind == ast.KindExpressionStatement && proof.l.isPanicCall(node.Expression()) {
			return nil, nil
		}
		if predicateEffects(node) {
			// Unknown calls and all writes can invalidate discriminants through an alias. Starting
			// over also prevents the checker from retaining stale readonly-property facts.
			for index := range paths {
				paths[index] = predicateFlowPath{flow: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
			}
		}
		return paths, nil
	default:
		return nil, proof.refused(node, fmt.Sprintf("return paths through %s are not verified", node.Kind))
	}
}

// check recognizes only independently trusted narrowing expressions, with no calls, aliases,
// casts or computed property accesses. The checker still decides what each branch proves.
func (proof *predicateFlowProof) check(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPrefixUnaryExpression &&
		node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		return proof.check(node.AsPrefixUnaryExpression().Operand)
	}
	if proof.assertion && proof.target == predicateTrueType(proof.l.checker) && proof.reference(node) {
		return true
	}
	if proof.l.isArrayPredicateCall(node) {
		return proof.reference(ast.SkipParentheses(node.AsCallExpression().Arguments.Nodes[0])) &&
			(proof.l.arrayPredicateDomain(proof.declared) || proof.declared.Flags()&checker.TypeFlagsUnknown != 0)
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
		return proof.check(binary.Left) && proof.check(binary.Right)
	case ast.KindInstanceOfKeyword:
		if !proof.reference(ast.SkipParentheses(binary.Left)) || binary.Right.Kind != ast.KindIdentifier {
			return false
		}
		symbol := proof.l.checker.GetSymbolAtLocation(binary.Right)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = proof.l.checker.GetAliasedSymbol(symbol)
		}
		return symbol != nil && symbol.Flags&ast.SymbolFlagsClass != 0
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		if proof.tagCanLie(binary.Left) || proof.tagCanLie(binary.Right) {
			return false
		}
		return proof.checkedValue(binary.Left) && proof.literal(binary.Right) ||
			proof.checkedValue(binary.Right) && proof.literal(binary.Left)
	}
	return false
}

// tagCanLie reports whether value reads a field off the parameter whose tags can lie in the union
// that reaches the body: a comparison on it proves no layout, whatever it's compared with
// (#drwm5pj). The checker would still narrow by it, so it's refused before the checker is asked.
func (proof *predicateFlowProof) tagCanLie(value *ast.Node) bool {
	value = ast.SkipParentheses(value)
	if value.Kind != ast.KindPropertyAccessExpression || !proof.checkedValue(value) {
		return false
	}
	incoming := proof.declared
	if proof.admitted != nil {
		incoming = proof.admitted
	}
	field := value.AsPropertyAccessExpression().Name().Text()
	hazard := proof.l.enumTagHazardOf(incoming, field)
	if hazard == enumTagNoHazard {
		return false
	}
	proof.lyingTag = predicateTagHazardReason(field, hazard)
	return true
}

// untrusted refuses a check, naming the tag that can lie when that was why, and reason otherwise.
func (proof *predicateFlowProof) untrusted(node *ast.Node, reason string) error {
	if proof.lyingTag == "" {
		return proof.refused(node, reason)
	}
	refused := proof.refused(node, proof.lyingTag)
	refused.lyingTag = proof.lyingTag
	return refused
}

func (proof *predicateFlowProof) reference(node *ast.Node) bool {
	return node.Kind == ast.KindIdentifier &&
		proof.l.checker.GetSymbolAtLocation(node) == proof.l.checker.GetSymbolAtLocation(proof.parameter.Name())
}

func (proof *predicateFlowProof) checkedValue(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if proof.reference(node) {
		return true
	}
	if node.Kind == ast.KindTypeOfExpression {
		return proof.reference(ast.SkipParentheses(node.AsTypeOfExpression().Expression))
	}
	return node.Kind == ast.KindPropertyAccessExpression &&
		proof.reference(ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)) &&
		node.Flags&ast.NodeFlagsOptionalChain == 0
}

func (proof *predicateFlowProof) literal(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined" &&
			proof.l.checker.GetSymbolAtLocation(node) == proof.l.checker.GetUndefinedSymbol()
	}
	return false
}

func predicateEffects(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// A property read may dispatch a getter through a structural view.
		return true
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindBinaryExpression:
		if node.Kind != ast.KindBinaryExpression ||
			ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			return true
		}
	case ast.KindPostfixUnaryExpression, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			return true
		}
	}
	return node.ForEachChild(predicateEffects)
}

func predicateTrueType(c *checker.Checker) *checker.Type {
	for _, member := range c.GetBooleanType().Types() {
		if member.AsLiteralType().Value() == true {
			return member
		}
	}
	panic("checker boolean type has no true member")
}

// typePredicateRefusal checks a type predicate and prunes it: the walk's later rules never see one,
// and never enter it.
func (lowering *lowering) typePredicateRefusal(node *ast.Node) (bool, error) {
	if node.Kind != ast.KindTypePredicate {
		return false, nil
	}
	return true, lowering.predicateRefusal(node)
}

func init() {
	refusalRules.registerPruning(500, "typePredicateRefusal", (*lowering).typePredicateRefusal)
}

// The flow verifier keeps main's literal and nominal proofs. Body summaries add
// verified helpers and truthiness assertions; tag-only facts use checked views.
func (lowering *lowering) predicateRefusal(node *ast.Node) error {
	if node.Parent.Kind == ast.KindFunctionDeclaration &&
		node.Parent.Body() == nil &&
		lowering.censusImplementation(node.Parent) != nil {
		if lowering.checkedAssertionSource(node) ||
			lowering.predicateOverloadProven(lowering.censusImplementation(node.Parent), node.Parent) {
			return nil
		}
		return predicateFailure(lowering, node, "the overload implementation does not prove both directions")
	}
	if lowering.predicateParameter(node) != nil {
		return nil
	}
	if lowering.checkedFunctionTypeClaim(node) {
		return nil // Every call through it checks the claim it resolved to (predicates_function_type.go).
	}
	original := lowering.proveFlowPredicate(node)
	if original == nil {
		return nil
	}
	proof, provePredicateError := lowering.provePredicate(node)
	if provePredicateError != nil {
		if lowering.checkedAssertionSource(node) &&
			node.Parent.Kind == ast.KindFunctionDeclaration &&
			node.Parent.Body() != nil {
			return nil // Every direct call validates this claim; other calls stay refused.
		}
		if lowering.checkedPredicateClosure(node.Parent) {
			return nil // The closure checks its own claim at each return (#1w5bve2).
		}
		// A tag that can lie is the reason worth naming, whichever prover met it.
		var flow, body *Refused
		if errors.As(provePredicateError, &body) &&
			body.lyingTag != "" &&
			(!errors.As(original, &flow) || flow.lyingTag == "") {
			return body
		}
		return original
	}
	if proof.TaggedView {
		target := lowering.checker.GetTypeAtLocation(node.AsTypePredicateNode().Type)
		// The claimed parameter's type is what the view's value had before it, or none.
		var source *checker.Type
		if node.Parent.Type() == node {
			if parameter := lowering.checkedPredicateClaimed(node.Parent); parameter != nil {
				source = lowering.checker.GetTypeAtLocation(parameter.Name())
			}
		}
		members := []*checker.Type{target}
		if target.Flags()&checker.TypeFlagsUnion != 0 {
			members = target.Types()
		}
		for _, member := range members {
			for _, property := range lowering.checker.GetPropertiesOfType(member) {
				// In .ts the view checks an optional field where it's read (#1yr7rdh), and a write to
				// one through the view is NotYet, never an unchecked store (#mr87ptk): absent reads
				// undefined, and present is checked like any field. .a needs proof, and keeps the
				// refusal, which names the field rather than the return (#1w5bve2, Adamic's condition 5).
				if property.Flags&ast.SymbolFlagsOptional != 0 && !lowering.checkedAssertionSource(node) {
					what := "a tag-proven predicate whose target has an optional field (" +
						property.Name +
						" on " +
						lowering.checker.TypeToString(member) +
						") that the view can't check"
					return &Refused{
						Where: lowering.program.Where(node),
						What:  what,
						Fix:   "make the field required, or return a boolean and narrow at the caller (adamic/no-type-predicate)",
					}
				}
			}
			// A body proven by its tag alone leaves the layout to the view: this registration serves
			// a tag-proven predicate wherever no checked call wraps it. A checked call registers the
			// same view in predicateMembership, so neither alone is guarded by a test (#k1t2wp9).
			if _, viewError := lowering.view(node, nil, source, member); viewError != nil {
				return viewError
			}
			// The value enters the view wherever a caller's read is narrowed to it (view_widened.go).
			lowering.viewSources[len(lowering.viewSources)-1].narrows = true
		}
	}
	return nil
}

// A signature on a plain function parameter is an argument contract. The call
// visitor proves the actual function body before admitting any invocation. Live
// contracts also need a closed set of direct callers; an uncalled declaration or
// an escaped function value supplies no producer proof for its callback.
func (lowering *lowering) predicateParameter(node *ast.Node) *ast.Node {
	signature := node.Parent
	if signature == nil ||
		signature.Kind != ast.KindFunctionType ||
		signature.Parent == nil ||
		signature.Parent.Kind != ast.KindParameter {
		return nil
	}
	parameter := signature.Parent
	if parameter.AsParameterDeclaration().Initializer != nil ||
		parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return nil
	}
	function := parameter.Parent
	if function.Kind == ast.KindFunctionDeclaration && function.Body() == nil {
		implementation := lowering.censusImplementation(function)
		if implementation == nil {
			return nil
		}
		for index, candidate := range function.Parameters() {
			if candidate == parameter && index < len(implementation.Parameters()) {
				parameter = implementation.Parameters()[index]
				function = implementation
				break
			}
		}
	}
	if !ast.IsFunctionLike(function) || function.Body() == nil {
		return nil
	}
	symbol := lowering.checker.GetSymbolAtLocation(parameter.Name())
	changed := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier &&
			lowering.checker.GetSymbolAtLocation(n) == symbol &&
			ast.IsAssignmentTarget(n) {
			changed = true
		}
		n.ForEachChild(visit)
		return false
	}
	function.Body().ForEachChild(visit)
	if changed {
		return nil
	}
	// A bodyless overload's callback predicate is not the implementation's
	// contract when the implementation takes an ordinary boolean callback.
	if signature.Parent != parameter && plainCallbackType(parameter.Type()) {
		return parameter
	}
	if !lowering.predicateCallbackCallers(function, symbol) {
		return nil
	}
	return parameter
}

// plainCallbackType is whether written, an implementation parameter's type, is a function type
// claiming no predicate, in parentheses or not, alone or beside undefined: debug.ts's assertNode takes
// `((node: Node) => boolean) | undefined` where core.ts's filter takes `(x: T) => boolean` (#1w5bve2).
// An absent argument is never called, and a present one is exactly the function, so the body narrows
// nothing through it either way and the overload's argument keeps predicateArguments' proof. It reads
// the written type, not the checker's, so a union of callables or an alias stays out as before.
func plainCallbackType(written *ast.Node) bool {
	if written == nil {
		return false
	}
	written = ast.SkipTypeParentheses(written)
	if written.Kind == ast.KindUnionType {
		var defined *ast.Node
		for _, member := range written.AsUnionTypeNode().Types.Nodes {
			member = ast.SkipTypeParentheses(member)
			if member.Kind == ast.KindUndefinedKeyword {
				continue
			}
			if defined != nil {
				return false
			}
			defined = member
		}
		if defined == nil {
			return false
		}
		written = defined
	}
	return written.Kind == ast.KindFunctionType && written.Type() != nil && written.Type().Kind != ast.KindTypePredicate
}

// Every producer is checked by predicateArguments before annotations are admitted.
// Keep that proof within direct calls, and do not let a callback escape its body.
func (lowering *lowering) predicateCallbackCallers(function *ast.Node, callback *ast.Symbol) bool {
	if function.Kind != ast.KindFunctionDeclaration ||
		!ast.IsIdentifier(function.Name()) ||
		ast.HasSyntacticModifier(function, ast.ModifierFlagsExport) {
		return false
	}
	owner := lowering.symbol(function.Name())
	if owner == nil {
		return false
	}
	found, safe := false, true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && !ast.IsPartOfTypeNode(node) && !ast.IsDeclarationName(node) {
			symbol := lowering.symbol(node)
			if symbol == owner || symbol == callback {
				callee := node
				for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
					callee = callee.Parent
				}
				call := callee.Parent
				safe = safe &&
					call != nil &&
					call.Kind == ast.KindCallExpression &&
					call.AsCallExpression().Expression == callee
				if symbol == owner {
					found = true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range lowering.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return found && safe
}

// predicateArgumentsRefusal checks a call's predicate contracts again in the walk. refuse's contracts
// pass already checked every call, so this never refuses; it's kept for order.
func (lowering *lowering) predicateArgumentsRefusal(node *ast.Node) error {
	if node.Kind == ast.KindCallExpression {
		if predicateArgumentsError := lowering.predicateArguments(node); predicateArgumentsError != nil {
			return predicateArgumentsError
		}
	}
	return nil
}

func init() {
	refusalRules.register(1700, "predicateArgumentsRefusal", (*lowering).predicateArgumentsRefusal)
}

func (lowering *lowering) predicateArguments(node *ast.Node) error {
	signature := lowering.checker.GetResolvedSignature(node)
	if signature == nil {
		return nil
	}
	if overload := signature.Declaration(); overload != nil &&
		overload.Kind == ast.KindFunctionDeclaration &&
		overload.Body() == nil &&
		overload.Type() != nil &&
		overload.Type().Kind == ast.KindTypePredicate {
		if !lowering.directPredicateCallee(node.AsCallExpression().Expression, overload) {
			return lowering.notYet(node, "an indirect call of a checked predicate overload")
		}
	}
	if declaration := signature.Declaration(); declaration != nil &&
		declaration.Type() != nil &&
		declaration.Type().Kind == ast.KindTypePredicate &&
		(declaration.Body() != nil || lowering.censusImplementation(declaration) != nil) &&
		lowering.checkedAssertionSource(declaration) &&
		!lowering.checkedPredicateBodyProven(declaration.Type()) {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		// A call through a value calls the checked forwarder functionValue made, which checks the
		// claim's true side (#vdt9f1q); the walk owns its false side. Direct names, including
		// namespace and module qualifications, check both directions at the static call.
		named := declaration.Name() != nil && lowering.symbol(callee) == lowering.symbol(declaration.Name())
		direct := lowering.directPredicateCallee(callee, declaration)
		// A checked closure checks itself, but only a const name bound to it can hold nothing else.
		if !lowering.checkedPredicateClosureCallee(callee, declaration) &&
			(declaration.Kind != ast.KindFunctionDeclaration ||
				(!direct &&
					(named || callee.Kind == ast.KindPropertyAccessExpression && lowering.namespaceMember(callee)))) {
			return lowering.notYet(node, "an indirect call of a checked predicate")
		}
	}
	for index, parameter := range signature.Parameters() {
		if index >= len(node.AsCallExpression().Arguments.Nodes) {
			continue
		}
		var contract *ast.Node
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter &&
				declaration.Type() != nil &&
				declaration.Type().Kind == ast.KindFunctionType &&
				declaration.Type().Type() != nil &&
				declaration.Type().Type().Kind == ast.KindTypePredicate {
				contract = declaration.Type().Type()
			}
		}
		// A parameter whose claim is checked at every call through it takes any function of its type: a
		// lie stops at that call. A library's parameter is called where nothing checks it, so its
		// argument keeps the proof (checkedFunctionTypeClaim).
		if contract == nil || lowering.checkedFunctionTypeClaim(contract) {
			continue
		}
		argument := ast.SkipParentheses(node.AsCallExpression().Arguments.Nodes[index])
		implementation := argument
		if argument.Kind == ast.KindIdentifier {
			symbol := lowering.symbol(argument)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if ast.IsFunctionLike(declaration) && declaration.Body() != nil {
						implementation = declaration
					}
					if declaration.Kind == ast.KindVariableDeclaration &&
						declaration.AsVariableDeclaration().Initializer != nil {
						implementation = ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
					}
				}
			}
		}
		failure := func() *Refused {
			file := ast.GetSourceFileOfNode(argument)
			text := file.Text()[scanner.GetTokenPosOfNode(argument, file, false):argument.End()]
			return &Refused{
				Where: lowering.program.Where(argument),
				What: "an unproven predicate argument for parameter " +
					parameter.Name +
					" (argument " +
					fmt.Sprintf("%q", text) +
					")",
				Fix: "pass a named function or arrow whose body proves both predicate branches; return a boolean and narrow at the caller (adamic/no-type-predicate)",
			}
		}
		if !ast.IsFunctionLike(implementation) ||
			implementation.Body() == nil ||
			(implementation.Type() != nil && implementation.Type().Kind != ast.KindTypePredicate) {
			return failure()
		}
		if argument.Kind == ast.KindIdentifier && lowering.assignedSymbol(lowering.symbol(argument)) {
			return failure()
		}
		wanted := contract.AsTypePredicateNode()
		// An optional parameter's type is the function or undefined, which has no call signature of its
		// own. An absent argument is never called, and a present one is exactly the function, so the
		// claim to prove is the function's (getParseTreeNode's nodeTest, #1w5bve2).
		calls := lowering.checker.GetSignaturesOfType(
			lowering.withoutUndefined(lowering.checker.GetTypeOfSymbol(parameter)),
			checker.SignatureKindCall,
		)
		var target *checker.Type
		if len(calls) == 1 {
			if predicate := predicateOfSignature(lowering.checker, calls[0]); predicate != nil {
				target = predicate.Type()
			}
		}
		if wanted.Type == nil || target == nil {
			return failure()
		}
		if implementation.Type() == nil {
			if !lowering.proveInferredPredicate(implementation, target) {
				return failure()
			}
		} else {
			annotation := implementation.Type().AsTypePredicateNode()
			if annotation.Type == nil ||
				!checker.Checker_isTypeIdenticalTo(
					lowering.checker,
					lowering.checker.GetTypeAtLocation(annotation.Type),
					target,
				) {
				return failure()
			}
			// A checked predicate passed by name is its checked forwarder (functionValue), which
			// checks this claim's true side at every call through it (#vdt9f1q).
			// An arrow or function expression is itself the value, and checks its claim at each return.
			if !lowering.checkedPredicateArgumentProven(implementation.Type()) &&
				!(argument.Kind == ast.KindIdentifier && lowering.checkedPredicateDeclaration(implementation)) &&
				!lowering.checkedPredicateClosure(implementation) {
				refused := failure()
				var proof *Refused
				if predicateRefusalError := lowering.predicateRefusal(
					implementation.Type(),
				); errors.As(predicateRefusalError, &proof) && proof.lyingTag != "" {
					refused.What += ": " + proof.lyingTag
				}
				return refused
			}
		}
	}
	return nil
}

// The pinned checker resolves a predicate's target under overload type arguments.
// This recovers the claim to validate; it never supplies a body-proof fact.
//
//go:linkname predicateOfSignature github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getTypePredicateOfSignature
func predicateOfSignature(receiver *checker.Checker, signature *checker.Signature) *checker.TypePredicate

// Keep reference identity aligned with the pinned checker's flow analysis,
// including dotted/indexed spellings and stable computed indices.
//
//go:linkname predicateMatchingReference github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).isMatchingReference
func predicateMatchingReference(receiver *checker.Checker, source, target *ast.Node) bool

// assignedSymbol says whether any identifier in the program that resolves to symbol (nil included) is an assignment
// target. predicateArguments asked it for every identifier passed as a predicate argument, and each ask walked every
// file of the program: 24.9% of the meter's latent step on tsc's 79 files once predicate proof was memoized (#pv935yw).
// The answer is the program's and the checker's alone, so one walk collects every assigned symbol into
// l.assignedSymbols the first time it's asked, and nothing invalidates it.
func (lowering *lowering) assignedSymbol(symbol *ast.Symbol) bool {
	if lowering.assignedSymbols == nil {
		lowering.assignedSymbols = map[*ast.Symbol]bool{}
		var writes ast.Visitor
		writes = func(n *ast.Node) bool {
			if n.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(n) {
				lowering.assignedSymbols[lowering.symbol(n)] = true
			}
			n.ForEachChild(writes)
			return false
		}
		for _, file := range lowering.program.Files() {
			file.AsNode().ForEachChild(writes)
		}
	}
	return lowering.assignedSymbols[symbol]
}
