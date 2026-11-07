package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A plain dictionary read has no getter effects. Its receiver and key still run,
// but a value discarded before any observation need not fetch an inherited member.
func (lowering *lowering) recordReadDiscarded(node *ast.Node) bool {
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
	if declaration.AsVariableDeclaration().Initializer != at ||
		!ast.IsIdentifier(declaration.Name()) ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	symbol := lowering.checker.GetSymbolAtLocation(declaration.Name())
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
		for index, sibling := range statements {
			if sibling == statement && index+1 < len(statements) && statements[index+1].Kind == ast.KindIfStatement {
				check := statements[index+1].AsIfStatement()
				if lowering.recordOwnCheck(check.Expression, node) {
					guarded = check.ThenStatement
				}
			}
		}
	}
	safe := true
	var visit ast.Visitor
	visit = func(use *ast.Node) bool {
		if ast.IsIdentifier(use) && use != declaration.Name() && lowering.checker.GetSymbolAtLocation(use) == symbol {
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
func (lowering *lowering) recordReadOwnGuarded(node *ast.Node) bool {
	at := node
	for at.Parent != nil {
		parent := at.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if len(call.Arguments.Nodes) == 0 ||
				call.Arguments.Nodes[0] != at ||
				(!ast.IsIdentifier(callee) && !lowering.isConsole(callee)) {
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
			return check.ThenStatement == at && lowering.recordOwnCheck(check.Expression, node)
		default:
			return false
		}
		at = parent
	}
	return false
}

func (lowering *lowering) recordOwnCheck(condition, read *ast.Node) bool {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindCallExpression || read.Kind != ast.KindElementAccessExpression {
		return false
	}
	call := condition.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	ordinary := callee.Kind == ast.KindPropertyAccessExpression &&
		callee.Name().Text() == "hasOwn" &&
		lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object")
	if (!ordinary && lowering.detachedOwnCallReceiver(condition) == nil) || len(call.Arguments.Nodes) != 2 {
		return false
	}
	access := read.AsElementAccessExpression()
	same := func(left, right *ast.Node) bool {
		left, right = ast.SkipParentheses(left), ast.SkipParentheses(right)
		// Reading an identifier itself cannot call or mutate anything between the
		// snapshot and the immediately following library ownership check.
		if ast.IsIdentifier(left) && ast.IsIdentifier(right) {
			return lowering.checker.GetSymbolAtLocation(left) == lowering.checker.GetSymbolAtLocation(right)
		}
		literal := func(node *ast.Node) bool {
			return node.Kind == ast.KindStringLiteral ||
				node.Kind == ast.KindNoSubstitutionTemplateLiteral ||
				node.Kind == ast.KindNumericLiteral
		}
		return literal(left) && literal(right) && left.Text() == right.Text()
	}
	return same(call.Arguments.Nodes[0], access.Expression) && same(call.Arguments.Nodes[1], access.ArgumentExpression)
}

func (lowering *lowering) recordOwnRead(record, key ir.Expression, element ir.Type) ir.Expression {
	builder := lowering.libraryArrayBuilder([]ir.Expression{record, key})
	r, k := builder.read(builder.parameters[0]), builder.read(builder.parameters[1])
	of := ir.Maybe(element)
	return builder.finish("record_own_read", ir.Conditional{
		Condition: ir.RecordCall{Method: "hasOwn", Arguments: []ir.Expression{r, k}, Returns: ir.Boolean},
		WhenTrue: ir.RecordCall{
			Method:    "get",
			OwnOnly:   true,
			Arguments: []ir.Expression{r, k},
			Element:   element,
			Returns:   of,
		},
		WhenNot: fit(ir.Undefined{}, of), Of: of,
	})
}

// Strict equality with a present primitive cannot observe any Object.prototype
// value: those members are functions or the prototype object. This proves the
// compareDataObjects scalar bucket without guessing at any-valued object reads.
func (lowering *lowering) recordScalarComparison(node *ast.Node) (ir.Expression, bool, error) {
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
	if !lowering.recordTarget(access) {
		access, other, reversed = other, access, true
	}
	if !lowering.recordTarget(access) {
		return nil, false, nil
	}
	proven := lowering.checker.GetTypeAtLocation(other)
	if proven.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) == 0 ||
		lowering.includesUndefined(proven) {
		return nil, false, nil
	}
	r, k, element, recordAccessError := lowering.recordAccess(access, false)
	if recordAccessError != nil {
		return nil, true, recordAccessError
	}
	value, expressionError := lowering.expression(other)
	if expressionError != nil {
		return nil, true, expressionError
	}
	arguments := []ir.Expression{r, k}
	if reversed {
		arguments = []ir.Expression{value, r, k}
	}
	builder := lowering.libraryArrayBuilder(arguments)
	offset := 0
	if reversed {
		offset = 1
	}
	record, key := builder.read(builder.parameters[offset]), builder.read(builder.parameters[offset+1])
	present := builder.declare(
		"own",
		ir.RecordCall{Method: "hasOwn", Arguments: []ir.Expression{record, key}, Returns: ir.Boolean},
	)
	// Snapshot before evaluating the right operand, which may delete or overwrite.
	observed := builder.declare("value", ir.Conditional{Condition: builder.read(present), WhenTrue: ir.RecordCall{
		Method:    "get",
		OwnOnly:   true,
		Arguments: []ir.Expression{record, key},
		Element:   element,
		Returns:   ir.Maybe(element),
	}, WhenNot: fit(ir.Undefined{}, ir.Maybe(element)), Of: ir.Maybe(element)})
	var otherRead ir.Expression
	if reversed {
		otherRead = builder.read(builder.parameters[0])
	} else {
		otherRead = builder.read(builder.declare("other", value))
	}
	result, combineError := lowering.combine(node, operator, builder.read(observed), otherRead)
	if combineError != nil {
		return nil, true, combineError
	}
	return builder.finish("record_scalar_comparison", ir.Conditional{
		Condition: builder.read(present),
		WhenTrue:  result,
		WhenNot:   ir.BooleanConstant{Value: operator == ast.KindExclamationEqualsEqualsToken},
		Of:        ir.Boolean,
	}), true, nil
}
