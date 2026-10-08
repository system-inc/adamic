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

// Recover the full stored object union, including a never alias's initializer.
func (l *lowering) enumRemainderType(node *ast.Node, seen map[*ast.Node]bool) *checker.Type {
	node = ast.SkipParentheses(node)
	if seen[node] {
		return nil
	}
	seen[node] = true
	stored := l.enumStoredType(node)
	if stored != nil && stored.Flags()&checker.TypeFlagsNever == 0 {
		return stored
	}
	if symbol := l.flagValueSymbol(node); symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindVariableDeclaration {
		if initializer := symbol.ValueDeclaration.AsVariableDeclaration().Initializer; initializer != nil {
			return l.enumRemainderType(initializer, seen)
		}
	}
	return stored
}

func (l *lowering) enumObjectRemainder(node *ast.Node) (*checker.Type, string, *ast.Symbol) {
	return l.enumObjectRemainderSeen(node, map[*ast.Node]bool{})
}

func (l *lowering) enumObjectRemainderSeen(node *ast.Node, seen map[*ast.Node]bool) (*checker.Type, string, *ast.Symbol) {
	node = ast.SkipParentheses(node)
	if seen[node] {
		return nil, "", nil
	}
	seen[node] = true
	if symbol := l.flagValueSymbol(node); symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindVariableDeclaration {
		if initializer := symbol.ValueDeclaration.AsVariableDeclaration().Initializer; initializer != nil {
			if stored, field, identity := l.enumObjectRemainderSeen(initializer, seen); stored != nil {
				return stored, field, identity
			}
		}
	}
	stored := l.enumRemainderType(node, map[*ast.Node]bool{})
	if stored == nil || stored.Flags()&checker.TypeFlagsUnion == 0 {
		return nil, "", nil
	}
	for _, field := range l.checker.GetPropertiesOfType(stored) {
		var first *checker.Type
		unit, different, numeric := false, false, true
		for _, member := range stored.Types() {
			property := l.checker.GetPropertyOfType(member, field.Name)
			if member.Flags()&checker.TypeFlagsObject == 0 || property == nil {
				numeric = false
				break
			}
			tag := l.checker.GetTypeOfSymbol(property)
			unit = unit || tag.Flags()&checker.TypeFlagsUnit != 0
			numeric = numeric && l.numericEnum(l.enumIdentity(tag))
			if first == nil {
				first = tag
			} else {
				different = different || first != tag
			}
		}
		if unit && different && numeric && (l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever != 0 || l.enumTagValuesExcluded(node, l.flagValueSymbol(node), field.Name, stored)) {
			return stored, field.Name, l.enumIdentity(first)
		}
	}
	return nil, "", nil
}

// Follow the declared type, not the flow type that has already become never. A never alias
// keeps its initializer's enum identity, but never acquires a runtime member proof.
func (l *lowering) enumNeverIdentity(node *ast.Node, seen map[*ast.Node]bool) *ast.Symbol {
	node = ast.SkipParentheses(node)
	if seen[node] {
		return nil
	}
	seen[node] = true
	if _, _, identity := l.enumObjectRemainder(node); identity != nil {
		return identity
	}
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever == 0 {
		return nil
	}
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
	if value.Type() == ir.Object {
		_, field, _ := l.enumObjectRemainder(node)
		printed = ir.NumberToString{Value: ir.Property{Object: read, Name: field, Of: ir.Number}}
		of = ir.Object
	}
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
