package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// An intrinsic alias needs no runtime slot when every use preserves its known
// identity. Never replace a first-class intrinsic with an ordinary user closure:
// that would change its metadata, identity and explicit receiver semantics.
func (l *lowering) libraryArrayAliasDeclaration(declaration *ast.Node) bool {
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil || ast.SkipParentheses(initializer).Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := l.libraryArrayMethodName(initializer)
	if name == "" {
		return false
	}
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsBlockScoped == 0 {
		return false
	}
	statement := list.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement || ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
		return false
	}
	if statement.Parent == nil || statement.Parent.Kind != ast.KindSourceFile && statement.Parent.Kind != ast.KindBlock {
		return false // Case clauses and loop headers need separate initialization proofs.
	}
	symbol := l.symbol(declaration.Name())
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	proven := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !proven {
			return true
		}
		if node.Kind == ast.KindShorthandPropertyAssignment {
			value := l.checker.GetShorthandAssignmentValueSymbol(node)
			if value != nil && l.checker.GetExportSymbolOfSymbol(value) == symbol {
				proven = false // A shorthand stores the value, not its property symbol.
				return true
			}
		}
		if ast.IsIdentifier(node) && l.symbol(node) == symbol && node != declaration.Name() {
			outer := libraryArrayOuter(node)
			parent := outer.Parent
			// A type query has no value read, including before initialization.
			if parent != nil && parent.Kind == ast.KindTypeQuery {
				return false
			}
			if !libraryArrayInitializedUse(declaration, node) {
				proven = false
				return true
			}
			if parent != nil && parent.Kind == ast.KindTypeOfExpression {
				return false
			}
			if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == outer && parent.AsPropertyAccessExpression().QuestionDotToken == nil {
				switch parent.Name().Text() {
				case "name", "length", "prototype":
					if libraryArrayReadOnlyUse(parent) {
						return false
					}
				case "call":
					if name != "isArray" && called(parent) {
						return false
					}
				}
			}
			if name == "isArray" && called(outer) {
				return false
			}
			proven = false // Assignment, escape, identity or unsupported invocation.
			return true
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return proven
}

func (l *lowering) libraryArrayAliasName(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return ""
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return ""
	}
	declaration := symbol.Declarations[0]
	if !l.libraryArrayAliasDeclaration(declaration) {
		return ""
	}
	return l.libraryArrayMethodName(declaration.AsVariableDeclaration().Initializer)
}

func (l *lowering) libraryArrayAliasInitializer(node *ast.Node) bool {
	outer := libraryArrayOuter(node)
	return outer.Parent != nil && outer.Parent.Kind == ast.KindVariableDeclaration && outer.Parent.AsVariableDeclaration().Initializer == outer && l.libraryArrayAliasDeclaration(outer.Parent)
}

func libraryArrayOuter(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// Folding a read must never turn a store or update into an observation.
func libraryArrayReadOnlyUse(node *ast.Node) bool {
	outer := libraryArrayOuter(node)
	parent := outer.Parent
	if parent == nil {
		return true
	}
	if parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == outer {
		operator := parent.AsBinaryExpression().OperatorToken.Kind
		_, compound := compoundAssignments[operator]
		if operator == ast.KindEqualsToken || compound {
			return false
		}
	}
	if parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression {
		return false
	}
	if parent.Kind == ast.KindForInStatement || parent.Kind == ast.KindForOfStatement {
		return parent.AsForInOrOfStatement().Initializer != outer
	}
	return true
}

// Ordinary blocks initialize in source order. Hoisted function/method bodies
// can run before a later binding, so only nested function expressions created afterwards are
// admitted. Exported and cross-module aliases need a module initialization proof.
func libraryArrayInitializedUse(declaration, use *ast.Node) bool {
	if ast.GetSourceFileOfNode(declaration) != ast.GetSourceFileOfNode(use) || use.Pos() < declaration.End() {
		return false
	}
	owner := declaration.Parent
	for owner != nil && !ast.IsFunctionLike(owner) {
		owner = owner.Parent
	}
	for parent := use.Parent; parent != nil && parent != owner; parent = parent.Parent {
		if ast.IsFunctionLike(parent) && parent.Kind != ast.KindArrowFunction && parent.Kind != ast.KindFunctionExpression {
			return false
		}
	}
	return true
}
