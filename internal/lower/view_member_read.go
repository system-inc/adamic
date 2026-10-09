package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// readViewMember is the checked own-member boundary. Syntax supplies its field
// and receiver certificates before a read can expose the stored value.
func (l *lowering) readViewMember(node *ast.Node, property ir.Property, field *ast.Symbol, receiver *checker.Type) ir.Expression {
	property.Readiness = sourceExpression(node)
	if property.View == "" {
		property.View = sourceExpression(node)
	}
	property.ViewWhere = l.program.Where(node)
	if receiver != nil {
		property.ViewReceiverTypeID = int(receiver.Id())
	}
	if field != nil {
		declared := l.checker.GetTypeOfSymbol(field)
		property.ViewTypeID = int(declared.Id())
		property.ViewContract = l.result.ViewContractTypes[property.ViewTypeID]
		property.ViewType = l.checker.TypeToString(declared)
		property.ViewAllowed = l.viewLiterals(declared)
	}
	if field != nil {
		for _, declaration := range field.Declarations {
			if declaration.Kind == ast.KindPropertyDeclaration {
				initializer := declaration.AsPropertyDeclaration().Initializer
				if l.lazyAssertionInitializer(initializer) && !l.uninitializedInitializer(initializer) {
					property.Readiness = sourceExpression(initializer)
				}
			}
		}
	}
	if field == nil || field.Flags&ast.SymbolFlagsOptional == 0 {
		return property
	}
	property.Absent = true
	if declared, _ := l.representation(l.checker.GetTypeOfSymbol(field)); declared.IsMaybe() && property.Of == declared.Present() {
		property.Of = declared
		return fit(property, declared.Present())
	}
	return property
}
