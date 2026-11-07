package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A getter, a setter or a method in an object literal is part of its property, seen as the slot's property,
// and is judged against it (#wjqtypf). Before, nothing judged any of them, since none is an expression:
// `{ get p(): R { return r; } }` returned as `{ p: W }` let `make().p.t = 1` write r.t, which R says is
// readonly, and Node and native both printed 1. p01 to p05 are Language's getter probes (a plain return
// slot, `{ p: W } | undefined`, a const, an inferred return, an argument), and p06 a method `{ get(): R }`
// returned as `{ get(): W }`; each is refused, as { p: r } is, and so is m01, the method into a union it
// fits no one member of but tsc accepts by splitting its discriminant (#20nf0sm's shape). t01 to t06 are
// their honest twins, W's own field returned. A getter and a method are judged at each return statement's value, not their declared
// result, so a literal made right there stays fresh (Lower's ruling on 55507f73): f01, a getter declared
// `{ readonly t: number }` returning `{ t: 1 }` into `{ p: W }`, lowers, since every read is a new object
// no other name holds, and a write through W changes nothing a later read sees. A getter judges what the
// property reads as, and a setter is judged on its own, by what it
// takes, since tsc relates only the getter's type (Compiler's must on 0ee9e50b): d01, `{ get p(): Dog, set
// p(v: Dog) }` written into a `{ p: Animal }` slot, is refused, and its twin t07, taking Animal, lowers. d02,
// the same literal held by a const first and then seen as `{ p: Animal }`, was refused already. A setter
// alone is not yet wherever a slot names its property (s01), since the property reads as undefined there,
// unless what it takes is refused first (s02). g01, a getter alone seen as a writable `{ p: W }` and
// written, throws a TypeError on Node, strict as a module is; stage 0 stops it as not yet at the write, a
// property name whose descriptor has no setter, and its stop is pinned. n01, an arrow
// returning R seen as one returning W, and n02, a class instance whose method returns R seen through
// `{ get(): W }`, are the neighbours, refused already. m02, a generic method `{ get<T>(x: T): T }` seen as
// `{ get(x: R): W }`, is refused by its result as the slot instantiates it, T := R, since its return
// statements are typed by T (Lower's must on 29bc3e33). Main stops it at stage 0's "a function returning T" gap, as
// it does its twin t08, a writable V in R's place, which this head also lets through to that gap, so no generic twin
// can lower yet. u01, a literal method `log(x: unknown): void` written into a generic slot `{ log<U>(x: U): void }`,
// which widened can't relate, is not yet, held until every module has lowered so a more specific stop (a
// thenable's) wins where one fires; u02 is the same literal never called. Each program's whole stop is pinned in its .stop
// file, or "lowers", which runs it on release, sanitized native and the JavaScript backend against Node.
// Mutants, each measured on Pupa at 6a579777, each red row in .a and .ts: (a) the fail-closed clause dropped
// turns no program here red, since every kind a literal holds is judged, and TestUnlistedLiteralMemberKindFailsClosed;
// (b) the getter case removed, p01 to p05; (c) the method case removed, p06, m01, m02, u01 and u02; (d) the setter
// refusal answering nil, d01, s01 and s02; (e) the generic slot's not yet dropped, u01 and u02 lowering, and the three
// TestAsync* thenables refused; (f) a getter judged by its declared result, f01; (m) a method judged by its whole
// result, TestIteratorGapsAreExplicit; (gm) a generic method's return statements judged again, m02; (u) the generic
// and overloaded slots' not yet never held, u01 and u02 lowering; (i) the indexed slot's deferral dropped,
// TestNamedRecordWritesAreJudgedAsTheirDirectTwins/w10_computed_method; (n) methodSlotMembers' fallback dropped,
// TestMethodSlotMembersFallBackToEveryCallable; (o) literalSlotMembers' fallback dropped, m01 and
// spread_discriminant_split; (n and o) together, the rows of both.
func TestLiteralMemberSlots(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/literal_member_slot"
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
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
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
	if programs != 52 {
		t.Fatalf("found %d programs, want 52", programs)
	}
}
