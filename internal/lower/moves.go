package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The prototype selects mutable arrays of records without recursively testing
// their callback types. The construction proof below judges the actual graph.
// The prototype proves exclusivity by construction. It does not mistake the
// cycle proof's confinement, or reuse's root count, for whole-graph ownership.
func (l *lowering) parallelMoves(node *ast.Node) bool {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 2 {
		return false
	}
	items := l.concrete(l.checker.GetTypeAtLocation(arguments[0]))
	if !l.checker.IsArrayType(items) || l.isLibraryType(items, "ReadonlyArray") {
		return false
	}
	for _, element := range l.checker.GetTypeArguments(items) {
		if element.Flags()&checker.TypeFlagsObject != 0 {
			return true
		}
	}
	return false
}

func (l *lowering) moveRefused(where *ast.Node, what string) error {
	return &Refused{Where: l.program.Where(where), What: what, Fix: "give each task an independent fresh object graph and use only the returned results; the general ownership proof is not implemented"}
}

func moveBlock(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindBlock || parent.Kind == ast.KindSourceFile {
			return parent
		}
	}
	return nil
}

func movePath(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		return movePath(node.AsPropertyAccessExpression().Expression) + "." + node.Name().Text()
	}
	if node.Kind == ast.KindElementAccessExpression {
		return movePath(node.AsElementAccessExpression().Expression) + "[]"
	}
	if ast.IsIdentifier(node) {
		return node.Text()
	}
	return "items"
}

// Source references are deliberately counted across the whole function and its
// closures, not just after the call. No earlier alias or repeated transfer can
// disappear because its binding looks dead at the source boundary.
func (l *lowering) checkMove(node *ast.Node) error {
	arguments := node.AsCallExpression().Arguments.Nodes
	items := ast.SkipParentheses(arguments[0])
	proof := &parallelProof{l: l}
	declaration := proof.declaration(items)
	if !ast.IsIdentifier(items) || declaration == nil || declaration.Kind != ast.KindVariableDeclaration || !proof.immutable(declaration) {
		return l.moveRefused(items, "cannot move items: whole reachable ownership is not proven")
	}
	insideFunction := false
	for parent := declaration.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLike(parent) {
			insideFunction = true
			break
		}
	}
	block := moveBlock(declaration)
	if !insideFunction || block == nil || block.Kind != ast.KindBlock || moveBlock(node) != block || declaration.Pos() >= node.Pos() {
		return l.moveRefused(items, "cannot move "+items.Text()+": the owner must be a local declared in the call's block")
	}
	// A complete initializer has no evaluation after the transfer which might
	// throw or reach the source. Nested expression calls are outside this slice.
	if node.Parent.Kind != ast.KindVariableDeclaration || node.Parent.AsVariableDeclaration().Initializer != node {
		return l.moveRefused(node, "cannot move "+items.Text()+": parallelMap must be a complete variable initializer")
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil || ast.SkipParentheses(initializer).Kind != ast.KindArrayLiteralExpression {
		return l.moveRefused(items, "cannot move "+items.Text()+": whole reachable ownership is not proven")
	}
	symbol := proof.symbol(items)
	var found error
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if found != nil || ast.IsPartOfTypeNode(child) {
			return false
		}
		if parallelReference(child) && proof.symbol(child) == symbol && child != items {
			path := child
			for path.Parent != nil && (path.Parent.Kind == ast.KindPropertyAccessExpression || path.Parent.Kind == ast.KindElementAccessExpression) {
				path = path.Parent
			}
			if child.Pos() > node.End() {
				found = l.moveRefused(child, "use after move: "+movePath(path)+"; "+items.Text()+" was moved into parallelMap")
			} else {
				found = l.moveRefused(child, "cannot move "+movePath(path)+": another variable, field or closure may reach the graph")
			}
			return false
		}
		return child.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(node).AsNode().ForEachChild(visit)
	if found != nil {
		return found
	}
	elements := ast.SkipParentheses(initializer).AsArrayLiteralExpression().Elements.Nodes
	for index, element := range elements {
		path := fmt.Sprintf("%s[%d]", items.Text(), index)
		element = ast.SkipParentheses(element)
		if element.Kind != ast.KindObjectLiteralExpression {
			return l.moveRefused(element, "cannot move "+path+": element is not a fresh object literal with only scalar literal fields")
		}
		for _, property := range element.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind != ast.KindPropertyAssignment || !ast.IsIdentifier(property.Name()) {
				return l.moveRefused(property, "cannot move "+path+": spread, shorthand and computed fields do not prove ownership")
			}
			value := ast.SkipParentheses(property.AsPropertyAssignment().Initializer)
			if value.Kind != ast.KindNumericLiteral && value.Kind != ast.KindTrueKeyword && value.Kind != ast.KindFalseKeyword {
				return l.moveRefused(value, "cannot move "+path+"."+property.Name().Text()+": field is not a numeric or boolean literal")
			}
		}
	}
	return l.checkMoveWork(arguments[1])
}

