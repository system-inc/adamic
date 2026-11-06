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
// union of two object types, and a generator's yield (.ts, since .a refuses a generator written that way
// first). A spread into the same slots was the same hole (Lower's neighbor): a union's
// own properties are only those every member has, so { ...h } was judged against none, and it's pinned
// beside a spread into a plain { p: W }, refused on main already. The honest twin, a W through the same return, lowers
// and agrees with Node, and so does an honest spread. Mutants: the shorthand case reading GetPropertyOfType
// off the literal's slot again, under which the shorthand roads with a union slot lower; and
// literalSlotMembers answering the whole slot, under which the two union spreads lower; and its
// empty filter answered as empty, under which a spread into a union it fits only by splitting its
// discriminant (spread_discriminant_split, Compiler's road on d2a29ebf) lowers. Its shorthand twin is
// judged against the name's own slot, which reaches W through either member.
func TestShorthandIsJudgedAgainstItsPropertysSlot(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, where, seen, only string }{
		{"return_narrowed", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(p: R): { p: W } | undefined { if (p.t > 0) { return { p }; } return undefined; }\nconst r: R = { t: 2 };\nconst x = make(r);\nif (x) { x.p.t = 1; }\nconsole.log(String(r.t));\n", ":3:69", "R seen as W", ""},
		{"return", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(p: R): { p: W } | undefined { return { p }; }\nconst x = make({ t: 2 });\nif (x) { x.p.t = 1; }\nconsole.log(String(x !== undefined));\n", ":3:54", "R seen as W", ""},
		{"argument", "interface R { readonly t: number }\ninterface W { t: number }\nfunction take(o: { p: W } | undefined): void { if (o) { o.p.t = 1; } }\nfunction run(p: R): void { take({ p }); }\nrun({ t: 2 });\nconsole.log('done');\n", ":4:35", "R seen as W", ""},
		{"assignment", "interface R { readonly t: number }\ninterface W { t: number }\nconst p: R = { t: 2 };\nlet o: { p: W } | undefined = undefined;\no = { p };\nif (o) { o.p.t = 1; }\nconsole.log(String(p.t));\n", ":5:7", "R seen as W", ""},
		{"array_element", "interface R { readonly t: number }\ninterface W { t: number }\nconst p: R = { t: 2 };\nconst xs: ({ p: W } | undefined)[] = [{ p }];\nfor (const o of xs) { if (o) { o.p.t = 1; } }\nconsole.log(String(p.t));\n", ":4:41", "R seen as W", ""},
		{"object_union", "interface R { readonly t: number }\ninterface W { t: number }\nconst p: R = { t: 2 };\nconst o: { p: W } | { q: number } = { p };\nif ('p' in o) { o.p.t = 1; }\nconsole.log(String(p.t));\n", ":4:39", "R seen as W", ""},
		{"spread", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(h: { p: R }): { p: W } | undefined { return { ...h }; }\nconst r: R = { t: 2 };\nconst x = make({ p: r });\nif (x) { x.p.t = 1; }\nconsole.log(String(r.t));\n", ":3:61", "{ p: R; } seen as { p: W; }", ""},
		{"spread_object_union", "interface R { readonly t: number }\ninterface W { t: number }\nconst r: R = { t: 2 };\nconst h = { p: r };\nconst o: { p: W } | { q: number } = { ...h };\nif ('p' in o) { o.p.t = 1; }\nconsole.log(String(r.t));\n", ":5:39", "{ p: R; } seen as { p: W; }", ""},
		{"spread_no_union", "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(h: { p: R }): { p: W } { return { ...h }; }\nconst r: R = { t: 2 };\nmake({ p: r }).p.t = 1;\nconsole.log(String(r.t));\n", ":3:49", "{ p: R; } seen as { p: W; }", ""},
		{"yield", "interface R { readonly t: number }\ninterface W { t: number }\nfunction* each(p: R): Generator<{ p: W } | undefined> { yield { p }; }\nfor (const o of each({ t: 2 })) { if (o) { o.p.t = 1; } }\nconsole.log('done');\n", ":3:65", "R seen as W", ".ts"},
		{"spread_discriminant_split", "interface R { readonly t: number }\ninterface W { t: number }\ntype S = { k: 'x'; p: W } | { k: 'y'; p: W };\nfunction make(h: { p: R }, k: 'x' | 'y'): S { return { ...h, k }; }\nconst r: R = { t: 2 };\nmake({ p: r }, 'x').p.t = 1;\nconsole.log(String(r.t));\n", ":4:56", "{ p: R; } seen as { k: \"x\"; p: W; }", ""},
		{"discriminant_split", "interface R { readonly t: number }\ninterface W { t: number }\ntype S = { k: 'x'; p: W } | { k: 'y'; p: W };\nfunction make(p: R, k: 'x' | 'y'): S { return { p, k }; }\nconst r: R = { t: 2 };\nmake(r, 'x').p.t = 1;\nconsole.log(String(r.t));\n", ":4:49", "R seen as W", ""},
	} {
		for _, extension := range []string{".a", ".ts"} {
			if probe.only != "" && probe.only != extension {
				continue
			}
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := probe.where + ": Adamic 0.1 refuses a value of type " + probe.seen + ", whose readonly field t becomes writable"
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), want) || !strings.HasSuffix(err.Error(), "(adamic/invariant-mutable)") {
					t.Fatalf("want %q, got %v", want, err)
				}
			})
		}
	}
	t.Run("honest_spread", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(h: { p: W }): { p: W } | undefined { return { ...h }; }\nconst w: W = { t: 2 };\nconst x = make({ p: w });\nif (x) { x.p.t = 1; }\nconsole.log(String(w.t));\n", ".ts")
	})
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, "interface R { readonly t: number }\ninterface W { t: number }\nfunction make(p: W): { p: W } | undefined { if (p.t > 0) { return { p }; } return undefined; }\nconst w: W = { t: 2 };\nconst x = make(w);\nif (x) { x.p.t = 1; }\nconsole.log(String(w.t));\n", ".ts")
	})
}
