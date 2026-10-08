package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func unavailableViewFamily(family string) viewContractHook {
	return func(l *lowering, node *ast.Node, _ *checker.Type, _ viewContractBuilder) (ir.ViewContractID, error) {
		return 0, l.notYet(node, "a "+family+" checked-view contract")
	}
}

var viewUnionContractHook = unavailableViewFamily("union")
var viewDictionaryContractHook = unavailableViewFamily("dictionary")
var viewIntersectionContractHook = unavailableViewFamily("intersection")

// Erasure needs a separate all-reaching-allocations certificate in V6.
func viewErasureProof(*ir.Program, ir.Property) bool { return false }
