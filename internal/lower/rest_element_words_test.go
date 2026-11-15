package lower

import (
	"errors"
	"strings"
	"testing"
)

// Mutant run and restored: remove valueSlotKeeps' rest-element diagnostic branch.
// All rows then print the old weak sentence, as recorded in rest-element-words.mutants.md.
func TestRestElementWords(t *testing.T) {
	t.Parallel()
	const collect = "const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + \":\" + String(item)).join(\",\");\n"
	for _, probe := range []struct{ name, source, where, view, position, restName string }{
		{"r12", collect + "const viewed: (...values: number[]) => string = collect;\nconsole.log(viewed(1, 2));\n", ":2:49:", "(...values: number[]) => string", "1", "values"},
		{"r13", collect + "const viewed: (a: string, ...values: number[]) => string = collect;\nconsole.log(viewed(\"x\", 2));\n", ":2:60:", "(a: string, ...values: number[]) => string", "2", "values"},
		{"named-rest", collect + "const viewed: (...rest: number[]) => string = collect;\nconsole.log(viewed(1, 2));\n", ":2:47:", "(...rest: number[]) => string", "1", "rest"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + " stage 0 can't lower a (...items: (string | number)[]) => string seen as a " + probe.view + ", whose rest's element string | number is held otherwise than the number the view hands it at parameter " + probe.position + " (give the view's rest that element, ..." + probe.restName + ": (string | number)[], or declare the function's parameters one by one) yet"
				var later *NotYet
				if !errors.As(err, &later) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want the complete NotYet ending %q", err, want)
				}
				t.Log(err)
			})
		}
	}
}

func TestRestElementWeakControl(t *testing.T) {
	t.Parallel()
	const source = "import type { Weak } from 'adamic';\ninterface Box {\n\treadonly label: string;\n}\nfunction run(n: number): void {\n\tconst box: Box = { label: `b${n}` };\n\tconst f: (x: Box) => boolean = (x: Weak<Box>): boolean => x === box;\n\tconsole.log(`${f(box)}`);\n}\nrun(1);\n"
	_, err := lowerSource(t, source)
	want := "main.a:7:33: stage 0 can't lower a (x: Weak<Box>) => boolean seen as a (x: Box) => boolean (one keeps something weakly that the other keeps strongly) yet"
	var later *NotYet
	if !errors.As(err, &later) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want the unchanged weak sentence ending %q", err, want)
	}
}

func TestRestElementViewFix(t *testing.T) {
	t.Parallel()
	const source = "const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + \":\" + String(item)).join(\",\");\nconst viewed: (...values: (string | number)[]) => string = collect;\nconsole.log(viewed(1, \"x\"));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// The source change is taken from the diagnostic, not reconstructed from its type.
func TestRestElementPrintedFixLowers(t *testing.T) {
	t.Parallel()
	const collect = "const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + \":\" + String(item)).join(\",\");\n"
	for _, probe := range []struct{ name, source string }{
		{"r12", collect + "const viewed: (...values: number[]) => string = collect;\nconsole.log(viewed(1, 2));\n"},
		{"r13", collect + "const viewed: (a: string, ...values: number[]) => string = collect;\nconsole.log(viewed(\"x\", 2));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var later *NotYet
				if !errors.As(err, &later) {
					t.Fatalf("want a NotYet carrying the fix, got %v", err)
				}
				_, tail, found := strings.Cut(later.What, " (give the view's rest that element, ")
				parameter, _, ended := strings.Cut(tail, ", or declare the function's parameters one by one)")
				if !found || !ended {
					t.Fatalf("no rest parameter fix in %q", later.What)
				}
				fixed := strings.Replace(probe.source, "...values: number[]", parameter, 1)
				if _, err := lowerSourceAs(t, fixed, extension); err != nil {
					t.Fatalf("printed fix %q does not lower: %v", parameter, err)
				}
			})
		}
	}
}
