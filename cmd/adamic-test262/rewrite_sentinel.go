package main

import (
	"strconv"
	"strings"
)

// SameValue is strict equality when one operand is null or the global undefined:
// neither sentinel has NaN or signed-zero cases. The intrinsic Date.prototype
// is an object, so SameValue likewise compares it solely by identity. Keep both operands and the
// message in the original evaluation order, and keep the original failure text.
func rewriteSentinelAssertion(source string, index int, from string, undefinedSafe, dateSafe bool) (string, int, bool) {
	start := index + len(from)
	tokens := tokenize(source[start:])
	if len(tokens) == 0 || tokens[0].text != "(" {
		return "", index, false
	}
	depth := 1
	argumentStart := tokens[0].end
	var arguments []string
	end := 0
	for _, token := range tokens[1:] {
		switch token.text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
			if depth == 0 {
				arguments = append(arguments, strings.TrimSpace(source[start+argumentStart:start+token.start]))
				end = start + token.end
			}
		case ",":
			if depth == 1 {
				arguments = append(arguments, strings.TrimSpace(source[start+argumentStart:start+token.start]))
				argumentStart = token.end
			}
		}
		if end != 0 {
			break
		}
	}
	if end == 0 || len(arguments) < 2 || len(arguments) > 3 {
		return "", index, false
	}
	sentinel := func(text string) bool { return text == "null" || undefinedSafe && text == "undefined" }
	prototype := func(text string) bool {
		if !dateSafe {
			return false
		}
		tokens := tokenize(text)
		return len(tokens) == 3 && tokens[0].text == "Date" && tokens[1].text == "." && tokens[2].text == "prototype"
	}
	if !sentinel(arguments[0]) && !sentinel(arguments[1]) && !prototype(arguments[0]) && !prototype(arguments[1]) {
		return "", index, false
	}
	op := "==="
	failure := "Expected SameValue"
	if from == "assert.notSameValue" {
		op = "!=="
		failure = "Expected NotSameValue"
	}
	message := ""
	if len(arguments) == 3 {
		message = ", " + rewriteHarnessCallsWithSentinels(arguments[2], undefinedSafe, dateSafe)
	}
	expression := "(" + rewriteHarnessCallsWithSentinels(arguments[0], undefinedSafe, dateSafe) + ") " + op + " (" + rewriteHarnessCallsWithSentinels(arguments[1], undefinedSafe, dateSafe) + ")"
	return "assertIdentity(" + expression + ", " + strconv.Quote(failure) + message + ")", end, true
}

func globalUndefinedUnshadowed(source string) bool {
	return globalIntrinsicUnshadowed(source, "undefined")
}

func globalIntrinsicUnshadowed(source, name string) bool {
	parsed := parseBody(source)
	if parsed.failed || parsed.unsafeVars {
		return false
	}
	for _, statement := range parsed.varStmts {
		for _, declaration := range statement.decls {
			if declaration.name == name {
				return false
			}
		}
	}
	for _, reference := range parsed.refs {
		if reference.name == name && lookup(reference.scope, name) != nil {
			return false
		}
	}
	return true
}
