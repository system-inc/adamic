package lower

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func TestRegExpNativeRefusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function made(pattern: string): RegExp { return new RegExp(pattern); }",
		"const matches = /a{18446744073709551616}/.test('a');",
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
	t.Parallel()
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

func TestRegExpUnicodeClassSourceAgreesWithNode(t *testing.T) {
	t.Parallel()
	source := "const regex = new RegExp('[[/]/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('//')}`);"
	program := lowersAndAgreesWithNode(t, source)
	regexNativeAgreesWithNode(t, program, source)
}

func TestRegExpSlashClassSourceAgreesWithNode(t *testing.T) {
	t.Parallel()
	source := "const regex = new RegExp('[/]/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('//')}`);"
	program := lowersAndAgreesWithNode(t, source)
	regexNativeAgreesWithNode(t, program, source)
}

func TestRegExpEscapedSlashSourceAgreesWithNode(t *testing.T) {
	t.Parallel()
	source := "const regex = new RegExp('\\\\/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('/')}`);"
	program := lowersAndAgreesWithNode(t, source)
	regexNativeAgreesWithNode(t, program, source)
}

// JavaScript's RegExp recomputes source and flags from the original constructor
// arguments, so it cannot observe corrupted lowering metadata. Native consumes
// that metadata; hold its printed behavior to the same source Node observation.
func regexNativeAgreesWithNode(t *testing.T, program *ir.Program, source string) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "source.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	want := runAgreementNode(t, path)
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("native regex: %v; stdout %q; stderr %q", err, stdout.Bytes(), stderr.Bytes())
	}
	if !bytes.Equal(stdout.Bytes(), want.stdout) || stderr.Len() != 0 {
		t.Fatalf("native regex stdout = %q, source Node = %q; stderr %q", stdout.Bytes(), want.stdout, stderr.Bytes())
	}
}
