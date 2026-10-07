package lower

import (
	"reflect"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Class layouts need concrete checker types, not just a generic function's native
// representations. Keep these functions' caches separate for each actual type vector.
func (l *lowering) classGenericCall(node *ast.Node) (ir.Expression, bool, error) {
	symbol := l.symbol(ast.SkipParentheses(node.AsCallExpression().Expression))
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, false, nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.TypeParameters()) == 0 {
		return nil, false, nil
	}
	needsTypes := l.genericUsesClasses(declaration, map[*ast.Node]bool{})
	constrained := false
	for _, parameter := range declaration.TypeParameters() {
		constrained = constrained || parameter.AsTypeParameterDeclaration().Constraint != nil
	}
	if !needsTypes && !constrained {
		return nil, false, nil
	}
	resolved := l.checker.GetResolvedSignature(node)
	if resolved == nil {
		return nil, true, l.notYet(node, "a generic call whose concrete signature isn't known")
	}
	// Like instantiate.go's shim bridge, this belongs in the checker shim. Reflection
	// checks the field's existence instead of depending on the private struct's offsets.
	field := reflect.ValueOf(resolved).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return nil, true, l.notYet(node, "a generic call whose checker type mapper isn't exposed")
	}
	resolvedMapper := (*typeMapper)(field.UnsafePointer())
	sources, targets := []*checker.Type{}, []*checker.Type{}
	key := l.program.Where(declaration)
	for _, parameter := range declaration.TypeParameters() {
		source := l.checker.GetTypeAtLocation(parameter.Name())
		target := l.concrete(instantiateType(l.checker, source, resolvedMapper))
		sources = append(sources, source)
		targets = append(targets, target)
		key += "," + strconv.Itoa(int(target.Id()))
	}
	mapper := newTypeMapper(sources, targets)
	if err := l.nominalTypeArguments(declaration, targets, mapper, node); err != nil {
		return nil, true, err
	}
	if !needsTypes {
		return nil, false, nil
	}
	if l.classGenericInstances == nil {
		l.classGenericInstances = map[string]map[string]int{}
	}
	bucket := l.classGenericInstances[key]
	if bucket == nil {
		bucket = map[string]int{}
		l.classGenericInstances[key] = bucket
	}
	outerMapper, outerInstances := l.typeMapper, l.genericInstances
	l.typeMapper, l.genericInstances = mapper, bucket
	defer func() { l.typeMapper, l.genericInstances = outerMapper, outerInstances }()
	value, err := l.call(node)
	return value, true, err
}

func (l *lowering) genericUsesClasses(declaration *ast.Node, seen map[*ast.Node]bool) bool {
	if seen[declaration] {
		return false
	}
	seen[declaration] = true
	needed := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) {
			symbol := l.symbol(node)
			if symbol != nil && len(symbol.Declarations) == 1 {
				target := symbol.Declarations[0]
				if target.Kind == ast.KindClassDeclaration && len(target.TypeParameters()) > 0 {
					needed = true
					return true
				}
				if node.Parent != nil && node.Parent.Kind == ast.KindCallExpression && target.Kind == ast.KindFunctionDeclaration && len(target.TypeParameters()) > 0 && l.genericUsesClasses(target, seen) {
					needed = true
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	declaration.ForEachChild(visit)
	return needed
}

// A structural fake cannot satisfy a nominal bound, including a dependent bound
// such as U extends Box<T>. The concrete mapper preserves both parameters' meanings.
func (l *lowering) nominalTypeArguments(declaration *ast.Node, arguments []*checker.Type, mapper *typeMapper, where *ast.Node) error {
	for index, parameter := range declaration.TypeParameters() {
		node := parameter.AsTypeParameterDeclaration().Constraint
		if node == nil {
			continue
		}
		constraint := l.checker.GetTypeAtLocation(node)
		if mapper != nil {
			constraint = instantiateType(l.checker, constraint, mapper)
		}
		if mismatch := l.nominalMismatch(arguments[index], constraint, map[[2]*checker.Type]bool{}); mismatch != nil {
			return &Refused{Where: l.program.Where(where), What: "a generic argument without nominal ancestry seen as " + l.checker.TypeToString(mismatch), Fix: "use that class or a subclass as the argument; use an interface for structural constraints (adamic/nominal-class)"}
		}
	}
	return nil
}
