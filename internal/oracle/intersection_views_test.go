package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A checked view over an intersection of plain object types (#qqvq6n7) is described by the
// intersection's resolved properties, each with its refined type: a boolean literal refinement, a
// base field, a field narrowed from string | number to string, an added field, a nested
// intersection read field by field, and a write through it that reads back. Each agrees with Node
// natively, sanitized and on the JavaScript backend. The nested ones cast through an interface that
// extends its source with no tag of its own, the downcast .a refuses (#995z2fa), so each is .ts.
func TestIntersectionViewsAgreeWithNode(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/intersection_views"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		path := directory + "/" + entry.Name()
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			loomViewsDAgrees(test, path)
		})
	}
	if programs != 3 {
		test.Fatalf(
			"%d programs in %s, want 3: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// Where the object doesn't hold what the intersection refines a field to, the read stops with the
// view's message and exit 70, after the untagged call has printed what Node prints: a literal tag
// that holds another number, a boolean refined to true that holds false, and a field narrowed to a
// string that holds a number. A write stops where the slot holds what the view says it can't.
// Node, without the checks, prints on.
func TestIntersectionViewsStopAtTheRead(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/intersection_views_stops/"
	for _, stop := range []struct{ name, node, stdout, stderr string }{
		{
			"nested_literal.ts",
			"other\nnode:fs 12\n",
			"other\n",
			"field read failed: imported.argument.literal expected 11, found number 12",
		},
		{
			"nested_boolean.ts",
			"other\n4 false\n",
			"other\n",
			"field read failed: declaration.clause.isTypeOnly expected true, found boolean false",
		},
		{
			"nested_token.ts",
			"other\nnode:fs undefined\n",
			"other\n",
			"field read failed: imported.argument.token is not a string; expected string, found number",
		},
		{
			"nested_write.ts",
			"other\nstamped\nnumber\n",
			"other\n",
			"field read failed: <write>.flags is not a number; expected number, found string",
		},
	} {
		test.Run(stop.name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, directory+stop.name))
			if err != nil {
				test.Fatal(err)
			}
			if truth := onNode(test, path); truth.exitCode != 0 || string(truth.stdout) != stop.node {
				test.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			want := run{
				exitCode: 70,
				stdout:   []byte(stop.stdout),
				stderr:   []byte("adamic: panic: " + stop.stderr + "\n"),
			}
			sanitized, _ := nativelyUncached(test, program)
			for backend, got := range map[string]run{
				"native":     releasedUncached(test, program),
				"sanitized":  sanitized,
				"javascript": onJavaScriptBackend(test, program),
			} {
				if diff := disagreement(want, got); diff != "" {
					test.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
				}
			}
		})
	}
}

// The four stops are negative witnesses of the checked_boundary form (#qqvq6n7), listed in
// cloud/admission-corpus/negative-witnesses.json, and each one's repair, which stores what the view
// says the field holds, agrees with Node natively, sanitized and on JavaScript.
func TestIntersectionViewWitnessRepairsAgreeWithNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"nested_literal", "nested_boolean", "nested_token", "nested_write"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			loomViewsDAgrees(test, "internal/oracle/testdata/intersection_views_stops/"+name+"_repaired.ts")
		})
	}
}

