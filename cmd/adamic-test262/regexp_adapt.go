package main

import (
	"encoding/json"
	"strings"
)

// The matcher validates the current ECMAScript grammar and Unicode properties;
// the pinned TypeScript checker still rejects some of those literal spellings.
// Under the pristine intrinsic binding, an equivalent constant constructor has
// the same source, flags, identity, captures and evaluation position.
func adaptRegExpLiterals(source string) adapted {
	tokens := tokenize(source)
	counts := map[string]int{}
	if !regexpPristine(tokens) {
		return adapted{Source: source, Counts: counts}
	}
	var edits []edit
	for _, token := range tokens {
		if token.kind != "regexp" {
			continue
		}
		last := strings.LastIndex(token.text, "/")
		if last <= 0 {
			continue
		}
		pattern := token.text[1:last]
		flags := token.text[last+1:]
		p, _ := json.Marshal(pattern)
		f, _ := json.Marshal(flags)
		edits = append(edits, edit{token.start, token.end, "(new RegExp(" + string(p) + ", " + string(f) + "))"})
		counts["regex-literal-constructor"]++
	}
	return adapted{Source: applyEdits(source, edits), Counts: counts}
}

// These adapters preserve each assertion. Expected arrays with immediately
// assigned literal metadata become a fixed object holding the array and those
// fields. Match reads use a checked unwrap, never a cast asserting they exist.
// The original Node test is an independent prerequisite for every reported pass.
func adaptRegExp(source string) adapted {
	counts := map[string]int{}
	source = adaptRegExpNull(source, counts)
	source = adaptRegExpExpected(source, counts)
	source = adaptRegExpReads(source, counts)
	source = adaptRegExpCodePoint(source, counts)
	source = adaptRegExpStringCollector(source, counts)
	source = adaptRegExpHarnessRecord(source, counts)
	return adapted{Source: source, Counts: counts}
}

// Empty diagnostic collectors are otherwise typed never[] at their declaration.
// Prove all uses are string-producing pushes or length/join reads; no alias or
// reassignment is accepted. The annotation changes no evaluation or assertion.
func adaptRegExpStringCollector(source string, counts map[string]int) string {
	tokens := tokenize(source)
	var edits []edit
	for i := 0; i+5 < len(tokens); i++ {
		if (tokens[i].text != "const" && tokens[i].text != "let") ||
			tokens[i+1].kind != "ident" ||
			tokens[i+2].text != "=" ||
			tokens[i+3].text != "[" ||
			tokens[i+4].text != "]" ||
			tokens[i+5].text != ";" {
			continue
		}
		name := tokens[i+1].text
		valid, pushes := true, 0
		for j := 0; j < len(tokens); j++ {
			if tokens[j].text != name || j == i+1 {
				continue
			}
			if j <= i+5 || j+2 >= len(tokens) || tokens[j+1].text != "." {
				valid = false
				break
			}
			switch tokens[j+2].text {
			case "length":
				if j+3 < len(tokens) && (tokens[j+3].text == "=" ||
					tokens[j+3].text == "++" ||
					tokens[j+3].text == "--" ||
					tokens[j+3].text == "+=" ||
					tokens[j+3].text == "-=") {
					valid = false
				}
			case "push", "join":
				args, _, ok := tokenArguments(tokens, j+3)
				if !ok || len(args) != 1 || args[0][0] >= args[0][1] || tokens[args[0][0]].kind != "str" {
					valid = false
					break
				}
				if tokens[j+2].text == "push" {
					pushes++
					if args[0][1]-args[0][0] > 1 && tokens[args[0][0]+1].text != "+" {
						valid = false
					}
				} else if args[0][1]-args[0][0] != 1 {
					valid = false
				}
			default:
				valid = false
			}
			if !valid {
				break
			}
		}
		if valid && pushes > 0 {
			edits = append(edits, edit{tokens[i+1].end, tokens[i+1].end, ": string[]"})
			counts["regex-string-collector"]++
		}
	}
	return applyEdits(source, edits)
}

// Generated character-class tests format a code point from the nonempty string
// yielded by for-of. Retain an explicit runtime check, without a type assertion.
func adaptRegExpCodePoint(source string, counts map[string]int) string {
	tokens := tokenize(source)
	for _, token := range tokens {
		if token.text == "try" || token.text == "catch" {
			return source
		}
	}
	var edits []edit
	for index := 0; index+8 < len(tokens); index++ {
		if tokens[index].kind != "ident" ||
			(index > 0 && (tokens[index-1].text == "." || tokens[index-1].text == "?.")) ||
			tokens[index+1].text != "." ||
			tokens[index+2].text != "codePointAt" ||
			tokens[index+3].text != "(" ||
			tokens[index+4].text != "0" ||
			tokens[index+5].text != ")" ||
			tokens[index+6].text != "." ||
			tokens[index+7].text != "toString" ||
			tokens[index+8].text != "(" {
			continue
		}
		edits = append(edits, edit{
			tokens[index].start,
			tokens[index+5].end,
			"requiredCodePoint(" + source[tokens[index].start:tokens[index+5].end] + ")",
		})
		counts["regex-checked-code-point"]++
		index += 5
	}
	return applyEdits(source, edits)
}

