package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Admit one concrete object representation plus both nullish tags. Collections,
// closures, weak handles and erased structural primitives need their own proofs.
func (l *lowering) nullableObjectUnion(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion == 0 || !l.includesNull(proven) || !l.includesUndefined(proven) {
		return false
	}
	present := l.checker.GetNonNullableType(proven)
	held, known := l.representation(present)
	return known && held == ir.Object && present.Flags()&checker.TypeFlagsObject != 0
}

func unionObservation(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node, parent = parent, parent.Parent
	}
	if comparedWithUndefined(node) || parent != nil && parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		case ast.KindQuestionQuestionToken:
			return binary.Left == node
		}
	}
	return false
}

func (l *lowering) checkedNullableObject(node *ast.Node, read ir.Expression) ir.Expression {
	observed := l.concrete(l.checker.GetTypeAtLocation(node))
	narrowed, known := l.representation(observed)
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual != nil && l.nullableObjectUnion(contextual) {
		return read
	}
	if !known || narrowed == ir.Union || unionObservation(node) {
		return read
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	null := ir.Expression(ir.IsNull{Value: held})
	matches := ir.Expression(ir.Binary{Operator: ir.And,
		Left:  ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant("object")}},
		Right: ir.Unary{Operator: ir.Not, Operand: null}})
	if l.includesNull(observed) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: null}
	}
	if l.includesUndefined(observed) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	result := ir.Expression(ir.Narrow{Value: held, To: ir.Object})
	if l.includesNull(observed) {
		result = ir.Conditional{Condition: null, WhenTrue: ir.Null{Of: ir.Object}, WhenNot: result, Of: ir.Object}
	}
	call := b.finish("narrowed_nullable_object", result)
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return call
}

// A legacy object-or-null pointer uses NULL for null. Normalize it when a
// contextual tagged slot receives it, evaluating the source only once.
func (l *lowering) nullableObjectBoundary(node *ast.Node, value ir.Expression) ir.Expression {
	if value == nil || value.Type() != ir.Object {
		return value
	}
	own := l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.includesNull(own) || l.includesUndefined(own) {
		return value
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual == nil {
		return value
	}
	target, known := l.representation(contextual)
	if !known || target != ir.Union {
		return value
	}
	if _, literal := value.(ir.Null); literal {
		return value
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	return b.finish("nullable_object_to_union", ir.Conditional{Condition: ir.IsNull{Value: held}, WhenTrue: ir.Box{Value: ir.Null{}}, WhenNot: ir.Box{Value: held}, Of: ir.Union})
}

// Optional access skips both nullish tags; only an actual object reaches the read.
func (l *lowering) optionalNullableObject(read ir.Expression) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	absent := ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}
	return b.finish("optional_nullable_object", ir.Conditional{Condition: absent, WhenTrue: ir.Undefined{Of: ir.Object}, WhenNot: ir.Narrow{Value: held, To: ir.Object}, Of: ir.Object})
}

// A value boundary can normalize a pointer. Shared fields, elements and callable
// signatures cannot change their storage convention without a copy or adapter.
func (l *lowering) nullableObjectSharedView(from, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	from, to = l.present(l.concrete(from)), l.present(l.concrete(to))
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return false
	}
	visited[[2]*checker.Type{from, to}] = true
	shared := func(inside, viewed *checker.Type) bool {
		inside, viewed = l.concrete(inside), l.concrete(viewed)
		before, knownBefore := l.representation(inside)
		after, knownAfter := l.representation(viewed)
		if knownBefore && knownAfter && before != after &&
			(before == ir.Object && l.includesNull(inside) && after == ir.Union ||
				after == ir.Object && l.includesNull(viewed) && before == ir.Union) {
			return true
		}
		return l.nullableObjectSharedView(inside, viewed, visited)
	}
	a := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	b := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(a) > 0 && len(b) > 0 {
		for i, parameter := range a[0].Parameters() {
			if i < len(b[0].Parameters()) && shared(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(b[0].Parameters()[i])) {
				return true
			}
		}
		return shared(l.checker.GetReturnTypeOfSignature(a[0]), l.checker.GetReturnTypeOfSignature(b[0]))
	}
	if l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		a, b := l.typeArguments(from), l.typeArguments(to)
		for i := range a {
			if i < len(b) && shared(a[i], b[i]) {
				return true
			}
		}
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(to) {
		if field.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if previous := l.checker.GetPropertyOfType(from, field.Name); previous != nil && shared(l.checker.GetTypeOfSymbol(previous), l.checker.GetTypeOfSymbol(field)) {
			return true
		}
	}
	return false
}

func freshNullableObjectStorage(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindArrayLiteralExpression && node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, part := range literalParts(node) {
		if part.Kind == ast.KindSpreadElement || part.Kind == ast.KindSpreadAssignment {
			return false
		}
	}
	return true
}

func (l *lowering) nullableObjectViewHazard(node *ast.Node, contextual *checker.Type) bool {
	if !freshNullableObjectStorage(node) && l.nullableObjectSharedView(l.checker.GetTypeAtLocation(node), contextual, map[[2]*checker.Type]bool{}) {
		return true
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, part := range literalParts(node) {
			if part.Kind == ast.KindSpreadAssignment && l.nullableObjectSharedView(l.checker.GetTypeAtLocation(part.AsSpreadAssignment().Expression), contextual, map[[2]*checker.Type]bool{}) {
				return true
			}
		}
	}
	return false
}

// These existing typeof paths retain the lookup slot's presence until it is
// classified. An ordinary nullable pointer lookup loses that distinction.
func nullableObjectLookupTypeOf(node *ast.Node, value ir.Expression) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindTypeOfExpression {
		return false
	}
	switch value.(type) {
	case ir.ArrayIndex, ir.MapGet, ir.ArrayPop:
		return true
	}
	return false
}
