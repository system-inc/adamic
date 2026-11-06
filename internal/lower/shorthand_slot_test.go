package lower

import (
	"errors"
	"strings"
	"testing"
)

// A shorthand { p } is p seen as the literal's property p, judged as { p: p } is, against the slot its name
// has. Before #20nf0sm the shorthand read its property off the literal's slot, and a slot that was a union
// holding undefined, a return type `{ p: W } | undefined`, has no property p, since undefined has none: the
// shorthand was judged against nothing, and a readonly R went out as a writable W, whose write changed r.t
// from 2 to 1 natively and on Node alike (CLI's probe, short2.ts, at main fb12ef08). The named form was
// refused, since its initializer reads the checker's contextual type for the property, which looks past
// undefined. Narrowing had nothing to do with it: the same return outside the if lowered too (a probe on
// Home, Nov 6). Each road a shorthand goes into a slot is pinned with a union slot, in .a and .ts: a
// return inside an if and outside one, an argument, an assignment, an array literal's element, and a
// union of two object types. A spread into the same slots was the same hole (Lower's neighbor): a union's
// own properties are only those every member has, so { ...h } was judged against none, and it's pinned
// beside a spread into a plain { p: W }, refused on main already. A generator's yield stops first at the generator's own NotYet on main, so
// it's pinned by that stop until generators lower. The honest twin, a W through the same return, lowers
// and agrees with Node, and so does an honest spread. Mutants: the shorthand case reading GetPropertyOfType
// off the literal's slot again, under which the shorthand roads with a union slot lower; and
// literalSlotMembers answering the whole slot, under which the two union spreads lower.
func TestShorthandIsJudgedAgainstItsPropertysSlot(t *testing.T) {
	t.Parallel()
	const refusal = "Adamic 0.1 refuses a value of type R seen as W, whose readonly field t becomes writable"
	for _, probe := range []struct{ name, source, where string }{
		{"return_narrowed", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(p: R): { p: W } | undefined { if (p.t > 0) { return { p }; } return undefined; }\nconst r: R = { t: 2 };\nconst x = make(r);\nif (x) { x.p.t = 1; }\nconsole.log(String(r.t));\n", ":3:69"},
		{"return", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(p: R): { p: W } | undefined { return { p }; }\nconst x = make({ t: 2 });\nif (x) { x.p.t = 1; }\nconsole.log(String(x !== undefined));\n", ":3:54"},
		{"argument", "interface R { readonly t: number }\ninterface W { t: number }\nfunction take(o: { p: W } | undefined): void { if (o) { o.p.t = 1; } }\nfunction run(p: R): void { take({ p }); }\nrun({ t: 2 });\nconsole.log('done');\n", ":4:35"},
		{"assignment", "interface R { readonly t: number }\ninterface W { t: number }\nconst p: R = { t: 2 };\nlet o: { p: W } | undefined = undefined;\no = { p };\nif (o) { o.p.t = 1; }\nconsole.log(String(p.t));\n", ":5:7"},
		{"array_element", "interface R { readonly t: number }\ninterface W { t: number }\nconst p: R = { t: 2 };\nconst xs: ({ p: W } | undefined)[] = [{ p }];\nfor (const o of xs) { if (o) { o.p.t = 1; } }\nconsole.log(String(p.t));\n", ":4:41"},
		{"object_union", "interface R { readonly t: number }\ninterface W { t: number }\nconst p: R = { t: 2 };\nconst o: { p: W } | { q: number } = { p };\nif ('p' in o) { o.p.t = 1; }\nconsole.log(String(p.t));\n", ":4:39"},
		{"spread", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(h: { p: R }): { p: W } | undefined { return { ...h }; }\nconst r: R = { t: 2 };\nconst x = make({ p: r });\nif (x) { x.p.t = 1; }\nconsole.log(String(r.t));\n", ":3:61"},
		{"spread_object_union", "interface R { readonly t: number }\ninterface W { t: number }\nconst r: R = { t: 2 };\nconst h = { p: r };\nconst o: { p: W } | { q: number } = { ...h };\nif ('p' in o) { o.p.t = 1; }\nconsole.log(String(r.t));\n", ":5:39"},
		{"spread_no_union", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(h: { p: R }): { p: W } { return { ...h }; }\nconst r: R = { t: 2 };\nmake({ p: r }).p.t = 1;\nconsole.log(String(r.t));\n", ":3:49"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(refused.Where, probe.where) || !strings.HasPrefix(refused.What, refusal) || !strings.Contains(refused.What, "(adamic/invariant-mutable)") {
					t.Fatalf("want %q refused at %s, got %v", refusal, probe.where, err)
				}
			})
		}
	}
	t.Run("yield", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSourceAs(t, "interface R { readonly t: number }\ninterface W { t: number }\nfunction* each(p: R): Generator<{ p: W } | undefined> { yield { p }; }\nfor (const o of each({ t: 2 })) { if (o) { o.p.t = 1; } }\nconsole.log('done');\n", ".ts")
		if err == nil {
			t.Fatal("want the generator stopped, got it lowered")
		}
		t.Logf("yield: %v", err)
	})
	t.Run("honest_spread", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(h: { p: W }): { p: W } | undefined { return { ...h }; }\nconst w: W = { t: 2 };\nconst x = make({ p: w });\nif (x) { x.p.t = 1; }\nconsole.log(String(w.t));\n", ".ts")
	})
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(p: W): { p: W } | undefined { if (p.t > 0) { return { p }; } return undefined; }\nconst w: W = { t: 2 };\nconst x = make(w);\nif (x) { x.p.t = 1; }\nconsole.log(String(w.t));\n", ".ts")
	})
}
