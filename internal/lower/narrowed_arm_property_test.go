package lower

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
)

func init() { narrowedArmViewSites = &enumTagSwingRecord{} }

// A property narrowed to an object arm and kept by the checker across a call that may rewrite it (#8nj5pn4):
// if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } reads whatever the call left there through A's
// layout. That's a claim the walk can't prove whatever v's representation (Adamic's ruling on #p1p7k3g, which
// superseded the filter that let a read pass where every type a stale value could be holds v at one
// representation), so on .ts every such read goes through a checked view of the declared type, as the swing reads
// an open-enum refinement it can't prove, and on .a it's refused at compile time, pinned by the string and enum
// direct lies. Each lie lowers and stops loudly at the checked read, exit 70, on sanitized native and
// the JavaScript backend (STOP), and each honest twin, whose call writes an A, lowers and agrees with Node on both
// (AGREE). The lies are string tags and an open enum's, read directly and handed on, through an element by a literal
// index (Language's r08), a const index (e02) or a const enum member (e05), and a stale value whose field comes from
// a base class (-12). A stale B holding v at A's own representation agrees with Node on both backends (AGREE): the
// checked view checks the field being read, and the stale B holds it at the same representation, so the read
// returns Node's value. Those are both arms holding v as a string, a type no stale value can be declaring v as a
// string (the unrelated declarer), an unrelated literal's computed key, and the base class twin, whose inherited v is
// a number; each records a view at its read, the proof the floor reached it. The element held into a const after
// the call is still the checker's A (Language's e01), so the program's own re-check reads got.kind through A's view:
// Node goes on, since it re-checks, and prints changed, while native and the JavaScript backend stop at the
// re-check's read, exit 70, the checked view's loud boundary on a stale type, as the swing gives a re-checked stale
// refinement. A stale value made by a spread of a viewed type, the extended interface and
// intersection member lies and their twins, keeps a compile-time stop, the view's copy rule at make's spread
// (main.ts:8:43). A getter named n stops on .ts too, even with nothing between, since it answers each read itself.
// Rows with nothing between the check and the read, a held binding the call can't rewrite, and an unrelated getter
// lower as they stand and agree with Node. Each row the floor judges records a view at its read
// (narrowedArmViewSites) and each row it passes over records none, so a read moved across that line is seen even
// where the program prints the same. Mutants (narrowed_arm_property.mutants.md) hold each part: the property
// dispatch, the write test, the view itself and the superseded filter restored.
func TestNarrowedArmPropertyRead(t *testing.T) {
	t.Parallel()
	const enumTypes = "enum K { A = 1, B = 2 }\ninterface A { readonly kind: K.A; readonly v: number }\ninterface B { readonly kind: K.B; readonly v: string }\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: K.A, v: 1 };\nconst b: B = { kind: K.B, v: \"x\" };\nfunction reset(m: M): void { m.n = b; }\nfunction keep(m: M): void { m.n = a; }\nfunction use(x: A): string { return String(x.v + 1); }\n"
	const stringTypes = "type A = { readonly kind: 'a'; readonly v: number };\ntype B = { readonly kind: 'b'; readonly v: string };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', v: 1 };\nconst b: B = { kind: 'b', v: \"x\" };\nfunction reset(m: M): void { m.n = b; }\nfunction keep(m: M): void { m.n = a; }\nfunction use(x: A): string { return String(x.v + 1); }\n"
	const sameTypes = "type A = { readonly kind: 'a'; readonly v: string };\ntype B = { readonly kind: 'b'; readonly v: string };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', v: \"y\" };\nconst b: B = { kind: 'b', v: \"x\" };\nfunction reset(m: M): void { m.n = b; }\n"
	const numberTypes = "type A = { readonly kind: 'a'; readonly v: number };\ntype B = { readonly kind: 'b'; readonly v: number };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', v: 1 };\nconst b: B = { kind: 'b', v: 5 };\nfunction reset(m: M): void { m.n = b; }\n"
	const stop = "narrowed to A where a call may have rewritten it"
	const refusedInA = ": Adamic 0.1 refuses a read of property n narrowed to A where a call may have rewritten it: the narrowing is kept across the call, which nothing proves leaves it holding; test the property again after the call, or read it into a const before the call and narrow that"
	const viewCopy = "main.ts:8:43: stage 0 can't lower a whole-object copy of a value of type Base that may hold a checked view's value yet"
	const stringDirectLie = stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n"
	const enumDirectLie = enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { reset(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n"
	const inheritedTypes = "type A = { readonly kind: 'a'; readonly v: number };\ntype B = { readonly kind: 'b' };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', v: 1 };\n"
	const inheritedShow = "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n"
	const elements = "function resetFirst(x: N[]): void { x[0] = b; }\nfunction keepFirst(x: N[]): void { x[0] = a; }\n"
	for _, row := range []struct{ name, source, want, stdout string }{
		{"string direct lie", stringDirectLie, "STOP", ""},
		{"string handed on lie", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return use(m.n); } return 'b'; }\nconsole.log(show({ n: a }));\n", "STOP", ""},
		{"enum direct lie", enumDirectLie, "STOP", ""},
		{"enum handed on lie", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { reset(m); return use(m.n); } return 'b'; }\nconsole.log(show({ n: a }));\n", "STOP", ""},
		{"string honest twin", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { keep(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "AGREE", ""},
		{"same representation", sameTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return m.n.v; } return 'b'; }\nconsole.log(show({ n: a }));\n", "AGREE", ""},
		{"string nothing between", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\nconsole.log(show({ n: b }));\n", "", "2\nb\n"},
		{"enum nothing between", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "", "2\n"},
		{"getter named n", stringTypes + "const g = { get n(): N { return b; } };\nconsole.log(g.n.kind);\nfunction show(m: M): string { if (m.n.kind === 'a') { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "NotYet", ""},
		{"unrelated getter", stringTypes + "const g = { get other(): number { return 1; } };\nfunction show(m: M): string { if (m.n.kind === 'a') { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\nconsole.log(String(g.other));\n", "", "2\n1\n"},
		{"element lie", stringTypes + elements + "function show(arr: N[]): string { if (arr[0] !== undefined && arr[0].kind === 'a') { resetFirst(arr); return String(arr[0].v + 1); } return 'b'; }\nconsole.log(show([a]));\n", "STOP", ""},
		{"element honest twin", stringTypes + elements + "function show(arr: N[]): string { if (arr[0] !== undefined && arr[0].kind === 'a') { keepFirst(arr); return String(arr[0].v + 1); } return 'b'; }\nconsole.log(show([a]));\n", "AGREE", ""},
		{"element const index lie", stringTypes + elements + "const index = 0;\nfunction show(arr: N[]): string { if (arr[index] !== undefined && arr[index].kind === 'a') { resetFirst(arr); return String(arr[index].v + 1); } return 'b'; }\nconsole.log(show([a]));\n", "STOP", ""},
		{"element const enum index lie", stringTypes + elements + "const enum Slot { First = 0 }\nfunction show(arr: N[]): string { if (arr[Slot.First] !== undefined && arr[Slot.First].kind === 'a') { resetFirst(arr); return String(arr[Slot.First].v + 1); } return 'b'; }\nconsole.log(show([a]));\n", "STOP", ""},
		{"element held after the call", stringTypes + elements + "const index = 0;\nfunction show(arr: N[]): string { if (arr[index] !== undefined && arr[index].kind === 'a') { resetFirst(arr); const got = arr[index]; if (got !== undefined && got.kind === 'a') { return String(got.v + 1); } return 'changed'; } return 'b'; }\nconsole.log(show([a]));\n", "STOP", ""},
		{"element nothing between", stringTypes + elements + "function show(arr: N[]): string { if (arr[0] !== undefined && arr[0].kind === 'a') { return String(arr[0].v + 1); } return 'b'; }\nconsole.log(show([a]));\nconsole.log(show([b]));\n", "", "2\nb\n"},
		{"unrelated declarer", numberTypes + "interface Other { readonly v: string }\nconst other: Other = { v: \"o\" };\nfunction show(m: M): string { if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } return other.v; }\nconsole.log(show({ n: a }));\n", "AGREE", ""},
		{"computed key elsewhere", numberTypes + "const key: string = \"q\";\nconst bag: Record<string, number> = { [key]: 1 };\nfunction show(m: M): string { if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } return String(bag[key]); }\nconsole.log(show({ n: a }));\n", "AGREE", ""},
		{"base class lie", inheritedTypes + "class Base { readonly v: string = \"s\"; }\nclass Bin extends Base { readonly kind = 'b' as const; }\nfunction reset(m: M): void { m.n = new Bin(); }\n" + inheritedShow, "STOP", ""},
		{"base class twin", inheritedTypes + "class Base { readonly v: number = 5; }\nclass Bin extends Base { readonly kind = 'b' as const; }\nfunction reset(m: M): void { m.n = new Bin(); }\n" + inheritedShow, "AGREE", ""},
		{"extended interface lie", inheritedTypes + "interface Base { readonly v: string }\ninterface Bin extends Base { readonly kind: 'b' }\nfunction make(base: Base): Bin { return { ...base, kind: 'b' }; }\nfunction reset(m: M): void { m.n = make({ v: \"s\" }); }\n" + inheritedShow, "ViewCopy", ""},
		{"extended interface twin", inheritedTypes + "interface Base { readonly v: number }\ninterface Bin extends Base { readonly kind: 'b' }\nfunction make(base: Base): Bin { return { ...base, kind: 'b' }; }\nfunction reset(m: M): void { m.n = make({ v: 5 }); }\n" + inheritedShow, "ViewCopy", ""},
		{"intersection member lie", inheritedTypes + "type Base = { readonly v: string };\ntype Bin = { readonly kind: 'b' } & Base;\nfunction make(base: Base): Bin { return { ...base, kind: 'b' }; }\nfunction reset(m: M): void { m.n = make({ v: \"s\" }); }\n" + inheritedShow, "ViewCopy", ""},
		{"intersection member twin", inheritedTypes + "type Base = { readonly v: number };\ntype Bin = { readonly kind: 'b' } & Base;\nfunction make(base: Base): Bin { return { ...base, kind: 'b' }; }\nfunction reset(m: M): void { m.n = make({ v: 5 }); }\n" + inheritedShow, "ViewCopy", ""},
		{"string held binding", stringTypes + "function show(m: M): string { const n = m.n; if (n.kind === 'a') { reset(m); return String(n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "", "2\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			// Every row the floor judges is read through a view, and every row it passes over takes none, which
			// only its site shows where the program prints the same either way. The getter's row stops before any
			// view.
			sites := narrowedArmViewedAt(t, row.source)
			if viewed := row.want == "STOP" || row.want == "AGREE" || row.want == "ViewCopy"; viewed != (len(sites) > 0) {
				t.Fatalf("want a view taken %v, got sites %v", viewed, sites)
			}
			switch row.want {
			case "NotYet":
				_, err := lowerSourceAs(t, row.source, ".ts")
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), stop) {
					t.Fatalf("want a not yet with %q, got %v", stop, err)
				}
			case "ViewCopy":
				_, err := lowerSourceAs(t, row.source, ".ts")
				if err == nil || !strings.Contains(err.Error(), viewCopy) {
					t.Fatalf("want a not yet with %q, got %v", viewCopy, err)
				}
			case "STOP", "AGREE":
				nativeClass, javascriptClass, observed := narrowedArmViewClasses(t, row.source)
				if nativeClass != row.want || javascriptClass != row.want {
					t.Fatalf("want %s on both backends, got native %s and JavaScript %s; %s", row.want, nativeClass, javascriptClass, observed)
				}
			default:
				program := lowersAndAgreesWithNodeAs(t, row.source, ".ts")
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != row.stdout {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, row.stdout)
				}
			}
		})
	}
	for _, row := range []struct{ name, source string }{
		{"string direct lie in .a", stringDirectLie},
		{"enum direct lie in .a", enumDirectLie},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".a")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), refusedInA) {
				t.Fatalf("want a refusal ending %q, got %v", refusedInA, err)
			}
		})
	}
}

