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
	tokens := tokenize(source)
	if !regexpPristine(tokens) {
		return adapted{Source: source, Counts: counts}
	}
	for index, token := range tokens {
		if token.text == "assertRegExpSyntaxError" ||
			token.text == "globalThis" ||
			token.text == "Object" ||
			token.text == "Reflect" ||
			token.text == "eval" ||
			token.text == "Function" {
			return adapted{Source: source, Counts: counts}
		}
		if token.text == "SyntaxError" &&
			(index < 4 ||
				tokens[index-1].text != "(" ||
				tokens[index-2].text != "throws" ||
				tokens[index-3].text != "." ||
				tokens[index-4].text != "assert") {
			return adapted{Source: source, Counts: counts}
		}
		if token.text == "RegExp" && (index+1 >= len(tokens) || tokens[index+1].text != "(") {
			return adapted{Source: source, Counts: counts}
		}
		if token.text == "assert" &&
			(index+3 >= len(tokens) || tokens[index+1].text != "." || tokens[index+3].text != "(") {
			return adapted{Source: source, Counts: counts}
		}
	}
	var edits []edit
	for index := 0; index+3 < len(tokens); index++ {
		if tokens[index].text != "assert" || tokens[index+1].text != "." || tokens[index+2].text != "throws" {
			continue
		}
		args, end, ok := tokenArguments(tokens, index+3)
		if !ok ||
			len(args) < 2 ||
			len(args) > 3 ||
			args[0][1]-args[0][0] != 1 ||
			tokens[args[0][0]].text != "SyntaxError" {
			continue
		}
		a, b := args[1][0], args[1][1]
		if b-a >= 5 &&
			tokens[a].text == "function" &&
			tokens[a+1].text == "(" &&
			tokens[a+2].text == ")" &&
			tokens[a+3].text == "{" &&
			tokens[b-1].text == "}" {
			a += 4
			b--
		} else if b-a >= 4 && tokens[a].text == "(" && tokens[a+1].text == ")" && tokens[a+2].text == "=>" {
			a += 3
			if tokens[a].text == "{" && tokens[b-1].text == "}" {
				a++
				b--
			}
		} else {
			continue
		}
		if a < b && tokens[b-1].text == ";" {
			b--
		}
		// A sole local declaration cannot be observed when construction throws.
		if b-a >= 3 &&
			(tokens[a].text == "let" || tokens[a].text == "const" || tokens[a].text == "var") &&
			tokens[a+1].kind == "ident" &&
			tokens[a+2].text == "=" {
			a += 3
		}
		if a < b && tokens[a].text == "return" {
			a++
		}
		if a < b && tokens[a].text == "new" {
			a++
		}
		if a >= b || tokens[a].text != "RegExp" {
			continue
		}
		values, close, ok := tokenArguments(tokens, a+1)
		if !ok || close != b-1 || len(values) < 1 || len(values) > 2 {
			continue
		}
		literal := func(v [2]int) (string, bool) {
			if v[1]-v[0] != 1 || tokens[v[0]].kind != "str" {
				return "", false
			}
			s := scanner.NewScanner()
			s.SetText(tokens[v[0]].text)
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
		_, compileError := regex.Compile(pattern, flags)
		var syntax *regex.SyntaxError
		if !errors.As(compileError, &syntax) {
			continue
		}
		replacement := "assertRegExpSyntaxError(" + tokens[values[0][0]].text + ", "
		if len(values) == 2 {
			replacement += tokens[values[1][0]].text
		} else {
			replacement += `""`
		}
		if len(args) == 3 {
			if _, ok = literal(args[2]); !ok {
				continue
			}
			replacement += ", " + tokens[args[2][0]].text
		}
		replacement += ")"
		edits = append(edits, edit{tokens[index].start, tokens[end].end, replacement})
		counts["regex-intrinsic-syntax-assertion"]++
		index = end
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
