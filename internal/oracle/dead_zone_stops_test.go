package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Two dead-zone roads stage 0 can't keep yet, each wrong on both backends where Node throws or
// assigns (#aw65psm), stop as not yet, in .a and .ts: a class used from a function that may run
// before the class's declaration (c01; Node throws "Cannot access 'Late' before initialization"),
// and a logical assignment to a module binding from code above its declaration (g01; Node gives 4 where
// both backends left it undefined). Their controls lower and agree with Node: a class used from a
// function when nothing before the class reaches it (c02), an object literal of literals before it (c03)
// or classes extending each other before it (c04), ??= on a function's own local (g02), and ||= on a
// module binding from below its declaration, tsc's deferredX ||= ... (g03).
// The class's own static parts run inside its dead zone, so a function they reach stops: from a static
// initializer (c05), a static block (c06), a computed base (c07), and a static block making another
// class whose constructor names this one (c08, which both backends lowered, native panicking where Node
// and the JavaScript backend catch). A reference inside the class's own static block lowers and agrees
// (c09). A ??= in a hoisted function written below its let stops when code before the let calls it
// (g04), and one written above still stops when it only runs after the let (g05, undefined on both
// backends where Node prints 4). A function exported by an export list after the class, under another
// name (e01, with e01_class_exported_by_list/back calling it), stops when the code before the class
// makes any call, as one with the export modifier does. An instance field's initializer runs when its
// class is made, so a class named there stops when that class is made before the named one's
// declaration (c10, where all three backends printed late where Node throws), and lowers and agrees
// when it's made only after (c11). A library call on primitives can't reach program code, so a class
// after `const knownUnits = '...'.split(' ')`, used from an exported function, lowers and agrees (c12,
// stage1/cohere/css/print.ts's shape); a library call taking a program callback before the class still
// stops, by name (c13, [1].map(make)) or through another module ('1'.replace('1', callBack) in e02, with
// e02_class_exported_library_callback/back calling the exported function).
// Language's rows on that proof, each with its back module calling the exported function from a value
// that only looks primitive or a library that only looks unpatched: a string-typed label laundered
// through a generic constraint, holding an object whose toString makes the class, read into String()
// (l01), split (l02) and concat (l03), stops here, since labelOf is a program call (on main, without
// this rule, all three lower and differ from Node on every backend: #xywvn06); a patched
// String.prototype.split is refused where it's patched, as a method read as a value (l04); and a
// Symbol.split splitter object stops here, an object argument (l05).
// Mutants: classDeadZoneGap returning nil lowers c01; the logical assignment's module-binding stop
// dropped lowers g01; mayRunBefore without the class's own running parts turns c05 to c08 red; the
// may-run-before logical assignment stop dropped lowers g04; the export read from the modifier alone
// lowers e01 and e02; holdingStatement not counting an instance field's initializer lowers c10; every
// library call counted again stops c12; every value taken as primitive lowers e02.
func TestDeadZoneStopsArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/dead_zone_stops"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q: Node prints %q", stop, onNode(t, absolute).stdout)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			truth := onNode(t, absolute)
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 50 {
		t.Fatalf("found %d programs, want 50", programs)
	}
}