// narrowedArmViewedAt lowers source as main.ts and gives the sites where the floor read through checked views in
// place of its stop (narrowedArmViewSites), sorted, whatever the lowering's answer. Each call writes its program in a
// directory of its own, so the sites are that program's alone.
func narrowedArmViewedAt(t *testing.T, source string) []string {
	t.Helper()
	path := writeAgreementSource(t, "main.ts", source)
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_, _ = Lower(context.Background(), loaded) // The rows judge the lowering's answer themselves.
	narrowedArmViewSites.mutex.Lock()
	defer narrowedArmViewSites.mutex.Unlock()
	sites := slices.Clone(narrowedArmViewSites.sites[filepath.Dir(path)])
	slices.Sort(sites)
	return sites
}

// narrowedArmViewClasses lowers source as .ts, runs it on sanitized native and on the JavaScript backend, and names
// each run against Node's as the swing's rows do (enumTagSwingClass), with what each printed for a failure's words.
func narrowedArmViewClasses(t *testing.T, source string) (nativeClass, javascriptClass, observed string) {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("want the checked view to lower it, got %v", err)
	}
	node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
	nativeCode, nativeStdout, nativeStderr := brandedNative(t, program)
	generated := runAgreementNode(t, writeAgreementSource(t, "generated.mjs", javascript.JavaScript(program)))
	nativeClass = enumTagSwingClass(nativeCode, nativeStdout, nativeStderr, node)
	javascriptClass = enumTagSwingClass(generated.code, string(generated.stdout), string(generated.stderr), node)
	observed = fmt.Sprintf("native exit %d stdout %q stderr %.300q, JavaScript exit %d stdout %q stderr %.300q, Node exit %d stdout %q", nativeCode, nativeStdout, nativeStderr, generated.code, generated.stdout, generated.stderr, node.code, node.stdout)
	return nativeClass, javascriptClass, observed
}

