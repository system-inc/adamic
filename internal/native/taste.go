package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// toBoolean snapshots before testing: neither a call nor a lazy C read is evaluated twice.
func (nativeEmitter *emitter) toBoolean(of ir.Type, value string) string {
	value = nativeEmitter.snapshot(of, value)
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
		result := nativeEmitter.temporary()
		nativeEmitter.line("bool %s = %s != NULL && %s != &adamic_null;", result, value, value)
		nativeEmitter.line("if (%s) {", result)
		nativeEmitter.indent++
		nativeEmitter.line("switch (%s->kind) {", value)
		nativeEmitter.line(
			"case adamic_kind_number: %s = ((adamic_number_box *)%s)->number != 0.0 && !isnan(((adamic_number_box *)%s)->number); break;",
			result,
			value,
			value,
		)
		nativeEmitter.line("case adamic_kind_boolean: %s = ((adamic_boolean_box *)%s)->boolean; break;", result, value)
		nativeEmitter.line("case adamic_kind_string: %s = ((adamic_string *)%s)->length != 0; break;", result, value)
		nativeEmitter.line("default: break;")
		nativeEmitter.line("}")
		nativeEmitter.indent--
		nativeEmitter.line("}")
		return result
	default:
		return fmt.Sprintf("(%s != NULL)", value)
	}
}

func (nativeEmitter *emitter) logicalValue(expression ir.Logical) string {
	left := nativeEmitter.snapshot(expression.Left.Type(), nativeEmitter.value(expression.Left))
	condition := nativeEmitter.toBoolean(expression.Left.Type(), left)
	if !expression.KeepTruthy {
		condition = "!(" + condition + ")"
	}
	result := nativeEmitter.temporary()
	nativeEmitter.line("%s %s;", cType(expression.Of), result)
	nativeEmitter.line("if (%s) {", unwrap(condition))
	nativeEmitter.indent++
	from := expression.Left.Type()
	// The || branch has proved a Maybe present. && keeps its absent value too.
	if from.IsMaybe() && expression.KeepTruthy {
		left, from = left+"."+member(from.Present()), from.Present()
	}
	value, fresh := logicalConverted(from, expression.Of, left)
	if expression.Of.IsReference() && !fresh {
		value = retained(value)
	}
	nativeEmitter.line("%s = %s;", result, value)
	nativeEmitter.indent--
	text, right, owned := nativeEmitter.aside(expression.Right)
	nativeEmitter.line("} else {")
	nativeEmitter.out.WriteString(text)
	nativeEmitter.indent++
	if expression.Of.IsReference() {
		right = retained(right)
	}
	nativeEmitter.line("%s = %s;", result, right)
	for index := len(owned) - 1; index >= 0; index-- {
		nativeEmitter.line("adamic_release(%s);", owned[index])
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
	if expression.Of.IsReference() {
		nativeEmitter.owned = append(nativeEmitter.owned, result)
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
		case ir.MaybeNumber:
			// The checker typed the falsy union number | undefined: undefined is the missing pointer,
			// and anything else is a number's box (left is a snapshot, read once).
			return fmt.Sprintf(
				"(%s == NULL ? (adamic_maybe_number){false, 0.0} : (adamic_maybe_number){true, ((adamic_number_box *)%s)->number})",
				value,
				value,
			), false
		case ir.MaybeBoolean:
			return fmt.Sprintf(
				"(%s == NULL ? (adamic_maybe_boolean){false, false} : (adamic_maybe_boolean){true, ((adamic_boolean_box *)%s)->boolean})",
				value,
				value,
			), false
		default:
			return fmt.Sprintf("((%s)%s)", cType(to), value), false
		}
	}
	return converted(from, to, value)
}

func (nativeEmitter *emitter) effects(expression ir.Effects) string {
	savedOwned, savedAt := nativeEmitter.owned, nativeEmitter.at
	nativeEmitter.outerOwned = append(nativeEmitter.outerOwned, savedOwned)
	nativeEmitter.owned, nativeEmitter.at = nil, nil
	for _, statement := range expression.Body {
		nativeEmitter.statement(statement)
	}
	nativeEmitter.end()
	nativeEmitter.owned, nativeEmitter.at = savedOwned, savedAt
	nativeEmitter.outerOwned = nativeEmitter.outerOwned[:len(nativeEmitter.outerOwned)-1]
	return nativeEmitter.value(expression.Result)
}
