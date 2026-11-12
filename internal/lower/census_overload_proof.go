package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (l *lowering) censusHasUndefined(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUndefined != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			return true
		}
	}
	return false
}

// Presence checks are sufficient only when all present members already satisfy
// the promise. They must not bless a brand or a mutable representation change.
func (l *lowering) censusNullableOverloadResult(produced, promised *checker.Type) bool {
	produced, promised = l.concrete(produced), l.concrete(promised)
	if !l.censusHasUndefined(produced) || l.censusHasUndefined(promised) {
		return false
	}
	for _, member := range produced.Types() {
		if member.Flags()&checker.TypeFlagsUndefined == 0 && !l.censusRelated(member, promised) {
			return false
		}
	}
	return produced.Flags()&checker.TypeFlagsUnion != 0
}

// Prove every reachable return under the overload's admitted parameter types.
// Only bare parameter tests supply flow facts. A write anywhere, including a
// captured write in a nested function, disables the proof. Unknown control flow
// keeps the call-site check rather than guessing that the annotation is true.
func (l *lowering) censusProveOverloadResult(implementation, overload *ast.Node) bool {
	return l.censusReturnProof(implementation, overload, nil, nil)
}

func (l *lowering) censusReturnProof(implementation, overload *ast.Node, admitted []*checker.Type, target *checker.Type) bool {
	outer := l.typeMapper
	defer func() { l.typeMapper = outer }()
	sources, targets := []*checker.Type{}, []*checker.Type{}
	for index, parameter := range implementation.TypeParameters() {
		if index >= len(overload.TypeParameters()) {
			return false
		}
		sources = append(sources, l.checker.GetTypeAtLocation(parameter.Name()))
		targets = append(targets, l.checker.GetTypeAtLocation(overload.TypeParameters()[index].Name()))
	}
	if len(sources) > 0 {
		l.typeMapper = newTypeMapper(sources, targets)
	}
	promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
	if target != nil {
		promised = target
	}
	parameters := implementation.Parameters()
	initial := make([]*checker.Type, len(parameters))
	for index, parameter := range parameters {
		if !ast.IsIdentifier(parameter.Name()) || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return false
		}
		given, _, rest, err := l.censusOverloadParameter(overload.Parameters(), index)
		if err != nil || rest {
			return false
		}
		if given == nil {
			given = l.checker.GetUndefinedType()
		}
		initial[index] = given
		if admitted != nil {
			initial[index] = l.checker.GetUndefinedType()
			if index < len(admitted) {
				initial[index] = admitted[index]
			}
		}
	}
	unchanged := true
	var scan ast.Visitor
	scan = func(node *ast.Node) bool {
		if node.Kind == ast.KindAsExpression || node.Kind == ast.KindTypeAssertionExpression || node.Kind == ast.KindNonNullExpression {
			unchanged = false
		}
		if ast.IsIdentifier(node) && ast.IsAssignmentTarget(node) {
			for _, parameter := range parameters {
				if l.symbol(node) == l.symbol(parameter.Name()) {
					unchanged = false
				}
			}
		}
		node.ForEachChild(scan)
		return false
	}
	implementation.Body().ForEachChild(scan)
	if !unchanged {
		return false
	}
	narrowed := func(index int, flow *ast.FlowNode) *checker.Type {
		parameter := parameters[index]
		declared := l.checker.GetTypeAtLocation(parameter.Name())
		return l.concrete(predicateFlowType(l.checker, parameter.Name(), declared, initial[index], implementation, flow))
	}
	reachable := func(flow *ast.FlowNode) bool {
		for index := range parameters {
			if narrowed(index, flow).Flags()&checker.TypeFlagsNever != 0 {
				return false
			}
		}
		return true
	}
	returned := func(expression *ast.Node, flow *ast.FlowNode) bool {
		if !reachable(flow) {
			return true
		}
		if expression == nil {
			return l.censusRelated(l.checker.GetUndefinedType(), promised)
		}
		expression = ast.SkipParentheses(expression)
		if ast.IsIdentifier(expression) {
			for index, parameter := range parameters {
				if l.symbol(expression) == l.symbol(parameter.Name()) {
					return l.censusRelated(narrowed(index, flow), promised)
				}
			}
			// Local aliases can be rebound by a call; they are not flow evidence.
			if l.symbol(expression) != l.checker.GetUndefinedSymbol() {
				return false
			}
		}
		if predicateEffects(expression) {
			return false
		}
		return l.censusRelated(l.checker.GetTypeAtLocation(expression), promised)
	}
	var statement func(*ast.Node, []*ast.FlowNode) ([]*ast.FlowNode, bool)
	statement = func(node *ast.Node, flows []*ast.FlowNode) ([]*ast.FlowNode, bool) {
		if len(flows) > 256 {
			return nil, false
		}
		switch node.Kind {
		case ast.KindBlock:
			for _, child := range node.AsBlock().Statements.Nodes {
				var ok bool
				flows, ok = statement(child, flows)
				if !ok {
					return nil, false
				}
			}
			return flows, true
		case ast.KindIfStatement:
			branch := node.AsIfStatement()
			if !l.censusTrustedTest(branch.Expression, parameters, initial) {
				return nil, false
			}
			then, otherwise := []*ast.FlowNode{}, []*ast.FlowNode{}
			for _, flow := range flows {
				for _, truth := range []bool{true, false} {
					path := predicateBranch(predicateFlowPath{flow: flow}, branch.Expression, truth).flow
					if reachable(path) {
						if truth {
							then = append(then, path)
						} else {
							otherwise = append(otherwise, path)
						}
					}
				}
			}
			then, ok := statement(branch.ThenStatement, then)
			if !ok {
				return nil, false
			}
			if branch.ElseStatement != nil {
				otherwise, ok = statement(branch.ElseStatement, otherwise)
				if !ok {
					return nil, false
				}
			}
			return append(then, otherwise...), true
		case ast.KindReturnStatement:
			for _, flow := range flows {
				if !returned(node.AsReturnStatement().Expression, flow) {
					return nil, false
				}
			}
			return nil, true
		case ast.KindThrowStatement:
			return nil, true
		case ast.KindExpressionStatement, ast.KindVariableStatement, ast.KindEmptyStatement:
			return flows, true
		}
		return nil, false
	}
	remaining, ok := statement(implementation.Body(), []*ast.FlowNode{{Flags: ast.FlowFlagsStart}})
	if !ok {
		return false
	}
	for _, flow := range remaining {
		if !returned(nil, flow) {
			return false
		}
	}
	return true
}

