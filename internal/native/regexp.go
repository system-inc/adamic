package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
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
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	method := map[string]string{"test": "test", "exec": "exec", "match": "match", "matchAll": "match_all", "next": "next", "replace": "replace", "replaceAll": "replace", "split": "split", "search": "search"}[call.Method]
	if method == "replace" {
		arguments = append(arguments, fmt.Sprint(call.Method == "replaceAll"))
	}
	invocation := "adamic_regex_" + method + "(" + strings.Join(arguments, ", ") + ")"
	if call.Returns.IsReference() {
		return e.own(call.Returns, invocation)
	}
	return e.snapshot(call.Returns, invocation)
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

// regexName gives compiled bytecode a content identity independent of the literal table.
func (e *emitter) regexName(index int) string {
	program := e.program.Regexps[index]
	return stableName("adamic_regex", "bytecode", fmt.Sprintf("%q/%q", program.Pattern, program.Flags))
}

// regexDeclarations renames identifier tokens, preserving literal and comment bytes.
func (e *emitter) regexDeclarations(index int) string {
	source := e.program.Regexps[index].Declarations
	old, name := fmt.Sprintf("adamic_regex_%d", index), e.regexName(index)
	var out strings.Builder
	for i := 0; i < len(source); {
		start := i
		if source[i] == '"' || source[i] == '\'' {
			quote := source[i]
			i++
			for i < len(source) {
				if source[i] == '\\' {
					i += 2
					continue
				}
				if source[i] == quote {
					i++
					break
				}
				i++
			}
		} else if source[i] == '/' && i+1 < len(source) && source[i+1] == '/' {
			for i < len(source) && source[i] != '\n' {
				i++
			}
		} else if source[i] == '/' && i+1 < len(source) && source[i+1] == '*' {
			end := strings.Index(source[i+2:], "*/")
			if end < 0 {
				panic("native: unterminated regexp declaration comment")
			}
			i += end + 4
		} else if identifierCharacter(source[i]) {
			for i < len(source) && identifierCharacter(source[i]) {
				i++
			}
			token := source[start:i]
			if token == old || strings.HasPrefix(token, old+"_") {
				out.WriteString(name + strings.TrimPrefix(token, old))
				continue
			}
		} else {
			i++
		}
		out.WriteString(source[start:i])
	}
	return out.String()
}
func identifierCharacter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}
