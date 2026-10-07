package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A tag proof certifies membership in the tag set only. Admission must preserve
// a checked view for the target's remaining fields; this is not a shape proof.
type predicateProof struct{ TaggedView bool }

type predicateTruth uint8

const (
	predicateTrue   predicateTruth = 1
	predicateFalse  predicateTruth = 2
	predicateEither predicateTruth = 3
)

type predicateVerifier struct {
	l         *lowering
	field     string
	cells     []string
	summaries map[*ast.Node][]predicateTruth
}

type predicatePath struct {
	cell   int
	locals map[*ast.Symbol]predicateTruth
}

func predicateFailure(l *lowering, node *ast.Node, path string) error {
	where := ""
	if node != nil {
		where = l.program.Where(node)
	}
	return &Refused{Where: where, What: "an unproven type predicate", Fix: path + "; return a boolean and narrow at the caller (adamic/no-type-predicate)"}
}

// provePredicate starts with no helper facts. Only independently verified bodies
// become summaries. Iteration terminates when no new body can be discharged;
// cycles and opaque calls never acquire a summary from their annotations.
func (l *lowering) provePredicate(node *ast.Node) (predicateProof, error) {
	proof := predicateProof{}
	if node == nil || node.Kind != ast.KindTypePredicate || node.Parent == nil || !ast.IsFunctionLike(node.Parent) || node.Parent.Body() == nil {
		return proof, predicateFailure(l, node, "no implementation body proves a return path")
	}
	v := &predicateVerifier{l: l, summaries: map[*ast.Node][]predicateTruth{}}
	if node.AsTypePredicateNode().Type == nil {
		return proof, predicateFailure(l, node, "an assertion without a target type is not proved")
	}
	target := l.checker.GetTypeAtLocation(node.AsTypePredicateNode().Type)
	if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion) != 0 && v.kindTarget(target) {
		v.field = "kind"
		proof.TaggedView = true
	} else if !predicatePrimitiveTarget(target) {
		return proof, predicateFailure(l, node, "the target has an unsupported runtime contract")
	}
	if v.field == "" {
		v.cells = []string{"string", "number", "boolean", "undefined", "object", "function", "symbol", "bigint"}
	} else {
		v.cells = []string{"<other kind>"}
	}
	var declarations []*ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if ast.IsFunctionLike(n) && n.Type() != nil && n.Type().Kind == ast.KindTypePredicate && n.Body() != nil {
			declarations = append(declarations, n)
			if v.field != "" {
				v.collectKinds(n)
			}
		}
		n.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	var last error
	for {
		changed := false
		for _, declaration := range declarations {
			if _, ok := v.summaries[declaration]; ok {
				continue
			}
			summary, err := v.verify(declaration)
			if declaration == node.Parent {
				last = err
			}
			if err == nil {
				v.summaries[declaration] = summary
				changed = true
			}
		}
		if _, ok := v.summaries[node.Parent]; ok {
			return proof, nil
		}
		if !changed {
			break
		}
	}
	if last == nil {
		last = predicateFailure(l, node, "unconverged helper proof")
	}
	return predicateProof{}, last
}

func predicatePrimitiveTarget(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, m := range t.Types() {
			if !predicatePrimitiveTarget(m) {
				return false
			}
		}
		return true
	}
	// Literal refinements need value partitions, not just typeof partitions.
	return t.Flags()&(checker.TypeFlagsString|checker.TypeFlagsNumber|checker.TypeFlagsBoolean|checker.TypeFlagsUndefined) != 0 && t.Flags()&(checker.TypeFlagsLiteral|checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter) == 0
}

func (v *predicateVerifier) kindTarget(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, m := range t.Types() {
			if !v.kindTarget(m) {
				return false
			}
		}
		return true
	}
	return t.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(t) && v.l.fieldLiteral(t, "kind") != nil
}

func predicateLiteral(t *checker.Type) (string, bool) {
	if t.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral) == 0 {
		return "", false
	}
	return fmt.Sprintf("%T:%v", t.AsLiteralType().Value(), t.AsLiteralType().Value()), true
}

