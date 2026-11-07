package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A binding freezes its origin only when no write in the closed program replaces it.
// Indexed shape mutations select record storage and invalidate this allocation proof.
// Follow imported symbols and annotated aliases, but do not infer origins for parameters or calls.
func (lowering *lowering) enumerationProven(node *ast.Node, element *checker.Type, depth int) bool {
	if node == nil || depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	if lowering.enumObject(node) != nil {
		return true
	}
	if node.Kind == ast.KindAsExpression && ast.IsConstTypeReference(node.AsAsExpression().Type) {
		return lowering.enumerationProven(node.AsAsExpression().Expression, element, depth+1)
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		if lowering.entriesRecordLiteral(node) {
			return false
		}
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
				return false
			}
			if property.Kind == ast.KindPropertyAssignment &&
				lowering.uninitializedInitializer(property.AsPropertyAssignment().Initializer) {
				return false
			}
			name, known := lowering.methodName(property)
			if !known || name == "__proto__" {
				return false
			}
		}
		for _, field := range lowering.checker.GetPropertiesOfType(lowering.checker.GetTypeAtLocation(node)) {
			if field.Flags&ast.SymbolFlagsOptional != 0 || (!lowering.enumAssignable(
				lowering.checker.GetTypeOfSymbol(field),
				element,
			) || !lowering.checker.IsTypeAssignableTo(lowering.checker.GetTypeOfSymbol(field), element)) {
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := lowering.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	if lowering.enumerationBindingWritten(symbol) {
		return false
	}
	return lowering.enumerationProven(declaration.AsVariableDeclaration().Initializer, element, depth+1)
}

func (lowering *lowering) enumerationBindingWritten(symbol *ast.Symbol) bool {
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return true
	}
	written := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var target *ast.Node
		if node.Kind == ast.KindBinaryExpression &&
			ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
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
				if ast.IsIdentifier(part) && lowering.symbol(part) == symbol {
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
func (lowering *lowering) enumerationDescriptors(where *ast.Node, checked bool) error {
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return moduleOrderError
	}
	var failure error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
			failure = &Refused{
				Where: lowering.program.Where(where),
				What:  "Object enumeration with getter or accessor descriptors",
				Fix:   "use own data properties for this first reflection implementation",
			}
		}
		if node.Kind == ast.KindObjectLiteralExpression {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if property.Name() != nil &&
					property.Name().Kind == ast.KindComputedPropertyName &&
					lowering.checker.GetTypeAtLocation(
						property.Name().AsComputedPropertyName().Expression,
					).Flags()&checker.TypeFlagsESSymbolLike != 0 {
					failure = &Refused{
						Where: lowering.program.Where(where),
						What:  "Object enumeration with symbol keys",
						Fix:   "use string-keyed own data properties for this first reflection implementation",
					}
				}
				name, known := lowering.methodName(property)
				if known && (strings.ContainsRune(name, 0) || strings.HasPrefix(name, "#") || name == iteratorSlot) {
					failure = lowering.notYet(
						where,
						"Object enumeration of a literal with reserved or symbol key storage",
					)
				}
				if property.Kind == ast.KindSpreadAssignment &&
					lowering.includesUndefined(
						lowering.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression),
					) {
					failure = lowering.notYet(
						where,
						"Object enumeration of a nullable spread with synthetic absent slots",
					)
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
