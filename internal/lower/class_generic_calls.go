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
func (lowering *lowering) classGenericCall(node *ast.Node) (ir.Expression, bool, error) {
	symbol := lowering.symbol(ast.SkipParentheses(node.AsCallExpression().Expression))
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, false, nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.TypeParameters()) == 0 {
		return nil, false, nil
	}
	needsTypes := lowering.genericUsesClasses(declaration, map[*ast.Node]bool{})
	constrained := false
	for _, parameter := range declaration.TypeParameters() {
		constrained = constrained || parameter.AsTypeParameterDeclaration().Constraint != nil
	}
	if !needsTypes && !constrained {
		return nil, false, nil
	}
	resolved := lowering.checker.GetResolvedSignature(node)
	if resolved == nil {
		return nil, true, lowering.notYet(node, "a generic call whose concrete signature isn't known")
	}
	// Like instantiate.go's shim bridge, this belongs in the checker shim. Reflection
	// checks the field's existence instead of depending on the private struct's offsets.
	field := reflect.ValueOf(resolved).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return nil, true, lowering.notYet(node, "a generic call whose checker type mapper isn't exposed")
	}
	resolvedMapper := (*typeMapper)(field.UnsafePointer())
	sources, targets := []*checker.Type{}, []*checker.Type{}
	key := lowering.program.Where(declaration)
	for _, parameter := range declaration.TypeParameters() {
		source := lowering.checker.GetTypeAtLocation(parameter.Name())
		target := lowering.concrete(instantiateType(lowering.checker, source, resolvedMapper))
		sources = append(sources, source)
		targets = append(targets, target)
		key += "," + strconv.Itoa(int(target.Id()))
	}
	mapper := newTypeMapper(sources, targets)
	if nominalTypeArgumentsError := lowering.nominalTypeArguments(
		declaration,
		targets,
		mapper,
		node,
	); nominalTypeArgumentsError != nil {
		return nil, true, nominalTypeArgumentsError
	}
	if !needsTypes {
		return nil, false, nil
	}
	if lowering.classGenericInstances == nil {
		lowering.classGenericInstances = map[string]map[string]int{}
	}
	bucket := lowering.classGenericInstances[key]
	if bucket == nil {
		bucket = map[string]int{}
		lowering.classGenericInstances[key] = bucket
	}
	outerMapper, outerInstances := lowering.typeMapper, lowering.genericInstances
	lowering.typeMapper, lowering.genericInstances = mapper, bucket
	defer func() { lowering.typeMapper, lowering.genericInstances = outerMapper, outerInstances }()
	value, callError := lowering.call(node)
	return value, true, callError
}

func (lowering *lowering) genericUsesClasses(declaration *ast.Node, seen map[*ast.Node]bool) bool {
	if seen[declaration] {
		return false
	}
	seen[declaration] = true
	needed := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) {
			symbol := lowering.symbol(node)
			if symbol != nil && len(symbol.Declarations) == 1 {
				target := symbol.Declarations[0]
				if target.Kind == ast.KindClassDeclaration && len(target.TypeParameters()) > 0 {
					needed = true
					return true
				}
				if node.Parent != nil &&
					node.Parent.Kind == ast.KindCallExpression &&
					target.Kind == ast.KindFunctionDeclaration &&
					len(target.TypeParameters()) > 0 &&
					lowering.genericUsesClasses(target, seen) {
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
func (lowering *lowering) nominalTypeArguments(
	declaration *ast.Node,
	arguments []*checker.Type,
	mapper *typeMapper,
	where *ast.Node,
) error {
	for index, parameter := range declaration.TypeParameters() {
		node := parameter.AsTypeParameterDeclaration().Constraint
		if node == nil {
			continue
		}
		constraint := lowering.checker.GetTypeAtLocation(node)
		if mapper != nil {
			constraint = instantiateType(lowering.checker, constraint, mapper)
		}
		if mismatch := lowering.nominalMismatch(
			arguments[index],
			constraint,
			map[[2]*checker.Type]bool{},
		); mismatch != nil {
			return &Refused{
				Where: lowering.program.Where(where),
				What:  "a generic argument without nominal ancestry seen as " + lowering.checker.TypeToString(mismatch),
				Fix:   "use that class or a subclass as the argument; use an interface for structural constraints (adamic/nominal-class)",
			}
		}
	}
	return nil
}
