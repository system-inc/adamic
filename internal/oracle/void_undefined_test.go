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
// back, and control.a is the B | undefined | null spelling that was always right. Each is held to Node
// and to its pinned output on the release, sanitized, JavaScript and, where a sysroot is set, wasm32
// legs, in .a and .ts alike. Mutant (void-undefined.mutants.md): includesUndefined's void exclusion
// restored, which reddens witness, alias and field and leaves control green.
func TestVoidMemberCountsAsUndefined(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, want string }{
		{"witness", "true false true object\nfalse true true undefined\nfalse false false object\n"},
		{"alias", "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"},
		{"field", "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"},
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
// refuses (#x2gt39p), as on main. make returns 42 through a void type, so Node prints false, and the
// read is refused before either backend could print true.
func TestVoidResultReadStillRefused(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("void_read"+extension, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/void_undefined/void_read"+extension))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "false\n" || len(truth.stderr) != 0 {
				t.Fatalf("Node: exit %d stdout %q stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "a read of a void value, as an operand of ===" || !strings.HasPrefix(refused.Fix, "call it as a statement, or give the function a return type") {
				t.Fatalf("want H2's void read refused as an operand of ===, got %v", err)
			}
		})
	}
}
