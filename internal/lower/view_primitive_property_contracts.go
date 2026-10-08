package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The union receiver may be unsupported while this declared field has a
// complete selector. Reserve that field's original descriptor at its read.
func (l *lowering) prepareMixedPrimitiveProperty(node *ast.Node, declared *checker.Type, property *ir.Property) {
	if !l.objectPrimitiveViewType(declared) {
		return
	}
	if id, err := l.viewContract(node, declared); err == nil {
		property.ViewContract = id
	}
}

// This certificate covers a single structural object alternative and scalar
// alternatives. The read emitter tests every runtime kind before extraction.
// It does not certify the receiver union or any unread descendant field.
func mixedPrimitivePropertyChecks(program *ir.Program, id ir.ViewContractID) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	contract := program.ViewContracts[id-1]
	if contract.Unsupported != "" || contract.Kind != ir.ViewUnion || contract.Of != ir.Union {
		return false
	}
	objects, scalars := 0, 0
	for _, member := range contract.Members {
		if member <= 0 || int(member) > len(program.ViewContracts) {
			return false
		}
		child := program.ViewContracts[member-1]
		if child.Unsupported != "" {
			return false
		}
		switch child.Kind {
		case ir.ViewUndefined:
		case ir.ViewScalar:
			if child.Of != ir.String && child.Of != ir.Number && child.Of != ir.Boolean {
				return false
			}
			scalars++
		case ir.ViewObject:
			if child.Of != ir.Object || child.Nominal != "" {
				return false
			}
			objects++
		default:
			return false
		}
	}
	return objects == 1 && scalars > 0
}
