// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strings"
)

// aside emits work into a buffer of its own, one level in, with temporaries of its own, and returns
// the text, the value, and what it owns, so the caller can decide where the work belongs.
func (nativeEmitter *emitter) aside(expression ir.Expression) (string, string, []string) {
	return nativeEmitter.asideWith(func() string { return nativeEmitter.value(expression) })
}

// asideWith is aside for whatever emit emits.
func (nativeEmitter *emitter) asideWith(emit func() string) (string, string, []string) {
	savedOut, savedOwned := nativeEmitter.out, nativeEmitter.owned
	nativeEmitter.out, nativeEmitter.owned = strings.Builder{}, nil
	// A throw from inside lets go of what the statement owned before the aside too.
	nativeEmitter.outerOwned = append(nativeEmitter.outerOwned, savedOwned)
	nativeEmitter.indent++
	value := emit()
	nativeEmitter.indent--
	nativeEmitter.outerOwned = nativeEmitter.outerOwned[:len(nativeEmitter.outerOwned)-1]
	text, owned := nativeEmitter.out.String(), nativeEmitter.owned
	nativeEmitter.out, nativeEmitter.owned = savedOut, savedOwned
	return text, value, owned
}

// logical emits && and ||. When the right operand has nothing to run first, C's own operator
// short-circuits the same way; otherwise the right side runs only on the branch JavaScript takes.
func (nativeEmitter *emitter) logical(binary ir.Binary) string {
	left := nativeEmitter.value(binary.Left)
	text, right, owned := nativeEmitter.aside(binary.Right)
	operator := "&&"
	if binary.Operator == ir.Or {
		operator = "||"
	}
	if binary.Type() == ir.MaybeBoolean {
		result := nativeEmitter.temporary()
		nativeEmitter.line("%s %s = %s;", cType(ir.MaybeBoolean), result, left)
		test := result + ".present && " + result + ".boolean"
		if binary.Operator == ir.Or {
			test = "!(" + test + ")"
		}
		nativeEmitter.line("if (%s) {", test)
		nativeEmitter.out.WriteString(text)
		nativeEmitter.indent++
		nativeEmitter.line("%s = %s;", result, right)
		for index := len(owned) - 1; index >= 0; index-- {
			nativeEmitter.line("adamic_release(%s);", owned[index])
		}
		nativeEmitter.indent--
		nativeEmitter.line("}")
		return result
	}
	if text == "" && len(owned) == 0 {
		return fmt.Sprintf("(%s %s %s)", left, operator, right)
	}
	result := nativeEmitter.temporary()
	nativeEmitter.line("bool %s = %s;", result, left)
	if binary.Operator == ir.And {
		nativeEmitter.line("if (%s) {", result)
	} else {
		nativeEmitter.line("if (!%s) {", result)
	}
	nativeEmitter.out.WriteString(text)
	nativeEmitter.indent++
	nativeEmitter.line("%s = %s;", result, right)
	for index := len(owned) - 1; index >= 0; index-- {
		nativeEmitter.line("adamic_release(%s);", owned[index])
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
	return result
}

// conditional emits ?:, evaluating only the branch JavaScript takes.
func (nativeEmitter *emitter) conditional(conditional ir.Conditional) string {
	condition := nativeEmitter.value(conditional.Condition)
	trueText, whenTrue, trueOwned := nativeEmitter.aside(conditional.WhenTrue)
	notText, whenNot, notOwned := nativeEmitter.aside(conditional.WhenNot)
	if trueText == "" && notText == "" && len(trueOwned) == 0 && len(notOwned) == 0 {
		return fmt.Sprintf("(%s ? %s : %s)", condition, whenTrue, whenNot)
	}
	valueType := conditional.Type()
	result := nativeEmitter.temporary()
	nativeEmitter.line("%s %s;", cType(valueType), result)
	branch := func(text string, value string, owned []string) {
		nativeEmitter.out.WriteString(text)
		nativeEmitter.indent++
		if valueType.IsReference() {
			nativeEmitter.line("%s = %s;", result, retained(value))
		} else {
			nativeEmitter.line("%s = %s;", result, value)
		}
		for index := len(owned) - 1; index >= 0; index-- {
			nativeEmitter.line("adamic_release(%s);", owned[index])
		}
		nativeEmitter.indent--
	}
	nativeEmitter.line("if (%s) {", unwrap(condition))
	branch(trueText, whenTrue, trueOwned)
	nativeEmitter.line("} else {")
	branch(notText, whenNot, notOwned)
	nativeEmitter.line("}")
	if valueType.IsReference() {
		nativeEmitter.owned = append(nativeEmitter.owned, result)
	}
	return result
}

// unwrap drops one pair of parentheses around a whole expression, so a condition reads if (a == b)
// rather than if ((a == b)), which clang's -Wparentheses-equality rightly flags. It drops them only
// when the opening one closes at the very end, never from (a) == (b).
func unwrap(expression string) string {
	if !strings.HasPrefix(expression, "(") || !strings.HasSuffix(expression, ")") {
		return expression
	}
	depth := 0
	for index, character := range expression {
		switch character {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && index != len(expression)-1 {
				return expression
			}
		}
	}
	return expression[1 : len(expression)-1]
}

// coalesce emits value ?? fallback, the fallback evaluated only when the value is missing, and
// value ?? panic(message), which ends the program there.
func (nativeEmitter *emitter) coalesce(coalesce ir.Coalesce) string {
	value := nativeEmitter.value(coalesce.Value)
	present, unwrapped := value+" != NULL", value
	_, undefined := coalesce.Value.(ir.Undefined)
	_, null := coalesce.Value.(ir.Null)
	if undefined || null {
		// Literal absence is known without comparing sentinel addresses in C.
		present = "false"
	} else if coalesce.Value.Type() == ir.Union {
		present += " && " + value + " != &adamic_null"
	}
	if coalesce.Value.Type().IsMaybe() {
		present, unwrapped = value+".present", value+"."+member(coalesce.Value.Type().Present())
	}
	// What's present is made what ?? makes: text ?? count boxes a present number into the union.
	fresh := false
	if coalesce.Value.Type() == ir.Union && coalesce.Of != ir.Union {
		switch coalesce.Of {
		case ir.Number:
			unwrapped = fmt.Sprintf("((const adamic_number_box *)%s)->number", value)
		case ir.Boolean:
			unwrapped = fmt.Sprintf("((const adamic_boolean_box *)%s)->boolean", value)
		case ir.MaybeNumber:
			unwrapped = maybe(ir.MaybeNumber, fmt.Sprintf("((const adamic_number_box *)%s)->number", value))
		case ir.MaybeBoolean:
			unwrapped = maybe(ir.MaybeBoolean, fmt.Sprintf("((const adamic_boolean_box *)%s)->boolean", value))
		default:
			unwrapped = fmt.Sprintf("((%s)%s)", cType(coalesce.Of), value)
		}
	} else {
		unwrapped, fresh = converted(coalesce.Value.Type().Present(), coalesce.Of, unwrapped)
	}
	if coalesce.Panic != nil {
		text, message, _ := nativeEmitter.aside(coalesce.Panic)
		nativeEmitter.line("if (!(%s)) {", present)
		nativeEmitter.out.WriteString(text)
		nativeEmitter.line("\tadamic_panic((%s)->bytes, (%s)->length);", message, message)
		nativeEmitter.line("}")
		if fresh {
			return nativeEmitter.own(coalesce.Of, unwrapped)
		}
		return unwrapped
	}
	text, fallback, owned := nativeEmitter.aside(coalesce.Fallback)
	result := nativeEmitter.temporary()
	nativeEmitter.line("%s %s;", cType(coalesce.Of), result)
	nativeEmitter.line("if (%s) {", present)
	if coalesce.Of.IsReference() && !fresh && nativeEmitter.taken(unwrapped) {
		// The value the statement owns is passed on as what ?? makes; when it isn't there it's
		// undefined, and there was nothing to let go of.
		nativeEmitter.line("\t%s = %s;", result, unwrapped)
	} else if coalesce.Of.IsReference() && !fresh {
		nativeEmitter.line("\t%s = %s;", result, retained(unwrapped))
	} else {
		nativeEmitter.line("\t%s = %s;", result, unwrapped)
	}
	nativeEmitter.line("} else {")
	nativeEmitter.out.WriteString(text)
	nativeEmitter.indent++
	if index := slices.Index(owned, fallback); coalesce.Of.IsReference() && index >= 0 {
		// A fallback made here is passed on with its count.
		owned = slices.Delete(slices.Clone(owned), index, index+1)
		nativeEmitter.line("%s = %s;", result, fallback)
	} else if coalesce.Of.IsReference() {
		nativeEmitter.line("%s = %s;", result, retained(fallback))
	} else {
		nativeEmitter.line("%s = %s;", result, fallback)
	}
	for index := len(owned) - 1; index >= 0; index-- {
		nativeEmitter.line("adamic_release(%s);", owned[index])
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
	if coalesce.Of.IsReference() {
		nativeEmitter.owned = append(nativeEmitter.owned, result)
	}
	return result
}
