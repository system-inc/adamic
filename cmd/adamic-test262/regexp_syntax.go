package main

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	regex "github.com/system-inc/adamic/internal/regexp"
)

// Constructor identity is provable here without a general prototype model:
// literal strings cannot throw during argument evaluation, and the pristine
// intrinsic is the only operation in the callback. Its syntax failures are
// intrinsic SyntaxErrors. All other callbacks retain the harness refusal.
func adaptRegExpSyntaxAssertions(source string) adapted {
	counts := map[string]int{}
	t := tokenize(source)
	if !regexpPristine(t) {
		return adapted{Source: source, Counts: counts}
	}
	for i, token := range t {
		if token.text == "assertRegExpSyntaxError" || token.text == "globalThis" || token.text == "Object" || token.text == "Reflect" || token.text == "eval" || token.text == "Function" {
			return adapted{Source: source, Counts: counts}
		}
		if token.text == "SyntaxError" && (i < 4 || t[i-1].text != "(" || t[i-2].text != "throws" || t[i-3].text != "." || t[i-4].text != "assert") {
			return adapted{Source: source, Counts: counts}
		}
		if token.text == "RegExp" && (i+1 >= len(t) || t[i+1].text != "(") {
			return adapted{Source: source, Counts: counts}
		}
		if token.text == "assert" && (i+3 >= len(t) || t[i+1].text != "." || t[i+3].text != "(") {
			return adapted{Source: source, Counts: counts}
		}
	}
	var edits []edit
	for i := 0; i+3 < len(t); i++ {
		if t[i].text != "assert" || t[i+1].text != "." || t[i+2].text != "throws" {
			continue
		}
		args, end, ok := tokenArguments(t, i+3)
		if !ok || len(args) < 2 || len(args) > 3 || args[0][1]-args[0][0] != 1 || t[args[0][0]].text != "SyntaxError" {
			continue
		}
		a, b := args[1][0], args[1][1]
		if b-a >= 5 && t[a].text == "function" && t[a+1].text == "(" && t[a+2].text == ")" && t[a+3].text == "{" && t[b-1].text == "}" {
			a += 4
			b--
		} else if b-a >= 4 && t[a].text == "(" && t[a+1].text == ")" && t[a+2].text == "=>" {
			a += 3
			if t[a].text == "{" && t[b-1].text == "}" {
				a++
				b--
			}
		} else {
			continue
		}
		if a < b && t[b-1].text == ";" {
			b--
		}
		// A sole local declaration cannot be observed when construction throws.
		if b-a >= 3 && (t[a].text == "let" || t[a].text == "const" || t[a].text == "var") && t[a+1].kind == "ident" && t[a+2].text == "=" {
			a += 3
		}
		if a < b && t[a].text == "return" {
			a++
		}
		if a < b && t[a].text == "new" {
			a++
		}
		if a >= b || t[a].text != "RegExp" {
			continue
		}
		values, close, ok := tokenArguments(t, a+1)
		if !ok || close != b-1 || len(values) < 1 || len(values) > 2 {
			continue
		}
		literal := func(v [2]int) (string, bool) {
			if v[1]-v[0] != 1 || t[v[0]].kind != "str" {
				return "", false
			}
			s := scanner.NewScanner()
			s.SetText(t[v[0]].text)
			if s.Scan() != ast.KindStringLiteral {
				return "", false
			}
			text := s.TokenValue()
			return text, s.Scan() == ast.KindEndOfFile
		}
		pattern, ok := literal(values[0])
		if !ok {
			continue
		}
		flags := ""
		if len(values) == 2 {
			flags, ok = literal(values[1])
			if !ok {
				continue
			}
		}
		_, err := regex.Compile(pattern, flags)
		var syntax *regex.SyntaxError
		if !errors.As(err, &syntax) {
			continue
		}
		replacement := "assertRegExpSyntaxError(" + t[values[0][0]].text + ", "
		if len(values) == 2 {
			replacement += t[values[1][0]].text
		} else {
			replacement += `""`
		}
		if len(args) == 3 {
			if _, ok = literal(args[2]); !ok {
				continue
			}
			replacement += ", " + t[args[2][0]].text
		}
		replacement += ")"
		edits = append(edits, edit{t[i].start, t[end].end, replacement})
		counts["regex-intrinsic-syntax-assertion"]++
		i = end
	}
	return adapted{Source: applyEdits(source, edits), Counts: counts}
}

const regexpSyntaxPrelude = `
function assertRegExpSyntaxError(pattern: string, flags: string, message?: string): void {
 let threw = false;
 try { new RegExp(pattern, flags); }
 catch (caught) {
  threw = true;
  if (!(caught instanceof Error)) { throw new Error("RegExp threw a non-Error"); }
  if (caught.name !== "SyntaxError") { throw new Error("RegExp threw " + caught.name); }
 }
 if (!threw) { throw new Error(message ?? "RegExp did not throw SyntaxError"); }
}
`
