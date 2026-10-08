package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Only NonNullable<T>[] -> T[] for a module function parameter is deferred.
// Generic classes and unrelated mutable views keep the declaration-time check.
func (l *lowering) deferredNullableArrayView(from, to, source, target *checker.Type) bool {
	if !l.checker.IsArrayType(from) || !l.checker.IsArrayType(to) || target.Flags()&checker.TypeFlagsTypeParameter == 0 || !identicalTypes(l.checker, source, l.checker.GetNonNullableType(target)) {
		return false
	}
	symbol := target.Symbol()
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindTypeParameter && declaration.Parent != nil && declaration.Parent.Kind == ast.KindFunctionDeclaration {
			return true
		}
	}
	return false
}

// Revisit the declaration with its concrete mapper before emitting its body.
// Mutable element comparisons retain null/undefined, even though ordinary value
// views may omit undefined. A failed obligation belongs to the instantiating use.
func (l *lowering) refuseInstantiatedArrayViews(call, declaration *ast.Node, arguments map[*checker.Type]*checker.Type) error {
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if ast.IsFunctionLike(node) {
			return false
		}
		if err := l.refuseWidening(node); err != nil {
			found = err
			return true
		}
		return node.ForEachChild(visit)
	}
	declaration.Body().ForEachChild(visit)
	if found == nil {
		return nil
	}
	refused, ok := found.(*Refused)
	if !ok {
		return found
	}
	names := []string{}
	for _, parameter := range declaration.TypeParameters() {
		proven := l.checker.GetTypeAtLocation(parameter.Name())
		if argument := arguments[proven]; argument != nil {
			names = append(names, parameter.Name().Text()+" = "+l.checker.TypeToString(argument))
		}
	}
	return &Refused{Where: l.program.Where(call), What: "type argument " + strings.Join(names, ", ") + " breaks invariance: " + refused.What, Fix: "use a type argument that includes neither null nor undefined, or keep the array view readonly (adamic/invariant-mutable)"}
}
