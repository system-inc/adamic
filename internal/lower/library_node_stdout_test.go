package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeStdoutWriteAdmission(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import process from 'node:process'; process.stdout.write('plain');`,
		`import process from 'node:process'; ((process.stdout.write('é\ud800')));`,
		`import host from 'node:process'; host.stdout.write('clear\x1b[2J');`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNodeStdoutWriteRefusals(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ source, reason string }{
		{`const value=process.stdout.write('x');`, "its boolean is backpressure in Node and blocking success natively"},
		{`function f():boolean{return process.stdout.write('x');}`, "used as a value"},
		{`process.stdout.write(new Uint8Array(2));`, "non-string argument"},
		{`import {Buffer} from 'node:buffer'; process.stdout.write(Buffer.from('x'));`, "non-string argument"},
		{`process.stdout.write('x','utf8');`, "exactly one string argument"},
		{`process.stdout.write('x',()=>{});`, "exactly one string argument"},
		{`const args:[string]=['x'];process.stdout.write(...args);`, "spread arguments"},
		{`const stream=process.stdout;`, "stdout used as a value"},
		{`const write=process.stdout.write;`, "used as a value"},
		{`process.stdout?.write('x');`, "optional chain"},
	} {
		t.Run(one.reason+one.source, func(t *testing.T) {
			_, err := lowerSource(t, "import process from 'node:process';"+one.source)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(ny.What, one.reason) {
				t.Fatalf("want %s, got %v", one.reason, err)
			}
		})
	}
}

func TestNodeStdoutTTYRefusals(t *testing.T) {
	t.Parallel()
	for _, member := range []string{"isTTY", "columns"} {
		for _, ext := range []string{"a", "ts"} {
			t.Run(member+"."+ext, func(t *testing.T) {
				file := filepath.Join("testdata/stdout/refused", member+"."+ext)
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				_, err = lowerSourceAs(t, string(source), "."+ext)
				var ny *NotYet
				kind := "boolean"
				if member == "columns" {
					kind = "number"
				}
				want := "declared " + kind + ", but undefined off a TTY; adamic/host-tty"
				if !errors.As(err, &ny) || ny.What != want {
					t.Fatalf("want exact %q, got %v", want, err)
				}
			})
		}
	}
}

func TestNodeStdoutWriteExpressionForms(t *testing.T) {
	t.Parallel()
	const valueReason = "node:process.Process.stdout.write used as a value: its boolean is backpressure in Node and blocking success natively"
	for _, one := range []struct{ name, source, reason string }{
		{"expression_arrow", `const f = (s: string) => process.stdout.write(s);`, valueReason},
		{"void", `void process.stdout.write('x');`, valueReason},
		{"comma", `(process.stdout.write('x'), 0);`, valueReason},
		{"call", `process.stdout.write.call(process.stdout, 'x');`, valueReason},
		{"element", `process.stdout['write']('x');`, "node:process.Process.stdout used as a value"},
		{"parenthesized_receiver", `(process.stdout).write('x');`, ""},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(one.name+extension, func(t *testing.T) {
				_, err := lowerSourceAs(t, "import process from 'node:process';\n"+one.source, extension)
				if one.reason == "" {
					if err != nil {
						t.Fatalf("want admitted parenthesized receiver, got %v", err)
					}
					return
				}
				var notYet *NotYet
				if !errors.As(err, &notYet) || notYet.What != one.reason {
					t.Fatalf("want exact NotYet %q, got %v", one.reason, err)
				}
			})
		}
	}
}
