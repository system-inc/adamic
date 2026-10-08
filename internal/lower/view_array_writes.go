package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
)

func (l *lowering) viewArrayRecordSourceContract(node *ast.Node, target *checker.Type) ir.ViewContractID {
	if target == nil || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) || len(l.checker.GetIndexInfosOfType(target)) != 0 || l.callableViewContract(target) {
		return 0
	}
	// Filter before interning: an unsupported child must not eagerly refuse a read
	// or manufacture an array-view origin in an unrelated program.
	for _, field := range l.checker.GetPropertiesOfType(target) {
		child := l.checker.GetTypeOfSymbol(field)
		of, known := l.representation(child)
		if field.Flags&ast.SymbolFlagsOptional != 0 || !interfaceScalar(child) || l.includesUndefined(child) || !known || (of != ir.Number && of != ir.Boolean && of != ir.String) {
			return 0
		}
	}
	id := l.slotContract(node, target)
	if !ir.FlatArrayRecordContract(l.result, id) {
		return 0
	}
	return id
}

func (l *lowering) viewArrayLiteralSourceContract(node *ast.Node) ir.ViewContractID {
	source := l.checker.GetTypeAtLocation(node)
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && l.viewArrayBase(contextual) != nil {
		source = contextual
	}
	element := l.concrete(l.viewArrayElementType(source))
	if element != nil && isClassInstance(l.checker.GetNonNullableType(element)) {
		if of, known := l.representation(element); known && (of == ir.Object || of == ir.Union) {
			return l.mapNominalEntrySlot(node, element)
		}
	}
	if element != nil && interfaceScalar(element) {
		if of, known := l.representation(element); known && of == ir.Union {
			if err := l.preparePrimitiveViewDeclaration(node, element); err == nil && l.viewPrimitiveUnionRead(element) {
				return l.slotContract(node, element)
			}
		}
	}
	return l.viewArrayRecordSourceContract(node, element)
}

func (l *lowering) viewArrayObjectSourceContract(node *ast.Node, literal ir.ObjectLiteral) ir.ViewContractID {
	if literal.Class != 0 || literal.Tuple || literal.Spread != nil || len(literal.Methods) != 0 {
		return 0
	}
	for _, field := range literal.Fields {
		if field.Uninitialized {
			return 0
		}
	}
	source := l.checker.GetTypeAtLocation(node)
	id := l.viewArrayRecordSourceContract(node, source)
	if id == 0 {
		return 0
	}
	contract := l.result.ViewContracts[id-1]
	contract.Fields = append([]ir.ViewFieldContract(nil), contract.Fields...)
	changed := false
	for i := range contract.Fields {
		for _, field := range literal.Fields {
			if field.Name == contract.Fields[i].Name && field.Contract != contract.Fields[i].Contract {
				if field.Contract <= 0 || int(field.Contract) > len(l.result.ViewContracts) {
					return 0
				}
				child := l.result.ViewContracts[field.Contract-1]
				if child.Kind != ir.ViewScalar || child.Unsupported != "" || child.Undefined || (child.Of != ir.Number && child.Of != ir.Boolean && child.Of != ir.String) {
					return 0
				}
				contract.Fields[i].Contract = field.Contract
				changed = true
			}
		}
	}
	if !changed {
		return id
	}
	// Contextual scalar slots can be wider than a fresh initializer. Preserve
	// their real declarations without turning a readonly view into allocation
	// evidence about the fresh record's own mutable fields.
	contract.Name = "record " + contract.Name
	copyID := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	if !ir.FlatArrayRecordContract(l.result, copyID) {
		return 0
	}
	return copyID
}

// Keep scalar RHS declarations until all array-view origins are known. Activation
// happens in readiness, so ordinary programs do not acquire source-write guards.
func (l *lowering) markViewArrayScalarSlotWrite(node, valueNode *ast.Node, write ir.SetProperty) ir.SetProperty {
	if write.Uninitialized || (write.Value.Type() != ir.Number && write.Value.Type() != ir.Boolean && write.Value.Type() != ir.String) {
		return write
	}
	write.ArraySlotWriteContract = l.slotContract(valueNode, l.concrete(l.checker.GetTypeAtLocation(valueNode)))
	if write.WriteWhere == "" {
		write.WriteWhere = strings.Join(strings.Split(filepath.Base(l.program.Where(node)), ":")[:2], ":")
	}
	return write
}

func prepareViewArrayReferenceWrites(program *ir.Program) {
	if !ir.HasArrayViews(program) {
		return
	}
	names := map[string]bool{}
	collect := func(node any) bool {
		id := ir.ViewContractID(0)
		switch value := node.(type) {
		case ir.ArrayLiteral:
			id = value.ElementContract
		case ir.ObjectLiteral:
			id = value.ArrayWriteContract
		}
		if ir.FlatArrayRecordContract(program, id) {
			for _, field := range program.ViewContracts[id-1].Fields {
				names[field.Name] = true
			}
		}
		return true
	}
	walk(program.Main, collect)
	for _, function := range program.Functions {
		walk(function.Body, collect)
	}
	if program.CheckedFields == nil {
		program.CheckedFields = map[string]bool{}
	}
	if program.OptionalViewFields == nil {
		program.OptionalViewFields = map[string]bool{}
	}
	for name := range names {
		program.CheckedFields[name] = true
		program.OptionalViewFields[name] = true
	}
	rewrite := func(node any) any {
		if write, ok := node.(ir.SetProperty); ok && names[write.Name] && write.WriteContract == 0 && write.ArraySlotWriteContract != 0 {
			write.WriteContract = write.ArraySlotWriteContract
			return write
		}
		return node
	}
	program.Main = rewriteShapeStatements(program.Main, rewrite)
	for index := range program.Functions {
		program.Functions[index].Body = rewriteShapeStatements(program.Functions[index].Body, rewrite)
	}
}
