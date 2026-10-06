package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

var dateGetters = map[string]int{"getFullYear": 0, "getMonth": 1, "getDate": 2, "getDay": 3, "getHours": 4, "getMinutes": 5, "getSeconds": 6, "getMilliseconds": 7, "getTimezoneOffset": 8, "getYear": 9}
var dateSetters = map[string]int{"setFullYear": 0, "setMonth": 1, "setDate": 2, "setHours": 4, "setMinutes": 5, "setSeconds": 6, "setMilliseconds": 7, "setYear": 9, "setTime": 10}

func (e *emitter) dateCall(call ir.DateCall) string {
	receiver := "NULL"
	if call.Receiver != nil {
		receiver = e.value(call.Receiver)
	}
	args := []string{}
	for _, a := range call.Arguments {
		args = append(args, e.value(a))
	}
	numbers := "NULL"
	if len(args) > 0 {
		numbers = "(const double[]){" + strings.Join(args, ", ") + "}"
	}
	code := ""
	switch call.Method {
	case "UTC":
		code = fmt.Sprintf("adamic_date_utc(%d, %s)", len(args), numbers)
	case "components":
		code = fmt.Sprintf("adamic_date_new(adamic_date_utc(%d, %s))", len(args), numbers)
	case "new":
		code = "adamic_date_new(" + args[0] + ")"
	case "newString":
		code = "adamic_date_new(adamic_date_parse_iso(" + args[0] + "))"
	case "copy":
		code = "adamic_date_new(adamic_date_value(" + receiver + "))"
	case "parse":
		code = "adamic_date_parse_iso(" + args[0] + ")"
	case "getTime", "valueOf":
		code = "adamic_date_value(" + receiver + ")"
	case "toString", "toUTCString", "toDateString", "toTimeString":
		style := map[string]int{"toString": 0, "toUTCString": 1, "toDateString": 2, "toTimeString": 3}[call.Method]
		code = fmt.Sprintf("adamic_date_format(%s, %d)", receiver, style)
	case "toJSON":
		code = "adamic_date_json(" + receiver + ")"
	case "nullableTypeOf":
		code = "(" + receiver + " == NULL ? &adamic_typeof_object : &adamic_typeof_string)"
	case "nullableNumber":
		code = "(" + receiver + " == NULL ? 0.0 : adamic_number_from_string(" + receiver + "))"
	case "toISOString":
		code = "adamic_date_iso(" + receiver + ")"
	default:
		name := strings.Replace(call.Method, "UTC", "", 1)
		if field, ok := dateGetters[name]; ok {
			code = fmt.Sprintf("adamic_date_get(%s, %d)", receiver, field)
		} else if field, ok := dateSetters[name]; ok {
			code = fmt.Sprintf("adamic_date_set(%s, %d, %d, %s)", receiver, field, len(args), numbers)
		} else {
			panic("native: unknown Date method")
		}
	}
	if call.Returns.IsReference() {
		return e.own(call.Returns, code)
	}
	return e.snapshot(call.Returns, code)
}
