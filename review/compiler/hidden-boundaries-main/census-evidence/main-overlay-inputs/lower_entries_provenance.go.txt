package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A binding freezes its origin only when no write in the closed program replaces it.
// Indexed shape mutations select record storage and invalidate this allocation proof.
// Follow imported symbols and annotated aliases, but do not infer origins for parameters or calls.
func (l *lowering) enumerationProven(node *ast.Node, element *checker.Type, depth int) bool {
	if node == nil || depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	if l.enumObject(node) != nil {
		return true
	}
	if node.Kind == ast.KindAsExpression && ast.IsConstTypeReference(node.AsAsExpression().Type) {
		return l.enumerationProven(node.AsAsExpression().Expression, element, depth+1)
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		if l.entriesRecordLiteral(node) {
			return false
		}
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
				return false
			}
			if property.Kind == ast.KindPropertyAssignment && l.uninitializedInitializer(property.AsPropertyAssignment().Initializer) {
				return false
			}
			name, known := l.methodName(property)
			if !known || name == "__proto__" {
				return false
			}
		}
		for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(node)) {
			if field.Flags&ast.SymbolFlagsOptional != 0 || (!l.enumAssignable(l.checker.GetTypeOfSymbol(field), element) || !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(field), element)) {
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	if l.enumerationBindingWritten(symbol) {
		return false
	}
	return l.enumerationProven(declaration.AsVariableDeclaration().Initializer, element, depth+1)
}

func (l *lowering) enumerationBindingWritten(symbol *ast.Symbol) bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	written := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var target *ast.Node
		if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			target = node.AsBinaryExpression().Left
		}
		if node.Kind == ast.KindPrefixUnaryExpression {
			target = node.AsPrefixUnaryExpression().Operand
		}
		if node.Kind == ast.KindPostfixUnaryExpression {
			target = node.AsPostfixUnaryExpression().Operand
		}
		if node.Kind == ast.KindForOfStatement {
			target = node.AsForInOrOfStatement().Initializer
		}
		if target != nil {
			var inspect ast.Visitor
			inspect = func(part *ast.Node) bool {
				if ast.IsIdentifier(part) && l.symbol(part) == symbol {
					written = true
				}
				return part.ForEachChild(inspect)
			}
			inspect(target)
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return written
}

// Native reflection currently represents own data descriptors only. Do not let an
// erased accessor origin become a slot read; accessor and symbol layouts are refused.
func (l *lowering) enumerationDescriptors(where *ast.Node, checked bool) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var failure error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
			failure = &Refused{Where: l.program.Where(where), What: "Object enumeration with getter or accessor descriptors", Fix: "use own data properties for this first reflection implementation"}
		}
		if node.Kind == ast.KindObjectLiteralExpression {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if property.Name() != nil && property.Name().Kind == ast.KindComputedPropertyName && l.checker.GetTypeAtLocation(property.Name().AsComputedPropertyName().Expression).Flags()&checker.TypeFlagsESSymbolLike != 0 {
					failure = &Refused{Where: l.program.Where(where), What: "Object enumeration with symbol keys", Fix: "use string-keyed own data properties for this first reflection implementation"}
				}
				name, known := l.methodName(property)
				if known && (strings.ContainsRune(name, 0) || strings.HasPrefix(name, "#") || name == iteratorSlot) {
					failure = l.notYet(where, "Object enumeration of a literal with reserved or symbol key storage")
				}
				if property.Kind == ast.KindSpreadAssignment && l.includesUndefined(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {
					failure = l.notYet(where, "Object enumeration of a nullable spread with synthetic absent slots")
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return failure
}
