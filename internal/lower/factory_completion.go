package lower

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// FactoryCompletion is a bounded source proof, not an instruction or a runtime
// type certificate. Lowering must retain checks until layout and provenance are
// available. Unknown control flow deliberately clears every proof.
type FactoryCompletion struct {
	Required     []string
	BeforeEscape map[string]bool
	Escapes      []FactoryEscape
	Reads        []FactoryRead
	Unknown      bool
}

type FactoryEscape struct {
	Node    *ast.Node
	Missing []string
}

func (completion FactoryCompletion) Proven(field string) bool {
	return !completion.Unknown && len(completion.Escapes) > 0 && completion.BeforeEscape[field]
}

// FactoryRead preserves the state at the read, independently of a later completed return.
type FactoryRead struct {
	Node    *ast.Node
	Field   string
	Checked bool
}

type factoryState struct {
	aliases           map[*ast.Symbol]bool
	written           map[string]bool
	allocated, active bool
	escaped           bool
}

type factoryAnalysis struct {
	checker *checker.Checker
	fields  map[string]*checker.Type
	result  FactoryCompletion
}

// AnalyzeFactory follows one local literal allocation, typed writes, symbol
// aliases, straight-line blocks and if joins. Opaque calls can expose captured
// aliases even without an explicit argument. Helpers, loops, multiple allocations,
// closures and computed writes remain Unknown. No asserted type proves a write.
func AnalyzeFactory(factory *ast.Node, target *checker.Type, typeChecker *checker.Checker) FactoryCompletion {
	analysis := factoryAnalysis{
		checker: typeChecker,
		fields:  map[string]*checker.Type{},
		result:  FactoryCompletion{BeforeEscape: map[string]bool{}},
	}
	for _, field := range typeChecker.GetPropertiesOfType(target) {
		if field.Flags&ast.SymbolFlagsOptional == 0 {
			analysis.fields[field.Name] = typeChecker.GetTypeOfSymbol(field)
			analysis.result.Required = append(analysis.result.Required, field.Name)
			analysis.result.BeforeEscape[field.Name] = true
		}
	}
	sort.Strings(analysis.result.Required)
	if factory.Body() == nil || factory.Body().Kind != ast.KindBlock {
		analysis.result.Unknown = true
		return analysis.result
	}
	state := factoryState{aliases: map[*ast.Symbol]bool{}, written: map[string]bool{}, active: true}
	analysis.statement(factory.Body(), &state)
	if state.active || !state.allocated || len(analysis.result.Escapes) == 0 {
		analysis.result.Unknown = true
	}
	return analysis.result
}

func (analysis *factoryAnalysis) alias(node *ast.Node, state *factoryState) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	for node.Kind == ast.KindAsExpression {
		node = ast.SkipParentheses(node.AsAsExpression().Expression)
	}
	return ast.IsIdentifier(node) && state.aliases[analysis.checker.GetSymbolAtLocation(node)]
}

func (analysis *factoryAnalysis) escape(node *ast.Node, state *factoryState) {
	if !state.allocated {
		return
	}
	missing := []string{}
	for _, field := range analysis.result.Required {
		if !state.written[field] {
			missing = append(missing, field)
			analysis.result.BeforeEscape[field] = false
		}
	}
	analysis.result.Escapes = append(analysis.result.Escapes, FactoryEscape{Node: node, Missing: missing})
	state.escaped = true
}

func (analysis *factoryAnalysis) typed(value *ast.Node, field string) bool {
	wanted := analysis.fields[field]
	if wanted == nil {
		return false
	}
	forged := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindAsExpression {
			t := n.AsAsExpression().Type
			if t.Kind != ast.KindTypeReference || t.AsTypeReferenceNode().TypeName.Text() != "const" {
				forged = true
			}
		}
		if n.Kind == ast.KindTypeAssertionExpression || n.Kind == ast.KindNonNullExpression {
			forged = true
		}
		if !forged {
			n.ForEachChild(visit)
		}
		return forged
	}
	visit(value)
	actual := analysis.checker.GetTypeAtLocation(value)
	return !forged &&
		interfaceScalar(actual) &&
		interfaceScalar(wanted) &&
		actual.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter) == 0 &&
		analysis.checker.IsTypeAssignableTo(actual, wanted)
}