func (l *lowering) checkMoveWork(work *ast.Node) error {
	work = ast.SkipParentheses(work)
	if work.Kind != ast.KindArrowFunction || ast.HasSyntacticModifier(work, ast.ModifierFlagsAsync) || len(work.Parameters()) == 0 || len(work.Parameters()) > 2 || work.Body().Kind != ast.KindBlock {
		return l.moveRefused(work, "cannot move task result: work must be a synchronous inline arrow returning its owned item")
	}
	for _, parameter := range work.Parameters() {
		if !ast.IsIdentifier(parameter.Name()) || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return l.moveRefused(parameter, "cannot move task result: parameters must be plain bindings without defaults or rest")
		}
	}
	parameter := work.Parameters()[0]
	proof := &parallelProof{l: l}
	item := l.symbol(parameter.Name())
	var found error
	returned := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil || ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsFunctionLike(node) || node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindThrowStatement {
			found = l.moveRefused(node, "cannot move task result: calls, closures, this and throws are outside the move prototype")
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			value := node.AsReturnStatement().Expression
			if value == nil || !ast.IsIdentifier(ast.SkipParentheses(value)) || proof.symbol(ast.SkipParentheses(value)) != item {
				found = l.moveRefused(node, "cannot move task result: return the owned item")
			} else {
				returned = true
			}
			return false
		}
		if target := parallelWrite(node); target != nil {
			target = ast.SkipParentheses(target)
			local := proof.declaration(target)
			allowed := ast.IsIdentifier(target) && local != nil && parallelInside(local, work) && proof.symbol(target) != item
			if target.Kind == ast.KindPropertyAccessExpression {
				receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
				allowed = ast.IsIdentifier(receiver) && proof.symbol(receiver) == item
			}
			if !allowed {
				found = l.moveRefused(target, "cannot move task result: write only scalar item fields and task locals")
				return false
			}
		}
		if parallelReference(node) {
			local := proof.declaration(node)
			if proof.symbol(node) == item {
				if node.Parent.Kind != ast.KindPropertyAccessExpression || node.Parent.AsPropertyAccessExpression().Expression != node {
					found = l.moveRefused(node, "cannot move task result: the owned item may only be read through scalar fields or returned")
					return false
				}
			} else if local == nil || !parallelInside(local, work) {
				found = l.moveRefused(node, "cannot move task result: captures are outside the move prototype")
				return false
			}
		}
		if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindVariableDeclaration {
			scalar := l.checker.GetTypeAtLocation(node)
			if node.Kind == ast.KindVariableDeclaration {
				scalar = l.checker.GetTypeAtLocation(node.Name())
			}
			if scalar.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) == 0 {
				found = l.moveRefused(node, "cannot move task result: reference fields and locals are outside the move prototype")
				return false
			}
		}
		return node.ForEachChild(visit)
	}
	visit(work.Body())
	if found != nil {
		return found
	}
	if !returned {
		return l.moveRefused(work, "cannot move task result: return the owned item")
	}
	return nil
}
