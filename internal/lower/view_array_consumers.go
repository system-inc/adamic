package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A chain of assertions and parentheses exposes no intermediate writable alias.
// Judge it by its actual consumer. Any mutable final slot keeps the usual
// invariant checks, including a mutable assertion after a readonly one.
func (l *lowering) readonlyArrayConsumer(node *ast.Node) *checker.Type {
	outer := node
	for outer.Parent != nil {
		parent := outer.Parent
		if parent.Kind == ast.KindParenthesizedExpression || parent.Kind == ast.KindAsExpression && parent.AsAsExpression().Expression == outer {
			outer = parent
			continue
		}
		break
	}
	target := l.checker.GetContextualType(outer, checker.ContextFlagsNone)
	if target == nil {
		target = l.impliedTarget(outer)
	}
	if target == nil {
		return nil
	}
	target = l.phantomArrayView(l.withoutUndefined(target))
	if !l.checker.IsArrayType(target) || !l.isLibraryType(target, "ReadonlyArray") {
		return nil
	}
	return target
}

// Undefined is passed through only into a consumer whose declared result still
// admits it. Its array member is checked by the outer element view; neither the
// unknown[] bridge nor the mutable assertion escapes as a writable alias.
func (l *lowering) nullableReadonlyArrayCast(node *ast.Node, source, target *checker.Type) bool {
	if !l.includesUndefined(source) || !l.checker.IsArrayType(l.withoutUndefined(source)) || !l.checker.IsArrayType(target) || l.readonlyArrayConsumer(node) == nil {
		return false
	}
	outer := node
	for outer.Parent != nil && (outer.Parent.Kind == ast.KindParenthesizedExpression || outer.Parent.Kind == ast.KindAsExpression && outer.Parent.AsAsExpression().Expression == outer) {
		outer = outer.Parent
	}
	consumer := l.checker.GetContextualType(outer, checker.ContextFlagsNone)
	if consumer == nil {
		consumer = l.impliedTarget(outer)
	}
	return consumer != nil && l.includesUndefined(consumer)
}

// A never[] has no present elements in a checked program. Use scalar slots for
// its empty allocation; invariant checks prevent a writable wider alias.
func viewNeverArrayElement(element *checker.Type) (ir.Type, bool) {
	return ir.Number, element.Flags()&checker.TypeFlagsNever != 0
}

func (l *lowering) viewArrayString(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	element, err := l.elementType(node)
	if err != nil {
		return nil, err
	}
	if element != ir.Number && element != ir.Boolean && element != ir.String && element != ir.MaybeNumber && element != ir.Union {
		return nil, l.notYet(node, "an array template requiring recursive element string conversion")
	}
	return ir.ArrayJoin{Array: value, Separator: ir.StringConstant{Index: l.constant(",")}, Element: element, Stringify: true, ViewRead: l.viewArrayUse(node, node, element, false)}, nil
}

func (l *lowering) viewOptionalArrayComparator(call, comparator *ast.Node, element ir.Type) (bool, error) {
	optional := l.includesUndefined(l.checker.GetTypeAtLocation(comparator))
	if !optional {
		return false, nil
	}
	if element != ir.Number && element != ir.Boolean && element != ir.String && element != ir.MaybeNumber {
		return false, l.notYet(comparator, "an optional comparator requiring recursive default string conversion")
	}
	receiver := ast.SkipParentheses(call.AsCallExpression().Expression).AsPropertyAccessExpression().Expression
	array := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	declared := l.concrete(l.checker.GetElementTypeOfArrayType(array))
	if element == ir.String && l.includesUndefined(declared) {
		return false, l.notYet(comparator, "an optional comparator over undefined reference elements")
	}
	return true, nil
}
