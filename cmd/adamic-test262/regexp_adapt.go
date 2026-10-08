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
	for _, t := range tokens {
		if t.kind != "regexp" {
			continue
		}
		last := strings.LastIndex(t.text, "/")
		if last <= 0 {
			continue
		}
		pattern := t.text[1:last]
		flags := t.text[last+1:]
		p, _ := json.Marshal(pattern)
		f, _ := json.Marshal(flags)
		edits = append(edits, edit{t.start, t.end, "(new RegExp(" + string(p) + ", " + string(f) + "))"})
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
	t := tokenize(source)
	var edits []edit
	for i := 0; i+5 < len(t); i++ {
		if (t[i].text != "const" && t[i].text != "let") || t[i+1].kind != "ident" || t[i+2].text != "=" || t[i+3].text != "[" || t[i+4].text != "]" || t[i+5].text != ";" {
			continue
		}
		name := t[i+1].text
		valid, pushes := true, 0
		for j := 0; j < len(t); j++ {
			if t[j].text != name || j == i+1 {
				continue
			}
			if j <= i+5 || j+2 >= len(t) || t[j+1].text != "." {
				valid = false
				break
			}
			switch t[j+2].text {
			case "length":
				if j+3 < len(t) && (t[j+3].text == "=" || t[j+3].text == "++" || t[j+3].text == "--" || t[j+3].text == "+=" || t[j+3].text == "-=") {
					valid = false
				}
			case "push", "join":
				args, _, ok := tokenArguments(t, j+3)
				if !ok || len(args) != 1 || args[0][0] >= args[0][1] || t[args[0][0]].kind != "str" {
					valid = false
					break
				}
				if t[j+2].text == "push" {
					pushes++
					if args[0][1]-args[0][0] > 1 && t[args[0][0]+1].text != "+" {
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
			edits = append(edits, edit{t[i+1].end, t[i+1].end, ": string[]"})
			counts["regex-string-collector"]++
		}
	}
	return applyEdits(source, edits)
}

// Generated character-class tests format a code point from the nonempty string
// yielded by for-of. Retain an explicit runtime check, without a type assertion.
func adaptRegExpCodePoint(source string, counts map[string]int) string {
	t := tokenize(source)
	for _, token := range t {
		if token.text == "try" || token.text == "catch" {
			return source
		}
	}
	var edits []edit
	for i := 0; i+8 < len(t); i++ {
		if t[i].kind != "ident" || (i > 0 && (t[i-1].text == "." || t[i-1].text == "?.")) || t[i+1].text != "." || t[i+2].text != "codePointAt" || t[i+3].text != "(" || t[i+4].text != "0" || t[i+5].text != ")" || t[i+6].text != "." || t[i+7].text != "toString" || t[i+8].text != "(" {
			continue
		}
		edits = append(edits, edit{t[i].start, t[i+5].end, "requiredCodePoint(" + source[t[i].start:t[i+5].end] + ")"})
		counts["regex-checked-code-point"]++
		i += 5
	}
	return applyEdits(source, edits)
}

// The upstream helper reads this optional field only. A fresh literal can
// supply its absent value explicitly without exposing a changed object shape.
func adaptRegExpHarnessRecord(source string, counts map[string]int) string {
	t := tokenize(source)
	var edits []edit
	for i := 1; i < len(t); i++ {
		if (t[i].text == "testPropertyOfStrings" || t[i].text == "testExtendedCharacterClass") && (t[i-1].text == "function" || t[i-1].text == "const" || t[i-1].text == "let" || t[i-1].text == "var" || (i+1 < len(t) && t[i+1].text == "=")) {
			return source
		}
	}
	for i := 0; i+2 < len(t); i++ {
		if t[i].text != "testPropertyOfStrings" && t[i].text != "testExtendedCharacterClass" {
			continue
		}
		if i > 0 && t[i-1].text == "." {
			continue
		}
		args, end, ok := tokenArguments(t, i+1)
		if !ok || len(args) != 1 {
			continue
		}
		a := args[0]
		if a[1]-a[0] < 2 || t[a[0]].text != "{" || t[a[1]-1].text != "}" {
			continue
		}
		found, unsafe := false, false
		depth := 0
		for j := a[0] + 1; j < a[1]-1; j++ {
			if depth == 0 && (j == a[0]+1 || t[j-1].text == ",") {
				if t[j].text == "nonMatchStrings" {
					found = true
				}
				if t[j].kind != "ident" || j+1 >= a[1] || t[j+1].text != ":" {
					unsafe = true
				}
			}
			if matchingClose(t[j].text) != "" {
				depth++
			}
			if t[j].text == "}" || t[j].text == "]" || t[j].text == ")" {
				depth--
			}
		}
		if !found && !unsafe {
			comma := ", "
			if a[1]-a[0] == 2 || t[a[1]-2].text == "," {
				comma = " "
			}
			edits = append(edits, edit{t[a[1]-1].start, t[a[1]-1].start, comma + "nonMatchStrings: undefined "})
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
	for i := start; i < len(tokens); i++ {
		s := tokens[i].text
		if close := matchingClose(s); close != "" {
			stack = append(stack, close)
			continue
		}
		if s == ")" || s == "]" || s == "}" {
			if len(stack) == 0 || stack[len(stack)-1] != s {
				return nil, 0, false
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				if i > start {
					args = append(args, [2]int{start, i})
				}
				return args, i, true
			}
		}
		if s == "," && len(stack) == 1 {
			args = append(args, [2]int{start, i})
			start = i + 1
		}
	}
	return nil, 0, false
}

func adaptRegExpNull(source string, counts map[string]int) string {
	tokens := tokenize(source)
	var edits []edit
	for i := 0; i+3 < len(tokens); i++ {
		if tokens[i].text != "assert" || tokens[i+1].text != "." || (tokens[i+2].text != "sameValue" && tokens[i+2].text != "notSameValue") {
			continue
		}
		args, end, ok := tokenArguments(tokens, i+3)
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
		if tokens[i+2].text == "notSameValue" {
			op = " !== null"
		}
		replacement := "assert.sameValue((" + source[tokens[other[0]].start:tokens[other[1]-1].end] + ")" + op + ", true"
		if len(args) == 3 {
			a := args[2]
			replacement += ", " + source[tokens[a[0]].start:tokens[a[1]-1].end]
		}
		replacement += ")"
		edits = append(edits, edit{tokens[i].start, tokens[end].end, replacement})
		counts["regex-null-assert"]++
		i = end
	}
	return applyEdits(source, edits)
}

func adaptRegExpExpected(source string, counts map[string]int) string {
	t := tokenize(source)
	var edits []edit
	for i := 0; i+3 < len(t); i++ {
		if (t[i].text != "let" && t[i].text != "const") || t[i+2].text != "=" || t[i+3].text != "[" {
			continue
		}
		name := t[i+1].text
		end := i + 3
		depth := 0
		for ; end < len(t); end++ {
			if t[end].text == "[" {
				depth++
			}
			if t[end].text == "]" {
				depth--
				if depth == 0 {
					break
				}
			}
		}
		// The two writes must immediately follow, with literal RHSs. Evaluation
		// order is then the array first, index second, input third on both sides.
		if end+13 >= len(t) || t[end+1].text != ";" {
			continue
		}
		a := end + 2
		b := end + 8
		if t[a].text != name || t[a+1].text != "." || t[a+2].text != "index" || t[a+3].text != "=" || t[a+4].kind != "num" || t[a+5].text != ";" || t[b].text != name || t[b+1].text != "." || t[b+2].text != "input" || t[b+3].text != "=" || t[b+4].kind != "str" || t[b+5].text != ";" {
			continue
		}
		valid := true
		var reads []edit
		for j := 0; j < len(t); j++ {
			if t[j].text != name || j == i+1 || j == a || j == b {
				continue
			}
			if j < b+6 || j+1 >= len(t) {
				valid = false
				break
			}
			if t[j+1].text == "." && j+2 < len(t) {
				if (t[j+2].text == "index" || t[j+2].text == "input") && j+3 < len(t) && t[j+3].text != "=" {
					continue
				}
				if t[j+2].text != "length" {
					valid = false
					break
				}
			} else if t[j+1].text != "[" {
				valid = false
				break
			}
			reads = append(reads, edit{t[j].end, t[j].end, ".entries"})
		}
		if !valid {
			continue
		}
		replacement := "{ entries: " + source[t[i+3].start:t[end].end] + ", index: " + t[a+4].text + ", input: " + t[b+4].text + " }"
		edits = append(edits, edit{t[i+3].start, t[b+5].end, replacement + ";"})
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
	for _, t := range parsed.toks {
		if t.text == "try" || t.text == "catch" {
			return source
		}
	}
	t := parsed.toks
	var edits []edit
	for _, ref := range parsed.refs {
		binding := lookup(ref.scope, ref.name)
		if binding == nil || binding.poison || len(binding.assigns) != 0 || binding.init == nil || binding.init.op != "call" || binding.init.left == nil || binding.init.left.op != "member" {
			continue
		}
		method := binding.init.left.prop
		if method != "exec" && method != "match" {
			continue
		}
		j := ref.at
		if j < 0 || j+1 >= len(t) || (t[j+1].text != "." && t[j+1].text != "[") {
			continue
		}
		if t[j+1].text == "." && j+2 < len(t) && !strings.Contains("|length|index|input|groups|indices|", "|"+t[j+2].text+"|") {
			continue
		}
		fn := "requiredRegexExec"
		if method == "match" {
			fn = "requiredRegexMatch"
		}
		edits = append(edits, edit{t[j].start, t[j].end, fn + "(" + ref.name + ")"})
		counts["regex-checked-match-read"]++
	}
	return applyEdits(source, edits)
}

func regexpPristine(tokens []token) bool {
	for i, t := range tokens {
		if t.text != "RegExp" {
			continue
		}
		if i > 0 && (strings.Contains("|const|let|var|function|class|.|...|", "|"+tokens[i-1].text+"|") || (strings.Contains("|[|{|:|", "|"+tokens[i-1].text+"|") && (i+1 >= len(tokens) || tokens[i+1].text != "(")) || ((tokens[i-1].text == "(" || tokens[i-1].text == ",") && (i+1 >= len(tokens) || tokens[i+1].text != "("))) {
			return false
		}
		if i+1 < len(tokens) {
			switch tokens[i+1].text {
			case "=", "+=", "-=", "*=", "/=", "%=", "**=", "&=", "^=", "|=", "&&=", "||=", "??=", "<<=", ">>=", ">>>=", "=>", ".", "[", "++", "--":
				return false
			}
		}
	}
	return true
}
