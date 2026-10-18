package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Array reads have no field symbol. Recover their stored element type before flow narrowing.
func (l *lowering) enumStoredType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	if symbol := l.flagValueSymbol(node); symbol != nil {
		return l.checker.GetTypeOfSymbol(symbol)
	}
	if node.Kind == ast.KindElementAccessExpression {
		stored := l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression)
		if l.checker.IsArrayType(stored) {
			if arguments := l.checker.GetTypeArguments(stored); len(arguments) == 1 {
				return arguments[0]
			}
		}
	}
	return nil
}

// Follow the declared type, not the flow type that has already become never. A never alias
// keeps its initializer's enum identity, but never acquires a runtime member proof.
func (l *lowering) enumNeverIdentity(node *ast.Node, seen map[*ast.Node]bool) *ast.Symbol {
	node = ast.SkipParentheses(node)
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever == 0 || seen[node] {
		return nil
	}
	seen[node] = true
	symbol := l.flagValueSymbol(node)
	stored := l.enumStoredType(node)
	if stored == nil {
		return nil
	}
	declared := l.concrete(stored)
	for _, member := range l.definedMembers(declared) {
		identity := l.enumIdentity(member)
		if l.numericEnum(identity) {
			return identity
		}
	}
	if symbol != nil {
		if declaration := symbol.ValueDeclaration; declaration != nil && declaration.Kind == ast.KindVariableDeclaration {
			if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
				identity := l.enumIdentity(l.checker.GetTypeAtLocation(initializer))
				if l.numericEnum(identity) {
					return identity
				}
				return l.enumNeverIdentity(initializer, seen)
			}
		}
	}
	return nil
}

// Only immutable aliases of actual members prove membership. Parameters, mutable slots,
// numeric literals, casts, arithmetic and flag-domain proofs do not prove a declared member.
func (l *lowering) enumMemberOrigin(node *ast.Node, identity *ast.Symbol, seen map[*ast.Node]bool) bool {
	node = ast.SkipParentheses(node)
	if member := l.enumMember(node); member != nil {
		return l.symbol(member.Parent.Name()) == identity
	}
	if !ast.IsIdentifier(node) || seen[node] {
		return false
	}
	seen[node] = true
	if symbol := l.symbol(node); symbol != nil && symbol.ValueDeclaration != nil {
		declaration := symbol.ValueDeclaration
		if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
			if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
				return l.enumMemberOrigin(initializer, identity, seen)
			}
		}
	}
	return false
}

// An ordinary IR helper preserves evaluation order and reads the operand exactly once. Its
// panic is shared by both backends; never itself does not need a new machine representation.
func (l *lowering) enumNeverCheck(node *ast.Node, value ir.Expression, identity *ast.Symbol) ir.Expression {
	of := ir.Number
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
		if representation, known := l.representation(contextual); known && !slotless(representation) {
			of = representation
		}
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindReturnStatement && l.function != nil {
		of = l.function.Returns
	}
	function := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "enum_never_value", Type: value.Type(), Function: function})
	read := ir.Read{Local: local, Of: value.Type()}
	var printed ir.Expression = ir.NumberToString{Value: read}
	if value.Type().IsMaybe() {
		printed = ir.MaybeToString{Value: read}
	} else if value.Type() == ir.Union {
		printed = ir.UnionToString{Value: read}
	}
	message := ir.Concat{Parts: []ir.Expression{
		ir.StringConstant{Index: l.constant("unreachable value ")},
		printed,
		ir.StringConstant{Index: l.constant(" for numeric enum " + identity.Name)},
	}}
	l.result.Functions = append(l.result.Functions, ir.Function{
		Name: "enum_never", Parameters: []int{local}, Returns: of,
		Body: []ir.Statement{ir.Panic{Message: message}},
	})
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: of}
}
