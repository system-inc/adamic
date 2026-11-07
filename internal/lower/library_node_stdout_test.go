package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeStdoutWriteAdmission(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`import process from 'node:process'; process.stdout.write('plain');`,
		`import process from 'node:process'; ((process.stdout.write('é\ud800')));`,
		`import host from 'node:process'; host.stdout.write('clear\x1b[2J');`,
	} {
		if _, err := lowerSource(test, source); err != nil {
			test.Fatal(err)
		}
	}
}

func TestNodeStdoutWriteRefusals(test *testing.T) {
	test.Parallel()
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
		test.Run(one.reason+one.source, func(test *testing.T) {
			_, err := lowerSource(test, "import process from 'node:process';"+one.source)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(ny.What, one.reason) {
				test.Fatalf("want %s, got %v", one.reason, err)
			}
		})
	}
}

func TestNodeStdoutWriteExpressionForms(test *testing.T) {
	test.Parallel()
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
			test.Run(one.name+extension, func(test *testing.T) {
				_, err := lowerSourceAs(test, "import process from 'node:process';\n"+one.source, extension)
				if one.reason == "" {
					if err != nil {
						test.Fatalf("want admitted parenthesized receiver, got %v", err)
					}
					return
				}
				var notYet *NotYet
				if !errors.As(err, &notYet) || notYet.What != one.reason {
					test.Fatalf("want exact NotYet %q, got %v", one.reason, err)
				}
			})
		}
	}
}