func (v *predicateVerifier) addKind(t *checker.Type) {
	key, ok := predicateLiteral(t)
	if !ok {
		return
	}
	for _, cell := range v.cells {
		if cell == key {
			return
		}
	}
	v.cells = append(v.cells, key)
}

func (v *predicateVerifier) collectKinds(declaration *ast.Node) {
	var addTarget func(*checker.Type)
	addTarget = func(t *checker.Type) {
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			for _, m := range t.Types() {
				addTarget(m)
			}
			return
		}
		if literal := v.l.fieldLiteral(t, "kind"); literal != nil {
			v.addKind(literal)
		}
	}
	predicate := declaration.Type().AsTypePredicateNode()
	if predicate.Type != nil {
		addTarget(v.l.checker.GetTypeAtLocation(predicate.Type))
	}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindBinaryExpression {
			b := n.AsBinaryExpression()
			if b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
				v.addKind(v.l.checker.GetTypeAtLocation(b.Left))
				v.addKind(v.l.checker.GetTypeAtLocation(b.Right))
			}
		}
		n.ForEachChild(visit)
		return false
	}
	declaration.Body().ForEachChild(visit)
}

func (v *predicateVerifier) wanted(t *checker.Type, cell string) (bool, bool) {
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		answer := false
		for _, m := range t.Types() {
			yes, ok := v.wanted(m, cell)
			if !ok {
				return false, false
			}
			answer = answer || yes
		}
		return answer, true
	}
	if v.field != "" {
		if !v.kindTarget(t) {
			return false, false
		}
		key, ok := predicateLiteral(v.l.fieldLiteral(t, v.field))
		return key == cell, ok
	}
	if !predicatePrimitiveTarget(t) {
		return false, false
	}
	kinds := map[string]checker.TypeFlags{"string": checker.TypeFlagsString, "number": checker.TypeFlagsNumber, "boolean": checker.TypeFlagsBoolean, "undefined": checker.TypeFlagsUndefined}
	return t.Flags()&kinds[cell] != 0, true
}

func (v *predicateVerifier) parameter(declaration *ast.Node) *ast.Symbol {
	predicate := declaration.Type().AsTypePredicateNode()
	for _, p := range declaration.Parameters() {
		if ast.IsIdentifier(p.Name()) && p.Name().Text() == predicate.ParameterName.Text() {
			return v.l.checker.GetSymbolAtLocation(p.Name())
		}
	}
	return nil
}

func (v *predicateVerifier) verify(declaration *ast.Node) ([]predicateTruth, error) {
	parameter := v.parameter(declaration)
	if parameter == nil || len(declaration.TypeParameters()) != 0 {
		return nil, predicateFailure(v.l, declaration, "a generic or non-parameter target has no body proof")
	}
	for _, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
		if declared.Initializer != nil || declared.DotDotDotToken != nil || !ast.IsIdentifier(parameter.Name()) {
			return nil, predicateFailure(v.l, parameter, "parameter initialization may mutate the tested input before this return path")
		}
	}
	predicate := declaration.Type().AsTypePredicateNode()
	if predicate.Type == nil {
		return nil, predicateFailure(v.l, declaration, "an assertion without a target type is not proved")
	}
	target := v.l.checker.GetTypeAtLocation(predicate.Type)
	summary := make([]predicateTruth, len(v.cells))
	for i, cell := range v.cells {
		wanted, ok := v.wanted(target, cell)
		if !ok {
			return nil, predicateFailure(v.l, declaration, "the target has an unsupported runtime contract")
		}
		path := predicatePath{cell: i, locals: map[*ast.Symbol]predicateTruth{}}
		check := func(n *ast.Node, result predicateTruth) error {
			if predicate.AssertsModifier != nil {
				if !wanted {
					return predicateFailure(v.l, n, "normal return does not establish the asserted target")
				}
			} else {
				if result&predicateTrue != 0 && !wanted {
					return predicateFailure(v.l, n, "true return does not establish the target")
				}
				if result&predicateFalse != 0 && wanted {
					return predicateFailure(v.l, n, "false return does not exclude the target")
				}
			}
			summary[i] |= result
			return nil
		}
		body := declaration.Body()
		if body.Kind != ast.KindBlock {
			result, err := v.expression(body, parameter, path)
			if err != nil {
				return nil, err
			}
			if err = check(body, result); err != nil {
				return nil, err
			}
			continue
		}
		paths, err := v.statements(body.AsBlock().Statements.Nodes, parameter, []predicatePath{path}, check)
		if err != nil {
			return nil, err
		}
		if len(paths) > 0 {
			if predicate.AssertsModifier == nil {
				return nil, predicateFailure(v.l, body, "normal return has no boolean result")
			}
			if err = check(body, predicateTrue); err != nil {
				return nil, err
			}
		}
	}
	return summary, nil
}

