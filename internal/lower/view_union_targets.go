package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"reflect"
)

// A broad object can enter a tagged union only when each arm preserves its
// source slots and a common, immutable tag uniquely selects that arm. Payload
// contracts remain lazy obligations of the existing shared view entry point.
func (l *lowering) viewUnionTargetProof(node *ast.Node, source, target *checker.Type) (*castProof, error) {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsUnion == 0 {
		return nil, nil
	}
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) || !l.checker.IsTypeAssignableTo(member, source) || l.widened(member, source, map[[2]*checker.Type]bool{}) != nil {
			return nil, nil
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(source) {
		if !l.checker.IsReadonlySymbol(property) || accessorSymbol(property) {
			continue
		}
		proof := castProof{field: property.Name}
		valid := true
		for _, member := range target.Types() {
			tag := l.checker.GetPropertyOfType(member, property.Name)
			literal := l.fieldLiteral(member, property.Name)
			if tag == nil || !l.checker.IsReadonlySymbol(tag) || accessorSymbol(tag) || literal == nil {
				valid = false
				break
			}
			for _, previous := range proof.allowed {
				mask := checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBooleanLiteral
				if previous.Flags()&mask != literal.Flags()&mask || reflect.DeepEqual(previous.AsLiteralType().Value(), literal.AsLiteralType().Value()) {
					valid = false
				}
			}
			proof.allowed = append(proof.allowed, literal)
		}
		if valid && len(proof.allowed) == len(target.Types()) {
			if _, err := l.viewSchema(node, target); err != nil {
				return nil, err
			}
			return &proof, nil
		}
	}
	return nil, nil
}
