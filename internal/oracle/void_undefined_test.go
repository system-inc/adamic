package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/wasitest"
)

// A reference union written with void for its undefined, B | null | void, holds null and undefined
// apart, as B | null | undefined does (#7yt238q). It was held as one pointer, so undefined read as null
// natively and value === undefined folded to false on both backends, since includesUndefined didn't
// count a void member. witness.a is Library's witness as the task gives it, alias.a spells it
// (B | null) | void and through a named alias, field.a writes a field of the type each way and reads it
// back, field_alias.a is the same field typed M = (B | null) | void, nullish.a reads the type through ??
// and ?. holding undefined, null and an object, and control.a is the B | undefined | null spelling that
// was always right. Each is held to Node and to its pinned output on the release, sanitized, JavaScript
// and, where a sysroot is set, wasm32 legs, in .a and .ts alike. Mutant (void-undefined.mutants.md):
// includesUndefined's void exclusion restored, which reddens witness, alias, field and field_alias.
// nullish and control stay green under it: ?? and ?. read null and undefined alike, so one pointer for
// both never misread them, and nullish pins that the boxed road reads them as Node does.
func TestVoidMemberCountsAsUndefined(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, want string }{
		{"witness", "true false true object\nfalse true true undefined\nfalse false false object\n"},
		{"alias", "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"},
		{"field", "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"},
		{"field_alias", "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"},
		{"nullish", "9 undefined -1\n9 undefined -1\n1 1 1\n9 undefined\n9 undefined\n2 2\n"},
		{"control", "true false true object\nfalse true true undefined\nfalse false false object\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/void_undefined", row.name+extension))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != row.want || len(truth.stderr) != 0 {
					t.Fatalf("Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatalf("Lower: %v", err)
				}
				sanitized, binary := natively(t, program)
				legs := map[string]run{"release": released(t, program), "sanitized": sanitized, "JavaScript": onJavaScriptBackend(t, program)}
				if _, runs := wasitest.Sysroot(t, "ADAMIC_ORACLE_WASI", "void_undefined/"+row.name+extension+" on wasm32-wasi"); runs {
					legs["wasm32"] = onWASI(t, native.C(program))
				}
				for leg, got := range legs {
					if difference := disagreement(truth, got); difference != "" {
						t.Errorf("%s: %s: exit %d, stdout %q, stderr %q; Node %q", leg, difference, got.exitCode, got.stdout, got.stderr, truth.stdout)
					}
				}
				if leaked := leaks(t, program, binary); leaked != "" {
					t.Errorf("leaks: %s", leaked)
				}
			})
		}
	}
}

// The void-return road didn't move: a value typed exactly void is still a discarded result whose read H2
// refuses (#x2gt39p), as on main, so a void-origin value never reaches a B | null | void slot to be boxed
// as undefined (Adamic's ruling on #7yt238q, :321 and :322). void_read.a reads one through ===: make
// returns 42 through a void type, so Node prints false, and the read is refused before either backend
// could print true. void_argument.a, void_binding.a and void_field.a hand a void function's result to a
// B | null | void parameter, binding and field. Each refusal is held to the words main gives it, in .a and .ts.
func TestVoidResultReadStillRefused(t *testing.T) {
	t.Parallel()
	// void_field's literal meets the void widening rule (adamic/no-void-widening) first, which refuses the
	// void value at the initializer as main does, before H2's read. Either refusal keeps a void-origin value
	// out of the slot, which is the condition, so each row pins the refusal main gives it, measured.
	const h2Fix = "call it as a statement, or give the function a return type"
	for _, row := range []struct{ name, stdout, what, fix string }{
		{"void_read", "false\n", "a read of a void value, as an operand of ===", h2Fix},
		{"void_argument", "undefined\n", "a read of a void value, as an argument", h2Fix},
		{"void_binding", "undefined\n", "a read of a void value, as a binding's initializer", h2Fix},
		{"void_field", "undefined\n", "a void value at the initializer site through .value seen as void | B | null", "declare the function's real return type"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/void_undefined", row.name+extension))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != row.stdout || len(truth.stderr) != 0 {
					t.Fatalf("Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				_, err = lowered(t, path)
				var refused *lower.Refused
				if !errors.As(err, &refused) || refused.What != row.what || !strings.HasPrefix(refused.Fix, row.fix) {
					t.Fatalf("want %q refused (fix %q), got %v", row.what, row.fix, err)
				}
			})
		}
	}
}
