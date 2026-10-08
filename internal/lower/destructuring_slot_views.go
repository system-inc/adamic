package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A readonly view can widen a field without converting its stored value. Fresh
// literals fit their fields, but existing objects need checked storage views.
// Keep this boundary at the owned read until that representation is available.
func (l *lowering) destructuringSlotView(field string, of ir.Type) bool {
	if of != ir.MaybeBoolean && of != ir.Union {
		return false
	}
	var differs func(*checker.Type, *checker.Type, map[[2]*checker.Type]bool, int) bool
	differs = func(from, to *checker.Type, seen map[[2]*checker.Type]bool, depth int) bool {
		if from == nil || to == nil || from == to || seen[[2]*checker.Type{from, to}] {
			return false
		}
		if depth > 16 {
			return true
		}
		seen[[2]*checker.Type{from, to}] = true
		if from.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range from.Types() {
				if differs(member, to, seen, depth+1) {
					return true
				}
			}
			return false
		}
		if to.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range to.Types() {
				if l.checker.IsTypeAssignableTo(from, member) && differs(from, member, seen, depth+1) {
					return true
				}
			}
			return false
		}
		from, to = l.concrete(from), l.concrete(to)
		if from.Flags()&checker.TypeFlagsObject == 0 || to.Flags()&checker.TypeFlagsObject == 0 {
			return false
		}
		if (l.checker.IsArrayType(from) || checker.IsTupleType(from)) && (l.checker.IsArrayType(to) || checker.IsTupleType(to)) {
			left, right := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
			for index, inside := range left {
				at := index
				if !checker.IsTupleType(to) {
					at = 0
				}
				if at < len(right) && differs(inside, right[at], seen, depth+1) {
					return true
				}
			}
		}
		leftCalls := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
		rightCalls := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
		if len(leftCalls) == 1 && len(rightCalls) == 1 {
			if differs(l.checker.GetReturnTypeOfSignature(leftCalls[0]), l.checker.GetReturnTypeOfSignature(rightCalls[0]), seen, depth+1) {
				return true
			}
			left, right := leftCalls[0].Parameters(), rightCalls[0].Parameters()
			for index := 0; index < len(left) && index < len(right); index++ {
				if differs(l.checker.GetTypeOfSymbol(right[index]), l.checker.GetTypeOfSymbol(left[index]), seen, depth+1) {
					return true
				}
			}
		}
		for _, target := range l.checker.GetPropertiesOfType(to) {
			if target.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}
			source := l.checker.GetPropertyOfType(from, target.Name)
			if source == nil {
				continue
			}
			inside, viewed := l.checker.GetTypeOfSymbol(source), l.checker.GetTypeOfSymbol(target)
			stored, known := l.representation(inside)
			expected, represented := l.representation(viewed)
			if target.Name == field && expected == of && (!known || !represented || stored != expected) {
				return true
			}
			if differs(inside, viewed, seen, depth+1) {
				return true
			}
		}
		return false
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if (ast.IsExpressionNode(node) || node.Kind == ast.KindSpreadAssignment) && node.Kind != ast.KindObjectLiteralExpression && node.Kind != ast.KindArrayLiteralExpression && node.Kind != ast.KindConditionalExpression && node.Kind != ast.KindParenthesizedExpression {
			from := l.checker.GetTypeAtLocation(node)
			to := l.checker.GetContextualType(node, checker.ContextFlagsNone)
			if node.Kind == ast.KindAsExpression {
				from, to = l.checker.GetTypeAtLocation(node.AsAsExpression().Expression), l.checker.GetTypeAtLocation(node)
			}
			if node.Kind == ast.KindSpreadAssignment {
				from = l.checker.GetTypeAtLocation(node.AsSpreadAssignment().Expression)
				to = l.checker.GetContextualType(node.Parent, checker.ContextFlagsNone)
			}
			if differs(from, to, map[[2]*checker.Type]bool{}, 0) {
				found = true
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range l.program.Files() {
		if !load.IsLibrary(file) && file.AsNode().ForEachChild(visit) {
			break
		}
	}
	return found
}
