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

// An unsupported result representation must not hide an already-provable type
// argument refusal (for example an array of string | null). Check explicit
// arguments before value lowering asks how the call result is represented.
func (l *lowering) explicitGenericArrayViews(node *ast.Node) error {
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	if call.TypeArguments == nil {
		return nil
	}
	callee := ast.SkipParentheses(call.Expression)
	var declaration *ast.Node
	if symbol := l.symbol(callee); symbol != nil {
		for _, candidate := range symbol.Declarations {
			if candidate.Kind == ast.KindFunctionDeclaration && candidate.Body() != nil && len(candidate.TypeParameters()) > 0 {
				declaration = candidate
				break
			}
		}
	}
	if declaration == nil || len(declaration.TypeParameters()) != len(call.TypeArguments.Nodes) {
		return nil
	}
	sources, targets := []*checker.Type{}, []*checker.Type{}
	arguments := map[*checker.Type]*checker.Type{}
	for index, parameter := range declaration.TypeParameters() {
		source := l.checker.GetTypeAtLocation(parameter.Name())
		target := l.concrete(l.checker.GetTypeAtLocation(call.TypeArguments.Nodes[index]))
		if l.unboundArrayViewArgument(target, map[*checker.Type]bool{}) {
			return nil
		}
		sources = append(sources, source)
		targets = append(targets, target)
		arguments[source] = target
	}
	outer := l.typeMapper
	l.typeMapper = newTypeMapper(sources, targets)
	defer func() { l.typeMapper = outer }()
	return l.refuseInstantiatedArrayViews(node, declaration, arguments)
}

func (l *lowering) unboundArrayViewArgument(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if seen[proven] {
		return false
	}
	seen[proven] = true
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return true
	}
	var members []*checker.Type
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		members = proven.Types()
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		members = l.checker.GetTypeArguments(proven)
	}
	for _, member := range members {
		if l.unboundArrayViewArgument(member, seen) {
			return true
		}
	}
	return false
}
