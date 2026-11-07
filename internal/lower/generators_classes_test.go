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
func lowerModules(test *testing.T, modules map[string]string) (*ir.Program, error) {
	test.Helper()
	directory := test.TempDir()
	for name, source := range modules {
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		test.Fatalf("Load: %v", loadError)
	}
	return Lower(context.Background(), program)
}

// (a) Refusals still walk dead code: an unused class whose generator method yields an any cast reads
// Refused, the walk's, though lowering never reaches the method. Mutant: the refusal walk pruned at
// a module's class declarations; the program lowers, and this goes red.
func TestUnusedClassGeneratorMethodIsStillWalked(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		"class Box { *each(items: number[]): Generator<number, void, unknown> { for (const x of items) { yield (x as any); } } }\nconsole.log('control');\n",
	)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) {
		test.Fatalf("got %v, want the walk's refusal of the any cast", lowerSourceError)
	}
}

// (b) Unused is decided across the whole program: a class exported from one module, imported and
// constructed in another, with its generator method called, stops in generatorForm's words, and so
// does a class reached only through another class's static method that calls the generator method.
// The same class exported and used nowhere lowers. Mutant: signature's NotYet for a form it doesn't
// lower dropped; the rows whose method has no yield compile, the for-of rows stop in other words,
// and this goes red.
func TestGeneratorMethodUseIsDecidedAcrossModules(test *testing.T) {
	test.Parallel()
	const generator = "Generator<number, void, unknown>"
	for _, probe := range []struct{ name, library, main string }{
		{
			"constructed and stepped in another module",
			"export class Box { *each(): " + generator + " { yield 1; } }\n",
			"import { Box } from './library';\nfor (const value of new Box().each()) console.log(String(value));\n",
		},
		{
			"constructed and dropped in another module",
			"export class Box { *each(): " + generator + " { return; } }\n",
			"import { Box } from './library';\nnew Box().each();\nconsole.log('control');\n",
		},
		{
			"reached through a static member that steps it",
			"export class Box { *each(): " +
				generator +
				" { yield 1; } }\nexport class Holder { static total(): number { let sum = 0; for (const value of new Box().each()) { sum += value; } return sum; } }\n",
			"import { Holder } from './library';\nconsole.log(String(Holder.total()));\n",
		},
		{
			"reached through a static member that drops it",
			"export class Box { *each(): " +
				generator +
				" { return; } }\nexport class Holder { static run(): void { new Box().each(); } }\n",
			"import { Holder } from './library';\nHolder.run();\nconsole.log('control');\n",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerModulesError := lowerModules(
				test,
				map[string]string{"library.ts": probe.library, "main.ts": probe.main},
			)
			var gap *NotYet
			if !errors.As(lowerModulesError, &gap) ||
				!strings.Contains(gap.What, "a generator function written as a method") {
				test.Fatalf("got %v, want a NotYet saying it is written as a method", lowerModulesError)
			}
		})
	}
	test.Run("exported and used nowhere", func(test *testing.T) {
		test.Parallel()
		if _, lowerModulesError := lowerModules(test, map[string]string{
			"library.ts": "export class Box { *each(): " +
				generator +
				" { yield 1; } }\nexport function label(): string { return 'library'; }\n",
			"main.ts": "import { label } from './library';\nconsole.log(label());\n",
		}); lowerModulesError != nil {
			test.Fatalf("an exported class nothing constructs: %v", lowerModulesError)
		}
	})
}

// (d) Compiler's both ways: an unused class with a generator method lowers, and the same class
// constructed with the method called stops in generatorForm's words. Mutants: the walk stopping a
// generator method (asyncGeneratorRefusal also taking one) turns the unused half red; signature's
// NotYet dropped turns the used half red, its dropped call compiling.
func TestGeneratorMethodBothWays(test *testing.T) {
	test.Parallel()
	const class = "class Box { *each(): Generator<number, void, unknown> { return; } }\n"
	if _, lowerSourceError := lowerSource(test, class+"console.log('control');\n"); lowerSourceError != nil {
		test.Errorf("an unused class with a generator method: %v", lowerSourceError)
	}
	_, lowerSourceError := lowerSource(test, class+"new Box().each();\nconsole.log('control');\n")
	var gap *NotYet
	if !errors.As(lowerSourceError, &gap) || !strings.Contains(gap.What, "a generator function written as a method") {
		test.Errorf(
			"the class constructed and its generator method called: got %v, want a NotYet saying it is written as a method",
			lowerSourceError,
		)
	}
}

// (f) A class method overloaded around a generator implementation: tsc refuses `*` on an overload
// signature (TS1222), so only the implementation carries it, and the class writes each overload's
// signature into the implementation's function first (class.go). The gate reads the implementation
// from the first overload, so the method, stepped or dropped, stops in generatorForm's words, never
// as an overload's ordinary signature. Mutant: signature's overload check (generatorImplementation)
// taken out; the first overload is written as an ordinary method and stops in other words, and this
// goes red.
//
// (e), an object literal's generator method on a used object, called, is
// TestGeneratorFormsStopAtLowering's "object literal method" rows.
func TestOverloadedGeneratorMethodStopsAtItsImplementation(test *testing.T) {
	test.Parallel()
	const overloads = "class Box {\n    each(limit: number): Generator<number, void, unknown>;\n    each(limit: string): Generator<number, void, unknown>;\n    *each(limit: number | string): Generator<number, void, unknown> { BODY }\n}\n"
	for _, probe := range []struct{ name, body, use string }{
		{"stepped", "yield 1;", "for (const value of new Box().each(1)) console.log(String(value));\n"},
		{"dropped", "return;", "new Box().each('a');\nconsole.log('control');\n"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, strings.ReplaceAll(overloads, "BODY", probe.body)+probe.use)
			var gap *NotYet
			if !errors.As(lowerSourceError, &gap) ||
				!strings.Contains(gap.What, "a generator function written as a method") {
				test.Fatalf("got %v, want a NotYet saying it is written as a method", lowerSourceError)
			}
		})
	}
}
