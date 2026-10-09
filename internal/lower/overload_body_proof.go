package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Step 05's readonly structural result hatch is TypeScript-only. The existing
// fresh writable scalar-record boundary remains a checked conversion.
func (l *lowering) overloadBodyRequired(node *ast.Node, produced, promised *checker.Type) bool {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return false
	}
	failure := l.overloadResultFailure(produced, promised, "result", map[[2]*checker.Type]bool{})
	return strings.HasPrefix(failure.relation, "readonly covariance")
}

// Each path owns its bindings. The checker still supplies types and Adamic's
// relation decides results; only independently verified tests prune paths.
type overloadBodyState map[*ast.Symbol]*checker.Type

type overloadBodyProof struct {
	l                        *lowering
	implementation, overload *ast.Node
	promised                 *checker.Type
	parameters               map[*ast.Symbol]bool
	aliases                  map[*ast.Symbol]*ast.Node
	failure                  string
	steps                    int
}

func (l *lowering) overloadBodyResult(implementation, overload *ast.Node) (bool, string) {
	p := &overloadBodyProof{l: l, implementation: implementation, overload: overload,
		promised:   l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))),
		parameters: map[*ast.Symbol]bool{}, aliases: map[*ast.Symbol]*ast.Node{}}
	if len(implementation.TypeParameters()) != 0 || len(overload.TypeParameters()) != 0 {
		return false, ""
	}
	state := overloadBodyState{}
	for index, parameter := range implementation.Parameters() {
		if !ast.IsIdentifier(parameter.Name()) || parameter.Name().Text() == "this" || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return false, ""
		}
		given, _, rest, err := l.censusOverloadParameter(overload.Parameters(), index)
		if err != nil || rest {
			return false, ""
		}
		if given == nil {
			given = l.checker.GetUndefinedType()
		}
		symbol := l.symbol(parameter.Name())
		p.parameters[symbol] = true
		state[symbol] = given
	}
	// Parameter writes and captured writes invalidate the entry assumptions. A
	// local assignment in this body is instead interpreted at its reaching path.
	valid := true
	var scan func(*ast.Node, bool)
	scan = func(node *ast.Node, nested bool) {
		nested = nested || ast.IsFunctionLike(node)
		if ast.IsIdentifier(node) && ast.IsAssignmentTarget(node) && (p.parameters[l.symbol(node)] || nested) {
			valid = false
		}
		if node.Kind == ast.KindThisKeyword {
			valid = false
		}
		node.ForEachChild(func(child *ast.Node) bool { scan(child, nested); return false })
	}
	implementation.Body().ForEachChild(func(node *ast.Node) bool { scan(node, false); return false })
	if !valid {
		return false, ""
	}
	remaining, ok := p.statement(implementation.Body(), []overloadBodyState{state})
	if !ok {
		return false, p.failure
	}
	for range remaining {
		if !p.returned(nil, implementation.Body(), nil) {
			return false, p.failure
		}
	}
	return true, ""
}

func overloadBodyCopy(state overloadBodyState) overloadBodyState {
	result := overloadBodyState{}
	for symbol, held := range state {
		result[symbol] = held
	}
	return result
}

func (p *overloadBodyProof) expression(node *ast.Node, state overloadBodyState) (*checker.Type, bool) {
	if node == nil {
		return p.l.checker.GetUndefinedType(), true
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindIdentifier:
		if held := state[p.l.symbol(node)]; held != nil {
			return held, true
		}
		if p.l.symbol(node) != p.l.checker.GetUndefinedSymbol() {
			return nil, false
		}
	case ast.KindCallExpression:
		if !p.safeCall(node, map[*ast.Node]bool{}) {
			return nil, false
		}
		for _, argument := range node.AsCallExpression().Arguments.Nodes {
			if _, ok := p.expression(argument, state); !ok {
				return nil, false
			}
		}
	case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindNewExpression, ast.KindArrowFunction, ast.KindFunctionExpression:
		return nil, false
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		held, ok := p.expression(access.Expression, state)
		if !ok {
			return nil, false
		}
		property := p.l.checker.GetPropertyOfType(held, access.Name().Text())
		if property == nil || accessorSymbol(property) || p.possibleAccessor(access.Name().Text()) || node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return nil, false
		}
		return p.l.concrete(p.l.checker.GetTypeOfSymbol(property)), true
	default:
		if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) || node.Kind == ast.KindPostfixUnaryExpression {
			return nil, false
		}
		if node.Kind == ast.KindPrefixUnaryExpression {
			op := node.AsPrefixUnaryExpression().Operator
			if op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
				return nil, false
			}
		}
		safe := true
		node.ForEachChild(func(child *ast.Node) bool {
			// Property names are not variable reads. Only value children need evidence.
			if ast.IsDeclarationName(child) || ast.IsPartOfTypeNode(child) {
				return false
			}
			_, ok := p.expression(child, state)
			safe = safe && ok
			return false
		})
		if !safe {
			return nil, false
		}
	}
	return p.l.concrete(p.l.checker.GetTypeAtLocation(node)), true
}