// The upstream helper reads this optional field only. A fresh literal can
// supply its absent value explicitly without exposing a changed object shape.
func adaptRegExpHarnessRecord(source string, counts map[string]int) string {
	tokens := tokenize(source)
	var edits []edit
	for index := 1; index < len(tokens); index++ {
		if (tokens[index].text == "testPropertyOfStrings" || tokens[index].text == "testExtendedCharacterClass") &&
			(tokens[index-1].text == "function" ||
				tokens[index-1].text == "const" ||
				tokens[index-1].text == "let" ||
				tokens[index-1].text == "var" ||
				(index+1 < len(tokens) && tokens[index+1].text == "=")) {
			return source
		}
	}
	for i := 0; i+2 < len(tokens); i++ {
		if tokens[i].text != "testPropertyOfStrings" && tokens[i].text != "testExtendedCharacterClass" {
			continue
		}
		if i > 0 && tokens[i-1].text == "." {
			continue
		}
		args, end, ok := tokenArguments(tokens, i+1)
		if !ok || len(args) != 1 {
			continue
		}
		a := args[0]
		if a[1]-a[0] < 2 || tokens[a[0]].text != "{" || tokens[a[1]-1].text != "}" {
			continue
		}
		found, unsafe := false, false
		depth := 0
		for j := a[0] + 1; j < a[1]-1; j++ {
			if depth == 0 && (j == a[0]+1 || tokens[j-1].text == ",") {
				if tokens[j].text == "nonMatchStrings" {
					found = true
				}
				if tokens[j].kind != "ident" || j+1 >= a[1] || tokens[j+1].text != ":" {
					unsafe = true
				}
			}
			if matchingClose(tokens[j].text) != "" {
				depth++
			}
			if tokens[j].text == "}" || tokens[j].text == "]" || tokens[j].text == ")" {
				depth--
			}
		}
		if !found && !unsafe {
			comma := ", "
			if a[1]-a[0] == 2 || tokens[a[1]-2].text == "," {
				comma = " "
			}
			edits = append(
				edits,
				edit{tokens[a[1]-1].start, tokens[a[1]-1].start, comma + "nonMatchStrings: undefined "},
			)
			counts["regex-harness-optional-field"]++
		}
		i = end
	}
	return applyEdits(source, edits)
}

// tokenArguments returns half-open token ranges, with nested delimiters intact.
func tokenArguments(tokens []token, open int) ([][2]int, int, bool) {
	if open >= len(tokens) || tokens[open].text != "(" {
		return nil, 0, false
	}
	stack := []string{")"}
	start := open + 1
	var args [][2]int
	for index := start; index < len(tokens); index++ {
		text := tokens[index].text
		if close := matchingClose(text); close != "" {
			stack = append(stack, close)
			continue
		}
		if text == ")" || text == "]" || text == "}" {
			if len(stack) == 0 || stack[len(stack)-1] != text {
				return nil, 0, false
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				if index > start {
					args = append(args, [2]int{start, index})
				}
				return args, index, true
			}
		}
		if text == "," && len(stack) == 1 {
			args = append(args, [2]int{start, index})
			start = index + 1
		}
	}
	return nil, 0, false
}

func adaptRegExpNull(source string, counts map[string]int) string {
	tokens := tokenize(source)
	var edits []edit
	for tokenIndex := 0; tokenIndex+3 < len(tokens); tokenIndex++ {
		if tokens[tokenIndex].text != "assert" ||
			tokens[tokenIndex+1].text != "." ||
			(tokens[tokenIndex+2].text != "sameValue" && tokens[tokenIndex+2].text != "notSameValue") {
			continue
		}
		args, end, ok := tokenArguments(tokens, tokenIndex+3)
		if !ok || len(args) < 2 || len(args) > 3 {
			continue
		}
		null := -1
		for j := 0; j < 2; j++ {
			if args[j][1]-args[j][0] == 1 && tokens[args[j][0]].text == "null" {
				null = j
			}
		}
		if null < 0 {
			continue
		}
		other := args[1-null]
		if other[0] == other[1] {
			continue
		}
		op := " === null"
		if tokens[tokenIndex+2].text == "notSameValue" {
			op = " !== null"
		}
		replacement := "assert.sameValue((" +
			source[tokens[other[0]].start:tokens[other[1]-1].end] +
			")" +
			op +
			", true"
		if len(args) == 3 {
			a := args[2]
			replacement += ", " + source[tokens[a[0]].start:tokens[a[1]-1].end]
		}
		replacement += ")"
		edits = append(edits, edit{tokens[tokenIndex].start, tokens[end].end, replacement})
		counts["regex-null-assert"]++
		tokenIndex = end
	}
	return applyEdits(source, edits)
}

