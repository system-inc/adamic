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

// spreadPrototypeRefusal is the words a .a spread of built, held as Source, is refused with when class,
// which reaches it, holds member on its prototype as a kind declared on holder, while Source declares member
// a field (checkSpreadPrototypeMembers).
func spreadPrototypeRefusal(class, member, kind, holder string) string {
	return spreadPrototypeRefusalOf("built", heldAs("Source", member, "a field"), class, member, kind, holder)
}

// spreadPrototypeRefusalOf is the words a .a spread of source is refused with when the type its copy is held
// as claims member as claim says, after "a spread of <source>, ".
func spreadPrototypeRefusalOf(source, claim, class, member, kind, holder string) string {
	inherited := ""
	if holder != class {
		inherited = ", inherited from " + holder
	}
	return "Adamic 0.1 refuses a spread of " + source + ", " + claim + ", while class " + class + ", which may reach it, holds " + member + " as a " + kind + " on its prototype" + inherited + ": a spread copies own properties only, so the copy has no " + member + " though its type says it does; copy the fields you mean explicitly, or declare " + member + " a field of " + holder
}

// heldAs is the claim of a copy held as its literal's contextual type, held, declaring member in form.
func heldAs(held, member, form string) string {
	return "whose copy is held as " + held + ", which declares " + member + " " + form
}

// heldArm is the claim of a copy held as a union, held, whose arm declares member in form.
func heldArm(held, arm, member, form string) string {
	return "whose copy is held as " + held + ", whose arm " + arm + " declares " + member + " " + form
}

// ownClaim is the claim of a copy held as its literal's own type, declaring member in form.
func ownClaim(member, form string) string {
	return "whose copy's own type declares " + member + " " + form
}