// The swing and the floor read the same way (Adamic's seam ruling of Nov 27 on #8nj5pn4): an open-enum refinement the
// walk can't prove is read through each narrowed member's checked view (enums.go's narrowedCheckedViews), and so is
// the floor's stale arm in .ts (narrowedArmView), both keyed by the declared type and the member, so a read both
// judge registers each view once. These are the swing's own types (enumTagSwingViewTypes's shape) with v a number in
// A and a string in B, which the floor judges too. The honest twin, whose call writes an A, lowers and agrees with
// Node on both backends, and the lie, whose call writes a B, lowers and stops at the checked read with exit 70 on
// sanitized native and the JavaScript backend. The floor once passed over a read the swing took; dropping that
// exemption left both rows green (J-6jsr1da's M6), since the floor's views are the swing's, so it was deleted.
func TestNarrowedArmPropertyReadSwingExempt(t *testing.T) {
	t.Parallel()
	const types = `enum K { A = 1, B = 2 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly v: number }
interface B { readonly kind: K.B; readonly v: string }
type N = A | B;
interface M { n: N }
const a: A = { kind: K.A, flag: F.X, v: 1 };
const b: B = { kind: K.B, v: "x" };
function reset(m: M): void { m.n = b; }
function keep(m: M): void { m.n = a; }
`
	for _, side := range []struct{ name, call, want string }{
		{"honest", "keep", "AGREE"},
		{"lie", "reset", "STOP"},
	} {
		t.Run(side.name, func(t *testing.T) {
			t.Parallel()
			source := types + "function show(m: M): string { if (m.n.kind === K.A) { " + side.call + "(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n"
			program, err := lowerSourceAs(t, source, ".ts")
			if err != nil {
				t.Fatalf("want the swing's checked view to lower it, got %v", err)
			}
			node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
			nativeCode, nativeStdout, nativeStderr := brandedNative(t, program)
			generated := runAgreementNode(t, writeAgreementSource(t, "generated.mjs", javascript.JavaScript(program)))
			nativeClass := enumTagSwingClass(nativeCode, nativeStdout, nativeStderr, node)
			javascriptClass := enumTagSwingClass(generated.code, string(generated.stdout), string(generated.stderr), node)
			if nativeClass != side.want || javascriptClass != side.want {
				t.Fatalf("want %s on both backends, got native %s (exit %d stdout %q stderr %q) and JavaScript %s (exit %d stdout %q stderr %q); Node exit %d stdout %q", side.want, nativeClass, nativeCode, nativeStdout, nativeStderr, javascriptClass, generated.code, generated.stdout, generated.stderr, node.code, node.stdout)
			}
		})
	}
}
