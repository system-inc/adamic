package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeStdoutWriteAdmission(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import process from 'node:process'; process.stdout.write('plain');`,
		`import process from 'node:process'; ((process.stdout.write('é\ud800')));`,
		`import host from 'node:process'; host.stdout.write('clear\x1b[2J');`,
		`import process from 'node:process'; process.stdout.write(new Uint8Array(2));`,
		`import process from 'node:process'; const b = new Uint8Array(4); process.stdout.write(b.subarray(1, 3));`,
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
		{`function w(either: string | Uint8Array): void { process.stdout.write(either); }`, "non-string argument other than a Uint8Array"},
		// A Buffer is held as a Uint8Array, so it meets the host view stop first: Buffer seen as string | Uint8Array.
		{`import {Buffer} from 'node:buffer'; process.stdout.write(Buffer.from('x'));`, "Buffer or Hash viewed as another object type"},
		{`process.stdout.write('x','utf8');`, "exactly one string argument"},
		{`process.stdout.write('x',()=>{});`, "exactly one string argument"},
		{`const args:[string]=['x'];process.stdout.write(...args);`, "spread arguments"},
		{`let stream=process.stdout;`, "stdout used as a value"},
		{`const {write}=process.stdout;`, "stdout used as a value"},
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

// #1trc7d3: const <name> = process.stdout is stdout itself. The declaration lowers to nothing, and the name is
// admitted as the receiver of write, isTTY and columns, as process.stdout is; anywhere else it stops by name, so
// stdout stays a receiver and never a value. Mutants: the const requirement dropped, and TestNodeStdoutWriteRefusals'
// let row lowers; the alias's other uses admitted rather than stopped by name, and the passed, returned, stored and
// typeof rows go red.
func TestNodeStdoutAlias(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const stdout = process.stdout; stdout.write('a'); stdout.write(new Uint8Array(1));`,
		`const out = (process.stdout); function f(): void { out.write('b'); } f();`,
		`const stdout = process.stdout; const tty = stdout.isTTY === true; const width = stdout.columns;`,
		`const stdout = process.stdout;`,
	} {
		if _, err := lowerSource(t, "import process from 'node:process';"+source); err != nil {
			t.Fatalf("%s: %v", source, err)
		}
	}
	const named = "node:process.Process.stdout, bound to stdout, used other than as the receiver of write, isTTY or columns"
	for _, one := range []struct{ name, source, reason string }{
		{"passed", `const stdout = process.stdout; function f(s: typeof stdout): void {} f(stdout);`, named},
		{"returned", `const stdout = process.stdout; function g(): typeof stdout { return stdout; }`, named},
		{"stored", `const stdout = process.stdout; const box = { out: stdout };`, named},
		{"typeof", `const stdout = process.stdout; const kind = typeof stdout;`, named},
		{"written", `const stdout = process.stdout; stdout.write = (): boolean => true;`, ""},
	} {
		t.Run(one.name, func(t *testing.T) {
			_, err := lowerSource(t, "import process from 'node:process';"+one.source)
			var ny *NotYet
			var refused *Refused
			if !errors.As(err, &ny) && !errors.As(err, &refused) {
				t.Fatalf("want a stop, got %v", err)
			}
			if one.reason != "" && (ny == nil || ny.What != one.reason) {
				t.Fatalf("want NotYet %q, got %v", one.reason, err)
			}
		})
	}
	// A parameter typed as stdout's stream stops at its own storage, before any receiver rule sees its write: no
	// value of stdout's type exists in a lowered program but process.stdout itself. That is what makes the mutant
	// "admit any identifier of the stream's type as a receiver" unreachable today (J-hcbw31a: it survives). The
	// day this storage stop lifts, this row goes red, and the receiver rule needs a pin that reaches it with that
	// mutant.
	_, err := lowerSource(t, "import process from 'node:process'; function w(s: typeof process.stdout): void { s.write('x'); }")
	var ny *NotYet
	// The field's name carries the checker's internal symbol id, which varies, so it's matched by its parts. The
	// error's own words end in " yet", which the NotYet's What doesn't hold.
	if !errors.As(err, &ny) || !strings.Contains(ny.What, "a value of type WriteStream & { fd: 1; }, an intersection whose field") || !strings.Contains(ny.What, "captureRejectionSymbol") || !strings.Contains(ny.What, "is a function, which its record storage doesn't hold") || !strings.HasSuffix(ny.Where, ":1:48") {
		t.Fatalf("a stream-typed parameter must stop at its storage, :1:48, got %v", err)
	}
}
