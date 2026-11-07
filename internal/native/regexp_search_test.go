package native

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

// Both the generic VM and generated straight-line runners use Node's captures.
func runRegExpSearchNodeUnit(test *testing.T, wasmOnly bool) {
	test.Helper()
	if wasmOnly && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	probes := []struct{ pattern, flags, input string }{
		{"^(a)\\1$", "", "a"},
		{"^(a)\\1$", "", "aa"},
		{"(?=a)b", "", "b"},
		{"(a|aa)(a?)", "", "aa"},
		{"a.*?b", "", "axbb"},
		{"a.*b", "", "axbb"},
		{"a.*z|b", "", "abzz"},
		{"a.*z|b", "", "abxx"},
		{"(a?){2,4}", "", "a"},
		{"(a?){2,4}?", "", "a"},
		{"\\b(?:if|interface)\\b", "", "x interface if"},
		{"a$", "m", "a\nb"},
		{"\\bK\\b", "iu", "K"},
		{"(a)", "", "xxa"},
		{"(a)", "g", "xxa"},
		{"(a)", "y", "a"},
		{"(a)+b", "", strings.Repeat("a", 200) + "b"},
		{"(?:a|b){200}", "", strings.Repeat("a", 200)},
		{"^shadow$", "", "shadow"},
		{"^shadow$", "", "xshadow"},
		{"^a", "m", "x\na"},
		{"a|b", "", "xxb"},
		{"a?", "", ""},
		{"(?:a|)", "", "x"},
		{"literal", "", "litxx literal"},
		{"literal", "", "\u006cliteral"},
		{"[ac]", "", "xxc"},
		{"[a-z]+", "", "123abc"},
		{"a(?=b)b", "", "xxab"},
		{"(?<=(a))b", "", "xxab"},
		{"[\\q{|ab}]", "v", "x"},
		{"k", "iu", "K"},
		{"k", "i", "K"},
		{"(.)\\1", "iu", "KK"},
		{"(.)\\1", "i", "KK"},
		{"(a|(b))+", "", "aba"},
		{"(a|(b))+?c", "", "abbac"},
		{"[\\uDF0D]", "", "🌍"},
		{"[\\uDF0D]", "u", "🌍"},
		{"(a)", "gu", "🌍a"},
		{"[a]", "yu", "🌍a"},
		{"word$", "", "wordx"},
		{"a", "y", "xa"},
		{"", "g", ""},
		{"[AEIOUY]", "", "xxxY"},
		{"[a-z]", "i", "Z"},
		{"\\u00C5", "", "Ł"},
		{"[a-z]cfg($|[A-Z0-9])", "", "xxacfgA"},
		{"([a-z])Ms($|[A-Z])", "", "xxaMsB"},
		{"[a-z]cfg", "g", "xxacfgacfg"},
		{"[a-z]cfg", "y", "xxacfg"},
		{"a$b", "", "ab"},
		{"a$", "y", "xa"},
		{"[a-z]", "i", "z"},
		{"az", "i", "aZ"},
		{"^(.*):", "", "abc:tail"},
		{"^(.*):", "", "abc\n:def"},
		{"^([a-z]*):", "", "ab:"},
		{"^([a-z]*):", "", ":tail"},
		{"^([a-z]*):", "", "AB:"},
		{"^([a-z]*):", "i", "AB:"},
		{"^(.*):", "g", "a:a:"},
		{"^(.*?):", "y", "a:a:"},
		{"az", "i", "az"},
		{"^a", "", "b"},
	}
	var cases []regexCase
	for _, p := range probes {
		cases = append(cases, regexCase{Pattern: p.pattern, Flags: p.flags, Input: utf16.Encode([]rune(p.input))})
	}
	for _, pattern := range []string{"^.", "^🌍", "^"} {
		cases = append(cases, regexCase{Pattern: pattern, Flags: "gu", Input: utf16.Encode([]rune("🌍")), LastIndex: 1})
	}
	runRegexNodeCasesForTarget(test, cases, wasmOnly)
}

func runRegexNodeCases(test *testing.T, cases []regexCase) {
	test.Helper()
	runRegexNodeCasesForTarget(test, cases, false)
}

func runRegexNodeCasesForTarget(test *testing.T, cases []regexCase, wasmOnly bool) {
	test.Helper()
	data, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		"node",
		"-e",
		`const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));if(!process.version.startsWith('v24.'))throw Error('Node 24 required');for(const c of cases){const r=new RegExp(c.pattern,c.flags+'d');r.lastIndex=c.lastIndex;const m=r.exec(String.fromCharCode(...c.input));c.expected={captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};}process.stdout.write(JSON.stringify(cases));`,
	)
	command.Stdin = bytes.NewReader(data)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, output)
	}
	if decodeError := json.Unmarshal(output, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	runRegexCasesForTarget(test, cases, wasmOnly)
}

// Freeze every benchmark input as a per-case oracle, beyond timing checksums.
func TestRegExpLintPatternsNode(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile("../../bench/regex/cases.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var probes []struct {
		Name, Pattern, Flags string
		Inputs               []string
	}
	if decodeError := json.Unmarshal(data, &probes); decodeError != nil {
		test.Fatal(decodeError)
	}
	var cases []regexCase
	for _, p := range probes {
		for _, input := range p.Inputs {
			cases = append(cases, regexCase{Pattern: p.Pattern, Flags: p.Flags, Input: utf16.Encode([]rune(input))})
		}
	}
	test.Logf("%d benchmark patterns, %d individual inputs", len(probes), len(cases))
	runRegexNodeCases(test, cases)
}

func TestRegExpSearchNode(test *testing.T) { test.Parallel(); runRegExpSearchNodeUnit(test, false) }
func TestRegExpSearchWASM(test *testing.T) { test.Parallel(); runRegExpSearchNodeUnit(test, true) }
