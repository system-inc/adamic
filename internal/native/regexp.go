package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) regexCall(call ir.RegExpCall) string {
	if call.Replacement != nil {
		return nativeEmitter.regexReplacement(call)
	}
	if call.Method == "iteratorDone" {
		object := nativeEmitter.value(call.Value)
		value := nativeEmitter.snapshot(ir.MaybeBoolean, "adamic_regex_done("+object+")")
		if call.Returns.IsMaybe() {
			return value
		}
		nativeEmitter.line("if (!%s.present) {", value)
		nativeEmitter.indent++
		message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
		nativeEmitter.line("adamic_panic(%s, %d);", cString(message), len(message))
		nativeEmitter.indent--
		nativeEmitter.line("}")
		return nativeEmitter.snapshot(ir.Boolean, value+".boolean")
	}
	arguments := []string{nativeEmitter.value(call.Value)}
	for _, argument := range call.Arguments {
		arguments = append(arguments, nativeEmitter.value(argument))
	}
	method := map[string]string{
		"test":       "test",
		"exec":       "exec",
		"match":      "match",
		"matchAll":   "match_all",
		"next":       "next",
		"replace":    "replace",
		"replaceAll": "replace",
		"split":      "split",
		"search":     "search",
	}[call.Method]
	if method == "replace" {
		arguments = append(arguments, fmt.Sprint(call.Method == "replaceAll"))
	}
	invocation := "adamic_regex_" + method + "(" + strings.Join(arguments, ", ") + ")"
	if nativeEmitter.program.GlobalFlagMayThrow(call) {
		// A RegExp without the g flag leaves Node's TypeError pending, tested once the result, an
		// empty string or no iterator, is owned (ir.Program's GlobalFlagMayThrow).
		result := nativeEmitter.own(call.Returns, invocation)
		nativeEmitter.checkThrown()
		return result
	}
	if call.Returns.IsReference() {
		return nativeEmitter.own(call.Returns, invocation)
	}
	return nativeEmitter.snapshot(call.Returns, invocation)
}
func (nativeEmitter *emitter) regexProperty(property ir.RegExpProperty) string {
	array := nativeEmitter.value(property.Array)
	if property.Of == ir.MaybeNumber {
		missing := array + "->properties == NULL"
		if property.Optional {
			missing = array + " == NULL || " + missing
		}
		value := fmt.Sprintf("adamic_regex_property(%s, %s).number", array, cString(property.Name))
		return nativeEmitter.snapshot(
			ir.MaybeNumber,
			fmt.Sprintf("(%s ? %s : %s)", missing, zero(ir.MaybeNumber), maybe(ir.MaybeNumber, value)),
		)
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
		return nativeEmitter.own(property.Of, "adamic_retain("+value+")")
	}
	return nativeEmitter.snapshot(property.Type(), value)
}

// Capture-group descriptors come from the regexp serializer rather than shapeOf.
// Include their real names in counted startup, even if no match ever allocates one.
func (nativeEmitter *emitter) regexpShapeNames() []string {
	var names []string
	if nativeEmitter.program == nil {
		return names
	}
	for index := range nativeEmitter.program.Regexps {
		tokens, cTokensError := cTokens(nativeEmitter.regexDeclarations(index))
		if cTokensError != nil {
			panic("native: malformed generated regexp declarations: " + cTokensError.Error())
		}
		for index := 0; index+4 < len(tokens); index++ {
			if tokens[index].text == "static" &&
				tokens[index+1].text == "const" &&
				tokens[index+2].text == "adamic_shape" &&
				tokens[index+4].text == "=" {
				names = append(names, tokens[index+3].text)
			}
		}
	}
	return names
}

// regexName gives compiled bytecode a content identity independent of the literal table.
func (nativeEmitter *emitter) regexName(index int) string {
	program := nativeEmitter.program.Regexps[index]
	return stableName("adamic_regex", "bytecode", fmt.Sprintf("%q/%q", program.Pattern, program.Flags))
}

// regexDeclarations renames identifier tokens, preserving literal and comment bytes.
func (nativeEmitter *emitter) regexDeclarations(index int) string {
	source := nativeEmitter.program.Regexps[index].Declarations
	old, name := fmt.Sprintf("adamic_regex_%d", index), nativeEmitter.regexName(index)
	var out strings.Builder
	for index := 0; index < len(source); {
		start := index
		if source[index] == '"' || source[index] == '\'' {
			quote := source[index]
			index++
			for index < len(source) {
				if source[index] == '\\' {
					index += 2
					continue
				}
				if source[index] == quote {
					index++
					break
				}
				index++
			}
		} else if source[index] == '/' && index+1 < len(source) && source[index+1] == '/' {
			for index < len(source) && source[index] != '\n' {
				index++
			}
		} else if source[index] == '/' && index+1 < len(source) && source[index+1] == '*' {
			end := strings.Index(source[index+2:], "*/")
			if end < 0 {
				panic("native: unterminated regexp declaration comment")
			}
			index += end + 4
		} else if identifierCharacter(source[index]) {
			for index < len(source) && identifierCharacter(source[index]) {
				index++
			}
			token := source[start:index]
			if token == old || strings.HasPrefix(token, old+"_") {
				out.WriteString(name + strings.TrimPrefix(token, old))
				continue
			}
		} else {
			index++
		}
		out.WriteString(source[start:index])
	}
	return out.String()
}
func identifierCharacter(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9' ||
		character == '_'
}
