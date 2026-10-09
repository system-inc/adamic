// Package decorators reads TypeScript decorators and the declared/actual nullability that the
// decorator parity rules compare.
//
// Four rules in the base layer ask the same two questions in different vocabularies:
// `graphql-nullable-parity`, `orm-column-nullable-parity`, `serializable-nullable-parity` and
// `cohere-optional-parity` each read a boolean flag off a decorator call and compare it against
// whether a TypeScript type admits nothing. The flag's name differs per rule; the judgment does not.
//
// So the judgment lives here rather than four times over. The original TypeScript layer reached the
// same conclusion from the other direction: its own comment says these helpers "consolidate the
// three drifting unwrap implementations across base-lint", which is the drift already having
// happened once.
package decorators

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
)

// NullableTypeFlags is the mask that decides whether a type counts as nullable for parity.
//
// Wider than null and undefined, and the width is the interesting part. `any` and `unknown` are
// included because they ERASE the distinction rather than answering it: a decorator saying
// `nullable: false` beside a value typed `any` is a claim nothing checked, so the parity rules treat
// it as unmet rather than as satisfied. `void` is included for the same reason on the return side.
//
// Copied from the TypeScript layer's `nullableTypeFlags`, whose comment states exactly this
// reasoning. Widening or narrowing it changes all four parity rules at once, which is the argument
// for it living in one place.
const NullableTypeFlags = checker.TypeFlagsNull |
	checker.TypeFlagsUndefined |
	checker.TypeFlagsAny |
	checker.TypeFlagsUnknown |
	checker.TypeFlagsVoid

// IsNullableType answers whether a type admits nothing, recursing through unions.
//
// A union is nullable when ANY member is, which is the opposite of how the union arm reads in a
// rule about what a value may be: here the question is whether the value can be absent, and one
// absent-admitting member is enough.
func IsNullableType(subject *checker.Type) bool {
	if subject == nil {
		return false
	}
	if type_checking.IsTypeFlagSet(subject, NullableTypeFlags) {
		return true
	}
	if subject.IsUnion() {
		for _, member := range subject.AsUnionType().Types() {
			if IsNullableType(member) {
				return true
			}
		}
	}
	return false
}

// CallName reads the identifier a decorator calls, or "" when the decorator is not a plain call.
//
// `@Foo()` answers "Foo". `@Foo` without parentheses, `@ns.Foo()`, and anything else answer "",
// because every rule using this keys on a bare identifier and a qualified name is a different
// symbol that happens to end in the same word.
func CallName(decorator *ast.Node) string {
	if decorator == nil || decorator.Kind != ast.KindDecorator {
		return ""
	}
	expression := decorator.AsDecorator().Expression
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return ""
	}
	callee := expression.AsCallExpression().Expression
	if callee == nil || !ast.IsIdentifier(callee) {
		return ""
	}
	return callee.Text()
}

// BooleanOption reads a boolean literal from an object argument of a decorator call.
//
// Three outcomes, and the difference between the second and third is a real decision rather than a
// nicety:
//
//	found=true,  value=v   the key is present and its value is a boolean literal
//	found=true,  value is meaningless with skip=true
//	                       the key is present and its value is NOT a boolean literal
//	found=false            the key is absent
//
// The middle case is what makes `nullable: 'items'` skippable: a list-nullability spelling is not a
// claim this rule can check, so the caller declines rather than reading it as false. The TypeScript
// original expresses the same three outcomes as `boolean | null` with absent folded into `false`,
// which conflates "absent" with "explicitly false". That conflation is deliberate there and
// reproduced by the callers here rather than by this function, so a caller that wants to tell the
// two apart can.
func BooleanOption(decorator *ast.Node, keyName string) (value bool, skip bool, found bool) {
	if decorator == nil || decorator.Kind != ast.KindDecorator {
		return false, false, false
	}
	expression := decorator.AsDecorator().Expression
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return false, false, false
	}
	call := expression.AsCallExpression()
	if call.Arguments == nil {
		return false, false, false
	}

	for _, argument := range call.Arguments.Nodes {
		if argument == nil || argument.Kind != ast.KindObjectLiteralExpression {
			continue
		}
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			if property == nil || property.Kind != ast.KindPropertyAssignment {
				continue
			}
			assignment := property.AsPropertyAssignment()
			name := assignment.Name()
			if name == nil || !ast.IsIdentifier(name) || name.Text() != keyName {
				continue
			}
			switch assignment.Initializer.Kind {
			case ast.KindTrueKeyword:
				return true, false, true
			case ast.KindFalseKeyword:
				return false, false, true
			}
			// Present but not a boolean literal, which the callers read as "not my question".
			return false, true, true
		}
	}
	return false, false, false
}

// HasDecoratorInSet answers whether a node carries any decorator whose call name is in the set.
//
// The node is anything the parser gives modifiers to: a method, a property, a parameter, a class.
// A node with no modifiers answers false rather than erroring, because "no decorators" and "no
// matching decorator" are the same answer to this question.
func HasDecoratorInSet(node *ast.Node, names map[string]struct{}) bool {
	for _, decorator := range Of(node) {
		if _, matched := names[CallName(decorator)]; matched {
			return true
		}
	}
	return false
}

// Of returns the decorators on a node, in source order.
//
// typescript-go keeps decorators in the MODIFIER list rather than in a field of their own, which is
// the shape difference that matters when porting from an estree rule: there is no `.decorators`
// to read, and a modifier list holds `public`, `readonly` and `async` alongside them.
func Of(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	modifiers := node.Modifiers()
	if modifiers == nil {
		return nil
	}

	var found []*ast.Node
	for _, modifier := range modifiers.Nodes {
		if modifier != nil && modifier.Kind == ast.KindDecorator {
			found = append(found, modifier)
		}
	}
	return found
}