func predicateNot(t predicateTruth) predicateTruth {
	return (t&predicateTrue)<<1 | (t&predicateFalse)>>1
}
func predicateCombine(a, b predicateTruth, and bool) predicateTruth {
	var result predicateTruth
	for _, x := range []predicateTruth{predicateTrue, predicateFalse} {
		for _, y := range []predicateTruth{predicateTrue, predicateFalse} {
			if a&x == 0 || b&y == 0 {
				continue
			}
			yes := x == predicateTrue || y == predicateTrue
			if and {
				yes = x == predicateTrue && y == predicateTrue
			}
			if yes {
				result |= predicateTrue
			} else {
				result |= predicateFalse
			}
		}
	}
	return result
}

func (v *predicateVerifier) expression(n *ast.Node, parameter *ast.Symbol, path predicatePath) (predicateTruth, error) {
	n = ast.SkipParentheses(n)
	switch n.Kind {
	case ast.KindTrueKeyword:
		return predicateTrue, nil
	case ast.KindFalseKeyword:
		return predicateFalse, nil
	case ast.KindIdentifier:
		if value, ok := path.locals[v.l.checker.GetSymbolAtLocation(n)]; ok {
			return value, nil
		}
		return predicateEither, nil
	case ast.KindPrefixUnaryExpression:
		u := n.AsPrefixUnaryExpression()
		if u.Operator == ast.KindExclamationToken {
			t, err := v.expression(u.Operand, parameter, path)
			return predicateNot(t), err
		}
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		if ast.IsAssignmentOperator(b.OperatorToken.Kind) {
			return 0, predicateFailure(v.l, n, "mutation invalidates the facts before this return path")
		}
		if b.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || b.OperatorToken.Kind == ast.KindBarBarToken {
			a, err := v.expression(b.Left, parameter, path)
			if err != nil {
				return 0, err
			}
			z, err := v.expression(b.Right, parameter, path)
			if err != nil {
				return 0, err
			}
			return predicateCombine(a, z, b.OperatorToken.Kind == ast.KindAmpersandAmpersandToken), nil
		}
		if b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			for _, pair := range [][2]*ast.Node{{b.Left, b.Right}, {b.Right, b.Left}} {
				test := ast.SkipParentheses(pair[0])
				constant := pair[1]
				matched := false
				yes := false
				if v.field == "" && test.Kind == ast.KindTypeOfExpression && v.sameParameter(test.AsTypeOfExpression().Expression, parameter) && constant.Kind == ast.KindStringLiteral {
					matched = true
					yes = v.cells[path.cell] == constant.Text()
				}
				if v.field != "" && test.Kind == ast.KindPropertyAccessExpression && test.Name().Text() == v.field && v.sameParameter(test.AsPropertyAccessExpression().Expression, parameter) {
					key, ok := predicateLiteral(v.l.checker.GetTypeAtLocation(constant))
					if ok && v.constant(constant) {
						matched = true
						yes = v.cells[path.cell] == key
					}
				}
				if matched {
					if b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
						yes = !yes
					}
					if yes {
						return predicateTrue, nil
					}
					return predicateFalse, nil
				}
			}
		}
		// Pure comparisons outside the known partition provide no narrowing fact.
		if b.OperatorToken.Kind == ast.KindGreaterThanToken || b.OperatorToken.Kind == ast.KindLessThanToken {
			if v.pure(b.Left) && v.pure(b.Right) {
				return predicateEither, nil
			}
		}
	case ast.KindCallExpression:
		call := n.AsCallExpression()
		if len(call.Arguments.Nodes) != 1 || !v.sameParameter(call.Arguments.Nodes[0], parameter) || !ast.IsIdentifier(call.Expression) {
			break
		}
		symbol := v.l.symbol(call.Expression)
		if symbol != nil {
			for _, d := range symbol.Declarations {
				if values, ok := v.summaries[d]; ok {
					return values[path.cell], nil
				}
			}
		}
		return 0, predicateFailure(v.l, n, "unconverged or opaque helper on this return path")
	}
	return 0, predicateFailure(v.l, n, "opaque test or possible mutation on this return path")
}

