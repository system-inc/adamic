package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A primitive seen through a program's structural object view cannot use its own
// library member dispatch yet (#thrpakr). Keep primitive intersections (brands):
// they retain the primitive representation. Keep unions naming the primitive:
// it enters its own member. Keep library-declared targets, owned by library
// dispatch, and empty object types, which have no member to read through a view.
// Array.from without a callback and object ToPrimitive conversions retain their
// own existing NotYet stops.
// Keep class targets so classViewRefusal reports the stronger nominal ancestry
// refusal, even for a class with only public optional fields.
// A generic binding is no view of its own: monomorphization keeps len<string>'s
// string representation and its length read. Unsupported instantiated member
// reads, such as String(x.toFixed), already stop in member lowering. Index-only
// targets Record<string, unknown> and { [k: string]: number } reject strings
// in the checker before this pass.
func (l *lowering) primitiveObjectViewGap(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
		return nil
	}
	// A logical operand is not a slot value: || and ?? discard some members,
	// and && keeps only its left's falsy members. Judge the whole expression's
	// checker type at its own site instead. Parentheses do not create a new site.
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	if parent := child.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindBarBarToken, ast.KindQuestionQuestionToken, ast.KindAmpersandAmpersandToken:
				return nil
			case ast.KindCommaToken:
				if binary.Left == child {
					return nil
				}
			}
		case ast.KindVoidExpression:
			return nil
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition == child {
				return nil
			}
		}
	}
	// Keep impliedTarget's carried conditional branches, array elements, return
	// values and arrow expression bodies. Its callable context can also reach a
	// conditional's discarded condition, excluded above. It currently gives no
	// target to a comma's left or a void operand; keep those excluded explicitly.
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
			return l.notYet(node, "a "+primitive+" seen as the object type "+l.checker.TypeToString(target)+" (read it as "+primitive+")")
		}
	}
	return nil
}

func init() {
	// Before optional widening: a primitive's representation is NotYet, not hidden fields.
	refusalRules.register(3290, "primitiveObjectViewGap", (*lowering).primitiveObjectViewGap)
}
