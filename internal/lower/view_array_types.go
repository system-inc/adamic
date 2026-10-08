package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"slices"
)

// Array ancestry fixes physical storage; additional own fields stay separate
// contracts. An overridden intrinsic or index signature is never erased.
func (l *lowering) viewArrayBase(target *checker.Type) *checker.Type {
	return l.viewArrayBaseSeen(l.concrete(target), map[*checker.Type]bool{})
}

func (l *lowering) viewArrayBaseSeen(target *checker.Type, seen map[*checker.Type]bool) *checker.Type {
	if target == nil || checker.IsTupleType(target) || isClassInstance(target) || seen[target] {
		return nil
	}
	if l.checker.IsArrayType(target) {
		return target
	}
	seen[target] = true
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		return l.viewReadonlyArrayUnionBase(target, seen)
	}
	var bases []*checker.Type
	if target.Flags()&checker.TypeFlagsIntersection != 0 {
		bases = target.Types()
	} else if viewInterfaceType(target) {
		declaration := l.classNodeFor(target)
		if declaration == nil || declaration.Kind != ast.KindInterfaceDeclaration {
			return nil
		}
		for _, member := range declaration.Members() {
			if member.Kind == ast.KindIndexSignature {
				return nil
			}
		}
		original := target
		if target.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.Target() != nil {
			original = target.Target()
		}
		for _, inherited := range l.checker.GetBaseTypes(original) {
			if target != original {
				inherited = instantiateType(l.checker, inherited, l.typeMapperOf(declaration, target))
			}
			bases = append(bases, inherited)
		}
	} else {
		return nil
	}
	var result *checker.Type
	for _, inherited := range bases {
		if array := l.viewArrayBaseSeen(inherited, seen); array != nil {
			if result != nil && !checker.Checker_isTypeIdenticalTo(l.checker, result, array) {
				return nil
			}
			result = array
		}
	}
	if result == nil {
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if inherited := l.checker.GetPropertyOfType(result, property.Name); inherited != nil && !slices.Equal(property.Declarations, inherited.Declarations) {
			return nil
		}
	}
	return result
}

func (l *lowering) viewArrayOwnProperties(target, base *checker.Type) []*ast.Symbol {
	var fields []*ast.Symbol
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if l.checker.GetPropertyOfType(base, property.Name) == nil {
			fields = append(fields, property)
		}
	}
	return fields
}

func (l *lowering) viewArrayElementType(target *checker.Type) *checker.Type {
	target = l.checker.GetNonNullableType(target)
	if element := l.viewArrayUnionElementType(target); element != nil {
		return element
	}
	base := l.viewArrayBase(target)
	if base == nil {
		return nil
	}
	return l.checker.GetElementTypeOfArrayType(base)
}
