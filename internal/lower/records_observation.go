package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A plain dictionary read has no getter effects. Its receiver and key still run,
// but a value discarded before any observation need not fetch an inherited member.
func (l *lowering) recordReadDiscarded(node *ast.Node) bool {
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if at.Parent == nil {
		return false
	}
	if at.Parent.Kind == ast.KindExpressionStatement {
		return true
	}
	if at.Parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := at.Parent
	if declaration.AsVariableDeclaration().Initializer != at || !ast.IsIdentifier(declaration.Name()) || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(declaration.Name())
	if symbol == nil || ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return false
	}
	// A following own check must run immediately, before any code can change the
	// dictionary. Only observations inside its true branch can use this snapshot.
	var guarded *ast.Node
	list := declaration.Parent
	if len(list.AsVariableDeclarationList().Declarations.Nodes) != 1 {
		return false
	}
	statement := list.Parent
	if statement != nil && statement.Parent != nil && statement.Parent.Kind == ast.KindBlock {
		statements := statement.Parent.AsBlock().Statements.Nodes
		for i, sibling := range statements {
			if sibling == statement && i+1 < len(statements) && statements[i+1].Kind == ast.KindIfStatement {
				check := statements[i+1].AsIfStatement()
				if l.recordOwnCheck(check.Expression, node) {
					guarded = check.ThenStatement
				}
			}
		}
	}
	safe := true
	var visit ast.Visitor
	visit = func(use *ast.Node) bool {
		if ast.IsIdentifier(use) && use != declaration.Name() && l.checker.GetSymbolAtLocation(use) == symbol {
			inside := false
			for parent := use.Parent; parent != nil; parent = parent.Parent {
				if parent == guarded {
					inside = true
					break
				}
			}
			if !inside {
				safe = false
			}
		}
		return use.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(node).AsNode().ForEachChild(visit)
	return safe
}

// A guard can also dominate the read itself. Walk only evaluation paths with
// no intervening side effect: the first statement, a return or binding, and the
// first argument of a call whose callee is a plain identifier or console.
func (l *lowering) recordReadOwnGuarded(node *ast.Node) bool {
	at := node
	for at.Parent != nil {
		parent := at.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != at || (!ast.IsIdentifier(callee) && !l.isConsole(callee)) {
				return false
			}
		case ast.KindReturnStatement, ast.KindExpressionStatement:
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer != at {
				return false
			}
		case ast.KindVariableDeclarationList:
			if len(parent.AsVariableDeclarationList().Declarations.Nodes) != 1 {
				return false
			}
		case ast.KindVariableStatement:
		case ast.KindBlock:
			statements := parent.AsBlock().Statements.Nodes
			if len(statements) == 0 || statements[0] != at {
				return false
			}
		case ast.KindIfStatement:
			check := parent.AsIfStatement()
			return check.ThenStatement == at && l.recordOwnCheck(check.Expression, node)
		default:
			return false
		}
		at = parent
	}
	return false
}

func (l *lowering) recordOwnCheck(condition, read *ast.Node) bool {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindCallExpression || read.Kind != ast.KindElementAccessExpression {
		return false
	}
	call := condition.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "hasOwn" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") || len(call.Arguments.Nodes) != 2 {
		return false
	}
	access := read.AsElementAccessExpression()
	same := func(a, b *ast.Node) bool {
		a, b = ast.SkipParentheses(a), ast.SkipParentheses(b)
		// Reading an identifier itself cannot call or mutate anything between the
		// snapshot and the immediately following library ownership check.
		if ast.IsIdentifier(a) && ast.IsIdentifier(b) {
			return l.checker.GetSymbolAtLocation(a) == l.checker.GetSymbolAtLocation(b)
		}
		literal := func(n *ast.Node) bool {
			return n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral || n.Kind == ast.KindNumericLiteral
		}
		return literal(a) && literal(b) && a.Text() == b.Text()
	}
	return same(call.Arguments.Nodes[0], access.Expression) && same(call.Arguments.Nodes[1], access.ArgumentExpression)
}

func (l *lowering) recordOwnRead(record, key ir.Expression, element ir.Type) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{record, key})
	r, k := b.read(b.parameters[0]), b.read(b.parameters[1])
	of := ir.Maybe(element)
	return b.finish("record_own_read", ir.Conditional{
		Condition: ir.RecordCall{Method: "hasOwn", Arguments: []ir.Expression{r, k}, Returns: ir.Boolean},
		WhenTrue:  ir.RecordCall{Method: "get", OwnOnly: true, Arguments: []ir.Expression{r, k}, Element: element, Returns: of},
		WhenNot:   fit(ir.Undefined{}, of), Of: of,
	})
}

// Strict equality with a present primitive cannot observe any Object.prototype
// value: those members are functions or the prototype object. This proves the
// compareDataObjects scalar bucket without guessing at any-valued object reads.
func (l *lowering) recordScalarComparison(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	comparison := node.AsBinaryExpression()
	operator := comparison.OperatorToken.Kind
	if operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
		return nil, false, nil
	}
	access, other := comparison.Left, comparison.Right
	reversed := false
	if !l.recordTarget(access) {
		access, other, reversed = other, access, true
	}
	if !l.recordTarget(access) {
		return nil, false, nil
	}
	proven := l.checker.GetTypeAtLocation(other)
	if proven.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) == 0 || l.includesUndefined(proven) {
		return nil, false, nil
	}
	r, k, element, err := l.recordAccess(access, false)
	if err != nil {
		return nil, true, err
	}
	value, err := l.expression(other)
	if err != nil {
		return nil, true, err
	}
	arguments := []ir.Expression{r, k}
	if reversed {
		arguments = []ir.Expression{value, r, k}
	}
	b := l.libraryArrayBuilder(arguments)
	offset := 0
	if reversed {
		offset = 1
	}
	record, key := b.read(b.parameters[offset]), b.read(b.parameters[offset+1])
	present := b.declare("own", ir.RecordCall{Method: "hasOwn", Arguments: []ir.Expression{record, key}, Returns: ir.Boolean})
	// Snapshot before evaluating the right operand, which may delete or overwrite.
	observed := b.declare("value", ir.Conditional{Condition: b.read(present), WhenTrue: ir.RecordCall{Method: "get", OwnOnly: true, Arguments: []ir.Expression{record, key}, Element: element, Returns: ir.Maybe(element)}, WhenNot: fit(ir.Undefined{}, ir.Maybe(element)), Of: ir.Maybe(element)})
	var otherRead ir.Expression
	if reversed {
		otherRead = b.read(b.parameters[0])
	} else {
		otherRead = b.read(b.declare("other", value))
	}
	result, err := l.combine(node, operator, b.read(observed), otherRead)
	if err != nil {
		return nil, true, err
	}
	return b.finish("record_scalar_comparison", ir.Conditional{Condition: b.read(present), WhenTrue: result, WhenNot: ir.BooleanConstant{Value: operator == ast.KindExclamationEqualsEqualsToken}, Of: ir.Boolean}), true, nil
}
