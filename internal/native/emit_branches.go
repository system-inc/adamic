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
func (e *emitter) aside(expression ir.Expression) (string, string, []string) {
	return e.asideWith(func() string { return e.value(expression) })
}

// asideWith is aside for whatever emit emits.
func (e *emitter) asideWith(emit func() string) (string, string, []string) {
	savedOut, savedOwned := e.out, e.owned
	e.out, e.owned = strings.Builder{}, nil
	// A throw from inside lets go of what the statement owned before the aside too.
	e.outerOwned = append(e.outerOwned, savedOwned)
	e.indent++
	value := emit()
	e.indent--
	e.outerOwned = e.outerOwned[:len(e.outerOwned)-1]
	text, owned := e.out.String(), e.owned
	e.out, e.owned = savedOut, savedOwned
	return text, value, owned
}

// logical emits && and ||. When the right operand has nothing to run first, C's own operator
// short-circuits the same way; otherwise the right side runs only on the branch JavaScript takes.
func (e *emitter) logical(binary ir.Binary) string {
	left := e.value(binary.Left)
	text, right, owned := e.aside(binary.Right)
	operator := "&&"
	if binary.Operator == ir.Or {
		operator = "||"
	}
	if binary.Type() == ir.MaybeBoolean {
		result := e.temporary()
		e.line("%s %s = %s;", cType(ir.MaybeBoolean), result, left)
		test := result + ".present && " + result + ".boolean"
		if binary.Operator == ir.Or {
			test = "!(" + test + ")"
		}
		e.line("if (%s) {", test)
		e.out.WriteString(text)
		e.indent++
		e.line("%s = %s;", result, right)
		for index := len(owned) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", owned[index])
		}
		e.indent--
		e.line("}")
		return result
	}
	if text == "" && len(owned) == 0 {
		return fmt.Sprintf("(%s %s %s)", left, operator, right)
	}
	result := e.temporary()
	e.line("bool %s = %s;", result, left)
	if binary.Operator == ir.And {
		e.line("if (%s) {", result)
	} else {
		e.line("if (!%s) {", result)
	}
	e.out.WriteString(text)
	e.indent++
	e.line("%s = %s;", result, right)
	for index := len(owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", owned[index])
	}
	e.indent--
	e.line("}")
	return result
}

// conditional emits ?:, evaluating only the branch JavaScript takes.
func (e *emitter) conditional(conditional ir.Conditional) string {
	condition := e.value(conditional.Condition)
	trueText, whenTrue, trueOwned := e.aside(conditional.WhenTrue)
	notText, whenNot, notOwned := e.aside(conditional.WhenNot)
	if trueText == "" && notText == "" && len(trueOwned) == 0 && len(notOwned) == 0 {
		return fmt.Sprintf("(%s ? %s : %s)", condition, whenTrue, whenNot)
	}
	valueType := conditional.Type()
	result := e.temporary()
	e.line("%s %s;", cType(valueType), result)
	branch := func(text string, value string, owned []string) {
		e.out.WriteString(text)
		e.indent++
		if valueType.IsReference() {
			e.line("%s = %s;", result, retained(value))
		} else {
			e.line("%s = %s;", result, value)
		}
		for index := len(owned) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", owned[index])
		}
		e.indent--
	}
	e.line("if (%s) {", unwrap(condition))
	branch(trueText, whenTrue, trueOwned)
	e.line("} else {")
	branch(notText, whenNot, notOwned)
	e.line("}")
	if valueType.IsReference() {
		e.owned = append(e.owned, result)
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
func (e *emitter) coalesce(coalesce ir.Coalesce) string {
	value := e.value(coalesce.Value)
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
	if coalesce.Value.Type() == ir.Union && coalesce.Of.IsReference() && coalesce.Of != ir.Union {
		// The present arm excludes both nullish tags; its proven reference is borrowed.
		unwrapped = fmt.Sprintf("((%s)%s)", cType(coalesce.Of), unwrapped)
	} else {
		unwrapped, fresh = converted(coalesce.Value.Type().Present(), coalesce.Of, unwrapped)
	}
	if coalesce.Panic != nil {
		text, message, _ := e.aside(coalesce.Panic)
		e.line("if (!(%s)) {", present)
		e.out.WriteString(text)
		e.line("\tadamic_panic((%s)->bytes, (%s)->length);", message, message)
		e.line("}")
		if fresh {
			return e.own(coalesce.Of, unwrapped)
		}
		return unwrapped
	}
	text, fallback, owned := e.aside(coalesce.Fallback)
	result := e.temporary()
	e.line("%s %s;", cType(coalesce.Of), result)
	e.line("if (%s) {", present)
	if coalesce.Of.IsReference() && !fresh && e.taken(unwrapped) {
		// The value the statement owns is passed on as what ?? makes; when it isn't there it's
		// undefined, and there was nothing to let go of.
		e.line("\t%s = %s;", result, unwrapped)
	} else if coalesce.Of.IsReference() && !fresh {
		e.line("\t%s = %s;", result, retained(unwrapped))
	} else {
		e.line("\t%s = %s;", result, unwrapped)
	}
	e.line("} else {")
	e.out.WriteString(text)
	e.indent++
	if index := slices.Index(owned, fallback); coalesce.Of.IsReference() && index >= 0 {
		// A fallback made here is passed on with its count.
		owned = slices.Delete(slices.Clone(owned), index, index+1)
		e.line("%s = %s;", result, fallback)
	} else if coalesce.Of.IsReference() {
		e.line("%s = %s;", result, retained(fallback))
	} else {
		e.line("%s = %s;", result, fallback)
	}
	for index := len(owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", owned[index])
	}
	e.indent--
	e.line("}")
	if coalesce.Of.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}