func adaptRegExpExpected(source string, counts map[string]int) string {
	tokens := tokenize(source)
	var edits []edit
	for i := 0; i+3 < len(tokens); i++ {
		if (tokens[i].text != "let" && tokens[i].text != "const") ||
			tokens[i+2].text != "=" ||
			tokens[i+3].text != "[" {
			continue
		}
		name := tokens[i+1].text
		end := i + 3
		depth := 0
		for ; end < len(tokens); end++ {
			if tokens[end].text == "[" {
				depth++
			}
			if tokens[end].text == "]" {
				depth--
				if depth == 0 {
					break
				}
			}
		}
		// The two writes must immediately follow, with literal RHSs. Evaluation
		// order is then the array first, index second, input third on both sides.
		if end+13 >= len(tokens) || tokens[end+1].text != ";" {
			continue
		}
		a := end + 2
		b := end + 8
		if tokens[a].text != name ||
			tokens[a+1].text != "." ||
			tokens[a+2].text != "index" ||
			tokens[a+3].text != "=" ||
			tokens[a+4].kind != "num" ||
			tokens[a+5].text != ";" ||
			tokens[b].text != name ||
			tokens[b+1].text != "." ||
			tokens[b+2].text != "input" ||
			tokens[b+3].text != "=" ||
			tokens[b+4].kind != "str" ||
			tokens[b+5].text != ";" {
			continue
		}
		valid := true
		var reads []edit
		for j := 0; j < len(tokens); j++ {
			if tokens[j].text != name || j == i+1 || j == a || j == b {
				continue
			}
			if j < b+6 || j+1 >= len(tokens) {
				valid = false
				break
			}
			if tokens[j+1].text == "." && j+2 < len(tokens) {
				if (tokens[j+2].text == "index" || tokens[j+2].text == "input") &&
					j+3 < len(tokens) &&
					tokens[j+3].text != "=" {
					continue
				}
				if tokens[j+2].text != "length" {
					valid = false
					break
				}
			} else if tokens[j+1].text != "[" {
				valid = false
				break
			}
			reads = append(reads, edit{tokens[j].end, tokens[j].end, ".entries"})
		}
		if !valid {
			continue
		}
		replacement := "{ entries: " +
			source[tokens[i+3].start:tokens[end].end] +
			", index: " +
			tokens[a+4].text +
			", input: " +
			tokens[b+4].text +
			" }"
		edits = append(edits, edit{tokens[i+3].start, tokens[b+5].end, replacement + ";"})
		edits = append(edits, reads...)
		counts["regex-expected-metadata"]++
		i = b + 5
	}
	return applyEdits(source, edits)
}

func adaptRegExpReads(source string, counts map[string]int) string {
	parsed := parseBody(source)
	if parsed.failed {
		return source
	}
	// A changed exception constructor could be observed by a catch. Such
	// programs stay unchanged and can be refused by the checker.
	for _, tok := range parsed.toks {
		if tok.text == "try" || tok.text == "catch" {
			return source
		}
	}
	tokens := parsed.toks
	var edits []edit
	for _, reference := range parsed.refs {
		binding := lookup(reference.scope, reference.name)
		if binding == nil ||
			binding.poison ||
			len(binding.assigns) != 0 ||
			binding.init == nil ||
			binding.init.op != "call" ||
			binding.init.left == nil ||
			binding.init.left.op != "member" {
			continue
		}
		method := binding.init.left.prop
		if method != "exec" && method != "match" {
			continue
		}
		index := reference.at
		if index < 0 || index+1 >= len(tokens) || (tokens[index+1].text != "." && tokens[index+1].text != "[") {
			continue
		}
		if tokens[index+1].text == "." &&
			index+2 < len(tokens) &&
			!strings.Contains("|length|index|input|groups|indices|", "|"+tokens[index+2].text+"|") {
			continue
		}
		fn := "requiredRegexExec"
		if method == "match" {
			fn = "requiredRegexMatch"
		}
		edits = append(edits, edit{tokens[index].start, tokens[index].end, fn + "(" + reference.name + ")"})
		counts["regex-checked-match-read"]++
	}
	return applyEdits(source, edits)
}

func regexpPristine(tokens []token) bool {
	for index, token := range tokens {
		if token.text != "RegExp" {
			continue
		}
		if index > 0 && (strings.Contains("|const|let|var|function|class|.|...|", "|"+tokens[index-1].text+"|") ||
			(strings.Contains("|[|{|:|", "|"+tokens[index-1].text+"|") &&
				(index+1 >= len(tokens) || tokens[index+1].text != "(")) ||
			((tokens[index-1].text == "(" || tokens[index-1].text == ",") &&
				(index+1 >= len(tokens) || tokens[index+1].text != "("))) {
			return false
		}
		if index+1 < len(tokens) {
			switch tokens[index+1].text {
			case "=", "+=", "-=", "*=", "/=", "%=", "**=", "&=", "^=", "|=", "&&=", "||=", "??=", "<<=", ">>=", ">>>=", "=>", ".", "[", "++", "--":
				return false
			}
		}
	}
	return true
}