// genericClaim is how a spread of a T types its copy: as what T stands for, every member of it claimed.
func genericClaim(class, member string) string {
	return "whose copy's own type is T, which stands for " + class + " here, so the copy is typed with " + member
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
//
// Diagnostics' review of head 2 (#vjvcnnj) found two fail-opens and an over-refusal, each a row here. A
// spread of a T types its copy as T, so every member of the class T stands for is claimed, a method and an
// accessor included: generic_method_call_a printed 20 natively where Node throws, generic_method_typeof_a
// said function where Node says undefined, and generic_setter_a stopped on the compiler bug; each is refused
// now, as is generic_getter_a, which stopped at the read. A union view's names are each member's, not only
// those every member has: union_setter_a stopped on the compiler bug and union_getter_a at the read, and
// both are refused. optional_setter_a's view is Source | undefined, whose copy the checker makes partial, so
// it claims no name present and lowers and agrees. over_refusal_a's only producer is Plain, whose fields
// are own, so a closed world holding Sized's getter no longer refuses it, and its read stops as at the base,
// since spreadCopyRead judges the copy by its type alone. literal_with_class_a's producer is
// an object literal whose getter is its own, which a spread copies, so the rule passes it and the read stops
// as c13b's does, while literal_accessor_a reads only label from such a copy and lowers and agrees.
//
// By Adamic's ruling (1) on #vjvcnnj every form a view declares a member in counts, since a spread copies
// only own enumerable data and every prototype member the copy's type still claims is a lie about the copy:
// method_signature_a's weigh(): number, called on the copy, panicked uncatchably on both backends where Node
// throws, and method_signature_label_a, which reads only label, is refused by the same rule; interface_getter_a's
// get weight() signature stopped at the read. Each is refused.
//
// By her ruling (1b) what is judged is the type the copy is held as: the literal's contextual type where the
// expression's type is that context (a declared slot, a parameter, a return), else its own type, and a
// generic or unresolved context falls back to the own type. tree_peek_a's copy has no context, so its own
// type claims peek and it's refused, as review/agree's fxspptb_aa17d3c_spread_method, whose copy is returned
// as Plain, lowers. generic_partial_a's return Partial<T> mentions T, so it falls back to the own type, T,
// which claims every member of Real. union_held_a is held as Plain | Tree, refused on its Tree arm. Language's
// s01 to s05 attack the premise (#vjvcnnj): s01's satisfies keeps the literal's own type, which claims peek,
// so it's refused; s02's overload passes the copy as Plain and reads peek only through in; s03 destructures a
// copy held as Plain; s04 widens a Plain to a type with an optional peek, which the optional-widening rule
// refuses; and s05's callback returns into map's U, a type parameter, so it falls back to the own type and is
// refused, an over-refusal that fails closed, held so a later relaxation sees it.
// Mutants: spreadCopyRead answering nil lowers c13b again (internal/lower/spread_copy_reads.mutants.md); the
// .a refusal answering nil lowers the refused .a rows, the type parameter's arm answering false lowers the
// generic rows, the union walk dropped lowers the union rows, prototypeFree answering false refuses
// over_refusal_a, the accessor literal's clause dropped refuses literal_with_class_a and literal_accessor_a,
// the signature forms dropped lowers the rows whose claim is a method or getter signature, the contextual type
// judged even when generic lowers generic_partial_a and s05_a, and the satisfies target judged lowers s01_a
// (internal/lower/spread_prototype_members.mutants.md).
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
		{"method_field_a.a", "t\n", "", spreadPrototypeRefusalOf("built", heldAs("Weighs", "weigh", "a field"), "Scale", "weigh", "method", "Scale")},
		{"base_getter_a.a", "undefined undefined\n", "", spreadPrototypeRefusal("Sized", "weight", "getter", "Base")},
		{"own_field_a.a", "number\nnumber 20\n", "", ""},
		{"undeclared_getter_a.a", "t\n", "", ""},
		{"generic_method_call_a.a", "threw\n", "", spreadPrototypeRefusalOf("source", genericClaim("Scale", "weigh"), "Scale", "weigh", "method", "Scale")},
		{"generic_method_typeof_a.a", "t undefined\n", "", spreadPrototypeRefusalOf("source", genericClaim("Scale", "weigh"), "Scale", "weigh", "method", "Scale")},
		{"generic_setter_a.a", "undefined undefined\n", "", spreadPrototypeRefusalOf("source", genericClaim("Sized", "weight"), "Sized", "weight", "setter", "Sized")},
		{"generic_getter_a.a", "undefined undefined\n", "", spreadPrototypeRefusalOf("source", genericClaim("Sized", "weight"), "Sized", "weight", "getter", "Sized")},
		{"generic_partial_a.a", "undefined\n", "", spreadPrototypeRefusalOf("x", genericClaim("Real", "peek"), "Real", "peek", "method", "Real")},
		{"union_setter_a.a", "undefined undefined\n", "", spreadPrototypeRefusalOf("source", heldArm("Heavy | Light", "Heavy", "weight", "a field"), "Sized", "weight", "setter", "Sized")},
		{"union_getter_a.a", "undefined undefined\n", "", spreadPrototypeRefusalOf("source", heldArm("Heavy | Light", "Heavy", "weight", "a field"), "Sized", "weight", "getter", "Sized")},
		{"union_held_a.a", "none\n", "", spreadPrototypeRefusalOf("tree", heldArm("Plain | Tree", "Tree", "peek", "a method"), "Real", "peek", "method", "Real")},
		{"optional_setter_a.a", "undefined undefined\n", "", ""},
		{"over_refusal_a.a", "20\nnumber 2\n", spreadCopyReadStop, ""},
		{"literal_with_class_a.a", "20\nnumber 7\n", spreadCopyReadStop, ""},
		{"literal_accessor_a.a", "20\nl\n", "", ""},
		{"method_signature_a.a", "threw\n", "", spreadPrototypeRefusalOf("built", heldAs("Weighs", "weigh", "a method"), "Scale", "weigh", "method", "Scale")},
		{"method_signature_label_a.a", "t\n", "", spreadPrototypeRefusalOf("built", heldAs("Weighs", "weigh", "a method"), "Scale", "weigh", "method", "Scale")},
		{"interface_getter_a.a", "undefined undefined\n", "", spreadPrototypeRefusalOf("built", heldAs("Source", "weight", "a getter"), "Sized", "weight", "getter", "Sized")},
		{"tree_peek_a.a", "threw\n", "", spreadPrototypeRefusalOf("tree", ownClaim("peek", "a method"), "Real", "peek", "method", "Real")},
		{"s01_a.a", "undefined\n", "", spreadPrototypeRefusalOf("tree", ownClaim("peek", "a method"), "Real", "peek", "method", "Real")},
		{"s02_a.a", "no peek\n", "", ""},
		{"s03_a.a", "tag\n", "", ""},
		{"s04_a.a", "undefined\n", "", ""},
		{"s05_a.a", "1\n", "", spreadPrototypeRefusalOf("t", ownClaim("peek", "a method"), "Real", "peek", "method", "Real")},
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