// A fixed ordinary helper contract has already been checked by lowering. Only
// helpers without calls, writes, getters or captures may preserve our facts.
// Reject recursion rather than assuming the summary being established.
func (p *overloadBodyProof) safeCall(node *ast.Node, active map[*ast.Node]bool) bool {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsIdentifier(callee) {
		return false
	}
	signature := p.l.checker.GetResolvedSignature(node)
	if signature == nil || len(signature.TypeParameters()) != 0 {
		return false
	}
	declaration := signature.Declaration()
	if declaration == nil || declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil || len(declaration.TypeParameters()) != 0 || active[declaration] {
		return false
	}
	symbol := p.l.symbol(callee)
	if symbol == nil {
		return false
	}
	for _, candidate := range symbol.Declarations {
		if candidate.Kind != ast.KindFunctionDeclaration || candidate.Body() == nil {
			return false
		}
	}
	active[declaration] = true
	defer delete(active, declaration)
	safe := true
	var visit ast.Visitor
	visit = func(current *ast.Node) bool {
		if ast.IsFunctionLike(current) || current.Kind == ast.KindAsExpression || current.Kind == ast.KindTypeAssertionExpression || current.Kind == ast.KindNonNullExpression || current.Kind == ast.KindThisKeyword || current.Kind == ast.KindNewExpression {
			safe = false
			return false
		}
		propertyName := current.Parent != nil && current.Parent.Kind == ast.KindPropertyAccessExpression && current.Parent.Name() == current
		if ast.IsIdentifier(current) && !propertyName && !ast.IsDeclarationName(current) && !ast.IsPartOfTypeNode(current) {
			own := p.l.symbol(current)
			if own != p.l.checker.GetUndefinedSymbol() {
				local := false
				if own != nil {
					for _, binding := range own.Declarations {
						for parent := binding.Parent; parent != nil; parent = parent.Parent {
							if parent == declaration {
								local = true
								break
							}
							if ast.IsFunctionLike(parent) {
								break
							}
						}
					}
				}
				// Direct helper callees are validated separately, never read as values.
				if !local && !(current.Parent != nil && current.Parent.Kind == ast.KindCallExpression && current.Parent.AsCallExpression().Expression == current) {
					safe = false
				}
			}
		}
		if ast.IsAssignmentTarget(current) {
			safe = false
		}
		if current.Kind == ast.KindCallExpression {
			safe = safe && p.safeCall(current, active)
		}
		if current.Kind == ast.KindPropertyAccessExpression {
			property := p.l.checker.GetSymbolAtLocation(current)
			if property == nil || accessorSymbol(property) || p.possibleAccessor(current.Name().Text()) {
				safe = false
			}
		}
		current.ForEachChild(visit)
		return false
	}
	declaration.Body().ForEachChild(visit)
	return safe
}

// An interface field can hide a class accessor. Without receiver provenance,
// conservatively reject matching accessors anywhere in the closed source tree.
func (p *overloadBodyProof) possibleAccessor(name string) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if (node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor) && node.Name() != nil {
			if !ast.IsIdentifier(node.Name()) || node.Name().Text() == name {
				found = true
			}
		}
		if !found {
			node.ForEachChild(visit)
		}
		return false
	}
	for _, file := range p.l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return found
}

func (p *overloadBodyProof) branch(node *ast.Node, state overloadBodyState, truth bool) (overloadBodyState, bool) {
	node = ast.SkipParentheses(node)
	if _, ok := p.expression(node, state); !ok {
		return nil, false
	}
	if ast.IsIdentifier(node) {
		if alias := p.aliases[p.l.symbol(node)]; alias != nil {
			return p.branch(alias, state, truth)
		}
	}
	if node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		return p.branch(node.AsPrefixUnaryExpression().Operand, state, !truth)
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			equal := truth == (binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken)
			left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
			if p.literal(left) {
				left, right = right, left
			}
			if p.literal(right) {
				return p.selectValue(left, p.l.checker.GetTypeAtLocation(right), state, equal)
			}
		}
	}
	// An unrecognized pure condition splits both ways without narrowing.
	held, ok := p.expression(node, state)
	if !ok {
		return nil, false
	}
	if held.Flags()&checker.TypeFlagsBooleanLiteral != 0 && held.AsLiteralType().Value() != truth {
		return nil, true
	}
	return overloadBodyCopy(state), true
}

