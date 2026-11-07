package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Array reads have no field symbol. Recover their stored element type before flow narrowing.
func (lowering *lowering) enumStoredType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	if symbol := lowering.flagValueSymbol(node); symbol != nil {
		return lowering.checker.GetTypeOfSymbol(symbol)
	}
	if node.Kind == ast.KindElementAccessExpression {
		stored := lowering.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression)
		if lowering.isArrayType(stored) {
			if arguments := lowering.checker.GetTypeArguments(stored); len(arguments) == 1 {
				return arguments[0]
			}
		}
	}
	return nil
}

// Follow the declared type, not the flow type that has already become never. A never alias
// keeps its initializer's enum identity, but never acquires a runtime member proof.
func (lowering *lowering) enumNeverIdentity(node *ast.Node, seen map[*ast.Node]bool) *ast.Symbol {
	node = ast.SkipParentheses(node)
	if lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever == 0 || seen[node] {
		return nil
	}
	seen[node] = true
	symbol := lowering.flagValueSymbol(node)
	stored := lowering.enumStoredType(node)
	if stored == nil {
		return nil
	}
	declared := lowering.concrete(stored)
	for _, member := range lowering.definedMembers(declared) {
		identity := lowering.enumIdentity(member)
		if lowering.numericEnum(identity) {
			return identity
		}
	}
	if symbol != nil {
		if declaration := symbol.ValueDeclaration; declaration != nil &&
			declaration.Kind == ast.KindVariableDeclaration {
			if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
				identity := lowering.enumIdentity(lowering.checker.GetTypeAtLocation(initializer))
				if lowering.numericEnum(identity) {
					return identity
				}
				return lowering.enumNeverIdentity(initializer, seen)
			}
		}
	}
	return nil
}

// Only immutable aliases of actual members prove membership. Parameters, mutable slots,
// numeric literals, casts, arithmetic and flag-domain proofs do not prove a declared member.
func (lowering *lowering) enumMemberOrigin(node *ast.Node, identity *ast.Symbol, seen map[*ast.Node]bool) bool {
	node = ast.SkipParentheses(node)
	if member := lowering.enumMember(node); member != nil {
		return lowering.symbol(member.Parent.Name()) == identity
	}
	if !ast.IsIdentifier(node) || seen[node] {
		return false
	}
	seen[node] = true
	if symbol := lowering.symbol(node); symbol != nil && symbol.ValueDeclaration != nil {
		declaration := symbol.ValueDeclaration
		if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
			if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
				return lowering.enumMemberOrigin(initializer, identity, seen)
			}
		}
	}
	return false
}

// An ordinary IR helper preserves evaluation order and reads the operand exactly once. Its
// panic is shared by both backends; never itself does not need a new machine representation.
func (lowering *lowering) enumNeverCheck(node *ast.Node, value ir.Expression, identity *ast.Symbol) ir.Expression {
	of := ir.Number
	if contextual := lowering.slotType(node); contextual != nil {
		if representation, known := lowering.representation(contextual); known && !slotless(representation) {
			of = representation
		}
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindReturnStatement && lowering.function != nil {
		of = lowering.function.Returns
	}
	function := len(lowering.result.Functions)
	local := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "enum_never_value", Type: value.Type(), Function: function},
	)
	read := ir.Read{Local: local, Of: value.Type()}
	var printed ir.Expression = ir.NumberToString{Value: read}
	if value.Type().IsMaybe() {
		printed = ir.MaybeToString{Value: read}
	} else if value.Type() == ir.Union {
		printed = ir.UnionToString{Value: read}
	}
	message := ir.Concat{Parts: []ir.Expression{
		ir.StringConstant{Index: lowering.constant("unreachable value ")},
		printed,
		ir.StringConstant{Index: lowering.constant(" for numeric enum " + identity.Name)},
	}}
	lowering.result.Functions = append(lowering.result.Functions, ir.Function{
		Name: "enum_never", Parameters: []int{local}, Returns: of,
		Body: []ir.Statement{ir.Panic{Message: message}},
	})
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: of}
}