// A checked view over an intersection admits a value the way a view over an interface does, so every road by
// which a view's value is seen as another type and read (#80cs7rz) must stop the same way when the value is an
// intersection's (Compiler's gate on #qqvq6n7). The intersection is entered as nested.ts enters one, a field
// read out of an interface view: each cast road of view_widened_test.go is rewritten so Lit is
// Node & { readonly kind: 9; readonly text: string }, held in Outer's inner, and the cast n as Lit becomes
// (n as Outer).inner, which keeps every line and moves only what follows the cast on its line. Each is NotYet
// at the same read with the same words. Mutants: checkViewWidenedReads returning nil at once admits every one;
// intersectionViewFamily always "intersection" (the contract this change builds dropped) stops each at another
// place, so the intersection's own contract is what these reach.
func TestIntersectionViewWidenedReadNotYet(test *testing.T) {
	test.Parallel()
	const declared = "interface Lit extends Node { readonly kind: 9; readonly text: string; }"
	const nested = "interface Outer extends Node { readonly kind: 9; readonly inner: Lit; } type Lit = Node & { readonly kind: 9; readonly text: string };"
	const lying = "const v = { kind: 9, pos: 1, text: 5 };"
	const lyingInner = "const v = { kind: 9, pos: 1, inner: { kind: 9, pos: 1, text: 5 } };"
	const cast, nestedCast = "n as Lit", "(n as Outer).inner"
	for _, probe := range []struct {
		file, through, node string
		line, column        int
	}{
		{"view_widened.ts", "HasText", "5\n", 5, 106},
		{"view_widened_array.ts", "HasText", "5\n", 5, 131},
		{"view_widened_map.ts", "HasText", "5\n", 5, 181},
		{"view_widened_closure.ts", "HasText", "5\n", 5, 118},
		{"view_widened_parameter.ts", "HasText", "5\n", 5, 46},
		{"view_widened_return.ts", "HasText", "5\n", 6, 90},
		{"view_widened_field.ts", "HasText", "5\n", 6, 112},
		{"view_widened_destructure.ts", "HasText", "5\n", 5, 102},
		{"view_widened_template.ts", "HasText", "t=5\n", 5, 111},
		{"view_widened_conditional.ts", "HasText", "5\n", 5, 166},
		{"view_widened_class_field.ts", "HasText", "5\n", 5, 84},
		{"view_widened_satisfies.ts", "HasText", "5\n", 5, 138},
		{"view_widened_optional.ts", "HasText", "5\n", 5, 144},
		{
			"view_widened_anonymous.ts",
			"{ readonly kind: 9; readonly pos: number; readonly text: string; }",
			"5\n",
			5,
			164,
		},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				test.Fatal(err)
			}
			lines := strings.Split(string(source), "\n")
			if len(lines) < 4 || lines[1] != declared || lines[3] != lying {
				test.Fatalf("%s isn't the shape the rewrite expects: the rewrite would test nothing", probe.file)
			}
			lines[1], lines[3] = nested, lyingInner
			line, column := probe.line, probe.column
			casts := 0
			for index, text := range lines {
				at := strings.Index(text, cast)
				if at < 0 {
					continue
				}
				casts += strings.Count(text, cast)
				lines[index] = strings.Replace(text, cast, nestedCast, 1)
				if index+1 == line && at+1 < column {
					column += len(nestedCast) - len(cast)
				}
			}
			if casts != 1 {
				test.Fatalf("%s has %d casts to Lit, want 1", probe.file, casts)
			}
			path := filepath.Join(test.TempDir(), probe.file)
			if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				test.Fatal(err)
			}
			if truth := onNode(test, path); disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(test, path)
			at := probe.file + ":" + strconv.Itoa(line) + ":" + strconv.Itoa(column)
			what := "a read of field text through type " + probe.through + ", which may hold a checked view's value"
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), at) || !strings.Contains(err.Error(), what) {
				test.Fatalf("want the widened read NotYet at %s, got %v", at, err)
			}
		})
	}
}

// An intersection can't be the entry itself: a cast from Node straight to Node & { ... } is no tagged member
// of a union, so it's refused as a cast the runtime can't check, and a predicate claiming one is NotYet as an
// overload target. Both fail closed before any view exists, so the roads above enter through a field.
func TestIntersectionDirectEntryFailsClosed(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, source, at, words string
		refused                 bool
	}{
		{
			"cast.ts",
			"interface Node { readonly kind: number; readonly pos: number; }\ntype Lit = Node & { readonly kind: 9; readonly text: string; };\nfunction show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } }\nshow({ kind: 9, pos: 1 });\n",
			"cast.ts:3:62",
			"a cast the runtime can't check",
			true,
		},
		{
			"predicate.ts",
			"interface Node { readonly kind: number; readonly pos: number; }\ntype Lit = Node & { readonly kind: 9; readonly text: string; };\nfunction isLit(n: Node): n is Lit { return n.kind === 9; }\nfunction show(n: Node): void { if (isLit(n)) { console.log(n.text); } }\nshow({ kind: 9, pos: 1 });\n",
			"predicate.ts:3:1",
			"checked predicate overload target Lit",
			false,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), probe.name)
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				test.Fatal(err)
			}
			_, err := lowered(test, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			kind := errors.As(err, &notYet)
			if probe.refused {
				kind = errors.As(err, &refused)
			}
			if !kind || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), probe.words) {
				test.Fatalf("want %q at %s (refused %v), got %v", probe.words, probe.at, probe.refused, err)
			}
		})
	}
}
