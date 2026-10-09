package main

import (
	"encoding/json"
	"errors"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	regex "github.com/system-inc/adamic/internal/regexp"
)

func regexpStringToken(t token) (string, bool) {
	if t.kind != "str" {
		return "", false
	}
	s := scanner.NewScanner()
	s.SetText(t.text)
	if s.Scan() != ast.KindStringLiteral {
		return "", false
	}
	value := s.TokenValue()
	return value, s.Scan() == ast.KindEndOfFile && utf8.ValidString(value)
}

// Only literal arguments and fresh intrinsic RegExp constructions are pure.
// This proof is also used to remove diagnostic operations after a constructor
// whose exact Go reference result is an ECMAScript SyntaxError.
func regexpPureConstructor(t []token, start, end, depth int) (string, string, int, bool) {
	if depth > 16 {
		return "", "", 0, false
	}
	if start < end && t[start].text == "new" {
		start++
	}
	if start >= end || t[start].text != "RegExp" {
		return "", "", 0, false
	}
	args, close, ok := tokenArguments(t, start+1)
	if !ok || close >= end || len(args) > 2 {
		return "", "", 0, false
	}
	pattern, flags := "", ""
	if len(args) > 0 {
		a := args[0]
		if a[1]-a[0] == 1 && t[a[0]].text == "undefined" {
		} else if a[1]-a[0] == 1 {
			pattern, ok = regexpStringToken(t[a[0]])
			if !ok {
				return "", "", 0, false
			}
		} else {
			var inner int
			pattern, flags, inner, ok = regexpPureConstructor(t, a[0], a[1], depth+1)
			if !ok || inner != a[1]-1 {
				return "", "", 0, false
			}
			program, err := regex.Compile(pattern, flags)
			if err != nil {
				return "", "", 0, false
			}
			if _, err = program.NativeDeclarations("pure"); err != nil {
				return "", "", 0, false
			}
		}
	}
	if len(args) > 1 {
		a := args[1]
		if a[1]-a[0] != 1 {
			return "", "", 0, false
		}
		if t[a[0]].text != "undefined" {
			flags, ok = regexpStringToken(t[a[0]])
			if !ok {
				return "", "", 0, false
			}
		}
	}
	return pattern, flags, close, true
}

// Sputnik's legacy tests build a failure diagnostic after the invalid
// constructor. Once the intrinsic constructor is proven to throw, neither
// RegExp-to-string coercion nor exec in that diagnostic can execute. Normalize
// this exact complete test to the equivalent upstream assert.throws form;
// untouched Node still executes the original diagnostic and constructor check.
func adaptRegExpConstructorErrors(source string) adapted {
	counts := map[string]int{}
	unchanged := adapted{Source: source, Counts: counts}
	t := tokenize(source)
	if !regexpPristine(t) || len(t) < 22 {
		return unchanged
	}
	prefix := []string{"try", "{", "throw", "new", "Test262Error", "("}
	for i, s := range prefix {
		if t[i].text != s {
			return unchanged
		}
	}
	payload, throwClose, ok := tokenArguments(t, 5)
	if !ok || len(payload) != 1 {
		return unchanged
	}
	a, b := payload[0][0], payload[0][1]
	if b-a < 3 || t[a].kind != "str" || t[a+1].text != "+" {
		return unchanged
	}
	a += 2
	for a < b && t[a].text == "(" {
		wrapped, close, ok := tokenArguments(t, a)
		if !ok || close != b-1 || len(wrapped) != 1 {
			break
		}
		a++
		b--
	}
	pattern, flags, close, ok := regexpPureConstructor(t, a, b, 0)
	if !ok {
		return unchanged
	}
	next := close + 1
	if next < b {
		if next+2 >= b || t[next].text != "." || t[next+1].text != "exec" {
			return unchanged
		}
		args, end, ok := tokenArguments(t, next+2)
		if !ok || len(args) != 1 || end != b-1 || args[0][1]-args[0][0] != 1 {
			return unchanged
		}
		if _, ok = regexpStringToken(t[args[0][0]]); !ok {
			return unchanged
		}
		next = end + 1
	}
	if next != b {
		return unchanged
	}
	var syntax *regex.SyntaxError
	if _, err := regex.Compile(pattern, flags); !errors.As(err, &syntax) {
		return unchanged
	}
	at := throwClose + 1
	if at < len(t) && t[at].text == ";" {
		at++
	}
	if at+7 >= len(t) || t[at].text != "}" || t[at+1].text != "catch" || t[at+2].text != "(" || t[at+3].kind != "ident" || t[at+4].text != ")" || t[at+5].text != "{" || t[at+6].text != "assert" || t[at+7].text != "." {
		return unchanged
	}
	name := t[at+3].text
	at += 8
	if at+1 >= len(t) || t[at].text != "sameValue" {
		return unchanged
	}
	args, end, ok := tokenArguments(t, at+1)
	if !ok || len(args) != 3 {
		return unchanged
	}
	first := args[0]
	if first[1]-first[0] != 3 || t[first[0]].text != name || t[first[0]+1].text != "instanceof" || t[first[0]+2].text != "SyntaxError" {
		return unchanged
	}
	if args[1][1]-args[1][0] != 1 || t[args[1][0]].text != "true" || args[2][1]-args[2][0] != 1 || t[args[2][0]].kind != "str" {
		return unchanged
	}
	end++
	if end < len(t) && t[end].text == ";" {
		end++
	}
	if end != len(t)-1 || t[end].text != "}" {
		return unchanged
	}
	p, _ := json.Marshal(pattern)
	f, _ := json.Marshal(flags)
	counts["regex-unreachable-constructor-diagnostic"] = 1
	return adapted{Source: "assert.throws(SyntaxError, function() { new RegExp(" + string(p) + ", " + string(f) + "); }, " + t[args[2][0]].text + ");", Counts: counts}
}
