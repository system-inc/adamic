package lower

import (
	"errors"
	"strings"
	"testing"
)

// #r11wppy: a fresh literal reached through an assignment chain, an || or a ?? keeps its freshness
// when every handle in the chain declares what the slot reads, as tsc's own sources write it.
func TestFreshAssignmentValueLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"or assignment", "const flag = Date.now() < 0;\nlet b: number[] | undefined = flag ? [7] : undefined;\n(b || (b = [])).push(1);\nconsole.log(String(b.length));\n"},
		{"chain", "const lists: number[][] = [];\nlet list = lists[0];\nif (!list) lists[0] = list = [];\nlist.push(2);\nconsole.log(String(lists[0]?.[0]));\n"},
		{"nullish field", "class Holder { r?: number[]; }\nconst o = new Holder();\nconst r = o.r ?? (o.r = []);\nr.push(4);\nconsole.log(String(o.r.length));\n"},
		{"argument", "const m = new Map<string, number[]>();\nlet d = m.get('k');\nif (!d) m.set('k', d = []);\nd.push(5);\nconsole.log(String(m.get('k')?.length));\n"},
		{"or return", "function g(m: Map<string, number[]>, k: string): number[] { return m.get(k) || []; }\nconsole.log(String(g(new Map<string, number[]>(), 'x').length));\n"},
		{"class field", "class C { items?: string[]; add(s: string): void { (this.items || (this.items = [])).push(s); } }\nconst c = new C();\nc.add('a');\nconsole.log(String(c.items?.length));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

const freshAssignmentInvariant = " Adamic 0.1 refuses a value of type never[] seen as number[], which can write number where never is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"

// Each stays refused at the outer assignment, where main refuses it. Mutant: judge only the outermost
// handle in freshAssignmentHandles, and "disagreeing handles" moves to the inner site a = [] (:3:26),
// which judges a against b's number[] on its own. Every other row is unchanged by that mutant.
func TestFreshAssignmentValueRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, where string }{
		{"disagreeing handles", "let a: never[] = [];\nlet b: number[] = [];\nconst v: number[] = (b = a = []);\nv.push(1);\nconst first = a[0];\nconsole.log(typeof first);\n", ":3:22:"},
		{"narrower handle", "let a: (1 | 2)[] | undefined;\nconst v: number[] = (a = []);\nv.push(3);\nconst first = a[0];\nconsole.log(String(first));\n", ":2:22:"},
		{"wider handle", "let a: (number | string)[] = [];\nconst v: number[] = (a = []);\na.push('s');\nconst first = v[0];\nconsole.log(typeof first);\n", ":2:22:"},
		{"held literal", "const lit: never[] = [];\nlet b: number[] | undefined;\nconst v: number[] = (b = lit);\nv.push(4);\nconst first = lit[0];\nconsole.log(typeof first);\n", ":3:22:"},
		{"evolving let", "let e;\nconst v: number[] = (e = []);\nv.push(5);\nconsole.log(String(v.length));\n", ":2:22:"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + freshAssignmentInvariant
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want a refusal ending %q", err, want)
				}
			})
		}
	}
}

// Language's attack roads on #r11wppy, each one tsc accepts and Node fails on, a number read where a string
// is: a chain whose targets disagree, a setter whose write type isn't its read type, a logical assignment
// used as a value, a destructuring field fed by another assignment, and an argument also assigned. Each must
// stay stopped, by this rule or an earlier one; the words are logged, so a run shows which rule holds each.
func TestFreshAssignmentLanguageAttacksStayStopped(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"f01 chain disagrees", "let a: string[] = ['x'];\nlet b: (string | number)[] = [];\nb = a = [];\nb.push(7);\nconst first = a[0];\nconsole.log(first === undefined ? 'none' : first.toUpperCase());\n"},
		{"f02 setter write type", "class H {\n\tprivate reads: string[] = [];\n\tprivate writes: (string | number)[] = [];\n\tget list(): string[] { return this.reads; }\n\tset list(v: (string | number)[]) { this.writes = v; }\n\tcount(): number { return this.writes.length; }\n}\nconst h = new H();\nconst s: string[] = (h.list = []);\nconsole.log(String(s.length + h.count()));\n"},
		{"f03 nullish assignment", "let x: string[] | undefined;\nconst wide: (string | number)[] = (x ??= []);\nwide.push(7);\nconsole.log(typeof x[0]);\n"},
		{"f03 or assignment", "let x: string[] | undefined;\nconst wide: (string | number)[] = (x ||= []);\nwide.push(7);\nconsole.log(typeof x[0]);\n"},
		{"f05 destructuring field", "let a: string[] = [];\nlet b: (string | number)[] = [];\n({ x: a } = { x: b = [] });\nb.push(7);\nconsole.log(typeof a[0]);\n"},
		{"f06 argument also assigned", "const m = new Map<string, (string | number)[]>();\nlet list: string[] = [];\nm.set('k', list = []);\nm.get('k')?.push(7);\nconsole.log(typeof list[0]);\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				var notYet *NotYet
				if !errors.As(err, &refused) && !errors.As(err, &notYet) {
					t.Fatalf("want it stopped, got %v", err)
				}
				t.Logf("%s stops: %v", probe.name+extension, err)
			})
		}
	}
}
