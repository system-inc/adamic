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

// checkedViewMembers holds the source once and checks declared own slots before
// exposing their presence or copying values. Copies retain hidden fields and tags.
func (l *lowering) checkedViewMembers(node, source *ast.Node, value ir.Expression, member *ast.Node) (ir.Expression, error) {
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(source))
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	checks := []ir.Statement{}
	for _, field := range l.checker.GetPropertiesOfType(receiver) {
		if member != nil && field.Name != member.Text() {
			continue
		}
		if accessorSymbol(field) {
			continue
		}
		prototype := false
		for _, declaration := range field.Declarations {
			if declaration.Kind == ast.KindMethodDeclaration || declaration.Kind == ast.KindMethodSignature || declaration.Name() != nil && declaration.Name().Kind == ast.KindPrivateIdentifier {
				prototype = true
			}
		}
		if prototype {
			continue
		}
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || of == ir.Weak || censusFieldSlotless(of) {
			return nil, l.notYet(node, "a view member operation with an unsupported field representation")
		}
		read := l.readViewMember(node, ir.Property{Object: held, Name: field.Name, Of: of, Absent: field.Flags&ast.SymbolFlagsOptional != 0, View: sourceExpression(node) + " (field " + field.Name + ")"}, field, receiver)
		checks = append(checks, ir.Evaluate{Value: read})
	}
	if len(checks) == 0 {
		return value, nil
	}
	if l.includesUndefined(l.checker.GetTypeAtLocation(source)) {
		b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: held}}, Then: checks})
	} else {
		b.body = append(b.body, checks...)
	}
	return b.finish("checked_view_members", held), nil
}
