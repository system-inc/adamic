package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An object union needs a shared finite tag before a member can be selected.
// Register every member's fields: testing the tag never proves its payload.
func (l *lowering) viewUnionFields(node *ast.Node, target *checker.Type, fields map[string]bool, seen map[*checker.Type]bool) error {
	if seen[target] {
		return nil
	}
	seen[target] = true
	of, known := l.representation(target)
	if !known || of != ir.Object {
		return l.notYet(node, "a nullish or mixed representation checked union")
	}
	tag := false
	for _, property := range l.checker.GetPropertiesOfType(target) {
		declared := l.checker.GetTypeOfSymbol(property)
		if property.Flags&ast.SymbolFlagsOptional == 0 && len(l.viewLiterals(declared)) != 0 {
			tag = true
			break
		}
	}
	if !tag {
		return l.notYet(node, "an object union checked view without a common finite discriminant")
	}
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsObject == 0 {
			return l.notYet(node, "a nullish member of an object checked view")
		}
		if err := l.viewObjectFields(node, member, fields, seen, true); err != nil {
			return err
		}
	}
	return nil
}
