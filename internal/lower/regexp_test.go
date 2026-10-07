package lower

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"testing"
)

func TestRegExpNativeRefusals(t *testing.T) {
	for _, source := range []string{
		"const matches = /a{18446744073709551616}/.test('a');",
		"console.log('a'.replace(/a/, (value: string) => ({value})));",
		"try { console.log('a'.replaceAll(/a/, 'b')); } catch {}",
		"const regex = /a/; regex.exec = (input: string): RegExpExecArray | null => null;",
		"const regex = /a/; const copy = {...regex};",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		var refused *Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			t.Fatalf("expected loud NotYet for %s: %v", source, err)
		}
	}
}

func TestRegExpSourceNode(t *testing.T) {
	cases := [][2]string{{"", ""}, {"[/]", ""}, {"[[/]/", ""}, {"[[a]--[b]]/", "v"}, {"\\[/", ""}, {"\\/", ""}, {"\\\\/", ""}, {"[\\]/]", ""}, {"(?<x>/)", ""}, {"\\\n", ""}, {"\r\n", "u"}, {"\u2028\u2029", ""}, {"K🌍", "iu"}}
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "-e", `if(!process.version.startsWith('v24.'))throw Error('Node 24 required');const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));const lone=new RegExp(String.fromCharCode(0xd800)).source;if(lone.length!==1 || lone.charCodeAt(0)!==0xd800)throw Error('unexpected lone-surrogate source');process.stdout.write(JSON.stringify(cases.map(([p,f])=>new RegExp(p,f).source)));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node source oracle: %v\n%s", err, output)
	}
	var expected []string
	if err = json.Unmarshal(output, &expected); err != nil {
		t.Fatal(err)
	}
	for index, c := range cases {
		if got := escapeRegexSource(c[0], c[1]); got != expected[index] {
			t.Errorf("source differs: %q/%s got=%q Node=%q", c[0], c[1], got, expected[index])
		}
	}
	lone := string([]byte{0xed, 0xa0, 0x80})
	if escapeRegexSource(lone, "") != lone {
		t.Error("source differs: lone surrogate bytes changed")
	}
}

func TestRegExpRuntimeConstructionLowers(t *testing.T) {
	for _, source := range []string{"function made(pattern: string): RegExp { return new RegExp(pattern); }", "function made(flags: string): RegExp { return new RegExp('a',flags); }"} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