// censusTrustedTest says whether censusReturnProof may take an if's test as flow evidence. The branch
// types stay the checker's own narrowing of the whole test (predicateBranch), so a test is trusted only
// when every leaf narrows soundly or narrows nothing: a tree of !, &&, || and parentheses over a bare
// parameter compared with a literal by === or !==, a bare parameter by truthiness, or a parameter's
// length compared with a number literal where every member of every type the proof starts the parameter
// from has a length that is exactly number (censusLengthPlain). A call, in, instanceof or a user predicate stops the proof,
// and a write inside the test is the body's, which the proof's scan already refuses (#z5d0jye step 2,
// tsc's getSupportedExtensions and concatenate). predicateFlowProof.check (predicates.go) walks the same
// tree for a predicate's body proof over its own leaves.
func (l *lowering) censusTrustedTest(test *ast.Node, parameters []*ast.Node, initial []*checker.Type) bool {
	test = ast.SkipParentheses(test)
	parameterOf := func(node *ast.Node) int {
		node = ast.SkipParentheses(node)
		for index, parameter := range parameters {
			proof := predicateFlowProof{l: l, parameter: parameter}
			if proof.reference(node) {
				return index
			}
		}
		return -1
	}
	switch test.Kind {
	case ast.KindIdentifier:
		return parameterOf(test) >= 0
	case ast.KindPrefixUnaryExpression:
		unary := test.AsPrefixUnaryExpression()
		return unary.Operator == ast.KindExclamationToken && l.censusTrustedTest(unary.Operand, parameters, initial)
	case ast.KindBinaryExpression:
		binary := test.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			return l.censusTrustedTest(binary.Left, parameters, initial) && l.censusTrustedTest(binary.Right, parameters, initial)
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			for _, parameter := range parameters {
				proof := predicateFlowProof{l: l, parameter: parameter}
				if proof.reference(ast.SkipParentheses(binary.Left)) && proof.literal(binary.Right) || proof.reference(ast.SkipParentheses(binary.Right)) && proof.literal(binary.Left) {
					return true
				}
			}
			starts := func(index int) []*checker.Type {
				return []*checker.Type{l.checker.GetTypeAtLocation(parameters[index].Name()), initial[index]}
			}
			return l.censusLengthLeaf(binary.Left, binary.Right, parameterOf, starts) || l.censusLengthLeaf(binary.Right, binary.Left, parameterOf, starts)
		}
	}
	return false
}

