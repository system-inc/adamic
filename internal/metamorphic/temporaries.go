package metamorphic

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// temporarySites split and inline. A split takes a statement's own expression (an initializer, a
// returned value, a called expression statement, an assignment's value) and moves the operands it
// computes, each into a const declared just before the statement, in the order they were evaluated: a
// binary operator's two sides (not && || ??, whose right side may not run), a call's arguments, a
// template's parts. An inline is the other way round: a const with no annotation, read exactly once,
// in the next statement, where nothing between could run it twice or not at all, goes back in
// parentheses where it was read.
func temporarySites(p *program) []site {
	found := p.references()
	var sites []site
	visit(p.file.AsNode(), func(node *ast.Node) bool {
		if !inStatementList(node) {
			return true
		}
		if change, ok := p.splitSite(node); ok {
			sites = append(sites, change)
		}
		if change, ok := p.inlineSite(node, found); ok {
			sites = append(sites, change)
		}
		return true
	})
	sort.SliceStable(sites, func(i, j int) bool { return sites[i].edits[0].start < sites[j].edits[0].start })
	return sites
}

// splitSite moves the operands a statement's root expression computes into temporaries.
func (p *program) splitSite(statement *ast.Node) (site, bool) {
	var root *ast.Node
	switch statement.Kind {
	case ast.KindVariableStatement:
		declarations := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		if len(declarations) == 1 {
			root = declarations[0].Initializer()
		}
	case ast.KindReturnStatement:
		root = statement.Expression()
	case ast.KindExpressionStatement:
		root = statement.Expression()
		if root.Kind == ast.KindBinaryExpression {
			binary := root.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left.Kind == ast.KindIdentifier {
				root = binary.Right
			} else if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				root = nil
			}
		}
	}
	if root == nil {
		return site{}, false
	}
	for root.Kind == ast.KindParenthesizedExpression {
		root = root.Expression()
	}
	var operands []*ast.Node
	switch root.Kind {
	case ast.KindBinaryExpression:
		binary := root.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if ast.IsAssignmentOperator(operator) || ast.IsLogicalOrCoalescingBinaryOperator(operator) || operator == ast.KindCommaToken {
			return site{}, false
		}
		operands = []*ast.Node{binary.Left, binary.Right}
	case ast.KindCallExpression, ast.KindNewExpression:
		if root.Flags&ast.NodeFlagsOptionalChain != 0 || !stableCallee(root.Expression()) || root.ArgumentList() == nil {
			return site{}, false
		}
		operands = root.Arguments()
	case ast.KindTemplateExpression:
		for _, span := range root.AsTemplateExpression().TemplateSpans.Nodes {
			operands = append(operands, span.Expression())
		}
	default:
		return site{}, false
	}
	indentation := p.indentation(statement)
	var declarations strings.Builder
	var edits []edit
	for _, operand := range operands {
		if operand.Kind == ast.KindSpreadElement || !splittable(operand) {
			continue
		}
		name := p.fresh("metamorphicTemporary")
		declarations.WriteString("const " + name + " = " + p.source(operand) + ";\n" + indentation)
		edits = append(edits, edit{p.start(operand), operand.End(), name})
	}
	if len(edits) == 0 {
		return site{}, false
	}
	start := p.start(statement)
	return site{edits: append([]edit{{start, start, declarations.String()}}, edits...)}, true
}

// stableCallee is a callee whose evaluation reads only names: evaluated after its arguments instead of
// before, it's the same.
func stableCallee(callee *ast.Node) bool {
	switch callee.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword:
		return true
	case ast.KindPropertyAccessExpression:
		return callee.Flags&ast.NodeFlagsOptionalChain == 0 && stableCallee(callee.Expression())
	}
	return false
}

// splittable is an operand worth a temporary: something computed, not a name or a literal, and not a
// literal of an array, an object or a function, whose type comes from where it's used.
func splittable(operand *ast.Node) bool {
	switch operand.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindBinaryExpression, ast.KindTemplateExpression, ast.KindConditionalExpression, ast.KindParenthesizedExpression,
		ast.KindPrefixUnaryExpression, ast.KindTypeOfExpression:
		return true
	}
	return false
}

// inlineSite puts a const read once, in the very next statement, back where it was read.
func (p *program) inlineSite(statement *ast.Node, found *references) (site, bool) {
	if statement.Kind != ast.KindVariableStatement {
		return site{}, false
	}
	list := statement.AsVariableStatement().DeclarationList
	declarations := list.AsVariableDeclarationList().Declarations.Nodes
	if list.Flags&ast.NodeFlagsConst == 0 || len(declarations) != 1 {
		return site{}, false
	}
	declaration := declarations[0]
	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier || declaration.Type() != nil || declaration.Initializer() == nil {
		return site{}, false
	}
	symbol := found.symbolOf[name]
	if symbol == nil {
		return site{}, false
	}
	reads := found.reads(symbol)
	if len(reads) != 1 || reads[0].shorthand || reads[0].written {
		return site{}, false
	}
	next := nextStatement(statement)
	if next == nil || !within(reads[0].node, next) || !evaluatedOnce(reads[0].node, next) {
		return site{}, false
	}
	read := reads[0].node
	initializer := declaration.Initializer()
	return site{edits: []edit{
		{statement.Pos(), statement.End(), ""},
		{p.start(read), read.End(), "(" + p.source(initializer) + ")"},
	}}, true
}

func nextStatement(statement *ast.Node) *ast.Node {
	list := statement.Parent.StatementList()
	for index, sibling := range list.Nodes {
		if sibling == statement && index+1 < len(list.Nodes) {
			return list.Nodes[index+1]
		}
	}
	return nil
}

// evaluatedOnce says whether an expression inside a statement runs exactly once when the statement
// does: every step up to the statement is one that always evaluates the part it holds.
func evaluatedOnce(node *ast.Node, statement *ast.Node) bool {
	for child, parent := node, node.Parent; child != statement; child, parent = parent, parent.Parent {
		if parent == nil || ast.IsFunctionLike(parent) {
			return false
		}
		switch parent.Kind {
		case ast.KindVariableStatement, ast.KindVariableDeclarationList, ast.KindVariableDeclaration, ast.KindExpressionStatement,
			ast.KindReturnStatement, ast.KindThrowStatement, ast.KindParenthesizedExpression, ast.KindTemplateExpression,
			ast.KindTemplateSpan, ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindPropertyAssignment,
			ast.KindSpreadElement, ast.KindTypeOfExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
		case ast.KindCallExpression, ast.KindNewExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			// An optional chain may stop before a later part.
			if parent.Flags&ast.NodeFlagsOptionalChain != 0 && parent.Expression() != child {
				return false
			}
		case ast.KindPrefixUnaryExpression:
			operator := parent.AsPrefixUnaryExpression().Operator
			if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
				return false
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) && binary.Right == child {
				return false
			}
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left == child {
				return false
			}
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition != child {
				return false
			}
		case ast.KindIfStatement, ast.KindSwitchStatement:
			if parent.Expression() != child {
				return false
			}
		default:
			return false
		}
	}
	return true
}
