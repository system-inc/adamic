package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// Searches read only as far as their first match. Copies would lose both lazy
// failures and the physical source storage used by its other aliases.
func (e *emitter) emitViewArraySearch(search ir.ArraySearch) string {
	array := e.own(ir.Array, fmt.Sprintf("adamic_retain(%s)", e.value(search.Array)))
	value := e.value(search.Value)
	if search.Element.IsReference() {
		value = e.own(search.Element, fmt.Sprintf("adamic_retain(%s)", value))
	} else {
		value = e.snapshot(search.Element, value)
	}
	from := "0.0"
	if search.From != nil {
		from = e.snapshot(ir.Number, e.value(search.From))
	}
	count, index := e.temporary(), e.temporary()
	e.line("double %s = (double)%s->length;", count, array)
	start := e.temporary()
	if search.Last && search.From == nil {
		e.line("double %s = %s - 1;", start, count)
	} else {
		e.line("double %s = isnan(%s) ? 0 : trunc(%s);", start, from, from)
		e.line("if (%s < 0) %s += %s;", start, start, count)
	}
	if search.Last {
		e.line("if (%s >= %s) %s = %s - 1;", start, count, start, count)
	} else {
		e.line("if (%s < 0) %s = 0;", start, start)
	}
	result := e.snapshot(ir.Number, "-1.0")
	condition, update := index+" < "+count, index+"++"
	if search.Last {
		condition, update = index+" >= 0", index+"--"
	}
	e.line("for (double %s = %s; %s; %s) {", index, start, condition, update)
	e.indent++
	if !search.Includes {
		e.line("if (adamic_array_holes_at(%s, %s) == NULL) continue;", array, index)
	}
	slot := e.emitViewArrayRead(search.ViewRead.Index(), array, index)
	equal, missing := "false", "false"
	switch search.Element {
	case ir.Number:
		equal = slot + "->number == " + value
		if search.Includes {
			equal += " || (isnan(" + slot + "->number) && isnan(" + value + "))"
		}
	case ir.Boolean:
		equal = slot + "->boolean == " + value
	case ir.MaybeNumber:
		present := "adamic_maybe_number_unpack(" + slot + "->number)"
		equal = "(" + value + ".present && " + present + ".number == " + value + ".number)"
		if search.Includes {
			equal += " || (" + value + ".present && isnan(" + present + ".number) && isnan(" + value + ".number))"
		}
		missing = "!" + value + ".present"
	case ir.String:
		equal = "(" + value + " != NULL && adamic_string_equal(" + slot + "->reference, " + value + "))"
		missing = value + " == NULL"
	default:
		equal = slot + "->reference == " + value
		missing = value + " == NULL"
	}
	e.line("if (%s == NULL ? %s : (%s)) {", slot, missing, equal)
	e.line("    %s = %s;", result, index)
	e.line("    break;")
	e.line("}")
	e.indent--
	e.line("}")
	if search.Includes {
		return e.snapshot(ir.Boolean, fmt.Sprintf("%s != -1", result))
	}
	return result
}
