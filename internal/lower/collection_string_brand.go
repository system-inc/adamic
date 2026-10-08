package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// These two compiler brands are phantom by the step 19 ruling. An observable payload or
// another structural intersection is not a brand merely because it is attached to a string.
func (l *lowering) collectionBrandMarker(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(l.checker.GetIndexInfosOfType(proven)) != 0 {
		return false
	}
	fields := l.checker.GetPropertiesOfType(proven)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if (field.Name != "__pathBrand" && field.Name != "__escapedIdentifier") || field.Flags&ast.SymbolFlagsOptional != 0 || l.checker.GetTypeOfSymbol(field).Flags()&checker.TypeFlagsVoid == 0 {
			return false
		}
	}
	return true
}

func (l *lowering) collectionBrandIntersection(proven *checker.Type) (stringPart, voidPart bool) {
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return false, false
	}
	marker := false
	for _, part := range proven.Types() {
		switch {
		case part.Flags()&checker.TypeFlagsStringLike != 0:
			stringPart = true
		case part.Flags()&checker.TypeFlagsVoid != 0:
			voidPart = true
		case l.collectionBrandMarker(part):
			marker = true
		default:
			return false, false
		}
	}
	return marker && stringPart && !voidPart, marker && voidPart && !stringPart
}

// __String's branded void arm has no runtime inhabitant. A union containing ordinary void
// or undefined is different and keeps its existing optional representation.
func (l *lowering) collectionBrandString(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if text, _ := l.collectionBrandIntersection(proven); text {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	branded := false
	for _, part := range proven.Types() {
		text, empty := l.collectionBrandIntersection(part)
		if text {
			branded = true
			continue
		}
		if empty || part.Flags()&checker.TypeFlagsNever != 0 {
			continue
		}
		if part.Flags()&checker.TypeFlagsStringLike == 0 {
			return false
		}
	}
	return branded
}

// Construction can erase a brand only when the target admits every string. A branded
// literal intersection still has a value refinement and must not inherit this permission.
func (l *lowering) collectionBrandAcceptsString(proven *checker.Type) bool {
	members := castMembers(proven)
	for _, member := range members {
		if text, _ := l.collectionBrandIntersection(member); text {
			for _, part := range member.Types() {
				if part.Flags()&checker.TypeFlagsString != 0 {
					return true
				}
			}
		}
	}
	return false
}

func (l *lowering) checkedCollectionBrand(value ir.Expression, target *checker.Type) ir.Expression {
	if value.Type() == ir.String {
		return value
	}
	b := l.libraryArrayBuilder([]ir.Expression{fit(value, ir.Union)})
	held := b.read(b.parameters[0])
	matches := ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant("string")}}
	message := "brand boundary failed: expected string for " + l.checker.TypeToString(target)
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	return b.finish("checked_collection_brand", ir.Narrow{Value: held, To: ir.String})
}
