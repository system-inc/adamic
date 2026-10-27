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
	if parent.Kind == ast.KindPrefixUnaryExpression && parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
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
		return (operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken) && realpathTruthiness(parent)
	}
	return false
}

func (l *lowering) realpathInitializer(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if l.nodeRealpathNative(node) {
		return true
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		return l.realpathInitializer(conditional.WhenTrue) || l.realpathInitializer(conditional.WhenFalse)
	}
	return false
}

func (l *lowering) realpathIdentity(node *ast.Node) bool {
	node = realpathParent(node)
	if node.Parent == nil || node.Parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.Parent.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	other := binary.Left
	if other == node {
		other = binary.Right
	}
	other = ast.SkipParentheses(other)
	if l.nodeRealpathNative(other) {
		return true
	}
	if !ast.IsIdentifier(other) {
		return false
	}
	symbol := l.symbol(other)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	return declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil && l.realpathInitializer(declaration.AsVariableDeclaration().Initializer)
}

func (l *lowering) realpathUseError(node *ast.Node, use string) error {
	return l.notYet(node, "node:fs.realpathSync.native as a value is admitted only as a const called with one path; this use "+use+" could pass it extra arguments Node reads as options")
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
func (l *lowering) nodeRealpathReferenceRefusal(node *ast.Node) error {
	use := realpathParent(node)
	if parent := use.Parent; parent != nil && parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == use {
		return nil
	}
	if realpathTruthiness(use) || l.realpathIdentity(use) {
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
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || declaration.AsVariableDeclaration().Initializer != use {
		return l.realpathUseError(node, realpathUseName(use))
	}
	binding := declaration.AsVariableDeclaration()
	// An unsupported plain-walk arm owns its diagnostic even if another use of
	// this binding would also fail the native value boundary.
	if plain := l.realpathPlainArm(binding.Initializer); plain != nil {
		return l.notYet(plain, nodeRealpathWalkNotYet)
	}
	if declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return l.realpathUseError(node, "is declared with let or var")
	}
	if !ast.IsIdentifier(binding.Name()) {
		return l.realpathUseError(node, "is stored in a binding pattern")
	}
	if ast.HasSyntacticModifier(declaration.Parent.Parent, ast.ModifierFlagsExport) {
		return l.realpathUseError(node, "is exported")
	}
	scope := ast.GetContainingFunction(declaration)
	if ast.GetContainingFunction(node) != scope {
		return l.realpathUseError(node, "crosses a function scope")
	}
	root := scope
	if root == nil {
		root = ast.GetSourceFileOfNode(declaration).AsNode()
	}
	symbol := l.symbol(binding.Name())
	if symbol == nil {
		return l.realpathUseError(node, "has no provable binding symbol")
	}
	var found error
	var visit ast.Visitor
	visit = func(reference *ast.Node) bool {
		if found != nil {
			return true
		}
		if ast.IsIdentifier(reference) && reference != binding.Name() && !ast.IsPartOfTypeNode(reference) {
			target := l.symbol(reference)
			if reference.Parent != nil && ast.IsShorthandPropertyAssignment(reference.Parent) {
				target = l.checker.GetShorthandAssignmentValueSymbol(reference.Parent)
			}
			if target == symbol {
				if err := l.realpathConstUse(reference); err != nil {
					found = err
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
func (l *lowering) realpathConstUse(reference *ast.Node) error {
	outer := realpathParent(reference)
	parent := outer.Parent
	if parent == nil {
		return l.realpathUseError(reference, "escapes its binding")
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		if call.Expression != outer {
			return l.realpathUseError(reference, "is passed as an argument")
		}
		arguments := call.Arguments.Nodes
		if len(arguments) != 1 {
			return l.realpathUseError(reference, "is called with zero or multiple arguments")
		}
		if arguments[0].Kind == ast.KindSpreadElement {
			return l.realpathUseError(reference, "is called with spread arguments")
		}
		if l.checker.GetTypeAtLocation(arguments[0]).Flags()&checker.TypeFlagsStringLike == 0 {
			return l.realpathUseError(reference, "is called with a non-string path")
		}
		return nil
	case ast.KindBinaryExpression:
		if l.realpathIdentity(outer) || realpathTruthiness(outer) {
			return nil
		}
	case ast.KindPrefixUnaryExpression, ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindConditionalExpression:
		if realpathTruthiness(outer) {
			return nil
		}
	}
	return l.realpathUseError(reference, realpathUseName(outer))
}

func (l *lowering) realpathPlainArm(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if l.nodeRealpathNative(node) {
		return nil
	}
	if l.nodeLibraryMember(node) == "node:fs.realpathSync" {
		return node
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		if plain := l.realpathPlainArm(conditional.WhenTrue); plain != nil {
			return plain
		}
		return l.realpathPlainArm(conditional.WhenFalse)
	}
	return nil
}
