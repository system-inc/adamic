package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Literal partitions retain an outside cell. A helper's annotation never fills
// its summary: only its own effect-free returns supply values for these cells.
func predicateLiteralTarget(target *checker.Type) bool {
	if target == nil {
		return false
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !predicateLiteralTarget(member) {
				return false
			}
		}
		return true
	}
	_, ok := predicateLiteral(target)
	return ok
}
func predicateLiteralMember(target *checker.Type, cell string) (bool, bool) {
	if !predicateLiteralTarget(target) {
		return false, false
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			yes, _ := predicateLiteralMember(member, cell)
			if yes {
				return true, true
			}
		}
		return false, true
	}
	key, _ := predicateLiteral(target)
	return key == cell, true
}
func (v *predicateVerifier) kindType(target *checker.Type) *checker.Type {
	if target == nil || target.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	for _, field := range v.l.checker.GetPropertiesOfType(target) {
		if field.Name == "kind" {
			return v.l.checker.GetTypeOfSymbol(field)
		}
	}
	return nil
}
func (v *predicateVerifier) addLiteralKinds(target *checker.Type) {
	if target == nil {
		return
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			v.addLiteralKinds(member)
		}
		return
	}
	v.addKind(target)
}

// Follow only direct implementation identities reachable from this body. This
// bounds work and keeps summaries local to the original proof obligation.
func (v *predicateVerifier) dependencies(root *ast.Node) []*ast.Node {
	var declarations []*ast.Node
	seen := map[*ast.Node]bool{}
	var add func(*ast.Node)
	add = func(declaration *ast.Node) {
		if seen[declaration] || !ast.IsFunctionLike(declaration) || declaration.Body() == nil {
			return
		}
		seen[declaration] = true
		if declaration != root {
			if len(declaration.Parameters()) != 1 || len(declaration.TypeParameters()) != 0 {
				return
			}
			signature := v.l.checker.GetSignatureFromDeclaration(declaration)
			if signature == nil || v.l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsBooleanLike == 0 {
				return
			}
		}
		declarations = append(declarations, declaration)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsFunctionLike(node) {
				return false
			}
			if node.Kind == ast.KindCallExpression && len(node.AsCallExpression().Arguments.Nodes) == 1 {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if ast.IsIdentifier(callee) {
					if symbol := v.l.symbol(callee); symbol != nil {
						for _, helper := range symbol.Declarations {
							add(helper)
						}
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		if declaration.Body().Kind == ast.KindBlock {
			declaration.Body().ForEachChild(visit)
		} else {
			visit(declaration.Body())
		}
	}
	add(root)
	return declarations
}
func (v *predicateVerifier) scalarParameter(parameter *ast.Symbol) bool {
	if v.field == "" {
		return false
	}
	var scalar func(*checker.Type) bool
	scalar = func(target *checker.Type) bool {
		if target.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range target.Types() {
				if !scalar(member) {
					return false
				}
			}
			return true
		}
		return target.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) != 0
	}
	return scalar(v.l.checker.GetTypeOfSymbol(parameter))
}
func (v *predicateVerifier) delegated(node *ast.Node, parameter *ast.Symbol, path predicatePath) (predicateTruth, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if len(call.Arguments.Nodes) != 1 || !ast.IsIdentifier(callee) {
		return 0, predicateFailure(v.l, node, "a helper must test the same argument through a direct call")
	}
	symbol := v.l.symbol(callee)
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			summary, ok := v.summaries[declaration]
			if !ok {
				continue
			}
			tested := v.parameter(declaration)
			// Scalar tag helpers observe node.kind; object helpers observe node itself.
			scalar := v.scalarParameter(tested)
			matches := v.sameParameter(call.Arguments.Nodes[0], parameter)
			if v.field != "" {
				if scalar {
					matches = v.kindValue(call.Arguments.Nodes[0], parameter, path)
				} else if v.scalarParameter(parameter) {
					matches = false
				}
			}
			if !matches {
				return 0, predicateFailure(v.l, node, "a delegated helper tests a different argument identity")
			}
			return summary[path.cell], nil
		}
	}
	return 0, predicateFailure(v.l, node, "unconverged or opaque helper on this return path")
}

// Only fixed literal cases of the tested tag are discharged. Fallthrough is
// interpreted in source order; breaks return to the enclosing statement list.
func (v *predicateVerifier) switchPaths(node *ast.Node, parameter *ast.Symbol, path predicatePath, check func(*ast.Node, predicateTruth) error) ([]predicatePath, error) {
	statement := node.AsSwitchStatement()
	if v.field == "" || !v.kindValue(statement.Expression, parameter, path) {
		return nil, predicateFailure(v.l, node, "a switch does not test the unchanged discriminant")
	}
	clauses := statement.CaseBlock.AsCaseBlock().Clauses.Nodes
	start, otherwise := -1, -1
	for index, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			otherwise = index
			continue
		}
		expression := clause.AsCaseOrDefaultClause().Expression
		key, known := predicateLiteral(v.l.checker.GetTypeAtLocation(expression))
		if !known || !v.constant(expression) {
			return nil, predicateFailure(v.l, expression, "a switch case is not an independently fixed literal")
		}
		if key == v.cells[path.cell] && start < 0 {
			start = index
		}
	}
	if start < 0 {
		start = otherwise
	}
	if start < 0 {
		return []predicatePath{path}, nil
	}
	paths := []predicatePath{path}
	var finished []predicatePath
	for _, clause := range clauses[start:] {
		var statements []*ast.Node
		if clause.Kind == ast.KindDefaultClause {
			statements = clause.AsCaseOrDefaultClause().Statements.Nodes
		} else {
			statements = clause.AsCaseOrDefaultClause().Statements.Nodes
		}
		for _, statement := range statements {
			if statement.Kind == ast.KindBreakStatement && statement.AsBreakStatement().Label == nil {
				finished = append(finished, paths...)
				paths = nil
				break
			}
			var err error
			paths, err = v.statements([]*ast.Node{statement}, parameter, paths, check)
			if err != nil {
				return nil, err
			}
		}
		if len(paths) == 0 {
			break
		}
	}
	return append(finished, paths...), nil
}
