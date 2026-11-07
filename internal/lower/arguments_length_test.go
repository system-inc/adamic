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

func TestMixedTupleSpreadIsNotYet(test *testing.T) {
	test.Parallel()
	// Node prints 2:7:word8; one argument slot can't hold both a number and a string representation.
	_, lowerSourceError := lowerSource(
		test,
		"function read(number: number, text: string): string { return `${arguments.length}:${number}:${text}`; }\nconsole.log(read(...[7, `word${8}`]));\n",
	)
	var gap *NotYet
	if !errors.As(lowerSourceError, &gap) ||
		!strings.Contains(gap.What, "a call spreading a tuple with differently represented elements") {
		test.Fatalf("got %v; want the mixed-representation tuple spread refusal", lowerSourceError)
	}
}

func TestArgumentsLengthReadNeighbors(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"function read(value?: number): number { return (arguments).length + (value ?? 0); } console.log(`${read()}`);",
		"function read(): number { const count = arguments.length; const arrow = () => count; return arrow(); } console.log(`${read()}`);",
		"const object = { arguments: { length: 3 } }; const arrow = () => object.arguments.length; console.log(`${arrow()}`);",
		"function sum(...items: number[]): number { return items.length; } console.log(`${sum(1, 2)}`);",
		"function greet(name: string, greeting?: string): string { return `${greeting ?? 'hi'} ${name}`; } const run: (name: string) => string = greet; console.log(run('a'));",
	} {
		lowersAndAgreesWithNode(test, source)
	}
}

func TestArgumentsLengthRefusalFixtures(test *testing.T) {
	test.Parallel()
	const other = "arguments other than a read of arguments.length; name the parameters, or take a rest parameter; only arguments.length may be read"
	for _, name := range []string{"indexing", "aliasing", "passing", "returning", "spreading", "writing", "arrow"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			source, readError := os.ReadFile(
				filepath.Join("..", "oracle", "testdata", "arguments_length_refused", name+".a"),
			)
			if readError != nil {
				test.Fatal(readError)
			}
			message := other
			if name == "writing" {
				message = "writing arguments.length; keep the count read-only; write an explicit local number instead"
			}
			if name == "arrow" {
				message = "arguments inside an arrow function; read arguments.length in the enclosing non-arrow function and capture that number"
			}
			for _, extension := range []string{".a", ".ts"} {
				path := filepath.Join(test.TempDir(), "main"+extension)
				if writeError := os.WriteFile(path, source, 0644); writeError != nil {
					test.Fatal(writeError)
				}
				checked, loadError := load.Load([]string{path})
				if loadError != nil {
					test.Fatal(loadError)
				}
				_, lowerError := Lower(context.Background(), checked)
				var refused *Refused
				if !errors.As(lowerError, &refused) ||
					!strings.HasSuffix(lowerError.Error(), "Adamic 0.1 refuses "+message) {
					test.Fatalf("%s: got %v; want pinned refusal %q", extension, lowerError, message)
				}
			}
		})
	}
}

// An override that doesn't read arguments.length still takes the count its signature's other implementations read,
// as a local named argument_count whose index is in its C name (native's emit_locals.go). Six classes each override
// such a reader, so one pass adds six of those locals: in an order of Go's choosing, if the pass ranged over its map,
// and then the same program would lower, and compile, to different C from one run to the next (#m27xj48).
func TestArgumentsCountLocalsAreNumberedTheSameEveryTime(test *testing.T) {
	test.Parallel()
	var source strings.Builder
	for group := range 6 {
		fmt.Fprintf(
			&source,
			"class Reader%[1]d { read(value?: number): string { return `${arguments.length}:${value ?? %[1]d}`; } }\n",
			group,
		)
		fmt.Fprintf(
			&source,
			"class Other%[1]d extends Reader%[1]d { read(value?: number): string { return `other:${value ?? %[1]d}`; } }\n",
			group,
		)
		fmt.Fprintf(&source, "function call%[1]d(reader: Reader%[1]d): string { return reader.read(); }\n", group)
		fmt.Fprintf(&source, "console.log(`${call%[1]d(new Reader%[1]d())} ${call%[1]d(new Other%[1]d())}`);\n", group)
	}
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source.String()), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	// Each local named argument_count, with the function that holds it, in the order the program numbers them.
	numbering := func() string {
		program, lowerError := Lower(context.Background(), checked)
		if lowerError != nil {
			test.Fatal(lowerError)
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
		test.Fatalf("want an argument_count local in each of the 12 read methods, got %q", first)
	}
	// Go starts each range over a map at a random entry, so twenty runs in one order are no accident.
	for run := 1; run < 20; run++ {
		if again := numbering(); again != first {
			test.Fatalf(
				"run %d numbered the argument_count locals\n%s\nwhere the first run numbered them\n%s",
				run,
				again,
				first,
			)
		}
	}
}

func TestArgumentsLengthReadKeepsReaderFact(test *testing.T) {
	test.Parallel()
	program := lowersAndAgreesWithNode(
		test,
		"function read(): number { return arguments.length; } console.log(`${read()}`);",
	)
	// Behavior on this zero-argument call cannot observe the hidden reader calling convention.
	for _, function := range program.Functions {
		if function.Name != "read" {
			continue
		}
		if function.ArgumentsCount != 1 {
			test.Fatalf("read: ArgumentsCount = %d, want 1", function.ArgumentsCount)
		}
		if !function.ReadsArguments {
			test.Fatal("read: ArgumentsCount exists but ReadsArguments is false")
		}
		return
	}
	test.Fatal("lowering lost the read function")
}
