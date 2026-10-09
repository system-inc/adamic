package native

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf16"
)

// Both the generic VM and generated straight-line runners use Node's captures.
func TestRegExpSearchNode(t *testing.T) {
	t.Parallel()
	probes := []struct{ pattern, flags, input string }{
		{"(a)", "", "xxa"}, {"(a)", "g", "xxa"}, {"(a)", "y", "a"}, {"(a)+b", "", strings.Repeat("a", 200) + "b"}, {"(?:a|b){200}", "", strings.Repeat("a", 200)}, {"^shadow$", "", "shadow"}, {"^shadow$", "", "xshadow"},
		{"^a", "m", "x\na"}, {"a|b", "", "xxb"}, {"a?", "", ""}, {"(?:a|)", "", "x"},
		{"literal", "", "litxx literal"}, {"literal", "", "\u006cliteral"},
		{"[ac]", "", "xxc"}, {"[a-z]+", "", "123abc"},
		{"a(?=b)b", "", "xxab"}, {"(?<=(a))b", "", "xxab"},
		{"[\\q{|ab}]", "v", "x"}, {"k", "iu", "K"}, {"k", "i", "K"},
		{"(.)\\1", "iu", "KK"}, {"(.)\\1", "i", "KK"},
		{"(a|(b))+", "", "aba"}, {"(a|(b))+?c", "", "abbac"},
		{"(a)", "gu", "🌍a"}, {"[a]", "yu", "🌍a"},
	}
	var cases []regexCase
	for _, p := range probes {
		cases = append(cases, regexCase{Pattern: p.pattern, Flags: p.flags, Input: utf16.Encode([]rune(p.input))})
	}
	for _, pattern := range []string{"^.", "^🌍", "^"} {
		cases = append(cases, regexCase{Pattern: pattern, Flags: "gu", Input: utf16.Encode([]rune("🌍")), LastIndex: 1})
	}
	runRegexNodeCases(t, cases)
}

func runRegexNodeCases(t *testing.T, cases []regexCase) {
	t.Helper()
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "-e", `const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));if(!process.version.startsWith('v24.'))throw Error('Node 24 required');for(const c of cases){const r=new RegExp(c.pattern,c.flags+'d');r.lastIndex=c.lastIndex;const m=r.exec(String.fromCharCode(...c.input));c.expected={captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};}process.stdout.write(JSON.stringify(cases));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, output)
	}
	if err = json.Unmarshal(output, &cases); err != nil {
		t.Fatal(err)
	}
	runRegexCases(t, cases)
}

// Freeze every benchmark input as a per-case oracle, beyond timing checksums.
func TestRegExpLintPatternsNode(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../bench/regex/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var probes []struct {
		Name, Pattern, Flags string
		Inputs               []string
	}
	if err = json.Unmarshal(data, &probes); err != nil {
		t.Fatal(err)
	}
	var cases []regexCase
	for _, p := range probes {
		for _, input := range p.Inputs {
			cases = append(cases, regexCase{Pattern: p.Pattern, Flags: p.Flags, Input: utf16.Encode([]rune(input))})
		}
	}
	t.Logf("%d benchmark patterns, %d individual inputs", len(probes), len(cases))
	runRegexNodeCases(t, cases)
}
