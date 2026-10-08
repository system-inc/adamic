package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A dictionary read checks its immediate member and retains every reference
// child descriptor. Descendant reads still use the ordinary view machinery.
// This certificate is separate from the original union's admission descriptor.
func (l *lowering) dictionaryReferenceReadContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, bool) {
	if !l.primitiveDictionaryCandidate(target) {
		return 0, false
	}
	selected := ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Union, Name: l.checker.TypeToString(target)}
	for _, member := range target.Types() {
		id, err := l.viewContract(node, member)
		if err != nil {
			return 0, false
		}
		child := l.result.ViewContracts[id-1]
		if child.Unsupported != "" || child.Nominal != "" || child.FixedTuple {
			return 0, false
		}
		switch child.Kind {
		case ir.ViewScalar, ir.ViewNull, ir.ViewUndefined, ir.ViewArray, ir.ViewObject, ir.ViewDictionary:
		default:
			return 0, false
		}
		selected.Members = append(selected.Members, id)
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, selected)
	if _, ok := ir.DictionaryReadKinds(l.result, id); !ok {
		l.result.ViewContracts = l.result.ViewContracts[:len(l.result.ViewContracts)-1]
		return 0, false
	}
	return id, true
}

// A direct String consumer retains its primitive-only checked selector until
// the reference families have their own ToPrimitive implementation.
func (l *lowering) dictionaryPrimitiveReadContext(node *ast.Node) bool {
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	parent := at.Parent
	return parent != nil && parent.Kind == ast.KindCallExpression && l.isLibraryGlobal(ast.SkipParentheses(parent.AsCallExpression().Expression), "String")
}
