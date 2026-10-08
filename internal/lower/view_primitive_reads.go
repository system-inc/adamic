package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A representation exemption requires a complete declared primitive contract.
// No storage tag or unknown contract can make an unsupported read admissible.
func (l *lowering) viewPrimitiveUnionRead(target *checker.Type) bool {
	target = l.concrete(target)
	of, known := l.representation(target)
	if !known || of != ir.Union {
		return false
	}
	id := l.result.ViewContractTypes[int(target.Id())]
	if id <= 0 || int(id) > len(l.result.ViewContracts) {
		return false
	}
	contract := l.result.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return false
	}
	if contract.Kind == ir.ViewNullable {
		id = contract.Element
	}
	members, complete := ir.PrimitiveViewMembers(l.result, id)
	return complete && len(members) > 0
}

// Destructuring is a field read too, including inside a helper accepting ordinary
// and viewed values. Attach the same selector metadata as a property expression.
func (l *lowering) preparePrimitiveDestructuredRead(node *ast.Node, receiver, declared *checker.Type, property *ir.Property) {
	if !l.viewPrimitiveUnionRead(declared) {
		return
	}
	declared = l.concrete(declared)
	property.ViewWhere = l.program.Where(node)
	property.ViewReceiverTypeID = int(receiver.Id())
	property.ViewTypeID = int(declared.Id())
	property.ViewContract = l.result.ViewContractTypes[property.ViewTypeID]
	property.Nullish = true
	property.NullAllowed = l.includesNull(declared)
	property.UndefinedAllowed = l.includesUndefined(declared)
	property.NullishKinds = l.nullishViewKinds(declared)
}

// The scalar-to-box conversion is required even when no cast exists. Erasing
// it would read a primitive producer slot as a boxed pointer. This exemption is
// limited to binding reads with complete primitive declarations.
func primitiveBindingConversion(program *ir.Program, property ir.Property) bool {
	if !property.Nullish || property.Of != ir.Union || !strings.HasSuffix(property.View, " (field "+property.Name+")") {
		return false
	}
	id := property.ViewContract
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	contract := program.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return false
	}
	if contract.Kind == ir.ViewNullable {
		id = contract.Element
	}
	members, complete := ir.PrimitiveViewMembers(program, id)
	return complete && len(members) > 0
}
