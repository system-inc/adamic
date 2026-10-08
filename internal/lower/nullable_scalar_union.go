package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Nullable scalar and mixed-layout unions keep null in an immortal box.
// One nullable reference still uses the existing pointer representation.
func (l *lowering) nullableScalarUnion(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion == 0 || !l.includesNull(proven) {
		return false
	}
	allScalar, mixed := true, false
	var shared ir.Type
	for _, member := range proven.Types() {
		if member.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			allScalar = false
		}
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			continue
		}
		held, known := l.representation(member)
		if !known || held == ir.Weak {
			return false
		}
		if shared != 0 && shared != held {
			mixed = true
		}
		shared = held
	}
	return (allScalar && (shared != ir.String || l.includesUndefined(proven))) || mixed
}

// Evaluate a nullable pointer once, preserving its empty case at the union boundary.
func (l *lowering) boxNullableUnion(node *ast.Node, value ir.Expression) ir.Expression {
	if !value.Type().IsReference() || value.Type() == ir.Union || !l.includesNull(l.checker.GetTypeAtLocation(node)) {
		return value
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if parent := outer.Parent; parent != nil && parent.Kind == ast.KindCallExpression && l.isConsole(parent.AsCallExpression().Expression) {
		return value
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual == nil {
		return value
	}
	to, _ := l.representation(l.concrete(contextual))
	if to != ir.Union {
		return value
	}
	return l.nullableObservation("union", value, ir.Null{Of: ir.Union}, func(present ir.Expression) ir.Expression { return ir.Box{Value: present} })
}
