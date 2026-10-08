package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

var dateUTCFields = map[string]int{"getUTCFullYear": 0, "getUTCMonth": 1, "getUTCDate": 2, "getUTCDay": 3, "getUTCHours": 4, "getUTCMinutes": 5, "getUTCSeconds": 6, "getUTCMilliseconds": 7}

// Date uses the synchronous borrowed-host IR so ownership and exception boundaries
// remain the same as filesystem Date observations. Only ISO validation can fail.
func (e *emitter) dateCall(call ir.NodeFSFile) string {
	args := make([]string, len(call.Arguments))
	for i, argument := range call.Arguments {
		args[i] = e.value(argument)
	}
	name := strings.TrimPrefix(call.Operation, "date_")
	code := ""
	switch name {
	case "new":
		code = "adamic_date_new(" + args[0] + ")"
	case "new_now":
		code = "adamic_date_new(adamic_date_now())"
	case "copy":
		code = "adamic_date_new(adamic_date_value(" + args[0] + "))"
	case "new_iso":
		code = "adamic_date_new(adamic_date_parse_iso(" + args[0] + "))"
	case "time":
		code = "adamic_date_value(" + args[0] + ")"
	case "now":
		code = "adamic_date_now()"
	case "now_function":
		code = "adamic_date_now_function()"
	case "parse":
		code = "adamic_date_parse_iso(" + args[0] + ")"
	case "UTC":
		numbers := "NULL"
		if len(args) > 0 {
			numbers = "(const double[]){" + strings.Join(args, ", ") + "}"
		}
		code = fmt.Sprintf("adamic_date_utc(%d, %s)", len(args), numbers)
	case "toISOString":
		code = "adamic_date_iso(" + args[0] + ")"
	case "toUTCString":
		code = "adamic_date_utc_string(" + args[0] + ")"
	case "prototype_own", "constructor_own":
		code = fmt.Sprintf("adamic_date_has_own(%s, %t)", args[0], name == "prototype_own")
	default:
		field, known := dateUTCFields[name]
		if !known {
			panic("native: unknown Date operation " + name)
		}
		code = fmt.Sprintf("adamic_date_get(%s, %d)", args[0], field)
	}
	if call.Of.IsReference() {
		return e.own(call.Of, code)
	}
	return e.snapshot(call.Of, code)
}
