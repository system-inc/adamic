package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Restricted port of V8 13.6.233 Object::OrdinaryToPrimitive in
// src/objects/objects.cc. A proven member signature supplies IsCallable and
// its result representation. Hidden members and Symbol.toPrimitive are NotYet.
// Only statically callable own conversion members are admitted. If an object result needs
// a second member the type doesn't expose, refuse instead of guessing what an erased view hid.
func (l *lowering) objectStringByHint(node *ast.Node, value ir.Expression, hint string) (ir.Expression, error) {
	typeOf := l.checker.GetTypeAtLocation(node)
	function, reads := l.stringHelper("ordinary_"+hint, []ir.Expression{value})
	body := []ir.Statement{}
	names := []string{"toString", "valueOf"}
	if hint == "default" {
		names = []string{"valueOf", "toString"}
	}
	for _, name := range names {
		member := l.checker.GetTypeOfPropertyOfType(typeOf, name)
		if member == nil || l.inheritedLibrarySymbol(l.checker.GetPropertyOfType(typeOf, name)) {
			if name == "valueOf" && l.exactObject(node, 0) {
				continue
			}
			// Only a literal's complete shape proves the default Object.prototype.toString.
			if name == "toString" && l.exactObject(node, 0) {
				body = append(body, ir.Return{Value: ir.StringConstant{Index: l.constant("[object Object]")}})
				l.result.Functions[function].Body = body
				return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
			}
			return nil, l.notYet(node, "String ToPrimitive needs a conversion member hidden by the object view")
		}
		signatures := l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)
		if len(signatures) == 0 && member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			continue
		}
		if len(signatures) != 1 || len(signatures[0].Parameters()) != 0 || l.includesUndefined(member) {
			return nil, l.notYet(node, "String ToPrimitive with an optional, noncallable or parameterized "+name)
		}
		returned := l.checker.GetReturnTypeOfSignature(signatures[0])
		of, known := l.representation(returned)
		never := returned.Flags()&checker.TypeFlagsNever != 0
		if never {
			of, known = ir.Type(0), true
		}
		if !known {
			return nil, l.notYet(node, "String ToPrimitive with an unrepresented result")
		}
		call := ir.CallClosure{Closure: ir.Property{Object: reads[0], Name: name, Of: ir.Closure, Method: true}, Returns: of}
		var text ir.Expression
		switch of {
		case ir.String:
			text = call
			if l.includesUndefined(returned) || l.includesNull(returned) {
				missing := "undefined"
				if l.includesNull(returned) {
					missing = "null"
				}
				text = ir.Coalesce{Value: call, Fallback: ir.StringConstant{Index: l.constant(missing)}, Of: ir.String}
			}
		case ir.Number:
			text = ir.NumberToString{Value: call}
		case ir.Boolean:
			text = ir.BooleanToString{Value: call}
		case ir.MaybeNumber, ir.MaybeBoolean:
			text = ir.MaybeToString{Value: call}
		case ir.Union:
			if l.writable(returned) {
				text = ir.UnionToString{Value: call}
			}
			if text == nil && !l.includesNull(returned) && !l.includesUndefined(returned) {
				// Compiler boxes primitive-admitting {} results. Inspect that proven tag once;
				// an object result proceeds to valueOf rather than being printed prematurely.
				local := len(l.result.Locals)
				l.result.Locals = append(l.result.Locals, ir.Local{Name: "conversion_result", Type: ir.Union, Function: function})
				read := ir.Read{Local: local, Of: ir.Union}
				body = append(body, ir.Declare{Local: local, Value: call})
				kind := ir.TypeOf{Value: read}
				primitive := ir.Binary{Operator: ir.And,
					Left:  ir.Binary{Operator: ir.NotEqual, Left: kind, Right: ir.StringConstant{Index: l.constant("object")}},
					Right: ir.Binary{Operator: ir.NotEqual, Left: kind, Right: ir.StringConstant{Index: l.constant("function")}}}
				body = append(body, ir.If{Condition: primitive, Then: []ir.Statement{ir.Return{Value: ir.UnionToString{Value: read}}}})
				continue
			}
		}
		if text != nil {
			body = append(body, ir.Return{Value: text})
			l.result.Functions[function].Body = body
			return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
		}
		if of != ir.Object && of != ir.Array && of != ir.Map && of != ir.Closure && !never {
			return nil, l.notYet(node, "String ToPrimitive with a mixed primitive and object result")
		}
		body = append(body, ir.Evaluate{Value: call})
		if never {
			body = append(body, ir.Return{Value: ir.StringConstant{Index: l.constant("")}})
			l.result.Functions[function].Body = body
			return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
		}
	}
	body = append(body, ir.Throw{Value: ir.MakeError{Name: ir.StringConstant{Index: l.constant("TypeError")}, Message: ir.StringConstant{Index: l.constant("Cannot convert object to primitive value")}}})
	l.result.Functions[function].Body = body
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
}

func (l *lowering) stringObjectConversion(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	return l.objectStringByHint(node, value, "string")
}

// Both operands evaluate before either conversion, as AdditiveExpression does.
func (l *lowering) objectStringAddition(node *ast.Node, left, right ir.Expression) (ir.Expression, error) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, l.notYet(node, "object addition without proven operand types")
	}
	b := node.AsBinaryExpression()
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsStringLike == 0 || (left.Type() != ir.String && right.Type() != ir.String) {
		return nil, l.notYet(node, "object + without a proven string operand (numeric or mixed addition requires tagged primitive dispatch)")
	}
	function, reads := l.stringHelper("object_add", []ir.Expression{left, right})
	parts := make([]ir.Expression, 2)
	for i, operand := range []*ast.Node{b.Left, b.Right} {
		var err error
		if reads[i].Type() == ir.Object {
			parts[i], err = l.objectStringByHint(operand, reads[i], "default")
		} else {
			parts[i], err = l.stringConversionValue(operand, reads[i])
		}
		if err != nil {
			return nil, err
		}
	}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: ir.Concat{Parts: parts}}}
	return ir.Call{Function: function, Arguments: []ir.Expression{left, right}, Returns: ir.String}, nil
}
