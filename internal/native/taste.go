package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// toBoolean snapshots before testing: neither a call nor a lazy C read is evaluated twice.
func (e *emitter) toBoolean(of ir.Type, value string) string {
	value = e.snapshot(of, value)
	switch of {
	case ir.Boolean:
		return value
	case ir.Number:
		return fmt.Sprintf("(%s != 0.0 && !isnan(%s))", value, value)
	case ir.MaybeNumber:
		return fmt.Sprintf("(%s.present && %s.number != 0.0 && !isnan(%s.number))", value, value, value)
	case ir.MaybeBoolean:
		return fmt.Sprintf("(%s.present && %s.boolean)", value, value)
	case ir.String:
		return fmt.Sprintf("(%s != NULL && %s->length != 0)", value, value)
	case ir.Union:
		result := e.temporary()
		e.line("bool %s = %s != NULL && %s != &adamic_null;", result, value, value)
		e.line("if (%s) {", result)
		e.indent++
		e.line("switch (%s->kind) {", value)
		e.line("case adamic_kind_number: %s = ((adamic_number_box *)%s)->number != 0.0 && !isnan(((adamic_number_box *)%s)->number); break;", result, value, value)
		e.line("case adamic_kind_boolean: %s = ((adamic_boolean_box *)%s)->boolean; break;", result, value)
		e.line("case adamic_kind_string: %s = ((adamic_string *)%s)->length != 0; break;", result, value)
		e.line("default: break;")
		e.line("}")
		e.indent--
		e.line("}")
		return result
	default:
		return fmt.Sprintf("(%s != NULL)", value)
	}
}

func (e *emitter) logicalValue(expression ir.Logical) string {
	left := e.snapshot(expression.Left.Type(), e.value(expression.Left))
	condition := e.toBoolean(expression.Left.Type(), left)
	if !expression.KeepTruthy {
		condition = "!(" + condition + ")"
	}
	result := e.temporary()
	e.line("%s %s;", cType(expression.Of), result)
	e.line("if (%s) {", unwrap(condition))
	e.indent++
	from := expression.Left.Type()
	// The || branch has proved a Maybe present. && keeps its absent value too.
	if from.IsMaybe() && expression.KeepTruthy {
		left, from = left+"."+member(from.Present()), from.Present()
	}
	value, fresh := "", false
	if expression.AbsentString {
		value = zero(ir.String)
	} else {
		value, fresh = logicalConverted(from, expression.Of, left)
	}
	if expression.Of.IsReference() && !fresh {
		value = retained(value)
	}
	e.line("%s = %s;", result, value)
	e.indent--
	text, right, owned := e.aside(expression.Right)
	e.line("} else {")
	e.out.WriteString(text)
	e.indent++
	if expression.Of.IsReference() {
		right = retained(right)
	}
	e.line("%s = %s;", result, right)
	for index := len(owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", owned[index])
	}
	e.indent--
	e.line("}")
	if expression.Of.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}

func logicalConverted(from, to ir.Type, value string) (string, bool) {
	// && keeps a falsy left value. A reference other than a string is falsy only when it's missing,
	// and a Maybe result says the checker typed that missing value as undefined, so it's the absent
	// pair (item && item.ready, item an object or undefined).
	if from.IsReference() && from != ir.String && from != ir.Union && to.IsMaybe() {
		return zero(to), false
	}
	if from == ir.Union && to != ir.Union {
		switch to {
		case ir.Number:
			return fmt.Sprintf("((adamic_number_box *)%s)->number", value), false
		case ir.Boolean:
			return fmt.Sprintf("((adamic_boolean_box *)%s)->boolean", value), false
		default:
			return fmt.Sprintf("((%s)%s)", cType(to), value), false
		}
	}
	return converted(from, to, value)
}

func (e *emitter) effects(expression ir.Effects) string {
	savedOwned, savedAt := e.owned, e.at
	e.outerOwned = append(e.outerOwned, savedOwned)
	e.owned, e.at = nil, nil
	for _, statement := range expression.Body {
		e.statement(statement)
	}
	e.end()
	e.owned, e.at = savedOwned, savedAt
	e.outerOwned = e.outerOwned[:len(e.outerOwned)-1]
	return e.value(expression.Result)
}
