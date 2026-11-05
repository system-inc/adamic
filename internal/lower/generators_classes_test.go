package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A class's generator method is lowered only when lowering reaches the class's methods, so an
// unused class costs nothing at lowering, as an uninstantiated generic does (#qk8rztp, Lower's open
// question 7). These pins, Lower's and Compiler's, hold what that may and may not mean.

// lowerModules lowers a program of several modules, written to one directory, from main.ts, as
// TestGeneratorCalledOtherThanByItsNameIsNotYet does.
func lowerModules(t *testing.T, modules map[string]string) (*ir.Program, error) {
	t.Helper()
	directory := t.TempDir()
	for name, source := range modules {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return Lower(context.Background(), program)
}

// (a) Refusals still walk dead code: an unused class whose generator method yields an any cast reads
// Refused, the walk's, though lowering never reaches the method. Mutant: the refusal walk pruned at
// a module's class declarations; the program lowers, and this goes red.
func TestUnusedClassGeneratorMethodIsStillWalked(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "class Box { *each(items: number[]): Generator<number, void, unknown> { for (const x of items) { yield (x as any); } } }\nconsole.log('control');\n")
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("got %v, want the walk's refusal of the any cast", err)
	}
}

// (b) Unused is decided across the whole program: a class exported from one module, imported and
// constructed in another, with its generator method called, stops in generatorForm's words, and so
// does a class reached only through another class's static method that calls the generator method.
// The same class exported and used nowhere lowers. Mutant: signature's NotYet for a form it doesn't
// lower dropped; the rows whose method has no yield compile, the for-of rows stop in other words,
// and this goes red.
func TestGeneratorMethodUseIsDecidedAcrossModules(t *testing.T) {
	t.Parallel()
	const generator = "Generator<number, void, unknown>"
	for _, probe := range []struct{ name, library, main string }{
		{"constructed and stepped in another module", "export class Box { *each(): " + generator + " { yield 1; } }\n", "import { Box } from './library';\nfor (const value of new Box().each()) console.log(String(value));\n"},
		{"constructed and dropped in another module", "export class Box { *each(): " + generator + " { return; } }\n", "import { Box } from './library';\nnew Box().each();\nconsole.log('control');\n"},
		{"reached through a static member that steps it", "export class Box { *each(): " + generator + " { yield 1; } }\nexport class Holder { static total(): number { let sum = 0; for (const value of new Box().each()) { sum += value; } return sum; } }\n", "import { Holder } from './library';\nconsole.log(String(Holder.total()));\n"},
		{"reached through a static member that drops it", "export class Box { *each(): " + generator + " { return; } }\nexport class Holder { static run(): void { new Box().each(); } }\n", "import { Holder } from './library';\nHolder.run();\nconsole.log('control');\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, map[string]string{"library.ts": probe.library, "main.ts": probe.main})
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "a generator function written as a method") {
				t.Fatalf("got %v, want a NotYet saying it is written as a method", err)
			}
		})
	}
	t.Run("exported and used nowhere", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerModules(t, map[string]string{
			"library.ts": "export class Box { *each(): " + generator + " { yield 1; } }\nexport function label(): string { return 'library'; }\n",
			"main.ts":    "import { label } from './library';\nconsole.log(label());\n",
		}); err != nil {
			t.Fatalf("an exported class nothing constructs: %v", err)
		}
	})
}

// (d) Compiler's both ways: an unused class with a generator method lowers, and the same class
// constructed with the method called stops in generatorForm's words. Mutants: the walk stopping a
// generator method (asyncGeneratorRefusal also taking one) turns the unused half red; signature's
// NotYet dropped turns the used half red, its dropped call compiling.
func TestGeneratorMethodBothWays(t *testing.T) {
	t.Parallel()
	const class = "class Box { *each(): Generator<number, void, unknown> { return; } }\n"
	if _, err := lowerSource(t, class+"console.log('control');\n"); err != nil {
		t.Errorf("an unused class with a generator method: %v", err)
	}
	_, err := lowerSource(t, class+"new Box().each();\nconsole.log('control');\n")
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "a generator function written as a method") {
		t.Errorf("the class constructed and its generator method called: got %v, want a NotYet saying it is written as a method", err)
	}
}
