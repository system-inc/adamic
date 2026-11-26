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

// spreadPrototypeRefusal is the words a .a spread of built is refused with when class, which reaches it,
// holds member on its prototype as a kind declared on holder, while built's type declares member a field
// (checkSpreadPrototypeMembers).
func spreadPrototypeRefusal(class, member, kind, holder string) string {
	inherited := ""
	if holder != class {
		inherited = ", inherited from " + holder
	}
	return "Adamic 0.1 refuses a spread of built, whose type declares " + member + " a field, while class " + class + ", which may reach it, holds " + member + " as a " + kind + " on its prototype" + inherited + ": a spread copies own properties only, so the copy has no " + member + " though its type says it does; copy the fields you mean explicitly, or declare " + member + " a field of " + holder
}

// A spread copies own properties only, so a class instance's prototype getter isn't in the copy, while the
// checker types the copy with it when the source is seen as a type declaring it (#7mjgxkc). Main read the
// slot the copy doesn't have: c13b printed garbage natively where Node prints undefined (ASan silent), and
// c13c stopped on a compiler bug. By Compiler's ruling (B) each now stops at compile time at the read,
// naming the field, the spread and the class, in .ts. c13 reads it after an untagged predicate that skips
// weight, since its source declares it (ruling 11, #4cvdkk3): the predicate's own view takes weight into its
// contract, so a getter of that name reaching it stops first, as the tagged c13t does on main.
// written_after, the getter's name written after the spread, stops at the literal, since natively the
// copy has no slot for it (main stopped its read on the same compiler bug). own_field (the class's weight an
// own data field) and method (a class's method, which the checker drops from the spread's type) lower and agree with
// Node.
//
// In .a the spread itself is refused, by Adamic's amendment of ruling 91 (#vjvcnnj), when its source's type
// declares as a field a member a class reaching it holds on its prototype: c13b_a's getter, setter_a's setter,
// method_field_a's method seen as a function-typed field, and base_getter_a's getter on a base class, named
// through the subclass that reaches the spread. method_field_a reads only label, since a read of weigh through
// the view stops first, at the read, as a class method through a view that erases its prototype origin.
// own_field_a (an own data field of the name) and
// undeclared_getter_a (a getter the source's type doesn't declare) lower and agree with Node.
// Mutants: spreadCopyRead answering nil lowers c13b again (internal/lower/spread_copy_reads.mutants.md), and
// the .a refusal answering nil lowers the refused .a rows (internal/lower/spread_prototype_members.mutants.md).
func TestSpreadPrototypeField(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, stop, refused string }{
		{"c13.ts", "number\ntarget undefined undefined\n", "c13.ts:14:2: stage 0 can't lower a getter in a checked field contract yet", ""},
		{"c13b.ts", "number\nundefined undefined\n", spreadCopyReadStop, ""},
		{"c13c.ts", "undefined undefined\n", spreadCopyReadStop, ""},
		{"c13b_a.a", "number\nundefined undefined\n", "", spreadPrototypeRefusal("Sized", "weight", "getter", "Sized")},
		{"own_field.ts", "number\nnumber 20\n", "", ""},
		{"written_after.ts", "number\nnumber 21\n", "stage 0 can't lower writing weight after the spread in { ...built, weight: built.weight + 1 }, a spread's copy of a Sized instance, whose weight is a getter on its prototype: the copy has no slot for weight yet", ""},
		{"method.ts", "t\n", "", ""},
		{"setter_a.a", "undefined undefined\n", "", spreadPrototypeRefusal("Sized", "weight", "setter", "Sized")},
		{"method_field_a.a", "t\n", "", spreadPrototypeRefusal("Scale", "weigh", "method", "Scale")},
		{"base_getter_a.a", "undefined undefined\n", "", spreadPrototypeRefusal("Sized", "weight", "getter", "Base")},
		{"own_field_a.a", "number\nnumber 20\n", "", ""},
		{"undeclared_getter_a.a", "t\n", "", ""},
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
			if probe.refused != "" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.refused) || !strings.Contains(err.Error(), probe.name) {
					t.Fatalf("want a refusal with %q, got %v", probe.refused, err)
				}
				return
			}
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
