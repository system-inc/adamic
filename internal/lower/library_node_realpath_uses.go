package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func realpathParent(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

func realpathTruthiness(node *ast.Node) bool {
	node = realpathParent(node)
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindPrefixUnaryExpression &&
		parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		return true
	}
	switch parent.Kind {
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindBinaryExpression:
		operator := parent.AsBinaryExpression().OperatorToken.Kind
		return (operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken) &&
			realpathTruthiness(parent)
	}
	return false
}

func (lowering *lowering) realpathInitializer(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if lowering.nodeRealpathNative(node) || lowering.nodeLibraryMember(node) == "node:fs.realpathSync" {
		return true
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		return lowering.realpathInitializer(conditional.WhenTrue) || lowering.realpathInitializer(conditional.WhenFalse)
	}
	return false
}

func (lowering *lowering) realpathIdentity(node *ast.Node) bool {
	node = realpathParent(node)
	if node.Parent == nil || node.Parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.Parent.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken &&
		binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	other := binary.Left
	if other == node {
		other = binary.Right
	}
	other = ast.SkipParentheses(other)
	if lowering.nodeRealpathNative(other) || lowering.nodeLibraryMember(other) == "node:fs.realpathSync" {
		return true
	}
	if !ast.IsIdentifier(other) {
		return false
	}
	symbol := lowering.symbol(other)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	return declaration.Kind == ast.KindVariableDeclaration &&
		declaration.Parent.Flags&ast.NodeFlagsConst != 0 &&
		declaration.AsVariableDeclaration().Initializer != nil &&
		lowering.realpathInitializer(declaration.AsVariableDeclaration().Initializer)
}

func (lowering *lowering) realpathUseError(node *ast.Node, use string) error {
	return lowering.notYet(
		node,
		"node:fs.realpathSync.native as a value is admitted only as a const called with one path; this use "+
			use+
			" could pass it extra arguments Node reads as options",
	)
}

func realpathUseName(node *ast.Node) string {
	node = realpathParent(node)
	parent := node.Parent
	if parent == nil {
		return "escapes its binding"
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		return "is passed as an argument"
	case ast.KindReturnStatement:
		return "is returned"
	case ast.KindPropertyAccessExpression:
		return "reads ." + parent.Name().Text()
	case ast.KindElementAccessExpression:
		return "reads a computed member"
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return "is stored in an object"
	case ast.KindArrayLiteralExpression:
		return "is stored in an array"
	case ast.KindVariableDeclaration:
		return "is stored in another binding"
	case ast.KindExportSpecifier, ast.KindExportAssignment:
		return "is exported"
	}
	return "appears in " + parent.Kind.String()
}

// Inspect the complete lexical scope, including nested functions so a capture
// cannot hide a callback escape. Symbols distinguish shadowed names; no shared
// lowering state or whole-program pass is needed.
func (lowering *lowering) nodeRealpathReferenceRefusal(node *ast.Node) error {
	use := realpathParent(node)
	if parent := use.Parent; parent != nil &&
		parent.Kind == ast.KindCallExpression &&
		parent.AsCallExpression().Expression == use {
		return nil
	}
	if realpathTruthiness(use) || lowering.realpathIdentity(use) {
		return nil
	}
	for use.Parent != nil && use.Parent.Kind == ast.KindConditionalExpression {
		conditional := use.Parent.AsConditionalExpression()
		if conditional.WhenTrue != use && conditional.WhenFalse != use {
			break
		}
		use = realpathParent(use.Parent)
	}
	declaration := use.Parent
	if declaration == nil ||
		declaration.Kind != ast.KindVariableDeclaration ||
		declaration.AsVariableDeclaration().Initializer != use {
		return lowering.realpathUseError(node, realpathUseName(use))
	}
	binding := declaration.AsVariableDeclaration()
	if declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return lowering.realpathUseError(node, "is declared with let or var")
	}
	if !ast.IsIdentifier(binding.Name()) {
		return lowering.realpathUseError(node, "is stored in a binding pattern")
	}
	if ast.HasSyntacticModifier(declaration.Parent.Parent, ast.ModifierFlagsExport) {
		return lowering.realpathUseError(node, "is exported")
	}
	scope := ast.GetContainingFunction(declaration)
	if ast.GetContainingFunction(node) != scope {
		return lowering.realpathUseError(node, "crosses a function scope")
	}
	root := scope
	if root == nil {
		root = ast.GetSourceFileOfNode(declaration).AsNode()
	}
	symbol := lowering.symbol(binding.Name())
	if symbol == nil {
		return lowering.realpathUseError(node, "has no provable binding symbol")
	}
	var found error
	var visit ast.Visitor
	visit = func(reference *ast.Node) bool {
		if found != nil {
			return true
		}
		if ast.IsIdentifier(reference) && reference != binding.Name() && !ast.IsPartOfTypeNode(reference) {
			target := lowering.symbol(reference)
			if reference.Parent != nil && ast.IsShorthandPropertyAssignment(reference.Parent) {
				target = lowering.checker.GetShorthandAssignmentValueSymbol(reference.Parent)
			}
			if target == symbol {
				if realpathConstUseError := lowering.realpathConstUse(reference); realpathConstUseError != nil {
					found = realpathConstUseError
					return true
				}
			}
		}
		reference.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	return found
}

// Only these use kinds consume the value without erasing its one-path calling
// convention. Every other syntax kind, including future kinds, fails closed.
func (lowering *lowering) realpathConstUse(reference *ast.Node) error {
	outer := realpathParent(reference)
	parent := outer.Parent
	if parent == nil {
		return lowering.realpathUseError(reference, "escapes its binding")
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		if call.Expression != outer {
			return lowering.realpathUseError(reference, "is passed as an argument")
		}
		arguments := call.Arguments.Nodes
		if len(arguments) != 1 {
			return lowering.realpathUseError(reference, "is called with zero or multiple arguments")
		}
		if arguments[0].Kind == ast.KindSpreadElement {
			return lowering.realpathUseError(reference, "is called with spread arguments")
		}
		if lowering.checker.GetTypeAtLocation(arguments[0]).Flags()&checker.TypeFlagsStringLike == 0 {
			return lowering.realpathUseError(reference, "is called with a non-string path")
		}
		return nil
	case ast.KindBinaryExpression:
		if lowering.realpathIdentity(outer) || realpathTruthiness(outer) {
			return nil
		}
	case ast.KindPrefixUnaryExpression, ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindConditionalExpression:
		if realpathTruthiness(outer) {
			return nil
		}
	}
	return lowering.realpathUseError(reference, realpathUseName(outer))
}
