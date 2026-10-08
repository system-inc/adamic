package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) regexReplacement(call ir.RegExpCall) string {
	values := "values"
	checks := []string{}
	kindCheck := func(value string, kind int) string {
		checks := []string{}
		if kind&64 != 0 {
			if kind&1 != 0 {
				checks = append(checks, value+" === undefined")
			}
			if kind&2 != 0 {
				checks = append(checks, "typeof "+value+" === 'string'")
			}
			if kind&4 != 0 {
				checks = append(checks, "typeof "+value+" === 'number'")
			}
			if kind&8 != 0 {
				checks = append(checks, "("+value+" !== null && typeof "+value+" === 'object')")
			}
		} else {
			if kind&16 != 0 {
				checks = append(checks, value+" === undefined")
			}
			switch kind & 15 {
			case 1:
				checks = append(checks, "typeof "+value+" === 'string'")
			case 2:
				checks = append(checks, "typeof "+value+" === 'number'")
			case 3:
				checks = append(checks, "("+value+" !== null && typeof "+value+" === 'object')")
			}
		}
		return "(" + strings.Join(checks, " || ") + ")"
	}
	for index, kind := range call.Replacement.Kinds {
		checks = append(checks, kindCheck(fmt.Sprintf("values[%d]", index), kind))
	}
	if call.Replacement.Rest != 0 {
		checks = append(checks, fmt.Sprintf("values.slice(%d).every(value => %s)", len(call.Replacement.Parameters), kindCheck("value", call.Replacement.RestKind)))
	}
	for _, group := range call.Replacement.Groups {
		if group.Optional {
			continue
		}
		value := fmt.Sprintf("values[%d]", group.Argument)
		condition := fmt.Sprintf("(%s === undefined || typeof %s !== 'object' || typeof %s[%s] === 'string')", value, value, value, quote(group.Name))
		if group.Rest {
			condition = fmt.Sprintf("values.slice(%d).every(value => value === undefined || typeof value !== 'object' || typeof value[%s] === 'string')", group.Argument, quote(group.Name))
		}
		checks = append(checks, condition)
	}
	guard := ""
	if len(checks) > 0 {
		guard = "if (!(" + strings.Join(checks, " && ") + ")) panic('RegExp replacement callback argument does not fit its declared type'); "
	}
	if call.Replacement.Rest != 0 {
		count := len(call.Replacement.Parameters)
		values = fmt.Sprintf("[...values.slice(0, %d), values.slice(%d)]", count, count)
	}
	return "((input, regex, callback) => input." + call.Method + "(regex, (...values) => { " + guard + "return adamicCall(callback, " + values + "); }))(" + e.value(call.Value) + ", " + e.value(call.Arguments[0]) + ", " + e.value(call.Arguments[1]) + ")"
}
