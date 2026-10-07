package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

const nullableTagReason = "a reference with both null and undefined (nullable reference needs an empty-case tag)"

// nullableObservation passes the operand once before testing and observing it. Ordinary IR keeps
// ownership and evaluation visible to both backends and every analysis.
func (l *lowering) nullableObservation(name string, value, empty ir.Expression, present func(ir.Expression) ir.Expression) ir.Expression {
	function := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "reference", Type: value.Type(), Function: function})
	read := ir.Read{Local: local, Of: value.Type()}
	result := ir.Conditional{Condition: ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: read}, Right: ir.IsUndefined{Value: read}}, WhenTrue: empty, WhenNot: present(read)}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "nullable_" + name, Parameters: []int{local}, Returns: empty.Type(), Body: []ir.Statement{ir.Return{Value: result}}})
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: empty.Type()}
}

func (l *lowering) nullableUse(node *ast.Node) error {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	own := l.concrete(l.checker.GetTypeAtLocation(node))
	if l.includesNull(own) && l.includesUndefined(own) {
		// typeof consumes a lookup's presence slot immediately. Its lowering accepts only
		// ArrayIndex, MapGet and ArrayPop; a stored value still needs an empty-case tag.
		lookup := ast.SkipParentheses(node)
		if outer.Parent != nil && outer.Parent.Kind == ast.KindTypeOfExpression &&
			(lookup.Kind == ast.KindElementAccessExpression || lookup.Kind == ast.KindCallExpression) {
			return nil
		}
		return l.notYet(node, nullableTagReason)
	}
	// Console observes its argument immediately as text; it never stores the wider declared type.
	if parent := outer.Parent; parent != nil && parent.Kind == ast.KindCallExpression && l.isConsole(parent.AsCallExpression().Expression) {
		return nil
	}
	// A contextual output type also propagates into operator operands. Those operands are
	// observed under their own static type, rather than stored through the output's view.
	if parent := outer.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			if parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
				return nil
			}
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindTypeOfExpression,
			ast.KindTemplateSpan, ast.KindConditionalExpression:
			return nil
		}
	}
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
		if !l.nullableViewsMatch(own, contextual, map[[2]*checker.Type]bool{}) {
			return l.notYet(node, "a nullable reference view from "+l.checker.TypeToString(own)+" to "+l.checker.TypeToString(l.concrete(contextual))+" changes its empty case or adds a second empty case (nullable reference needs an empty-case tag)")
		}
	}
	return nil
}

// Check empty cases before stripping unions. A structural view can hide this change in a field,
// element, map argument or function signature, even when the outer pointers have the same layout.
func (l *lowering) nullableViewsMatch(from, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	from, to = l.concrete(from), l.concrete(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	if (l.includesNull(from) && l.includesUndefined(to)) || (l.includesUndefined(from) && l.includesNull(to)) {
		return false
	}
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil {
		return true
	}
	a, b := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall), l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(a) > 0 && len(b) > 0 {
		ap, bp := a[0].Parameters(), b[0].Parameters()
		for i := 0; i < len(ap) && i < len(bp); i++ {
			if !l.nullableViewsMatch(l.checker.GetTypeOfSymbol(bp[i]), l.checker.GetTypeOfSymbol(ap[i]), visited) {
				return false
			}
		}
		return l.nullableViewsMatch(l.checker.GetReturnTypeOfSignature(a[0]), l.checker.GetReturnTypeOfSignature(b[0]), visited)
	}
	if l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		aa, ba := l.typeArguments(from), l.typeArguments(to)
		for i := 0; i < len(aa) && i < len(ba); i++ {
			if !l.nullableViewsMatch(aa[i], ba[i], visited) {
				return false
			}
		}
		return true
	}
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil {
			if viewed.Flags&ast.SymbolFlagsOptional != 0 && l.includesNull(l.checker.GetTypeOfSymbol(inside)) {
				return false
			}
			if !l.nullableViewsMatch(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed), visited) {
				return false
			}
		}
	}
	return true
}
