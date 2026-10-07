package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func regexReplacementKind(of ir.Type) int {
	switch of {
	case 0:
		return 0
	case ir.String:
		return 1
	case ir.Number:
		return 2
	case ir.Object:
		return 3
	case ir.Union:
		return 4
	case ir.Boolean:
		return 5
	}
	panic("unsupported regex replacement ABI")
}

func (e *emitter) regexReplacement(call ir.RegExpCall) string {
	input := e.value(call.Value)
	regex := e.value(call.Arguments[0])
	callback := e.value(call.Arguments[1])
	kinds := []string{}
	for _, kind := range call.Replacement.Kinds {
		kinds = append(kinds, fmt.Sprint(kind))
	}
	schema := "NULL"
	if len(kinds) > 0 {
		schema = "(const unsigned char[]){" + strings.Join(kinds, ",") + "}"
	}
	rest := 0
	if call.Replacement.Rest != 0 {
		rest = call.Replacement.RestKind
	}
	groupChecks := []string{}
	for _, group := range call.Replacement.Groups {
		groupChecks = append(groupChecks, fmt.Sprintf("{%d,%t,%s,%t}", group.Argument, group.Rest, cString(group.Name), group.Optional))
	}
	groups := "NULL"
	if len(groupChecks) > 0 {
		groups = "(const adamic_regex_replacement_group[]){" + strings.Join(groupChecks, ",") + "}"
	}
	result := e.temporary()
	e.line("adamic_string *%s = adamic_regex_replace_callback(%s, %s, %s, %t, %s, %d, %d, %d, %s, %d);", result, input, regex, callback, call.Method == "replaceAll", schema, len(kinds), rest, regexReplacementKind(call.Replacement.Returns), groups, len(groupChecks))
	e.closureThrown()
	return e.own(ir.String, result)
}
