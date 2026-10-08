package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Class/interface unions hold structural values without manufacturing class identity.
func classInterfaceUnion(t *checker.Type) bool {
	if t == nil || t.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	class, structural := false, false
	for _, member := range t.Types() {
		if isClassInstance(member) {
			class = true
		} else if member.Symbol() != nil && member.Symbol().Flags&ast.SymbolFlagsInterface != 0 {
			structural = true
		} else {
			return false
		}
	}
	return class && structural
}

func (l *lowering) classInterfaceLiteral(node *ast.Node, target *checker.Type) bool {
	if node.Kind != ast.KindObjectLiteralExpression || !classInterfaceUnion(target) {
		return false
	}
	source := l.checker.GetTypeAtLocation(node)
	for _, member := range target.Types() {
		if isClassInstance(member) {
			continue
		}
		compatible := true
		for _, property := range l.checker.GetPropertiesOfType(member) {
			actual := l.checker.GetPropertyOfType(source, property.Name)
			if actual == nil {
				compatible = property.Flags&ast.SymbolFlagsOptional != 0
				if !compatible {
					break
				}
				continue
			}
			from, to := l.checker.GetTypeOfSymbol(actual), l.checker.GetTypeOfSymbol(property)
			if !l.checker.IsTypeAssignableTo(from, to) || l.nominalMismatch(from, to, map[[2]*checker.Type]bool{}) != nil {
				compatible = false
				break
			}
		}
		if compatible {
			return true
		}
	}
	return false
}

// Follow inferred aliases and forwarding results too: inference does not create a tag.
func (l *lowering) classInterfaceSource(node *ast.Node) *checker.Type {
	return l.classInterfaceOrigin(node, map[*ast.Node]bool{})
}

func (l *lowering) classInterfaceOrigin(node *ast.Node, seen map[*ast.Node]bool) *checker.Type {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if seen[node] {
		return nil
	}
	seen[node] = true
	if own := l.checker.GetTypeAtLocation(node); classInterfaceUnion(own) {
		return own
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression:
		symbol := l.symbol(node)
		if symbol == nil {
			return nil
		}
		if declared := l.concrete(l.checker.GetTypeOfSymbol(symbol)); classInterfaceUnion(declared) {
			return declared
		}
		if len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			switch declaration.Kind {
			case ast.KindVariableDeclaration, ast.KindPropertyDeclaration, ast.KindPropertyAssignment:
				return l.classInterfaceOrigin(declaration.Initializer(), seen)
			}
		}
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		symbol := l.symbol(callee)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return nil
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind == ast.KindVariableDeclaration {
			declaration = declaration.Initializer()
		}
		if declaration == nil {
			return nil
		}
		switch declaration.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
			body := declaration.Body()
			if body == nil {
				return nil
			}
			if body.Kind != ast.KindBlock {
				return l.classInterfaceOrigin(body, seen)
			}
			var found *checker.Type
			var visit ast.Visitor
			visit = func(part *ast.Node) bool {
				if found != nil {
					return true
				}
				if ast.IsFunctionLike(part) {
					return false
				}
				if part.Kind == ast.KindReturnStatement {
					found = l.classInterfaceOrigin(part.AsReturnStatement().Expression, seen)
					return found != nil
				}
				return part.ForEachChild(visit)
			}
			body.ForEachChild(visit)
			return found
		}
	}
	return nil
}

// Only fixed public member names of these object unions are supported here. The runtime
// query includes class prototypes and never evaluates a getter.
func (l *lowering) classInterfaceIn(node *ast.Node) bool {
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindInKeyword || binary.Left.Kind != ast.KindStringLiteral {
		return false
	}
	source := l.checker.GetTypeAtLocation(binary.Right)
	if !classInterfaceUnion(source) {
		source = l.classInterfaceSource(binary.Right)
	}
	if !classInterfaceUnion(source) {
		return false
	}
	for _, member := range source.Types() {
		if l.checker.GetPropertyOfType(member, binary.Left.Text()) != nil {
			return true
		}
	}
	return false
}

func (l *lowering) classInterfaceMethod(node *ast.Node) bool {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && l.classInterfaceSource(callee.AsPropertyAccessExpression().Expression) != nil
}

// A contextual class view is a checked downcast even when control flow called the
// union a class after a property-presence test. Member reads themselves keep the value.
func (l *lowering) classInterfaceRead(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindAsExpression {
		return value, nil
	}
	source := l.classInterfaceSource(node)
	target := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if source == nil || target == nil || !isClassInstance(target) {
		return value, nil
	}
	if l.classInterfaceTarget(source, target) {
		return l.checkedClassCast(node, value, []*checker.Type{target}, "class/interface downcast: value lacks the class tag")
	}
	return nil, l.notYet(node, "a class/interface downcast whose class tag cannot prove its nominal target and generic arguments")
}

func (l *lowering) classInterfaceTarget(source, target *checker.Type) bool {
	if !classInterfaceUnion(source) || !isClassInstance(target) || len(l.checker.GetTypeArguments(target)) != 0 {
		return false
	}
	matched := false
	for _, member := range source.Types() {
		if !isClassInstance(member) {
			continue
		}
		ancestor := l.nominalAncestor(member, target, map[[2]*checker.Type]bool{})
		// One class tag cannot prove incompatible arguments of the same generic class.
		if l.classView(member, l.classNodeFor(target)) != nil && !ancestor {
			return false
		}
		matched = matched || ancestor
	}
	return matched
}

// The shared expression entry calls this separate path; worker-owned ordinary
// calls and property lowering retain their existing guards and implementation.
func (l *lowering) classInterfaceValue(node *ast.Node) (ir.Expression, bool, error) {
	node = ast.SkipParentheses(node)
	if l.classInterfaceIn(node) {
		binary := node.AsBinaryExpression()
		object, err := l.expression(binary.Right)
		if err != nil {
			return nil, true, err
		}
		key, err := l.expression(binary.Left)
		if err != nil {
			return nil, true, err
		}
		return ir.HasOwn{Object: object, Key: key, Prototype: true}, true, nil
	}
	if node.Kind == ast.KindCallExpression && l.classInterfaceMethod(node) {
		value, err := l.callClosure(node)
		if err == nil && value.Type() == 0 {
			return nil, true, l.notYet(node, "a void class/interface method used as a value")
		}
		return value, true, err
	}
	if node.Kind == ast.KindPropertyAccessExpression && isCallee(node) {
		access := node.AsPropertyAccessExpression()
		if l.classInterfaceSource(access.Expression) != nil {
			of, known := l.representation(l.checker.GetTypeAtLocation(node))
			if !known || of != ir.Closure {
				return nil, true, l.notYet(node, "a class/interface callee without a represented signature")
			}
			object, err := l.expression(access.Expression)
			if err != nil {
				return nil, true, err
			}
			if object.Type() != ir.Object {
				return nil, true, l.notYet(node, "a class/interface receiver without an object representation")
			}
			return ir.Property{Object: object, Name: access.Name().Text(), Of: ir.Closure, Method: true, Optional: access.QuestionDotToken != nil}, true, nil
		}
	}
	return nil, false, nil
}
