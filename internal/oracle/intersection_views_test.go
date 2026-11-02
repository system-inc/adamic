package oracle

import (
	"errors"
	"os"
	"path/filepath"
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
func TestIntersectionViewsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/intersection_views"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		programs++
		path := directory + "/" + entry.Name()
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, path)
		})
	}
	if programs != 3 {
		t.Fatalf("%d programs in %s, want 3: a program added or removed needs this count changed with it", programs, directory)
	}
}

// Where the object doesn't hold what the intersection refines a field to, the read stops with the
// view's message and exit 70, after the untagged call has printed what Node prints: a literal tag
// that holds another number, a boolean refined to true that holds false, and a field narrowed to a
// string that holds a number. A write stops where the slot holds what the view says it can't.
// Node, without the checks, prints on.
func TestIntersectionViewsStopAtTheRead(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/intersection_views_stops/"
	for _, stop := range []struct{ name, node, stdout, stderr string }{
		{"nested_literal.ts", "other\nnode:fs 12\n", "other\n", "field read failed: imported.argument.literal expected 11, found number 12"},
		{"nested_boolean.ts", "other\n4 false\n", "other\n", "field read failed: declaration.clause.isTypeOnly expected true, found boolean false"},
		{"nested_token.ts", "other\nnode:fs undefined\n", "other\n", "field read failed: imported.argument.token is not a string; expected string, found number"},
		{"nested_write.ts", "other\nstamped\nnumber\n", "other\n", "field read failed: <write>.flags is not a number; expected number, found string"},
	} {
		t.Run(stop.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, directory+stop.name))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); truth.exitCode != 0 || string(truth.stdout) != stop.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stdout: []byte(stop.stdout), stderr: []byte("adamic: panic: " + stop.stderr + "\n")}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stderr %q", backend, diff, got.stderr)
				}
			}
		})
	}
}

// The four stops are negative witnesses of the checked_boundary form (#qqvq6n7), listed in
// cloud/admission-corpus/negative-witnesses.json, and each one's repair, which stores what the view
// says the field holds, agrees with Node natively, sanitized and on JavaScript.
func TestIntersectionViewWitnessRepairsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"nested_literal", "nested_boolean", "nested_token", "nested_write"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loomViewsDAgrees(t, "internal/oracle/testdata/intersection_views_stops/"+name+"_repaired.ts")
		})
	}
}

// A checked view over an intersection admits a value the way a view over an interface does, so every road
// by which an interface view's value is seen as another type and read (#80cs7rz) must stop the same way
// when the cast's type is the intersection (Compiler's gate on #qqvq6n7): each of view_widened_test.go's
// roads whose Lit is interface Lit extends Node { ... }, rewritten as type Lit = Node & { ... }, is NotYet
// at the same read with the same words. The rewrite keeps every later line and column where they were.
// Mutants: checkViewWidenedReads returning nil at once admits every one of them; view_lazy.go's
// intersection contract dropped (the view's family unsupported again) stops each at its first read
// through Lit instead, with other words. Only the first is a hole; the second shows the cast is the
// intersection view this change builds.
func TestIntersectionViewWidenedReadNotYet(t *testing.T) {
	t.Parallel()
	const declared = "interface Lit extends Node { readonly kind: 9; readonly text: string; }"
	const intersected = "type Lit = Node & { readonly kind: 9; readonly text: string; };"
	for _, probe := range []struct{ file, at, through, node string }{
		{"view_widened.ts", "5:106", "HasText", "5\n"},
		{"view_widened_array.ts", "5:131", "HasText", "5\n"},
		{"view_widened_map.ts", "5:181", "HasText", "5\n"},
		{"view_widened_closure.ts", "5:118", "HasText", "5\n"},
		{"view_widened_parameter.ts", "5:46", "HasText", "5\n"},
		{"view_widened_return.ts", "6:90", "HasText", "5\n"},
		{"view_widened_field.ts", "6:112", "HasText", "5\n"},
		{"view_widened_destructure.ts", "5:102", "HasText", "5\n"},
		{"view_widened_template.ts", "5:111", "HasText", "t=5\n"},
		{"view_widened_conditional.ts", "5:166", "HasText", "5\n"},
		{"view_widened_class_field.ts", "5:84", "HasText", "5\n"},
		{"view_widened_satisfies.ts", "5:138", "HasText", "5\n"},
		{"view_widened_optional.ts", "5:144", "HasText", "5\n"},
		{"view_widened_anonymous.ts", "5:164", "{ readonly kind: 9; readonly pos: number; readonly text: string; }", "5\n"},
		{"view_widened_predicate.ts", "6:82", "HasText", "5\n"},
		{"view_widened_predicate_array.ts", "6:107", "HasText", "5\n"},
		{"view_widened_predicate_parameter.ts", "6:46", "HasText", "5\n"},
		{"view_widened_predicate_closure.ts", "6:94", "HasText", "5\n"},
		{"view_widened_predicate_field.ts", "7:88", "HasText", "5\n"},
		{"view_widened_predicate_anonymous.ts", "6:140", "{ readonly kind: 9; readonly pos: number; readonly text: string; }", "5\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(source), "\n")
			if len(lines) < 2 || lines[1] != declared {
				t.Fatalf("line 2 of %s isn't %q: the rewrite would test nothing", probe.file, declared)
			}
			lines[1] = intersected
			path := filepath.Join(t.TempDir(), probe.file)
			if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			what := "a read of field text through type " + probe.through + ", which may hold a checked view's value"
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.file+":"+probe.at) || !strings.Contains(err.Error(), what) {
				t.Fatalf("want the widened read NotYet at %s, got %v", probe.at, err)
			}
		})
	}
}
