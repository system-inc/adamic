package lower

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

func TestRegExpNativeRefusals(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"function made(pattern: string): RegExp { return new RegExp(pattern); }",
		"const regex = /a/; regex.exec = (input: string): RegExpExecArray | null => null;",
		"const regex = /a/; const copy = {...regex};",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		var refused *Refused
		if !errors.As(lowerSourceError, &notYet) && !errors.As(lowerSourceError, &refused) {
			test.Fatalf("expected loud NotYet for %s: %v", source, lowerSourceError)
		}
	}
}

func TestRegExpSourceNode(test *testing.T) {
	test.Parallel()
	cases := [][2]string{
		{"", ""},
		{"[/]", ""},
		{"[[/]/", ""},
		{"[[a]--[b]]/", "v"},
		{"\\[/", ""},
		{"\\/", ""},
		{"\\\\/", ""},
		{"[\\]/]", ""},
		{"(?<x>/)", ""},
		{"\\\n", ""},
		{"\r\n", "u"},
		{"\u2028\u2029", ""},
		{"K🌍", "iu"},
	}
	data, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	command := toolchain.Command(
		"node",
		"-e",
		`if(!process.version.startsWith('v24.'))throw Error('Node 24 required');const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));const lone=new RegExp(String.fromCharCode(0xd800)).source;if(lone.length!==1 || lone.charCodeAt(0)!==0xd800)throw Error('unexpected lone-surrogate source');process.stdout.write(JSON.stringify(cases.map(([p,f])=>new RegExp(p,f).source)));`,
	)
	command.Stdin = bytes.NewReader(data)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node source oracle: %v\n%s", combinedOutputError, output)
	}
	var expected []string
	if decodeError := json.Unmarshal(output, &expected); decodeError != nil {
		test.Fatal(decodeError)
	}
	for index, c := range cases {
		if got := escapeRegexSource(c[0], c[1]); got != expected[index] {
			test.Errorf("source differs: %q/%s got=%q Node=%q", c[0], c[1], got, expected[index])
		}
	}
	lone := string([]byte{0xed, 0xa0, 0x80})
	if escapeRegexSource(lone, "") != lone {
		test.Error("source differs: lone surrogate bytes changed")
	}
}

func TestRegExpUnicodeClassSourceAgreesWithNode(test *testing.T) {
	test.Parallel()
	// Only native exposes lowered RegExp source metadata; JavaScript recomputes it.
	lowersAndAgreesWithNodeNative(test, "RegExpUnicodeClassSource")
}

func TestRegExpSlashClassSourceAgreesWithNode(test *testing.T) {
	test.Parallel()
	// Only native exposes lowered RegExp source metadata; JavaScript recomputes it.
	lowersAndAgreesWithNodeNative(test, "RegExpSlashClassSource")
}

func TestRegExpEscapedSlashSourceAgreesWithNode(test *testing.T) {
	test.Parallel()
	// Only native exposes lowered RegExp source metadata; JavaScript recomputes it.
	lowersAndAgreesWithNodeNative(test, "RegExpEscapedSlashSource")
}

func TestRegExpV8QuantifierBoundsLower(test *testing.T) {
	test.Parallel()
	if _, lowerSourceError := lowerSource(
		test,
		"const matches = /a{18446744073709551616}/.test('a');",
	); lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
}
