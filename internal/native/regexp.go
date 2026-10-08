package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"math"
	"strings"
)

func (e *emitter) regexCall(call ir.RegExpCall) string {
	if call.Method == "iteratorDone" {
		object := e.value(call.Value)
		value := e.snapshot(ir.MaybeBoolean, "adamic_regex_done("+object+")")
		if call.Returns.IsMaybe() {
			return value
		}
		e.line("if (!%s.present) {", value)
		e.indent++
		message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
		e.line("adamic_panic(%s, %d);", cString(message), len(message))
		e.indent--
		e.line("}")
		return e.snapshot(ir.Boolean, value+".boolean")
	}
	arguments := []string{e.value(call.Value)}
	split := call.Method == "split" || call.Method == "symbol:split"
	defaultLimit := "false"
	for index, argument := range call.Arguments {
		if split && index == len(call.Arguments)-1 {
			if _, undefined := argument.(ir.Undefined); undefined {
				arguments = append(arguments, "4294967295.0")
				defaultLimit = "true"
				continue
			}
			if value, constant := argument.(ir.NumberConstant); constant && value.Value >= -2147483648 && value.Value <= 2147483647 && value.Value == math.Trunc(value.Value) {
				defaultLimit = "true"
			}
			if argument.Type() == ir.MaybeNumber {
				value := e.value(argument)
				arguments = append(arguments, "("+value+".present ? "+value+".number : 4294967295.0)")
				defaultLimit = "!" + value + ".present"
				continue
			}
		}
		arguments = append(arguments, e.value(argument))
	}
	if split {
		arguments = append(arguments, defaultLimit)
	}
	name := call.Method
	symbol := strings.HasPrefix(name, "symbol:")
	if symbol {
		name = strings.TrimPrefix(name, "symbol:")
		arguments[0], arguments[1] = arguments[1], arguments[0]
	}
	method := map[string]string{"test": "test", "exec": "exec", "toString": "to_string", "match": "match", "matchAll": "match_all", "next": "next", "replace": "replace", "replaceAll": "replace", "replaceCallback": "replace_callback", "replaceAllCallback": "replace_callback", "split": "split", "search": "search"}[name]
	if symbol && name == "matchAll" {
		method = "symbol_match_all"
	}
	if method == "replace" || method == "replace_callback" {
		arguments = append(arguments, fmt.Sprint(call.Method == "replaceAll" || call.Method == "replaceAllCallback"))
	}
	invocation := "adamic_regex_" + method + "(" + strings.Join(arguments, ", ") + ")"
	var result string
	if call.Returns.IsReference() {
		result = e.own(call.Returns, invocation)
	} else {
		result = e.snapshot(call.Returns, invocation)
	}
	if call.Method == "replaceAll" || call.Method == "matchAll" || strings.HasSuffix(call.Method, "Callback") {
		e.checkThrown()
	}
	return result
}
func (e *emitter) regexProperty(property ir.RegExpProperty) string {
	array := e.value(property.Array)
	if property.Of == ir.MaybeNumber {
		missing := array + "->properties == NULL"
		if property.Optional {
			missing = array + " == NULL || " + missing
		}
		value := fmt.Sprintf("adamic_regex_property(%s, %s).number", array, cString(property.Name))
		return e.snapshot(ir.MaybeNumber, fmt.Sprintf("(%s ? %s : %s)", missing, zero(ir.MaybeNumber), maybe(ir.MaybeNumber, value)))
	}
	value := fmt.Sprintf("adamic_regex_property(%s, %s).%s", array, cString(property.Name), member(property.Of))
	if property.Of.IsReference() {
		value = fmt.Sprintf("(%s)%s", cType(property.Of), value)
	}
	if property.Optional {
		if property.Of.IsReference() {
			value = fmt.Sprintf("(%s == NULL ? NULL : %s)", array, value)
		} else {
			value = fmt.Sprintf("(%s == NULL ? %s : %s)", array, zero(property.Type()), maybe(property.Type(), value))
		}
	}
	if property.Of.IsReference() {
		return e.own(property.Of, "adamic_retain("+value+")")
	}
	return e.snapshot(property.Type(), value)
}
