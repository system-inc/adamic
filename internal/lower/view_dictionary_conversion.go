package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Only readonly consumers can use a representation checked at lookup. A read
// certificate never grants mutation of a fixed object's storage.
func (l *lowering) dictionaryReferenceConversion(node *ast.Node) bool {
	if node.Kind != ast.KindAsExpression {
		return false
	}
	as := node.AsAsExpression()
	source, target := l.concrete(l.checker.GetTypeAtLocation(as.Expression)), l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.primitiveDictionaryCandidate(source) || !l.stringDictionary(target) {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(target)
	return len(infos) == 1 && infos[0].IsReadonly() && len(l.checker.GetPropertiesOfType(target)) == 0
}

func (l *lowering) checkedDictionaryReferenceConversion(node *ast.Node) (ir.Expression, error) {
	value, err := l.expression(node.AsAsExpression().Expression)
	if err != nil {
		return nil, err
	}
	// Carry the heap reference unchanged; source_read checks the representation
	// before following either a fixed slot or a record.c table.
	converted := ir.Narrow{Value: value, To: ir.Object, Dictionary: true}
	return l.view(node, converted, l.concrete(l.checker.GetTypeAtLocation(node)))
}
