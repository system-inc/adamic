package lower

import (
	"context"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArgumentsLengthRefusals(t *testing.T) {
	t.Parallel()
	const other = "arguments other than a read of arguments.length; name the parameters, or take a rest parameter; only arguments.length may be read"
	const writing = "writing arguments.length; keep the count read-only; write an explicit local number instead"
	for _, probe := range []struct{ name, body, message string }{
		{"indexing", "console.log(`${arguments[0]}`);", other},
		{"aliasing", "const alias = arguments; console.log(`${alias.length}`);", other},
		{"passing", "take(arguments);", other},
		{"returning", "return arguments;", other},
		{"spreading", "const copy = [...arguments]; console.log(`${copy.length}`);", other},
		{"writing object", "arguments[0] = 1;", other},
		{"writing length", "arguments.length = 1;", writing},
		{"incrementing length", "arguments.length++;", writing},
		{"compound length", "arguments.length += 1;", writing},
		{"arrow", "const read = () => arguments.length; console.log(`${read()}`);", "arguments inside an arrow function; read arguments.length in the enclosing non-arrow function and capture that number"},
		{"parenthesized write", "(arguments.length) = 1;", writing},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "function take(value: IArguments): void { console.log(`${value.length}`); }\nfunction rejected() { " + probe.body + " }\nrejected();\n"
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), "Adamic 0.1 refuses "+probe.message) {
				t.Fatalf("got %v; want pinned refusal %q", err, probe.message)
			}
		})
	}
}

func TestMixedTupleSpreadIsNotYet(t *testing.T) {
	t.Parallel()
	// Node prints 2:7:word8; one argument slot can't hold both a number and a string representation.
	_, err := lowerSource(t, "function read(number: number, text: string): string { return `${arguments.length}:${number}:${text}`; }\nconsole.log(read(...[7, `word${8}`]));\n")
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "a call spreading a tuple with differently represented elements") {
		t.Fatalf("got %v; want the mixed-representation tuple spread refusal", err)
	}
}

func TestArgumentsLengthReadNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function read(value?: number): number { return (arguments).length + (value ?? 0); } console.log(`${read()}`);",
		"function read(): number { const count = arguments.length; const arrow = () => count; return arrow(); } console.log(`${read()}`);",
		"const object = { arguments: { length: 3 } }; const arrow = () => object.arguments.length; console.log(`${arrow()}`);",
		"function sum(...items: number[]): number { return items.length; } console.log(`${sum(1, 2)}`);",
		"function greet(name: string, greeting?: string): string { return `${greeting ?? 'hi'} ${name}`; } const run: (name: string) => string = greet; console.log(run('a'));",
	} {
		lowersAndAgreesWithNode(t, source)
	}
}

func TestArgumentsLengthRefusalFixtures(t *testing.T) {
	t.Parallel()
	const other = "arguments other than a read of arguments.length; name the parameters, or take a rest parameter; only arguments.length may be read"
	for _, name := range []string{"indexing", "aliasing", "passing", "returning", "spreading", "writing", "arrow"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "arguments_length_refused", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			message := other
			if name == "writing" {
				message = "writing arguments.length; keep the count read-only; write an explicit local number instead"
			}
			if name == "arrow" {
				message = "arguments inside an arrow function; read arguments.length in the enclosing non-arrow function and capture that number"
			}
			for _, extension := range []string{".a", ".ts"} {
				path := filepath.Join(t.TempDir(), "main"+extension)
				if err := os.WriteFile(path, source, 0644); err != nil {
					t.Fatal(err)
				}
				checked, err := load.Load([]string{path})
				if err != nil {
					t.Fatal(err)
				}
				_, err = Lower(context.Background(), checked)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), "Adamic 0.1 refuses "+message) {
					t.Fatalf("%s: got %v; want pinned refusal %q", extension, err, message)
				}
			}
		})
	}
}

// An override that doesn't read arguments.length still takes the count its signature's other implementations read,
// as a local named argument_count whose index is in its C name (native's emit_locals.go). Six classes each override
// such a reader, so one pass adds six of those locals: in an order of Go's choosing, if the pass ranged over its map,
// and then the same program would lower, and compile, to different C from one run to the next (#m27xj48).
func TestArgumentsCountLocalsAreNumberedTheSameEveryTime(t *testing.T) {
	t.Parallel()
	var source strings.Builder
	for group := range 6 {
		fmt.Fprintf(&source, "class Reader%[1]d { read(value?: number): string { return `${arguments.length}:${value ?? %[1]d}`; } }\n", group)
		fmt.Fprintf(&source, "class Other%[1]d extends Reader%[1]d { read(value?: number): string { return `other:${value ?? %[1]d}`; } }\n", group)
		fmt.Fprintf(&source, "function call%[1]d(reader: Reader%[1]d): string { return reader.read(); }\n", group)
		fmt.Fprintf(&source, "console.log(`${call%[1]d(new Reader%[1]d())} ${call%[1]d(new Other%[1]d())}`);\n", group)
	}
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	// Each local named argument_count, with the function that holds it, in the order the program numbers them.
	numbering := func() string {
		program, err := Lower(context.Background(), checked)
		if err != nil {
			t.Fatal(err)
		}
		var locals []string
		for _, local := range program.Locals {
			if local.Name == "argument_count" {
				locals = append(locals, program.Functions[local.Function].Name)
			}
		}
		return strings.Join(locals, " ")
	}
	first := numbering()
	if strings.Count(first, "read") < 12 {
		t.Fatalf("want an argument_count local in each of the 12 read methods, got %q", first)
	}
	// Go starts each range over a map at a random entry, so twenty runs in one order are no accident.
	for run := 1; run < 20; run++ {
		if again := numbering(); again != first {
			t.Fatalf("run %d numbered the argument_count locals\n%s\nwhere the first run numbered them\n%s", run, again, first)
		}
	}
}

func TestArgumentsLengthReadKeepsReaderFact(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNode(t, "function read(): number { return arguments.length; } console.log(`${read()}`);")
	// Behavior on this zero-argument call cannot observe the hidden reader calling convention.
	for _, function := range program.Functions {
		if function.Name != "read" {
			continue
		}
		if function.ArgumentsCount != 1 {
			t.Fatalf("read: ArgumentsCount = %d, want 1", function.ArgumentsCount)
		}
		if !function.ReadsArguments {
			t.Fatal("read: ArgumentsCount exists but ReadsArguments is false")
		}
		return
	}
	t.Fatal("lowering lost the read function")
}
