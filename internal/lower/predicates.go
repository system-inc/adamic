package lower

import (
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
func predicateFlowType(receiver *checker.Checker, reference *ast.Node, declared, initial *checker.Type, container *ast.Node, flow *ast.FlowNode) *checker.Type

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
}

type predicateFlowPath struct {
	flow    *ast.FlowNode
	checked bool
}

func (p *predicateFlowProof) refused(node *ast.Node, reason string) error {
	return &Refused{Where: p.l.program.Where(node), What: "a type predicate whose return is not proven (" + reason + ")", Fix: "inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)"}
}

func (l *lowering) proveFlowPredicate(node *ast.Node) error {
	if l.censusPredicateMarkerContract(node) {
		return nil
	}
	annotation := node.AsTypePredicateNode()
	function := node.Parent
	if !ast.IsFunctionLike(function) || function.Body() == nil || annotation.ParameterName.Kind != ast.KindIdentifier {
		return (&predicateFlowProof{l: l}).refused(node, "there is no body proving this parameter")
	}
	var parameter *ast.Node
	for _, candidate := range function.Parameters() {
		if ast.IsIdentifier(candidate.Name()) && candidate.Name().Text() == annotation.ParameterName.Text() {
			parameter = candidate
			break
		}
	}
	p := &predicateFlowProof{l: l, function: function, parameter: parameter, assertion: annotation.AssertsModifier != nil}
	if parameter == nil || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return p.refused(node, "the predicate must name an unchanged plain parameter")
	}
	p.declared = l.checker.GetTypeAtLocation(parameter.Name())
	if annotation.Type != nil {
		p.target = l.checker.GetTypeFromTypeNode(annotation.Type)
	} else {
		// asserts cond is accepted only for a boolean parameter, with a proven true postcondition.
		if p.declared.Flags()&checker.TypeFlagsBooleanLike == 0 {
			return p.refused(node, "asserts cond needs a boolean parameter")
		}
		p.target = predicateTrueType(l.checker)
	}
	// A rebinding changes the value being tested, not the caller's argument. Include writes in
	// nested closures, even if the checker happens to retain a narrowing across their calls.
	var changed *ast.Node
	var visit ast.Visitor
	visit = func(current *ast.Node) bool {
		if current.Kind == ast.KindIdentifier && l.checker.GetSymbolAtLocation(current) == l.checker.GetSymbolAtLocation(parameter.Name()) && ast.IsAssignmentTarget(current) {
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

func (p *predicateFlowProof) proveBody(node *ast.Node) error {
	function := p.function
	paths := []predicateFlowPath{{flow: &ast.FlowNode{Flags: ast.FlowFlagsStart}}}
	var remaining []predicateFlowPath
	var err error
	if function.Body().Kind == ast.KindBlock {
		remaining, err = p.statement(function.Body(), paths)
	} else {
		err = p.returned(function.Body(), function.Body(), paths)
	}
	if err != nil {
		return err
	}
	for _, path := range remaining {
		if !p.assertion {
			return p.refused(function.Body(), "implicit return is not a proven boolean")
		}
		if err := p.normalReturn(function.Body(), path); err != nil {
			return err
		}
	}
	if !p.assertion && len(p.trueTypes) > 0 && !checker.Checker_isTypeIdenticalTo(p.l.checker, p.l.checker.GetUnionType(p.trueTypes), p.target) {
		return p.refused(p.lastReturn, "the body's true narrowing does not match "+p.l.checker.TypeToString(p.target))
	}
	return nil
}

func (p *predicateFlowProof) narrowed(path predicateFlowPath, initial *checker.Type) *checker.Type {
	if path.flow.Node != nil && p.l.isArrayPredicateCall(path.flow.Node) && p.reference(ast.SkipParentheses(ast.SkipParentheses(path.flow.Node).AsCallExpression().Arguments.Nodes[0])) {
		return p.l.arrayPredicateMembers(initial, path.flow.Flags&ast.FlowFlagsTrueCondition != 0)
	}
	if p.admitted != nil && initial == p.declared {
		initial = p.admitted
	}
	return p.l.concrete(predicateFlowType(p.l.checker, p.parameter.Name(), p.declared, initial, p.function, path.flow))
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
func (p *predicateFlowProof) admittedValueProof() bool {
	return p.admitted != nil && p.target.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined) != 0
}

func (p *predicateFlowProof) normalReturn(node *ast.Node, path predicateFlowPath) error {
	narrowed := p.narrowed(path, p.declared)
	if (!path.checked && !p.admittedValueProof()) || !checker.Checker_isTypeIdenticalTo(p.l.checker, narrowed, p.target) || p.l.nominalMismatch(narrowed, p.target, map[[2]*checker.Type]bool{}) != nil {
		return p.refused(node, "normal return has not narrowed "+p.parameter.Name().Text()+" to "+p.l.checker.TypeToString(p.target))
	}
	return nil
}

func (p *predicateFlowProof) returned(node, expression *ast.Node, paths []predicateFlowPath) error {
	p.lastReturn = node
	if p.assertion {
		if expression != nil {
			return p.refused(node, "an assertion must return without a value")
		}
		for _, path := range paths {
			if err := p.normalReturn(node, path); err != nil {
				return err
			}
		}
		return nil
	}
	if expression == nil {
		return p.refused(node, "return has no boolean check")
	}
	expression = ast.SkipParentheses(expression)
	literal := expression.Kind == ast.KindTrueKeyword || expression.Kind == ast.KindFalseKeyword
	if !literal && !p.check(expression) {
		return p.refused(node, "return expression is not a trusted check on "+p.parameter.Name().Text())
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
				narrowed := p.narrowed(branch, p.declared)
				if (!branch.checked && !p.admittedValueProof()) || !p.l.checker.IsTypeAssignableTo(narrowed, p.target) || p.l.nominalMismatch(narrowed, p.target, map[[2]*checker.Type]bool{}) != nil {
					return p.refused(node, "true return narrows to "+p.l.checker.TypeToString(narrowed)+", not "+p.l.checker.TypeToString(p.target))
				}
				p.trueTypes = append(p.trueTypes, narrowed)
			} else if p.narrowed(branch, p.target).Flags()&checker.TypeFlagsNever == 0 {
				return p.refused(node, "false return can still contain "+p.l.checker.TypeToString(p.target))
			}
		}
	}
	return nil
}

