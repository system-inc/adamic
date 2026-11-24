package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// spreadCopyReadStop is the words a read of weight from c13's spread copy stops with (spreadCopyRead).
const spreadCopyReadStop = "stage 0 can't lower a read of weight from { ...built }, a spread's copy of a Sized instance, whose weight is a getter on its prototype: a spread copies own properties only, so the copy has no weight though its type says it does yet"

// A spread copies own properties only, so a class instance's prototype getter isn't in the copy, while the
// checker types the copy with it when the source is seen as a type declaring it (#7mjgxkc). Main read the
// slot the copy doesn't have: c13b printed garbage natively where Node prints undefined (ASan silent), and
// c13c stopped on a compiler bug. By Compiler's ruling (B) each now stops at compile time at the read,
// naming the field, the spread and the class, in .ts and .a; c13 stops earlier, at its predicate, as on main.
// written_after, the getter's name written after the spread, stops at the literal, since natively the
// copy has no slot for it (main stopped its read on the same compiler bug). own_field (the class's weight an
// own data field) and method (a class's method, which the checker drops from the spread's type) lower and agree with
// Node. Mutant (internal/lower/spread_copy_reads.mutants.md): spreadCopyRead answering nil lowers c13b again.
func TestSpreadPrototypeField(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, stop string }{
		{"c13.ts", "number\ntarget undefined undefined\n", "c13.ts:9:1: stage 0 can't lower checked predicate overload target Target yet"},
		{"c13b.ts", "number\nundefined undefined\n", spreadCopyReadStop},
		{"c13c.ts", "undefined undefined\n", spreadCopyReadStop},
		{"c13b_a.a", "number\nundefined undefined\n", spreadCopyReadStop},
		{"own_field.ts", "number\nnumber 20\n", ""},
		{"written_after.ts", "number\nnumber 21\n", "stage 0 can't lower writing weight after the spread in { ...built, weight: built.weight + 1 }, a spread's copy of a Sized instance, whose weight is a getter on its prototype: the copy has no slot for weight yet"},
		{"method.ts", "t\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/spread_prototype", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.node)}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if probe.stop != "" {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.stop) || !strings.Contains(err.Error(), probe.name) {
					t.Fatalf("want a not yet with %q, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
