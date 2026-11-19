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

// A checked view over an intersection admits a value the way a view over an interface does, so every road by
// which a view's value is seen as another type and read (#80cs7rz) must stop the same way when the value is an
// intersection's (Compiler's gate on #qqvq6n7). The intersection is entered as nested.ts enters one, a field
// read out of an interface view: each cast road of view_widened_test.go is rewritten so Lit is
// Node & { readonly kind: 9; readonly text: string }, held in Outer's inner, and the cast n as Lit becomes
// (n as Outer).inner, which keeps every line and moves only what follows the cast on its line. #80cs7rz held
// each NotYet at the read; #zpzm9ta checks it where it's made, so each lowers, and its lie stops at the same read
// with the same located failure as view_widened_test.go's on sanitized native, release native and JavaScript, where
// Node prints the field. Mutants: checkViewWidenedReads returning nil at once, or readiness stripping a marked
// read's view, and every one reads the lie raw; intersectionViewFamily always "intersection" (the contract this
// change builds dropped) stops each at another place, so the intersection's own contract is what these reach.
func TestIntersectionViewWidenedReadIsChecked(t *testing.T) {
	t.Parallel()
	const declared = "interface Lit extends Node { readonly kind: 9; readonly text: string; }"
	const nested = "interface Outer extends Node { readonly kind: 9; readonly inner: Lit; } type Lit = Node & { readonly kind: 9; readonly text: string };"
	const lying = "const v = { kind: 9, pos: 1, text: 5 };"
	const lyingInner = "const v = { kind: 9, pos: 1, inner: { kind: 9, pos: 1, text: 5 } };"
	const cast, nestedCast = "n as Lit", "(n as Outer).inner"
	for _, probe := range []struct{ file, read, node string }{
		{"view_widened.ts", "h.text", "5\n"},
		{"view_widened_array.ts", "h.text", "5\n"},
		{"view_widened_map.ts", "h.text", "5\n"},
		{"view_widened_closure.ts", "h.text", "5\n"},
		{"view_widened_parameter.ts", "x.text", "5\n"},
		{"view_widened_return.ts", "h.text", "5\n"},
		{"view_widened_field.ts", "b.item.text", "5\n"},
		{"view_widened_destructure.ts", "h.text", "5\n"},
		{"view_widened_template.ts", "h.text", "t=5\n"},
		{"view_widened_conditional.ts", "h.text", "5\n"},
		{"view_widened_class_field.ts", "this.h.text", "5\n"},
		{"view_widened_satisfies.ts", "h2.text", "5\n"},
		{"view_widened_optional.ts", "h?.text", "5\n"},
		{"view_widened_anonymous.ts", "r.text", "5\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(source), "\n")
			if len(lines) < 4 || lines[1] != declared || lines[3] != lying {
				t.Fatalf("%s isn't the shape the rewrite expects: the rewrite would test nothing", probe.file)
			}
			lines[1], lines[3] = nested, lyingInner
			casts := 0
			for index, text := range lines {
				if !strings.Contains(text, cast) {
					continue
				}
				casts += strings.Count(text, cast)
				lines[index] = strings.Replace(text, cast, nestedCast, 1)
			}
			if casts != 1 {
				t.Fatalf("%s has %d casts to Lit, want 1", probe.file, casts)
			}
			path := filepath.Join(t.TempDir(), probe.file)
			if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, path); disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("want the widened read lowered as a checked read, got %v", err)
			}
			stopsAtFieldRead(t, program, "", "field read failed: "+probe.read+" is not a string; expected string, found number")
		})
	}
}

// An intersection can't be the entry itself: a cast from Node straight to Node & { ... } is no tagged member
// of a union, so it's refused as a cast the runtime can't check, and a predicate claiming one is NotYet as an
// overload target. Both fail closed before any view exists, so the roads above enter through a field.
func TestIntersectionDirectEntryFailsClosed(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, at, words string
		refused                 bool
	}{
		{"cast.ts", "interface Node { readonly kind: number; readonly pos: number; }\ntype Lit = Node & { readonly kind: 9; readonly text: string; };\nfunction show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } }\nshow({ kind: 9, pos: 1 });\n", "cast.ts:3:62", "a cast the runtime can't check", true},
		{"predicate.ts", "interface Node { readonly kind: number; readonly pos: number; }\ntype Lit = Node & { readonly kind: 9; readonly text: string; };\nfunction isLit(n: Node): n is Lit { return n.kind === 9; }\nfunction show(n: Node): void { if (isLit(n)) { console.log(n.text); } }\nshow({ kind: 9, pos: 1 });\n", "predicate.ts:3:1", "checked predicate overload target Lit", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name)
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			kind := errors.As(err, &notYet)
			if probe.refused {
				kind = errors.As(err, &refused)
			}
			if !kind || !strings.Contains(err.Error(), probe.at) || !strings.Contains(err.Error(), probe.words) {
				t.Fatalf("want %q at %s (refused %v), got %v", probe.words, probe.at, probe.refused, err)
			}
		})
	}
}
