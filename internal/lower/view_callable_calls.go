package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The read checks the declared set. The call uses only its resolved declaration,
// including the instantiated parameter types supplied by the checker.
func (l *lowering) convertResolvedCallableArguments(node *ast.Node, declared []*checker.Signature, arguments []ir.Expression) error {
	resolved := l.checker.GetResolvedSignature(node)
	if resolved == nil {
		return l.notYet(node, "a callable call without a resolved signature")
	}
	if len(declared) > 1 {
		member := false
		for _, signature := range declared {
			member = member || resolved.Declaration() != nil && resolved.Declaration() == signature.Declaration()
		}
		if !member {
			return l.notYet(node, "a resolved callable signature outside its declared overload set")
		}
	}
	for index, parameter := range resolved.Parameters() {
		if index >= len(arguments) {
			break
		}
		takes, known := l.censusCallableParameter(parameter)
		if !known {
			return l.notYet(node, "a resolved callable parameter without a representation")
		}
		arguments[index] = fit(arguments[index], takes)
		if arguments[index].Type() != takes {
			return l.notYet(node, "a resolved callable argument requiring another representation")
		}
	}
	return nil
}
