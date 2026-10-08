package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only an immediate intrinsic .call supplies the lost receiver explicitly.
// An empty literal's methods are the same intrinsics, with no operand effects.
func (l *lowering) libraryArrayExplicitMethod(node *ast.Node) string {
	name := l.libraryArrayMethodName(node)
	if name == "" {
		method := ast.SkipParentheses(node)
		if method.Kind == ast.KindPropertyAccessExpression && method.AsPropertyAccessExpression().QuestionDotToken == nil {
			receiver := ast.SkipParentheses(method.AsPropertyAccessExpression().Expression)
			if receiver.Kind == ast.KindArrayLiteralExpression && len(receiver.AsArrayLiteralExpression().Elements.Nodes) == 0 {
				if _, known := libraryArrayLengths[method.Name().Text()]; known {
					name = method.Name().Text()
				}
			}
		}
	}
	if name == "" || name == "isArray" {
		return ""
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindPropertyAccessExpression && outer.Parent.Name().Text() == "call" && outer.Parent.AsPropertyAccessExpression().QuestionDotToken == nil && called(outer.Parent) {
		return name
	}
	return ""
}

// Port the RequireObjectCoercible/ToObject boundary from V8's Array builtins.
// Primitive number/boolean receivers have no length; these result-only cases
// never expose their temporary wrapper identity or require a stored boxed slot.
func (l *lowering) libraryArrayReceiverCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" || callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		return nil, false, nil
	}
	method := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	name := l.libraryArrayExplicitMethod(method)
	if name == "" {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) == 0 {
		return nil, true, l.notYet(node, name+".call without an explicit receiver")
	}
	proven := l.checker.GetTypeAtLocation(written[0])
	nullish := proven.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0
	of, known := l.representation(proven)
	primitive := known && (of == ir.Number || of == ir.Boolean) && !l.includesNull(proven) && !l.includesUndefined(proven)
	stringMutation := known && of == ir.String && !l.includesNull(proven) && !l.includesUndefined(proven) && (name == "pop" || name == "push" || name == "shift" || name == "unshift")
	if !nullish && !primitive && !stringMutation {
		if name == "indexOf" || name == "lastIndexOf" {
			return nil, false, nil
		}
		return nil, true, l.notYet(node, name+".call requiring an array-like or boxed receiver representation")
	}
	if primitive && (name == "indexOf" || name == "lastIndexOf") {
		return nil, false, nil // Existing generic search owns coercion and bounds.
	}
	arguments := []ir.Expression{}
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(node, name+".call with spread arguments")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, value)
	}
	b := l.libraryArrayBuilder(arguments)
	if stringMutation {
		return l.libraryArrayStringMutation(node, name, b, len(written)-1), true, nil
	}
	if nullish {
		message := "Cannot convert undefined or null to object"
		switch name {
		case "indexOf", "find", "findIndex", "findLast", "findLastIndex", "concat", "every", "filter", "map", "reduce", "reduceRight", "some":
			message = "Array.prototype." + name + " called on null or undefined"
		}
		b.body = append(b.body, b.typeError(node, message)...)
		// The fallback return is unreachable; ordinary IR carries cleanup and throws
		// through every caller rather than turning this library error into a panic.
		return b.finish("array_nullish_"+name, ir.Undefined{}), true, nil
	}
	var result ir.Expression
	switch name {
	case "pop", "shift":
		result = ir.Undefined{}
	case "push", "unshift":
		// The fresh primitive wrapper is extensible, starts at length zero, and
		// cannot escape. Every insertion succeeds; only its length is returned.
		result = ir.NumberConstant{Value: float64(len(written) - 1)}
	case "includes":
		result = ir.BooleanConstant{Value: false}
	case "reduce", "reduceRight":
		if len(arguments) < 2 || arguments[1].Type() != ir.Closure {
			return nil, true, l.notYet(node, name+".call without a proven callable reducer")
		}
		if len(arguments) < 3 {
			b.body = append(b.body, b.typeError(node, "Reduce of empty array with no initial value")...)
			result = ir.Undefined{}
		} else {
			result = b.read(b.parameters[2])
		}
	case "every", "some", "find", "findIndex", "findLast", "findLastIndex", "forEach":
		if len(arguments) < 2 || arguments[1].Type() != ir.Closure {
			return nil, true, l.notYet(node, name+".call without a proven callable predicate")
		}
		switch name {
		case "every":
			result = ir.BooleanConstant{Value: true}
		case "some":
			result = ir.BooleanConstant{Value: false}
		case "find", "findLast", "forEach":
			result = ir.Undefined{}
		default:
			result = ir.NumberConstant{Value: -1}
		}
	case "join":
		// Even on zero elements join converts a supplied separator. Primitive
		// conversions cannot invoke user code; object conversion remains refused.
		if len(arguments) > 1 {
			separator := arguments[1]
			if separator.Type() != ir.Number && separator.Type() != ir.Boolean && separator.Type() != ir.String {
				if _, absent := separator.(ir.Undefined); !absent {
					if _, null := separator.(ir.Null); !null {
						return nil, true, l.notYet(node, "join.call with an object separator requiring ToPrimitive")
					}
				}
			}
		}
		result = ir.StringConstant{Index: l.constant("")}
	case "toLocaleString":
		result = ir.StringConstant{Index: l.constant("")}
	case "toString":
		tag := "[object Number]"
		if of == ir.Boolean {
			tag = "[object Boolean]"
		}
		result = ir.StringConstant{Index: l.constant(tag)}
	default:
		return nil, true, l.notYet(node, name+".call requiring a boxed result or array-like representation")
	}
	return b.finish("array_primitive_"+name, result), true, nil
}
