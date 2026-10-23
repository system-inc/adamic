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

var viewUnionContractHook viewContractHook

func init() { viewUnionContractHook = buildUnionReadContract }

// Reserved attachment point for V5 dictionary contracts; no current caller.
var viewDictionaryContractHook = unavailableViewFamily("dictionary")

// Reserved for V5 dictionary intersections; no current caller.
var viewIntersectionContractHook = unavailableViewFamily("intersection")

// Reserved fail-closed attachment point for V6 erasure; no current caller.
func viewErasureProof(*ir.Program, ir.Property) bool { return false }
