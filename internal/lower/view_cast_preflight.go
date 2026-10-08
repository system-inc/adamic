package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Deferred lowerers use one preflight mechanism. A candidate never admits a
// cast: lowering must return a checked value or its original refusal survives.
type castLoweringKind uint8

const (
	castLoweringNone castLoweringKind = iota
	castLoweringViews
	castLoweringPhantom
	castLoweringNodeProjection
	castLoweringOptionalPresence
	castLoweringScalar
	castLoweringGenericView
)

type castLowerer struct {
	kind           castLoweringKind
	beforeRelation bool
	candidate      func(*lowering, *ast.Node, *checker.Type, *checker.Type) bool
	lower          func(*lowering, *ast.Node, ir.Expression, *checker.Type, *checker.Type) (ir.Expression, error)
}

var deferredCastLowerers = []castLowerer{
	{castLoweringGenericView, true, (*lowering).genericViewCastCandidate, (*lowering).unresolvedGenericViewCast},
	{castLoweringScalar, true, (*lowering).scalarCastCandidate, (*lowering).scalarCast},
	{castLoweringOptionalPresence, true, (*lowering).optionalPresenceViewCandidate, (*lowering).optionalPresenceViewCast},
	{castLoweringViews, false, func(l *lowering, node *ast.Node, source, target *checker.Type) bool {
		return l.viewCastCandidate(source, target) || l.nullableReadonlyArrayCast(node, source, target)
	}, (*lowering).deferredViewCast},
	{castLoweringPhantom, true, func(l *lowering, node *ast.Node, source, target *checker.Type) bool {
		return l.phantomArrayBase(source) != nil || l.phantomArrayBase(target) != nil || l.phantomCast(source, target)
	}, (*lowering).deferredPhantomCast},
	{castLoweringNodeProjection, true, func(l *lowering, node *ast.Node, source, target *checker.Type) bool {
		return l.nodeRequirePerformanceProjection(node)
	}, func(l *lowering, node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
		return value, nil
	}},
}

func (l *lowering) deferredCastCandidate(node *ast.Node, source, target *checker.Type, beforeRelation bool) castLoweringKind {
	for _, handler := range deferredCastLowerers {
		if handler.beforeRelation == beforeRelation && handler.candidate(l, node, source, target) {
			return handler.kind
		}
	}
	return castLoweringNone
}
func (l *lowering) lowerDeferredCast(kind castLoweringKind, node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	for _, handler := range deferredCastLowerers {
		if handler.kind == kind {
			return handler.lower(l, node, value, source, target)
		}
	}
	return nil, nil
}
func (l *lowering) deferredPhantomCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	from, to := l.phantomArrayView(source), l.phantomArrayView(target)
	if l.checker.IsArrayType(from) && l.checker.IsArrayType(to) && l.isLibraryType(from, "ReadonlyArray") && !l.isLibraryType(to, "ReadonlyArray") {
		return nil, l.wideningRefusal(node, from, to, &widening{source: from, target: to, readonlyField: "[element]"})
	}
	if handled, err := l.phantomArrayCast(node, node.AsAsExpression().Expression, source, target); handled {
		return value, err
	}
	if l.phantomCast(source, target) {
		return value, nil
	}
	return nil, nil
}

// Deferral is not admission. Preserve the earlier any/double-assertion/upcast,
// writable-slot and nominal proofs, then let the actual view dispatch validate
// representations, fields and callable bodies when the expression is lowered.
func (l *lowering) viewCastCandidate(source, target *checker.Type) bool {
	if checker.IsTupleType(source) || checker.IsTupleType(target) {
		return false
	}
	if l.checker.IsArrayType(l.withoutUndefined(source)) && l.checker.IsArrayType(target) {
		return true
	}
	return source.Flags()&checker.TypeFlagsObject != 0 && target.Flags()&checker.TypeFlagsObject != 0 && !l.callableViewContract(source) && !l.callableViewContract(target) && l.checker.IsTypeAssignableTo(target, source)
}

func (l *lowering) deferredViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if l.readonlyArrayViewBridge(node, source, target) {
		return value, nil
	}
	if checked, err := l.viewArrayCast(node, value, source, target); checked != nil || err != nil {
		return checked, err
	}
	if checked, err := l.viewProvenClassCast(node, value, source, target); checked != nil || err != nil {
		if err != nil && isClassInstance(target) {
			return nil, nil
		}
		return checked, err
	}
	if checked, err := l.interfaceCast(node, value, source, target); checked != nil || err != nil {
		return checked, err
	}
	return l.structuralViewCast(node, value, source, target)
}

// A syntactically immediate unknown[] bridge has no writable alias when its
// only consumer is a readonly scalar view. The outer assertion supplies the
// lazy contract; the intermediate unknown element type is never exposed.
func (l *lowering) readonlyArrayViewBridge(node *ast.Node, source, target *checker.Type) bool {
	if !l.checker.IsArrayType(l.withoutUndefined(source)) || !l.checker.IsArrayType(target) || l.checker.GetElementTypeOfArrayType(target).Flags()&checker.TypeFlagsUnknown == 0 {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindAsExpression || parent.AsAsExpression().Expression != node {
		return false
	}
	final := l.concrete(l.checker.GetTypeAtLocation(parent))
	return (l.checker.IsArrayType(final) && l.isLibraryType(final, "ReadonlyArray") && interfaceScalar(l.checker.GetElementTypeOfArrayType(final))) || l.readonlyArrayConsumer(node) != nil
}
