package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A read of a structural object checks its own heap kind. Register every descendant
// field too: returning or aliasing that object must not lose the checks on its reads.
// Interface descendants share the same data checks, including inherited fields.
func (l *lowering) viewObjectFields(node *ast.Node, target *checker.Type, fields map[string]bool, seen map[*checker.Type]bool, descendant bool) error {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		return l.viewUnionFields(node, target, fields, seen)
	}
	if descendant && viewInterfaceType(target) {
		return l.viewInterfaceFields(node, target, fields, seen)
	}
	if seen[target] {
		return nil
	}
	if l.checker.IsArrayType(target) || checker.IsTupleType(target) {
		return l.viewArrayFields(node, target, fields, seen)
	}
	seen[target] = true
	if isClassInstance(target) {
		return l.notYet(node, "a nominal class field in a checked view")
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return l.notYet(node, "a dictionary checked view")
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if property.Flags&ast.SymbolFlagsOptional != 0 {
			l.optionalViewWriteField(property.Name)
		}
		declared := l.checker.GetTypeOfSymbol(property)
		if l.callableViewContract(declared) {
			fields[property.Name] = true
			if err := l.viewCallableFieldUses(node, target, property); err != nil {
				return err
			}
			continue
		}
		of, known := l.representation(declared)
		if !l.viewDataType(declared) || !known || (of < ir.Number || of > ir.Array) && of != ir.MaybeNumber && of != ir.MaybeBoolean {
			return l.notYet(node, "checked view field "+property.Name+" of type "+l.checker.TypeToString(declared))
		}
		fields[property.Name] = true
		if of == ir.Object {
			if len(l.checker.GetPropertiesOfType(l.checker.GetNonNullableType(declared))) == 0 {
				return l.notYet(node, "an empty structural object field in a checked view")
			}
			if err := l.viewObjectFields(node, l.checker.GetNonNullableType(declared), fields, seen, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *lowering) viewDataType(target *checker.Type) bool {
	if l.structuralViewIntersection(target) || l.viewArrayBase(target) != nil {
		return true
	}
	if base := l.phantomBase(target); base != nil {
		return interfaceScalar(base)
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if !l.viewDataType(member) {
				return false
			}
		}
		return true
	}
	return target.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 || interfaceScalar(target) || target.Flags()&checker.TypeFlagsObject != 0
}

func viewInterfaceType(target *checker.Type) bool {
	if target.ObjectFlags()&checker.ObjectFlagsInterface != 0 {
		return true
	}
	return target.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.Target() != nil && target.Target().ObjectFlags()&checker.ObjectFlagsInterface != 0
}

// An untagged structural cast installs the same checked reads without a tag test.
// Admission preserves nominal identity and the existing writable-slot restrictions.
func (l *lowering) structuralViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object || source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) || !l.checker.IsTypeAssignableTo(target, source) {
		return nil, nil
	}
	if len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return nil, l.notYet(node, "a dictionary checked view")
	}
	if err := l.widened(target, source, map[[2]*checker.Type]bool{}); err != nil {
		return nil, l.notYet(node, "a writable-slot checked view requiring source contract certification")
	}
	return l.view(node, value, target)
}