func (p *predicateFlowProof) statement(node *ast.Node, paths []predicateFlowPath) ([]predicateFlowPath, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	if len(paths) > 256 {
		return nil, p.refused(node, "too many return paths to prove")
	}
	switch node.Kind {
	case ast.KindBlock:
		for _, statement := range node.AsBlock().Statements.Nodes {
			var err error
			paths, err = p.statement(statement, paths)
			if err != nil {
				return nil, err
			}
		}
		return paths, nil
	case ast.KindIfStatement:
		statement := node.AsIfStatement()
		if !p.check(statement.Expression) {
			return nil, p.refused(node, "branch is not a trusted parameter check")
		}
		var then, otherwise []predicateFlowPath
		for _, path := range paths {
			then = append(then, predicateBranch(path, statement.Expression, true))
			otherwise = append(otherwise, predicateBranch(path, statement.Expression, false))
		}
		then, err := p.statement(statement.ThenStatement, then)
		if err != nil {
			return nil, err
		}
		if statement.ElseStatement != nil {
			otherwise, err = p.statement(statement.ElseStatement, otherwise)
			if err != nil {
				return nil, err
			}
		}
		return append(then, otherwise...), nil
	case ast.KindReturnStatement:
		return nil, p.returned(node, node.AsReturnStatement().Expression, paths)
	case ast.KindThrowStatement:
		return nil, nil
	case ast.KindEmptyStatement:
		return paths, nil
	case ast.KindExpressionStatement, ast.KindVariableStatement:
		if node.Kind == ast.KindExpressionStatement && p.l.isPanicCall(node.Expression()) {
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
		return nil, p.refused(node, fmt.Sprintf("return paths through %s are not verified", node.Kind))
	}
}

// check recognizes only independently trusted narrowing expressions, with no calls, aliases,
// casts or computed property accesses. The checker still decides what each branch proves.
func (p *predicateFlowProof) check(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		return p.check(node.AsPrefixUnaryExpression().Operand)
	}
	if p.assertion && p.target == predicateTrueType(p.l.checker) && p.reference(node) {
		return true
	}
	if p.l.isArrayPredicateCall(node) {
		return p.reference(ast.SkipParentheses(node.AsCallExpression().Arguments.Nodes[0])) && (p.l.arrayPredicateDomain(p.declared) || p.declared.Flags()&checker.TypeFlagsUnknown != 0)
	}
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
		return p.check(binary.Left) && p.check(binary.Right)
	case ast.KindInstanceOfKeyword:
		if !p.reference(ast.SkipParentheses(binary.Left)) || binary.Right.Kind != ast.KindIdentifier {
			return false
		}
		symbol := p.l.checker.GetSymbolAtLocation(binary.Right)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = p.l.checker.GetAliasedSymbol(symbol)
		}
		return symbol != nil && symbol.Flags&ast.SymbolFlagsClass != 0
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return p.checkedValue(binary.Left) && p.literal(binary.Right) || p.checkedValue(binary.Right) && p.literal(binary.Left)
	}
	return false
}

