// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// appendsTo is the parts after the first of an assignment text = text + ..., to a string local only
// this function can see: not a global, which a call among the parts could write, not captured, which a
// closure among them could, and not borrowed. Those are the assignments appendTo writes, where text's
// own reference goes to adamic_string_append, which may write in place.
func (e *emitter) appendsTo(statement ir.Assign) ([]ir.Expression, bool) {
	if _, suspended := e.asyncSlots[statement.Local]; suspended {
		return nil, false
	}
	declared := e.program.Locals[statement.Local]
	if declared.Type != ir.String || declared.Global || declared.Captured || declared.Borrowed || statement.Checked {
		return nil, false
	}
	concat, isConcat := statement.Value.(ir.Concat)
	if !isConcat || len(concat.Parts) < 2 {
		return nil, false
	}
	if read, isRead := concat.Parts[0].(ir.Read); !isRead || read.Local != statement.Local || read.Checked {
		return nil, false
	}
	return concat.Parts[1:], true
}

// appendTo is text = text + parts: the parts first, as JavaScript reads them after text, which only
// this function writes, then the append, which takes text's reference and gives one back.
func (e *emitter) appendTo(local int, parts []ir.Expression) {
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, e.value(part))
	}
	name := e.localName(local)
	e.line("%s = adamic_string_append(%s, %d, (adamic_string *const[]){%s});", name, name, len(values), strings.Join(values, ", "))
}

func (e *emitter) stringCall(call ir.StringCall) string {
	value := e.value(call.Value)
	arguments := make([]string, 0, len(call.Arguments))
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	switch call.Method {
	case "slice":
		start, end, hasEnd := "0.0", "0.0", false
		if len(arguments) > 0 {
			start = arguments[0]
		}
		if len(arguments) > 1 {
			end, hasEnd = arguments[1], true
		}
		return e.own(ir.String, fmt.Sprintf("adamic_string_slice(%s, %s, %s, %t)", value, start, end, hasEnd))
	case "codePointAt":
		result := e.temporary()
		e.line("adamic_maybe_number %s = adamic_string_code_point_at(%s, %s);", result, value, arguments[0])
		return result
	case "padStart", "padEnd":
		return e.own(ir.String, fmt.Sprintf("adamic_string_pad(%s, %s, %s, %t)", value, arguments[0], arguments[1], call.Method == "padStart"))
	case "repeat":
		return e.own(ir.String, fmt.Sprintf("adamic_string_repeat(%s, %s)", value, arguments[0]))
	case "split":
		return e.own(ir.Array, fmt.Sprintf("adamic_string_split(%s, %s)", value, arguments[0]))
	case "indexOf":
		if len(arguments) == 2 {
			return e.snapshot(ir.Number, fmt.Sprintf("adamic_string_index_of_from(%s, %s, %s)", value, arguments[0], arguments[1]))
		}
		return fmt.Sprintf("adamic_string_index_of(%s, %s)", value, arguments[0])
	case "lastIndexOf":
		return e.snapshot(ir.Number, fmt.Sprintf("adamic_string_last_index_of(%s, %s)", value, arguments[0]))
	case "trimStart", "trimEnd":
		return e.own(ir.String, fmt.Sprintf("adamic_string_trim_sides(%s, %t, %t)", value, call.Method == "trimStart", call.Method == "trimEnd"))
	case "at":
		return e.own(ir.String, fmt.Sprintf("adamic_string_at_relative(%s, %s)", value, arguments[0]))
	case "toUpperCase":
		return e.own(ir.String, fmt.Sprintf("adamic_string_to_upper(%s)", value))
	case "toLowerCase":
		return e.own(ir.String, fmt.Sprintf("adamic_string_to_lower(%s)", value))
	case "normalize":
		return e.own(ir.String, fmt.Sprintf("adamic_string_normalize(%s, %s)", value, arguments[0]))
	case "replace", "replaceAll":
		return e.own(ir.String, fmt.Sprintf("adamic_string_replace(%s, %s, %s, %t)", value, arguments[0], arguments[1], call.Method == "replaceAll"))
	case "includes":
		if len(arguments) == 2 {
			return e.snapshot(ir.Boolean, fmt.Sprintf("(adamic_string_index_of_from(%s, %s, %s) != -1)", value, arguments[0], arguments[1]))
		}
		return fmt.Sprintf("(adamic_string_index_of(%s, %s) != -1)", value, arguments[0])
	case "startsWith":
		return fmt.Sprintf("adamic_string_starts_with(%s, %s)", value, arguments[0])
	}
	return fmt.Sprintf("adamic_string_ends_with(%s, %s)", value, arguments[0])
}