func (p *overloadBodyProof) literal(node *ast.Node) bool {
	return (&predicateFlowProof{l: p.l}).literal(node)
}

func (p *overloadBodyProof) selectValue(reference *ast.Node, value *checker.Type, state overloadBodyState, equal bool) (overloadBodyState, bool) {
	root, field := reference, ""
	if reference.Kind == ast.KindPropertyAccessExpression {
		access := reference.AsPropertyAccessExpression()
		root, field = ast.SkipParentheses(access.Expression), access.Name().Text()
	}
	if !ast.IsIdentifier(root) || state[p.l.symbol(root)] == nil {
		return overloadBodyCopy(state), true
	}
	held := state[p.l.symbol(root)]
	members := []*checker.Type{held}
	if held.Flags()&checker.TypeFlagsUnion != 0 {
		members = held.Types()
	}
	kept := []*checker.Type{}
	for _, member := range members {
		actual := member
		if field != "" {
			property := p.l.checker.GetPropertyOfType(member, field)
			if property == nil || accessorSymbol(property) || !p.l.checker.IsReadonlySymbol(property) || property.Flags&ast.SymbolFlagsOptional != 0 {
				return nil, false
			}
			actual = p.l.checker.GetTypeOfSymbol(property)
		}
		// Only disjoint scalar types can exclude an equality path, and only
		// singleton literals can exclude its inequality path. Broad types survive.
		scalar := p.l.overloadScalar(actual) && p.l.overloadScalar(value)
		overlaps := !scalar || p.l.checker.IsTypeAssignableTo(actual, value) || p.l.checker.IsTypeAssignableTo(value, actual)
		singleton := actual.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
		identical := scalar && singleton && checker.Checker_isTypeIdenticalTo(p.l.checker, actual, value)
		if equal && overlaps || !equal && !identical {
			kept = append(kept, member)
		}
	}
	if len(kept) == 0 {
		return nil, true
	}
	result := overloadBodyCopy(state)
	result[p.l.symbol(root)] = p.l.checker.GetUnionType(kept)
	return result, true
}

func (p *overloadBodyProof) returned(expression, where *ast.Node, state overloadBodyState) bool {
	if expression != nil && ast.SkipParentheses(expression).Kind == ast.KindCallExpression && p.l.overloadReturnCall(ast.SkipParentheses(expression), p.promised) {
		return true
	}
	held, known := p.expression(expression, state)
	if known && p.l.censusRelated(held, p.promised) && p.l.overloadReturnStorage(held, p.promised) {
		return true
	}
	path := "result"
	if known {
		path = p.l.overloadResultPath(held, p.promised, "result", map[[2]*checker.Type]bool{})
	}
	text := "undefined"
	if expression != nil {
		file := ast.GetSourceFileOfNode(expression)
		text = strings.TrimSpace(file.Text()[scanner.GetTokenPosOfNode(expression, file, false):expression.End()])
	}
	p.failure = fmt.Sprintf("; reachable return %s at %s cannot prove %s", text, p.l.program.Where(where), path)
	return false
}