func (v *predicateVerifier) sameParameter(n *ast.Node, parameter *ast.Symbol) bool {
	n = ast.SkipParentheses(n)
	return ast.IsIdentifier(n) && v.l.checker.GetSymbolAtLocation(n) == parameter
}

func (v *predicateVerifier) constant(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	switch n.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindPropertyAccessExpression:
		symbol := v.l.checker.GetSymbolAtLocation(n)
		return symbol != nil && (symbol.Flags&ast.SymbolFlagsEnumMember != 0 || v.l.checker.IsReadonlySymbol(symbol))
	}
	return false
}

func (v *predicateVerifier) pure(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	return n.Kind == ast.KindIdentifier || n.Kind == ast.KindNumericLiteral || n.Kind == ast.KindStringLiteral
}

func copyPredicatePath(path predicatePath) predicatePath {
	copy := predicatePath{cell: path.cell, locals: map[*ast.Symbol]predicateTruth{}}
	for symbol, value := range path.locals {
		copy.locals[symbol] = value
	}
	return copy
}

func (v *predicateVerifier) statements(nodes []*ast.Node, parameter *ast.Symbol, paths []predicatePath, check func(*ast.Node, predicateTruth) error) ([]predicatePath, error) {
	for _, n := range nodes {
		var next []predicatePath
		for _, path := range paths {
			switch n.Kind {
			case ast.KindBlock:
				result, err := v.statements(n.AsBlock().Statements.Nodes, parameter, []predicatePath{path}, check)
				if err != nil {
					return nil, err
				}
				next = append(next, result...)
			case ast.KindReturnStatement:
				result := predicateTrue
				if expression := n.AsReturnStatement().Expression; expression != nil {
					value, err := v.expression(expression, parameter, path)
					if err != nil {
						return nil, err
					}
					result = value
				}
				if err := check(n, result); err != nil {
					return nil, err
				}
			case ast.KindIfStatement:
				branch := n.AsIfStatement()
				condition, err := v.expression(branch.Expression, parameter, path)
				if err != nil {
					return nil, err
				}
				for _, side := range []struct {
					truth predicateTruth
					body  *ast.Node
				}{{predicateTrue, branch.ThenStatement}, {predicateFalse, branch.ElseStatement}} {
					if condition&side.truth == 0 {
						continue
					}
					p := copyPredicatePath(path)
					if side.body == nil {
						next = append(next, p)
						continue
					}
					result, err := v.statements([]*ast.Node{side.body}, parameter, []predicatePath{p}, check)
					if err != nil {
						return nil, err
					}
					next = append(next, result...)
				}
			case ast.KindVariableStatement:
				for _, d := range n.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					variable := d.AsVariableDeclaration()
					if !ast.IsIdentifier(d.Name()) || variable.Initializer == nil {
						return nil, predicateFailure(v.l, n, "an unproven alias on this return path")
					}
					value, err := v.expression(variable.Initializer, parameter, path)
					if err != nil {
						return nil, err
					}
					path.locals[v.l.checker.GetSymbolAtLocation(d.Name())] = value
				}
				next = append(next, path)
			case ast.KindExpressionStatement:
				expression := n.AsExpressionStatement().Expression
				if v.l.isPanicCall(expression) {
					call := expression.AsCallExpression()
					for _, argument := range call.Arguments.Nodes {
						if !v.pure(argument) {
							return nil, predicateFailure(v.l, argument, "opaque panic argument may mutate this path")
						}
					}
					continue
				}
				_, err := v.expression(expression, parameter, path)
				if err != nil {
					return nil, err
				}
				next = append(next, path)
			default:
				return nil, predicateFailure(v.l, n, "unsupported control flow or mutation on this return path")
			}
		}
		paths = next
	}
	return paths, nil
}
