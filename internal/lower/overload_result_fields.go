package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Follow immutable result aliases only. Admission still proves or checks the
// return before this read; matching field storage never proves a result type.
func (l *lowering) overloadResultOrigin(node *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		signature := l.checker.GetResolvedSignature(node)
		if signature != nil && signature.Declaration() != nil && signature.Declaration().Body() == nil && l.censusImplementation(signature.Declaration()) != nil {
			return node
		}
	}
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && !seen[symbol] && len(symbol.Declarations) == 1 {
			seen[symbol] = true
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.AsVariableDeclaration().Initializer != nil {
				return l.overloadResultOrigin(declaration.AsVariableDeclaration().Initializer, seen)
			}
		}
	}
	return nil
}

// A combined property symbol can describe a tagged union even though the
// individual objects store unboxed fields. The ordinary property instruction
// has one storage kind: every admitted result member must use that kind.
func (l *lowering) overloadResultField(node *ast.Node) error {
	receiver := node.AsPropertyAccessExpression().Expression
	if l.overloadResultOrigin(receiver, map[*ast.Symbol]bool{}) == nil {
		return nil
	}
	result := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
	if result.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	name := node.Name().Text()
	combined := l.checker.GetTypeOfPropertyOfType(result, name)
	if combined == nil {
		return nil
	}
	storage, known := l.representation(combined)
	compatible := known && !censusFieldSlotless(storage)
	for _, member := range result.Types() {
		field := l.checker.GetTypeOfPropertyOfType(member, name)
		if field == nil {
			compatible = false
			break
		}
		held, known := l.representation(field)
		if !known || held != storage {
			compatible = false
			break
		}
	}
	if !compatible {
		return &Refused{Where: l.program.Where(node), What: "an overload result field at result." + name + " without a single storage representation", Fix: "narrow the result to a member before reading the field; matching storage does not replace the overload return proof"}
	}
	return nil
}
