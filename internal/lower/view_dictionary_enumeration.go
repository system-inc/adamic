package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Keys depend on actual own-property presence, never on element conformance.
// The runtime selects the existing record table or fixed-object key adapter.
func (l *lowering) dictionaryKeysCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if name != "keys" || len(args) != 1 || !l.stringDictionary(l.checker.GetTypeAtLocation(args[0])) {
		return nil, false, nil
	}
	receiver, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	if receiver.Type() != ir.Object && receiver.Type() != ir.Record {
		return nil, true, l.notYet(node, "dictionary keys without object storage")
	}
	return ir.RecordCall{Method: "keys", Arguments: []ir.Expression{receiver}, Returns: ir.Array, DictionaryKeys: true, ViewWhere: sourceExpression(node)}, true, nil
}
