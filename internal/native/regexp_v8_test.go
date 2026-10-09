package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	regex "github.com/system-inc/adamic/internal/regexp"
)

// Every refusal is counted explicitly. The Go reference still runs the entire
// corpus, while production emits neither backend for these known V8 shapes.
func regexCompatibleCases(t *testing.T, cases []regexCase) []regexCase {
	t.Helper()
	programs := map[string]error{}
	var accepted []regexCase
	refused := 0
	for _, c := range cases {
		flags := c.Flags
		if !strings.Contains(flags, "d") {
			flags += "d"
		}
		key := c.Pattern + flags + fmt.Sprint(c.PatternUnits)
		err, seen := programs[key]
		if !seen {
			var p *regex.Program
			if c.PatternUnits != nil {
				p, err = regex.CompileUTF16(c.PatternUnits, flags)
			} else {
				p, err = regex.Compile(c.Pattern, flags)
			}
			if err != nil {
				t.Fatal(err)
			}
			err = p.NativeCompatibility()
			programs[key] = err
			if err != nil {
				var divergence *regex.V8DivergenceError
				if !errors.As(err, &divergence) {
					t.Fatal(err)
				}
				t.Logf("REFUSED pattern=%q units=%v flags=%q: %v", c.Pattern, c.PatternUnits, flags, err)
			}
		}
		if err != nil {
			refused++
			continue
		}
		accepted = append(accepted, c)
	}
	t.Logf("compatibility totals: %d cases, %d accepted, %d explicitly refused", len(cases), len(accepted), refused)
	return accepted
}

func TestRegExpBytecodeV8Node(t *testing.T) {
	var cases []regexCase
	for _, pattern := range []string{`\B`, `\b`, `(?!\W)`, `(?=\W)`, `(?!.)`, `(?<!.)`, `(?:)`, `\B(a?)`, `\B.`, `(?!\W).`, `(?=\B)`} {
		for _, flags := range []string{"g", "y", "gu", "yu", "giu", "yiu", "gv", "yv"} {
			for _, input := range []string{"a🌍b", "🌍", " 🌍 ", "ab", ""} {
				for index := uint64(0); index < 6; index++ {
					cases = append(cases, regexCase{Pattern: pattern, Flags: flags, Input: utf16.Encode([]rune(input)), LastIndex: index})
				}
			}
		}
	}
	for _, row := range [][3]string{
		{`[\q{AB}]`, "iv", "ab"}, {`[\q{ab}]`, "iv", "AB"}, {`[\q{Ss}]`, "iv", "sſ"},
		{`[a\q{a}]`, "iv", "A"}, {`[\q{12}]`, "iv", "12"},
		{`(?i:a)b`, "v", "AB"}, {`(?i:a)(?:[b])`, "v", "AB"},
		{`(?i:a)[0-9]`, "v", "A1"}, {`(?i:a)[b]`, "u", "aB"},
		{`(?i:^)[\w\W]`, "u", "K"}, {`(?-i:^)[\w\W]`, "iu", "ſ"},
		{`(?s:a).`, "", "a\n"}, {`(?-s:a).`, "s", "a\n"},
		{`(?m:)^a`, "", "\na"}, {`(?-m:)^a`, "m", "\na"},
		{`(?-i:^)[\q{AB}]`, "iv", "ab"}, {`(?i:a)\w`, "", "aK"},
		{`[[\q{ab|a|}]--[\q{}]]`, "v", "a"}, {`[\q{ab|}]`, "v", "a"},
	} {
		cases = append(cases, regexCase{Pattern: row[0], Flags: row[1], Input: utf16.Encode([]rune(row[2]))})
	}
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "-e", `if(!process.version.startsWith('v24.'))throw Error('Node 24 required');const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));for(const c of cases){const r=new RegExp(c.pattern,c.flags+'d');r.lastIndex=c.lastIndex;const m=r.exec(String.fromCharCode(...c.input));c.expected={captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups??null};}process.stdout.write(JSON.stringify(cases));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, output)
	}
	if err := json.Unmarshal(output, &cases); err != nil {
		t.Fatal(err)
	}
	runRegexCases(t, cases)
}