// censusLengthLeaf says whether side reads a parameter's length and other is a number literal, where
// every member of the parameter's declared type has a length that is plain number. The checker narrows
// by length only where length is a discriminant, a member's length with a literal part (a tuple, or an
// object like { length: 1 }), so this leaf narrows nothing the proof reads (Lower's read on #z5d0jye).
// Both the declared type and the type the proof starts from are judged, since the narrowing runs from
// the overload's parameter or the call's argument (initial), not the implementation's declaration
// (Lower's must on 1fdc8083).
// p?.length narrows p and isn't this leaf.
func (l *lowering) censusLengthLeaf(side, other *ast.Node, parameterOf func(*ast.Node) int, starts func(int) []*checker.Type) bool {
	side, other = ast.SkipParentheses(side), ast.SkipParentheses(other)
	if side.Kind != ast.KindPropertyAccessExpression || other.Kind != ast.KindNumericLiteral {
		return false
	}
	access := side.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil || !ast.IsIdentifier(access.Name()) || access.Name().Text() != "length" {
		return false
	}
	index := parameterOf(access.Expression)
	if index < 0 {
		return false
	}
	for _, start := range starts(index) {
		if !l.censusLengthPlain(start, 0) {
			return false
		}
	}
	return true
}

// censusLengthPlain says whether every member of of, through unions, intersections and a type
// parameter's constraint, has a length whose type is exactly number. A nullish member has no length to
// narrow by. Anything it can't read fails closed.
func (l *lowering) censusLengthPlain(of *checker.Type, depth int) bool {
	if of == nil || depth > 16 {
		return false
	}
	switch {
	case of.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0:
		return true
	case of.Flags()&checker.TypeFlagsTypeParameter != 0:
		return l.censusLengthPlain(l.checker.GetBaseConstraintOfType(of), depth+1)
	case of.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range of.Types() {
			if !l.censusLengthPlain(member, depth+1) {
				return false
			}
		}
		return len(of.Types()) > 0
	case of.Flags()&checker.TypeFlagsIntersection != 0:
		found := false
		for _, part := range of.Types() {
			if l.checker.GetPropertyOfType(part, "length") == nil {
				continue
			}
			if !l.censusLengthPlain(part, depth+1) {
				return false
			}
			found = true
		}
		return found
	}
	property := l.checker.GetPropertyOfType(of, "length")
	return property != nil && l.checker.GetTypeOfSymbol(property).Flags() == checker.TypeFlagsNumber
}

// The checker retains the overload's declared result even when every admitted
// return is undefined. JSON needs no container metadata for that proven case,
// but still evaluates the original call, including its arguments and effects.
func (l *lowering) censusCallReturnsUndefined(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if !ast.IsIdentifier(ast.SkipParentheses(call.Expression)) {
		return false
	}
	resolved := l.checker.GetResolvedSignature(node)
	if resolved == nil || resolved.Declaration() == nil {
		return false
	}
	overload := resolved.Declaration()
	implementation := l.censusImplementation(overload)
	if implementation == nil {
		return false
	}
	admitted := make([]*checker.Type, len(call.Arguments.Nodes))
	for index, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return false
		}
		admitted[index] = l.checker.GetTypeAtLocation(argument)
		if admitted[index].Flags()&checker.TypeFlagsUndefined == 0 {
			return false
		}
	}
	return l.censusReturnProof(implementation, overload, admitted, l.checker.GetUndefinedType())
}