func (p *predicateFlowProof) reference(node *ast.Node) bool {
	return node.Kind == ast.KindIdentifier && p.l.checker.GetSymbolAtLocation(node) == p.l.checker.GetSymbolAtLocation(p.parameter.Name())
}

func (p *predicateFlowProof) checkedValue(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if p.reference(node) {
		return true
	}
	if node.Kind == ast.KindTypeOfExpression {
		return p.reference(ast.SkipParentheses(node.AsTypeOfExpression().Expression))
	}
	return node.Kind == ast.KindPropertyAccessExpression && p.reference(ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)) && node.Flags&ast.NodeFlagsOptionalChain == 0
}

func (p *predicateFlowProof) literal(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined" && p.l.checker.GetSymbolAtLocation(node) == p.l.checker.GetUndefinedSymbol()
	case ast.KindPropertyAccessExpression:
		// An enum member has an independently fixed scalar value. A readonly
		// property or getter is not an enum constant and supplies no such fact.
		symbol := p.l.checker.GetSymbolAtLocation(node)
		if symbol == nil || symbol.Flags&ast.SymbolFlagsEnumMember == 0 {
			return false
		}
		_, _, known := p.l.literalConstant(p.l.checker.GetTypeAtLocation(node))
		return known
	}
	return false
}

