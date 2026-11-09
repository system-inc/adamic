package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A primitive seen through a program's structural object view cannot use its own
// library member dispatch yet (#thrpakr). Keep primitive intersections (brands),
// unions naming the primitive itself, library-declared targets and empty object
// types exempt. Those are erased brands or roads owned by library dispatch.
func (l *lowering) primitiveObjectViewGap(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
		return nil
	}
	var target *checker.Type
	if viewSite(node) {
		target = l.slotType(node)
	}
	if target == nil {
		target = l.impliedTarget(node)
	}
	if target == nil {
		return nil
	}
	return l.primitiveObjectView(node, l.checker.GetTypeAtLocation(node), target)
}

func primitiveObjectKind(source *checker.Type) (string, checker.TypeFlags) {
	for _, kind := range []struct {
		name  string
		flags checker.TypeFlags
	}{
		{"string", checker.TypeFlagsStringLike},
		{"number", checker.TypeFlagsNumberLike},
		{"boolean", checker.TypeFlagsBooleanLike},
		{"bigint", checker.TypeFlagsBigIntLike},
		{"symbol", checker.TypeFlagsESSymbolLike},
	} {
		if source.Flags()&kind.flags != 0 {
			return kind.name, kind.flags
		}
	}
	return "", 0
}

func (l *lowering) primitiveObjectView(node *ast.Node, source, target *checker.Type) error {
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if err := l.primitiveObjectView(node, member, target); err != nil {
				return err
			}
		}
		return nil
	}
	if base := l.phantomBase(source); base != nil {
		source = base
	}
	primitive, flags := primitiveObjectKind(source)
	if primitive == "" {
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if member.Flags()&flags != 0 && l.checker.IsTypeAssignableTo(source, member) {
				return nil
			}
		}
		for _, member := range target.Types() {
			if l.checker.IsTypeAssignableTo(source, member) {
				if err := l.primitiveObjectView(node, source, member); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range target.Types() {
			if member.Flags()&flags != 0 {
				return nil
			}
		}
	}
	if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || isClassInstance(target) {
		return nil
	}
	for _, member := range l.checker.GetPropertiesOfType(target) {
		if programDeclared(member) {
			return l.notYet(node, "a "+primitive+" seen as the object type "+l.checker.TypeToString(target)+" isn't lowered yet; read it as "+primitive)
		}
	}
	return nil
}

func init() {
	// Before optional widening: a primitive's representation is NotYet, not hidden fields.
	refusalRules.register(3290, "primitiveObjectViewGap", (*lowering).primitiveObjectViewGap)
}