func (p *overloadBodyProof) statement(node *ast.Node, states []overloadBodyState) ([]overloadBodyState, bool) {
	if len(states) == 0 {
		return nil, true
	}
	p.steps++
	if p.steps > 1024 || len(states) > 256 {
		return nil, false
	}
	switch node.Kind {
	case ast.KindBlock:
		for _, child := range node.AsBlock().Statements.Nodes {
			var ok bool
			states, ok = p.statement(child, states)
			if !ok {
				return nil, false
			}
		}
		return states, true
	case ast.KindVariableStatement:
		for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if !ast.IsIdentifier(declaration.Name()) {
				return nil, false
			}
			initializer := declaration.AsVariableDeclaration().Initializer
			symbol := p.l.symbol(declaration.Name())
			for _, state := range states {
				held, ok := p.expression(initializer, state)
				if !ok {
					return nil, false
				}
				state[symbol] = held
			}
			if declaration.Parent.Flags&ast.NodeFlagsConst != 0 && initializer != nil {
				// Only parameter tests can be replayed: their roots never change.
				for parameter := range p.parameters {
					proof := predicateFlowProof{l: p.l, parameter: parameter.Declarations[0]}
					if proof.check(initializer) {
						p.aliases[symbol] = initializer
					}
				}
			}
		}
		return states, true
	case ast.KindExpressionStatement:
		expression := ast.SkipParentheses(node.Expression())
		if expression.Kind != ast.KindBinaryExpression || expression.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
			return nil, false
		}
		assignment := expression.AsBinaryExpression()
		if !ast.IsIdentifier(assignment.Left) {
			return nil, false
		}
		symbol := p.l.symbol(assignment.Left)
		if p.parameters[symbol] {
			return nil, false
		}
		for _, state := range states {
			if state[symbol] == nil {
				return nil, false
			}
			held, ok := p.expression(assignment.Right, state)
			if !ok {
				return nil, false
			}
			state[symbol] = held
		}
		delete(p.aliases, symbol)
		return states, true
	case ast.KindIfStatement:
		branch := node.AsIfStatement()
		then, otherwise := []overloadBodyState{}, []overloadBodyState{}
		for _, state := range states {
			yes, ok := p.branch(branch.Expression, state, true)
			if !ok {
				return nil, false
			}
			if yes != nil {
				then = append(then, yes)
			}
			no, ok := p.branch(branch.Expression, state, false)
			if !ok {
				return nil, false
			}
			if no != nil {
				otherwise = append(otherwise, no)
			}
		}
		then, ok := p.statement(branch.ThenStatement, then)
		if !ok {
			return nil, false
		}
		if branch.ElseStatement != nil {
			otherwise, ok = p.statement(branch.ElseStatement, otherwise)
			if !ok {
				return nil, false
			}
		}
		return append(then, otherwise...), true
	case ast.KindSwitchStatement:
		branch := node.AsSwitchStatement()
		for _, state := range states {
			if _, ok := p.expression(branch.Expression, state); !ok {
				return nil, false
			}
		}
		pending := states
		entries := make([][]overloadBodyState, len(branch.CaseBlock.AsCaseBlock().Clauses.Nodes))
		defaultIndex := -1
		for index, clause := range branch.CaseBlock.AsCaseBlock().Clauses.Nodes {
			if clause.Kind == ast.KindDefaultClause {
				defaultIndex = index
				continue
			}
			if !p.literal(clause.Expression()) {
				return nil, false
			}
			remaining := []overloadBodyState{}
			for _, state := range pending {
				yes, ok := p.selectValue(branch.Expression, p.l.checker.GetTypeAtLocation(clause.Expression()), state, true)
				if !ok {
					return nil, false
				}
				if yes != nil {
					entries[index] = append(entries[index], yes)
				}
				no, ok := p.selectValue(branch.Expression, p.l.checker.GetTypeAtLocation(clause.Expression()), state, false)
				if !ok {
					return nil, false
				}
				if no != nil {
					remaining = append(remaining, no)
				}
			}
			pending = remaining
		}
		if defaultIndex >= 0 {
			entries[defaultIndex] = pending
			pending = nil
		}
		fallen := []overloadBodyState{}
		exits := pending
		for index, clause := range branch.CaseBlock.AsCaseBlock().Clauses.Nodes {
			fallen = append(fallen, entries[index]...)
			var statements []*ast.Node
			if clause.Kind == ast.KindCaseClause {
				statements = clause.AsCaseOrDefaultClause().Statements.Nodes
			} else {
				statements = clause.AsCaseOrDefaultClause().Statements.Nodes
			}
			for _, child := range statements {
				if child.Kind == ast.KindBreakStatement {
					if child.AsBreakStatement().Label != nil {
						return nil, false
					}
					exits = append(exits, fallen...)
					fallen = nil
					break
				}
				var ok bool
				fallen, ok = p.statement(child, fallen)
				if !ok {
					return nil, false
				}
			}
		}
		return append(exits, fallen...), true
	case ast.KindReturnStatement:
		for _, state := range states {
			if !p.returned(node.AsReturnStatement().Expression, node, state) {
				return nil, false
			}
		}
		return nil, true
	case ast.KindThrowStatement:
		// Without a supported catch/finally here, evaluating or throwing the
		// operand can only leave abruptly. Neither outcome returns a value.
		return nil, true
	case ast.KindEmptyStatement:
		return states, true
	}
	return nil, false
}
