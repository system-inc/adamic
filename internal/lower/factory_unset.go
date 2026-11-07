package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Discover contracts before function bodies or global save locals are lowered.
// This retains the existing conservative field-name view propagation.
func (lowering *lowering) factoryFieldOrigin(name string) string {
	if lowering.result.FactoryFields == nil {
		lowering.result.FactoryFields = map[string]string{}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindAsExpression && lowering.constructionCast(node) {
				factory := "literal factory"
				for parent := node.Parent; parent != nil; parent = parent.Parent {
					if parent.Kind == ast.KindFunctionDeclaration ||
						parent.Kind == ast.KindFunctionExpression ||
						parent.Kind == ast.KindArrowFunction {
						if parent.Name() != nil {
							factory = parent.Name().Text()
						}
						break
					}
				}
				for _, field := range lowering.checker.GetPropertiesOfType(
					lowering.concrete(lowering.checker.GetTypeAtLocation(node)),
				) {
					if old := lowering.result.FactoryFields[field.Name]; old == "" {
						lowering.result.FactoryFields[field.Name] = factory
					} else if old != factory && !strings.Contains(old, factory) {
						lowering.result.FactoryFields[field.Name] = old + ", " + factory
					}
				}
			}
			return node.ForEachChild(visit)
		}
		for _, file := range lowering.program.Files() {
			file.AsNode().ForEachChild(visit)
		}

	}
	return lowering.result.FactoryFields[name]
}

func (lowering *lowering) savedFactoryOrigin(node *ast.Node, seen map[*ast.Symbol]bool) string {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindPropertyAccessExpression {
		if layout, _, known := lowering.nodeArrayLayoutOf(
			lowering.checker.GetNonNullableType(
				lowering.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression),
			),
		); known {
			if _, readError := layout.Read(node.Name().Text()); readError == nil {
				return "field '" + node.Name().Text() + "', factory 'NodeArray allocation'"
			}
		}
		if origin := lowering.factoryFieldOrigin(node.Name().Text()); origin != "" {
			return "field '" + node.Name().Text() + "', factory '" + origin + "'"
		}
	}
	if !ast.IsIdentifier(node) {
		return ""
	}
	symbol := lowering.symbol(node)
	if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
		return ""
	}
	seen[symbol] = true
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 ||
		declaration.Type() != nil ||
		declaration.Initializer() == nil {
		return ""
	}
	return lowering.savedFactoryOrigin(declaration.Initializer(), seen)
}

func untypedSave(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent != nil &&
		node.Parent.Kind == ast.KindVariableDeclaration &&
		node.Parent.Initializer() == node &&
		node.Parent.Type() == nil &&
		node.Parent.Parent.Flags&ast.NodeFlagsConst != 0
}

// Observation and a bounded save carry an existing tagged union. A receiving
// optional slot still checks T, permitting its real nullish members explicitly.
func (lowering *lowering) unsetObserved(node *ast.Node) bool {
	if comparedWithUndefined(node) || untypedSave(node) {
		return true
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindTypeOfExpression || parent.Kind == ast.KindTemplateSpan {
		return true
	}
	return parent.Kind == ast.KindCallExpression &&
		lowering.isLibraryGlobal(parent.AsCallExpression().Expression, "String")
}

func (lowering *lowering) unsetSpecialization(
	node *ast.Node,
	value ir.Expression,
	declared *checker.Type,
	origin string,
) (ir.Expression, error) {
	if lowering.unsetObserved(node) {
		return value, nil
	}
	present := lowering.checker.GetNonNullableType(declared)
	held, known := lowering.representation(present)
	if !known || held < ir.Number || held > ir.String {
		return nil, lowering.notYet(node, "an unset field use requiring a non-scalar result specialization")
	}
	b := lowering.libraryArrayBuilder([]ir.Expression{value})
	observed := b.read(b.parameters[0])
	name := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[held]
	matches := ir.Expression(ir.Binary{
		Operator: ir.Equal,
		Left:     ir.TypeOf{Value: observed},
		Right:    ir.StringConstant{Index: lowering.constant(name)},
	})
	if allowed := lowering.viewLiterals(declared); len(allowed) != 0 {
		var fits ir.Expression
		for _, allowedValue := range allowed {
			test := ir.Expression(
				ir.Binary{Operator: ir.Equal, Left: ir.Narrow{Value: observed, To: held}, Right: allowedValue},
			)
			if fits == nil {
				fits = test
			} else {
				fits = ir.Binary{Operator: ir.Or, Left: fits, Right: test}
			}
		}
		matches = ir.Binary{Operator: ir.And, Left: matches, Right: fits}
	}
	optional := lowering.includesUndefined(lowering.checker.GetTypeAtLocation(node))
	if contextual := lowering.slotType(node); contextual != nil && lowering.includesUndefined(contextual) {
		optional = true
	}
	if optional {
		matches = ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: observed}, Right: matches}
	}
	message := "unset use failed: " +
		origin +
		", use '" +
		sourceExpression(node) +
		"' at " +
		lowering.program.Where(node) +
		"; expected " +
		lowering.checker.TypeToString(declared)
	b.body = append(b.body, ir.If{
		Condition: ir.Unary{Operator: ir.Not, Operand: matches},
		Then:      []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: lowering.constant(message)}}},
	})
	result := ir.Expression(ir.Narrow{Value: observed, To: held})
	if optional && (held == ir.Number || held == ir.Boolean) {
		result = ir.Conditional{
			Condition: ir.IsUndefined{Value: observed},
			WhenTrue:  ir.MaybeOf{Of: ir.Maybe(held)},
			WhenNot:   ir.MaybeOf{Of: ir.Maybe(held), Value: result},
		}
	}
	return b.finish("factory_unset_specialization", result), nil
}

func (lowering *lowering) factoryFieldRead(
	node *ast.Node,
	object ir.Expression,
	name string,
) (ir.Expression, bool, error) {
	origin := lowering.factoryFieldOrigin(name)
	if origin == "" {
		return nil, false, nil
	}

	field := lowering.checker.GetSymbolAtLocation(node.Name())
	if field == nil || accessorSymbol(field) || field.Flags&ast.SymbolFlagsMethod != 0 {
		return nil, false, nil
	}
	declared := lowering.checker.GetTypeOfSymbol(field)
	present := lowering.checker.GetNonNullableType(declared)
	if !interfaceScalar(present) {
		return nil, false, nil
	}
	value := ir.DynamicProperty{Object: fit(object, ir.Union), Name: name}
	checked, unsetSpecializationError := lowering.unsetSpecialization(
		node,
		value,
		declared,
		"field '"+name+"', factory '"+origin+"'",
	)
	return checked, true, unsetSpecializationError
}
