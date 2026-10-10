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

var viewDictionaryContractHook = unavailableViewFamily("dictionary")
var viewIntersectionContractHook = unavailableViewFamily("intersection")

func viewErasureProof(*ir.Program, ir.Property) bool { return false }
func (l *lowering) viewObjectFields(node *ast.Node, target *checker.Type, fields map[string]bool, _ map[*checker.Type]bool, _ bool) error {
	all, err := l.viewSchema(node, target)
	for field := range all {
		fields[field] = true
	}
	return err
}
