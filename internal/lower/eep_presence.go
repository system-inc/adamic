package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Native tuples contain numbered slots, not an own length field. Keep unsupported
// reads out of ordinary field lowering until runtime tuple lengths are represented.
func (lowering *lowering) eepPropertyPresence(node *ast.Node) error {
	access := node.AsPropertyAccessExpression()
	name := node.Name().Text()
	receiver := lowering.concrete(lowering.checker.GetTypeAtLocation(access.Expression))
	if name == "length" && (access.QuestionDotToken != nil || !checker.IsTupleType(receiver)) {
		members := []*checker.Type{receiver}
		if receiver.Flags()&checker.TypeFlagsUnion != 0 {
			members = receiver.Types()
		}
		for _, member := range members {
			if checker.IsTupleType(member) {
				return lowering.notYet(
					node,
					"a union or optional tuple length; declare arrays where the values are constructed, or carry an explicit length alongside each tuple",
				)
			}
		}
	}
	// finishAccessors dispatches by name. Its fallback is an own-field load,
	// which cannot select a prototype method through a structural interface.
	if lowering.accessorNames[name] {
		symbol := lowering.checker.GetSymbolAtLocation(node.Name())
		if symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 {
			return lowering.notYet(
				node,
				"a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation",
			)
		}
	}
	// An earlier initializer can enter a static method before this own field
	// exists. Native live-parent lookup currently assumes the parent names it.
	if lowering.instance != nil &&
		lowering.instance.static &&
		ast.SkipParentheses(access.Expression).Kind == ast.KindThisKeyword {
		if symbol := lowering.checker.GetSymbolAtLocation(node.Name()); symbol != nil {
			for _, field := range symbol.Declarations {
				if field.Kind != ast.KindPropertyDeclaration ||
					!ast.HasSyntacticModifier(field, ast.ModifierFlagsStatic) ||
					field.Parent == nil ||
					field.Parent.Kind != ast.KindClassDeclaration {
					continue
				}
				for _, earlier := range field.Parent.Members() {
					if earlier == field {
						break
					}
					if earlier.Kind != ast.KindPropertyDeclaration ||
						!ast.HasSyntacticModifier(earlier, ast.ModifierFlagsStatic) {
						continue
					}
					initializer := earlier.AsPropertyDeclaration().Initializer
					if initializer != nil && staticInitializerCallsThis(initializer) {
						return lowering.notYet(
							node,
							"a static field read with an earlier initializer calling this; move that initializer after the fields read by static methods, or evaluate the call after class construction",
						)
					}
				}
			}
		}
	}
	return nil
}

func staticInitializerCallsThis(node *ast.Node) bool {
	found := false
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if child.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(child.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression &&
				ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression).Kind == ast.KindThisKeyword {
				found = true
			}
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	return found
}