func (analysis *factoryAnalysis) expression(node *ast.Node, state *factoryState) {
	if node == nil {
		return
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if analysis.alias(node.AsPropertyAccessExpression().Expression, state) &&
			analysis.fields[node.Name().Text()] != nil {
			analysis.result.Reads = append(analysis.result.Reads, FactoryRead{
				Node:    node,
				Field:   node.Name().Text(),
				Checked: !state.written[node.Name().Text()] || state.escaped,
			})
		}
	case ast.KindConditionalExpression:
		analysis.result.Unknown = true
		analysis.escape(node, state)
		state.written = map[string]bool{}
		return
	case ast.KindCallExpression, ast.KindNewExpression:
		analysis.result.Unknown = true
		// Evaluate nested effects before this call's escape. With no effect summary,
		// a call could reach the object through a captured or previously exposed alias.
		node.ForEachChild(func(node *ast.Node) bool { analysis.expression(node, state); return false })
		analysis.escape(node, state)
		state.written = map[string]bool{}
		return
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		analysis.result.Unknown = true
		analysis.escape(node, state)
		return
	case ast.KindBinaryExpression:
		expression := node.AsBinaryExpression()
		if expression.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
			expression.OperatorToken.Kind == ast.KindBarBarToken ||
			expression.OperatorToken.Kind == ast.KindQuestionQuestionToken {
			analysis.result.Unknown = true
			analysis.escape(node, state)
			state.written = map[string]bool{}
			return
		}
		if ast.IsAssignmentOperator(expression.OperatorToken.Kind) {
			if expression.OperatorToken.Kind != ast.KindEqualsToken {
				analysis.result.Unknown = true
				analysis.escape(node, state)
				return
			}
			left := ast.SkipParentheses(expression.Left)
			if left.Kind == ast.KindPropertyAccessExpression &&
				analysis.alias(left.AsPropertyAccessExpression().Expression, state) {
				analysis.expression(expression.Right, state)
				field := left.Name().Text()
				state.written[field] = analysis.typed(expression.Right, field)
				return
			}
			analysis.expression(expression.Right, state)
			if ast.IsIdentifier(left) {
				symbol := analysis.checker.GetSymbolAtLocation(left)
				if symbol == nil {
					analysis.result.Unknown = true
					return
				}
				if analysis.alias(expression.Right, state) {
					analysis.result.Unknown = true
					analysis.escape(node, state)
				}
				delete(state.aliases, symbol)
				return
			}
			analysis.escape(node, state)
			analysis.result.Unknown = true
			return
		}
	}
	node.ForEachChild(func(node *ast.Node) bool { analysis.expression(node, state); return false })
}

func cloneFactoryState(s factoryState) factoryState {
	c := factoryState{
		aliases:   map[*ast.Symbol]bool{},
		written:   map[string]bool{},
		allocated: s.allocated,
		active:    s.active,
		escaped:   s.escaped,
	}
	for key, value := range s.aliases {
		c.aliases[key] = value
	}
	for key, value := range s.written {
		c.written[key] = value
	}
	return c
}

// Look at initializer types, never an asserted/contextual interface as evidence of payload.
func interfaceLiteralFields(node *ast.Node) (map[string]*ast.Node, bool) {
	fields := map[string]*ast.Node{}
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
			return nil, false
		}
		name := property.Name()
		if name == nil || (name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral) {
			return nil, false
		}
		value := name
		if property.Kind == ast.KindPropertyAssignment {
			value = property.AsPropertyAssignment().Initializer
		}
		fields[name.Text()] = value
	}
	return fields, true
}

func (analysis *factoryAnalysis) statement(node *ast.Node, state *factoryState) {
	if node == nil || !state.active {
		return
	}
	switch node.Kind {
	case ast.KindBlock:
		for _, n := range node.AsBlock().Statements.Nodes {
			analysis.statement(n, state)
		}
	case ast.KindVariableStatement:
		for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			variableDeclaration := declaration.AsVariableDeclaration()
			if !ast.IsIdentifier(variableDeclaration.Name()) || variableDeclaration.Initializer == nil {
				analysis.result.Unknown = true
				continue
			}
			symbol := analysis.checker.GetSymbolAtLocation(variableDeclaration.Name())
			if symbol == nil {
				analysis.result.Unknown = true
				continue
			}
			if analysis.alias(variableDeclaration.Initializer, state) {
				state.aliases[symbol] = true
				continue
			}
			init := ast.SkipParentheses(variableDeclaration.Initializer)
			for init.Kind == ast.KindAsExpression {
				init = ast.SkipParentheses(init.AsAsExpression().Expression)
			}
			if init.Kind == ast.KindObjectLiteralExpression {
				if state.allocated {
					analysis.result.Unknown = true
					continue
				}
				analysis.expression(init, state)
				state.allocated = true
				state.aliases[symbol] = true
				fields, known := interfaceLiteralFields(init)
				if !known {
					analysis.result.Unknown = true
					continue
				}
				for field, value := range fields {
					state.written[field] = analysis.typed(value, field)
				}
			} else {
				analysis.expression(variableDeclaration.Initializer, state)
			}
		}
	case ast.KindExpressionStatement:
		analysis.expression(node.AsExpressionStatement().Expression, state)
	case ast.KindReturnStatement:
		value := node.AsReturnStatement().Expression
		analysis.expression(value, state)
		if !analysis.alias(value, state) {
			analysis.result.Unknown = true
		}
		analysis.escape(node, state)
		state.active = false
	case ast.KindIfStatement:
		branch := node.AsIfStatement()
		analysis.expression(branch.Expression, state)
		yes, no := cloneFactoryState(*state), cloneFactoryState(*state)
		analysis.statement(branch.ThenStatement, &yes)
		analysis.statement(branch.ElseStatement, &no)
		if !yes.active {
			*state = no
			return
		}
		if !no.active {
			*state = yes
			return
		}
		for field := range yes.written {
			yes.written[field] = yes.written[field] && no.written[field]
		}
		for symbol := range yes.aliases {
			if !no.aliases[symbol] {
				delete(yes.aliases, symbol)
			}
		}
		yes.escaped = yes.escaped || no.escaped
		if yes.allocated != no.allocated {
			analysis.result.Unknown = true
		}
		*state = yes
	default:
		analysis.result.Unknown = true
		analysis.escape(node, state)
		state.written = map[string]bool{}
	}
}
