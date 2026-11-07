package lower

import (
	"strings"
	"testing"
)

// A collator seen through a view that names compare stops by name at the relation, at any depth the
// host view walk reaches (nodeHostObject's Intl.Collator, #zjndrx5), since its compare isn't a slot of
// its shape. Seen as itself, it lowers on past the walk.
func TestIntlCollatorViewsStopByName(t *testing.T) {
	t.Parallel()
	const collator = `type F = (x: string, y: string) => number;
const c = new Intl.Collator(undefined, { usage: "sort", sensitivity: "variant", numeric: true });
`
	for _, probe := range []struct{ name, source string }{
		{"top", `const v: { compare: F } = c;
console.log(String(v.compare("a", "b")));
`},
		{"array", `const cs: readonly Intl.Collator[] = [c];
const vs: readonly { compare: F }[] = cs;
console.log(String(vs.length));
`},
		{"field", `const box: { readonly c: Intl.Collator } = { c };
const seen: { readonly c: { compare: F } } = box;
console.log(String(seen.c.compare("a", "b")));
`},
		{"callback", `function hand(take: (c: Intl.Collator) => void): void { take(c); }
hand((v: { compare: F }): void => { console.log(String(v.compare("a", "b"))); });
`},
		{"map", `const m = new Map<string, Intl.Collator>([["a", c]]);
const mv: ReadonlyMap<string, { compare: F }> = m;
console.log(String(mv.size));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, collator+probe.source, ".ts")
			t.Logf("%s stops at: %v", probe.name, err)
			if err == nil || !strings.Contains(err.Error(), "which reaches Collator.compare, a member no host registration names") {
				t.Fatalf("want the host view stop naming Collator.compare, got %v", err)
			}
		})
	}
	t.Run("itself", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSourceAs(t, collator+"const same: Intl.Collator = c;\nconsole.log(String(same.compare(\"a\", \"b\")));\n", ".ts")
		t.Logf("itself stops at: %v", err)
		if err != nil && strings.Contains(err.Error(), "a member no host registration names") {
			t.Fatalf("a collator seen as itself reached the host view stop: %v", err)
		}
	})
}
