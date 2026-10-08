package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// markViewRead preserves declared obligations until all bodies have been lowered.
// The runtime read independently checks presence, readiness and physical storage.
func (l *lowering) markViewRead(node *ast.Node, property ir.Property, declared *checker.Type) ir.Property {
	property.ViewType = l.checker.TypeToString(declared)
	property.ViewAllowed = l.viewLiterals(declared)
	property.ViewContract = l.result.ViewContractTypes[int(declared.Id())]
	property.ViewTypeID = int(declared.Id())
	property.ViewWhere = l.program.Where(node)
	return property
}