func predicateEffects(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// A property read may dispatch a getter through a structural view.
		return true
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindBinaryExpression:
		if node.Kind != ast.KindBinaryExpression || ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
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

// The flow verifier keeps main's literal and nominal proofs. Body summaries add
// verified helpers and truthiness assertions; tag-only facts use checked views.
func (l *lowering) predicateRefusal(node *ast.Node) error {
	if node.Parent.Kind == ast.KindFunctionDeclaration && node.Parent.Body() == nil && l.censusImplementation(node.Parent) != nil {
		if l.checkedAssertionSource(node) || l.predicateOverloadProven(l.censusImplementation(node.Parent), node.Parent) {
			return nil
		}
		return predicateFailure(l, node, "the overload implementation does not prove both directions")
	}
	if l.predicateParameter(node) != nil {
		return nil
	}
	original := l.proveFlowPredicate(node)
	if original == nil {
		return nil
	}
	proof, err := l.provePredicate(node)
	if err != nil {
		if l.checkedAssertionSource(node) && node.Parent.Kind == ast.KindFunctionDeclaration && node.Parent.Body() != nil {
			return nil // Every direct call validates this claim; other calls stay refused.
		}
		return original
	}
	if proof.TaggedView {
		target := l.checker.GetTypeAtLocation(node.AsTypePredicateNode().Type)
		members := []*checker.Type{target}
		if target.Flags()&checker.TypeFlagsUnion != 0 {
			members = target.Types()
		}
		for _, member := range members {
			if _, err := l.view(node, nil, member); err != nil {
				return err
			}
		}
	}
	return nil
}

// A signature on a plain function parameter is an argument contract. The call
// visitor proves the actual function body before admitting any invocation. Live
// contracts also need a closed set of direct callers; an uncalled declaration or
// an escaped function value supplies no producer proof for its callback.
func (l *lowering) predicateParameter(node *ast.Node) *ast.Node {
	signature := node.Parent
	if signature == nil || signature.Kind != ast.KindFunctionType || signature.Parent == nil || signature.Parent.Kind != ast.KindParameter {
		return nil
	}
	parameter := signature.Parent
	if parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return nil
	}
	function := parameter.Parent
	if function.Kind == ast.KindFunctionDeclaration && function.Body() == nil {
		implementation := l.censusImplementation(function)
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
	symbol := l.checker.GetSymbolAtLocation(parameter.Name())
	changed := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && l.checker.GetSymbolAtLocation(n) == symbol && ast.IsAssignmentTarget(n) {
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
	if signature.Parent != parameter && parameter.Type() != nil && parameter.Type().Kind == ast.KindFunctionType && parameter.Type().Type() != nil && parameter.Type().Type().Kind != ast.KindTypePredicate {
		return parameter
	}
	if !l.predicateCallbackCallers(function, symbol) {
		return nil
	}
	return parameter
}

// Every producer is checked by predicateArguments before annotations are admitted.
// Keep that proof within direct calls, and do not let a callback escape its body.
func (l *lowering) predicateCallbackCallers(function *ast.Node, callback *ast.Symbol) bool {
	if function.Kind != ast.KindFunctionDeclaration || !ast.IsIdentifier(function.Name()) || ast.HasSyntacticModifier(function, ast.ModifierFlagsExport) {
		return false
	}
	owner := l.symbol(function.Name())
	if owner == nil {
		return false
	}
	found, safe := false, true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && !ast.IsPartOfTypeNode(node) && !ast.IsDeclarationName(node) {
			symbol := l.symbol(node)
			if symbol == owner || symbol == callback {
				callee := node
				for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
					callee = callee.Parent
				}
				call := callee.Parent
				safe = safe && call != nil && call.Kind == ast.KindCallExpression && call.AsCallExpression().Expression == callee
				if symbol == owner {
					found = true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return found && safe
}

func (l *lowering) predicateArguments(node *ast.Node) error {
	signature := l.checker.GetResolvedSignature(node)
	if signature == nil {
		return nil
	}
	if overload := signature.Declaration(); overload != nil && overload.Kind == ast.KindFunctionDeclaration && overload.Body() == nil && overload.Type() != nil && overload.Type().Kind == ast.KindTypePredicate {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind != ast.KindIdentifier || l.symbol(callee) != l.symbol(overload.Name()) {
			return l.notYet(node, "an indirect call of a checked predicate overload")
		}
	}
	if declaration := signature.Declaration(); declaration != nil && declaration.Type() != nil && declaration.Type().Kind == ast.KindTypePredicate && (declaration.Body() != nil || l.censusImplementation(declaration) != nil) && l.checkedAssertionSource(declaration) && !l.checkedPredicateBodyProven(declaration.Type()) {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if declaration.Kind != ast.KindFunctionDeclaration || !ast.IsIdentifier(callee) || l.symbol(callee) != l.symbol(declaration.Name()) {
			return l.notYet(node, "an indirect call of a checked predicate")
		}
	}
	for index, parameter := range signature.Parameters() {
		if index >= len(node.AsCallExpression().Arguments.Nodes) {
			continue
		}
		var contract *ast.Node
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.Type() != nil && declaration.Type().Kind == ast.KindFunctionType && declaration.Type().Type() != nil && declaration.Type().Type().Kind == ast.KindTypePredicate {
				contract = declaration.Type().Type()
			}
		}
		if contract == nil {
			continue
		}
		argument := ast.SkipParentheses(node.AsCallExpression().Arguments.Nodes[index])
		implementation := argument
		if argument.Kind == ast.KindIdentifier {
			symbol := l.symbol(argument)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if ast.IsFunctionLike(declaration) && declaration.Body() != nil {
						implementation = declaration
					}
					if declaration.Kind == ast.KindVariableDeclaration && declaration.AsVariableDeclaration().Initializer != nil {
						implementation = ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
					}
				}
			}
		}
		failure := func() error {
			file := ast.GetSourceFileOfNode(argument)
			text := file.Text()[scanner.GetTokenPosOfNode(argument, file, false):argument.End()]
			return &Refused{Where: l.program.Where(argument), What: "an unproven predicate argument for parameter " + parameter.Name + " (argument " + fmt.Sprintf("%q", text) + ")", Fix: "pass a named function or arrow whose body proves both predicate branches; return a boolean and narrow at the caller (adamic/no-type-predicate)"}
		}
		if !ast.IsFunctionLike(implementation) || implementation.Body() == nil || (implementation.Type() != nil && implementation.Type().Kind != ast.KindTypePredicate) {
			return failure()
		}
		if argument.Kind == ast.KindIdentifier {
			symbol := l.symbol(argument)
			changed := false
			var writes ast.Visitor
			writes = func(n *ast.Node) bool {
				if n.Kind == ast.KindIdentifier && l.symbol(n) == symbol && ast.IsAssignmentTarget(n) {
					changed = true
				}
				n.ForEachChild(writes)
				return false
			}
			for _, file := range l.program.Files() {
				file.AsNode().ForEachChild(writes)
			}
			if changed {
				return failure()
			}
		}
		wanted := contract.AsTypePredicateNode()
		calls := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(parameter), checker.SignatureKindCall)
		var target *checker.Type
		if len(calls) == 1 {
			if predicate := predicateOfSignature(l.checker, calls[0]); predicate != nil {
				target = predicate.Type()
			}
		}
		if wanted.Type == nil || target == nil {
			return failure()
		}
		if implementation.Type() == nil {
			if !l.proveInferredPredicate(implementation, target) {
				return failure()
			}
		} else {
			annotation := implementation.Type().AsTypePredicateNode()
			if annotation.Type == nil || !checker.Checker_isTypeIdenticalTo(l.checker, l.checker.GetTypeAtLocation(annotation.Type), target) {
				return failure()
			}
			if !l.checkedPredicateArgumentProven(implementation.Type()) {
				return failure()
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
