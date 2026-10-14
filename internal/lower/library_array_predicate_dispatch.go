package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func libraryArrayPredicateMethod(name string) bool {
	return name == "filter" || name == "some" || name == "every"
}

// Keep the predicate dispatch separate from generic search and tail methods.
func (l *lowering) libraryArrayPredicateDispatch(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	if !libraryArrayPredicateMethod(name) && name != "findLast" && name != "findLastIndex" {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if !libraryArrayPredicateMethod(name) {
		if len(written) != 1 {
			return nil, true, l.notYet(node, name+" without one callback")
		}
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
		if len(signatures) == 1 && len(signatures[0].Parameters()) > 3 {
			return nil, true, l.notYet(node, name+" with more than three callback parameters")
		}
	}
	element, err := l.elementType(receiver)
	if err != nil {
		return nil, true, err
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if libraryArrayPredicateMethod(name) {
		return l.libraryArrayPredicate(node, array, element, name)
	}
	return l.libraryArrayFindLast(node, array, element, name)
}
