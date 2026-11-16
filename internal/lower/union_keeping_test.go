package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

const unionKeepingPrelude = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function make(): Strong { return { kind: 'v', v: { label: 'label' } }; }
`

func TestUnionKeepingPairs(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, types string }{
			{"p5", `const strong: Strong = make();
const viewed: Weakly = strong;
if (viewed.kind === 'v') console.log(viewed.v?.label ?? 'none');`, "Strong -> Weakly"},
			{"p6", `const strong: { readonly item: Strong } = { item: make() };
const viewed: { readonly item: Weakly } = strong;
if (viewed.item.kind === 'v') console.log(viewed.item.v?.label ?? 'none');`, "{ readonly item: Strong } -> { readonly item: Weakly }"},
			{"p7", `function read(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
const viewed: (x: Strong) => string = read;
console.log(viewed(make()));`, "(x: Weakly) => string -> (x: Strong) => string"},
			{"p8", `const strong: readonly Strong[] = [make()];
const viewed: readonly Weakly[] = strong;
const item = viewed[0]; if (item !== undefined && item.kind === 'v') console.log(item.v?.label ?? 'none');`, "readonly Strong[] -> readonly Weakly[]"},
			{"later_weak_member", `const strong: Strong = make();
const viewed: { readonly kind: 'w'; readonly w: number } | { readonly kind: 'v'; readonly v: Weak<Box> } = strong;
if (viewed.kind === 'v') console.log(viewed.v?.label ?? 'none');`, "Strong -> { readonly kind: 'w'; readonly w: number } | { readonly kind: 'v'; readonly v: Weak<Box> }"},
			{"whole_view_only", `function makeWide(): { readonly kind: 'v' | 'w'; readonly v: Box } { return { kind: 'v', v: { label: 'label' } }; }
const viewed: { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly v: Weak<Box> } = makeWide();`, "{ readonly kind: 'v' | 'w'; readonly v: Box } -> { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly v: Weak<Box> }"},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				source := unionKeepingPrelude + row.source
				_, err := lowerSourceAs(t, source, extension)
				var ny *NotYet
				if !errors.As(err, &ny) || !strings.Contains(err.Error(), "one keeps something weakly that the other keeps strongly") {
					t.Fatalf("%s: got %v, want retention NotYet at assignment", row.types, err)
				}
				line := strings.Count(source[:strings.Index(source, "const viewed:")], "\n") + 1
				if !strings.Contains(ny.Where, fmt.Sprintf("main%s:%d:", extension, line)) {
					t.Fatalf("%s: got %v, want assignment on line %d", row.types, err, line)
				}
			})
		}
	}
}

func TestUnionKeepingPairsAgree(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source string }{
			{"r09_control_swapped_weak", `import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
interface Other { readonly size: number }
type First = { readonly kind: "v"; readonly v: Weak<Box> } | { readonly kind: "w"; readonly w: Weak<Other> };
type Second = { readonly kind: "w"; readonly w: Weak<Other> } | { readonly kind: "v"; readonly v: Weak<Box> };
const held = box();
const first: First = { kind: "v", v: held };
const second: Second = first;
console.log(second.kind === "v" ? second.v?.label ?? "none" : "w");
`},
			{"r10_control_strong_same", `import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Strong = { readonly kind: "v"; readonly v: Box } | { readonly kind: "w"; readonly w: number };
type Twin = { readonly kind: "w"; readonly w: number } | { readonly kind: "v"; readonly v: Box };
const strong: Strong = { kind: "v", v: box() };
const twin: Twin = strong;
console.log(twin.kind === "v" ? twin.v.label : "w");
`},
			{"weak_member_control", `import type { Weak } from 'adamic';
interface A { readonly kind: 'a'; readonly label: string }
interface C { readonly kind: 'c'; readonly label: string }
interface Held { readonly item: Weak<A> | C }
interface Twin { readonly item: Weak<A> | { readonly kind: 'c'; readonly label: string } }
function view(held: Held): Twin { const same: Held = held; const twin: Twin = same; return twin; }
console.log('control');`},
			// A union with a Promise member, which takes the promise as that member. A union holding only a
			// structural PromiseLike<string> boxes it as an object, which shape_fit.go's promiseWhereObjectGoes
			// stops (#v5b6pzp: `'then' in` such a value panicked natively where Node prints true).
			{"promise_union_view", `const p: string | Promise<string> = Promise.resolve('x'); console.log(typeof p);`},
			{"narrowed_nullish_box", `interface Box { readonly label: string }
function read(box: Box | undefined): string {
    if (box === undefined) return 'none';
    const viewed: Box = box;
    return viewed.label;
}
console.log(read({label:'label'})); console.log(read(undefined));`},
			{"range_iterator", `class RangeIterator {
    index = 0;
    readonly limit: number;
    constructor(limit: number) { this.limit = limit; }
    [Symbol.iterator](): RangeIterator { return this; }
    next(): { readonly done: boolean; readonly value: number } {
        const value = this.index;
        this.index++;
        return { done: value >= this.limit, value };
    }
}
class Range {
    readonly limit: number;
    constructor(limit: number) { this.limit = limit; }
    [Symbol.iterator](): RangeIterator { return new RangeIterator(this.limit); }
}
console.log(Array.from(new Range(3)).join(','));`},
			{"weak_union_write", `import type { Weak } from 'adamic';
interface Named { readonly kind:'Named'; readonly name:string }
interface Counted { readonly kind:'Counted'; readonly count:number }
interface Holder { kept:Weak<Counted | Named> }
const named:Named = {kind:'Named',name:'label'};
const holder:Holder = {kept:named};
holder.kept = named;
const seen = holder.kept;
if (seen !== undefined && seen.kind === 'Named') console.log(seen.name);`},
			{"checked_downcast", `type Circle = { readonly kind: 'circle'; readonly radius: number }; type Square = { readonly kind: 'square'; readonly side: number };
function read(shape: Circle | Square): number { if (shape.kind === 'circle') return (shape as Circle).radius; return 0; }
console.log(String(read({kind:'circle',radius:3})));`},
			{"optional_name", `interface Box { readonly name: string | undefined } function read(box: Box | undefined): string { return box?.name ?? 'none'; } console.log(read({name:'label'})); console.log(read(undefined));`},
			{"same_union", `type U = { readonly kind:'v'; readonly v:string } | { readonly kind:'w'; readonly w:number }; function read(x:U):string { const y:U=x; return y.kind==='v'?y.v:String(y.w); } console.log(read({kind:'v',v:'label'}));`},
			{"regex_string", `function read(x:string|RegExp):string|RegExp { const y:string|RegExp=x; return y; } const value = read('/label/'); if (typeof value === 'string') console.log(value);`},
			{"recursive_union", `type Tree = {readonly kind:'leaf'; readonly label:string} | {readonly kind:'branch'; readonly child:Tree}; function read(x:Tree):string { const y:Tree=x; return y.kind==='leaf'?y.label:read(y.child); } console.log(read({kind:'branch',child:{kind:'leaf',label:'label'}}));`},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				source := row.source
				if row.name == "checked_downcast" && extension == ".ts" {
					// .ts checks the cast's discriminant; .a requires explicit narrowing.
					source = strings.Replace(source, "if (shape.kind === 'circle') return (shape as Circle).radius; return 0;", "return (shape as Circle).radius;", 1)
				}
				lowersAndAgreesWithNodeAs(t, source, extension)
			})
		}
	}
}

func TestUnionKeepingGrow(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			// Load the growing types without asking the checker to relate them: this
			// pins the keeping walk's bound, independently of the view pre-pass.
			program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, `interface Grow<T> { readonly next?: Grow<Grow<T>>; readonly v: T }
interface Twin<T> { readonly next?: Twin<Twin<T>>; readonly v: T }
type From = { readonly kind: 'grow'; readonly value: Grow<number> } | { readonly kind: 'end' };
type To = { readonly kind: 'grow'; readonly value: Twin<number> } | { readonly kind: 'end' };
declare const from: From;
declare const to: To;`)})
			if err != nil {
				t.Fatal(err)
			}
			checked, release := program.Checker(t.Context(), program.Files()[0])
			defer release()
			var from, to *checker.Type
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if ast.IsIdentifier(node) {
					switch node.Text() {
					case "from":
						from = checked.GetTypeAtLocation(node)
					case "to":
						to = checked.GetTypeAtLocation(node)
					}
				}
				return node.ForEachChild(visit)
			}
			program.Files()[0].Node.ForEachChild(visit)
			if from == nil || to == nil {
				t.Fatal("missing Grow union types")
			}
			l := &lowering{program: program, checker: checked}
			if l.sameKeepingWalk(from, to, map[[2]*checker.Type]bool{}) {
				t.Fatal("From with Grow<number> -> To with Twin<number>: want not same at the nesting bound")
			}
		})
	}
}

func TestUnionKeepingWeakMembers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, from, to string }{
			{"field", `import type { Weak } from 'adamic';
interface A { readonly kind: 'a'; readonly label: string }
interface B { readonly kind: 'b'; readonly label: string }
interface C { readonly kind: 'c'; readonly label: string }
let strongA: A | undefined = { kind: 'a', label: 'a'.repeat(2) };
let strongB: B | undefined = { kind: 'b', label: 'b'.repeat(2) };
interface Held { readonly item: Weak<A> | Weak<B> | C }
interface Viewed { readonly item: Weak<A> | B | C }
function make(): Held { return { item: strongB }; }
const held: Held = make();
const viewed: Viewed = held;
strongA = undefined;
strongB = undefined;
if (viewed.item !== undefined && viewed.item.kind === 'b') console.log(viewed.item.label);
`, "{ readonly item: Weak<A> | Weak<B> | C }", "{ readonly item: Weak<A> | B | C }"},
			{"top", `import type { Weak } from 'adamic';
interface A { readonly kind: 'a'; readonly label: string }
interface B { readonly kind: 'b'; readonly label: string }
interface C { readonly kind: 'c'; readonly label: string }
let strongA: A | undefined = { kind: 'a', label: 'a'.repeat(2) };
let strongB: B | undefined = { kind: 'b', label: 'b'.repeat(2) };
function make(): Weak<A> | Weak<B> { return strongB; }
const held: Weak<A> | Weak<B> = make();
const viewed: Weak<A> | B = held;
strongA = undefined;
strongB = undefined;
if (viewed !== undefined && viewed.kind === 'b') console.log(viewed.label);
`, "Weak<A> | Weak<B>", "Weak<A> | B"},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				_, err := lowerSourceAs(t, row.source, extension)
				var ny *NotYet
				if !errors.As(err, &ny) {
					t.Fatalf("%s -> %s: got %v, want NotYet", row.from, row.to, err)
				}
				// Below the top pair the member rule must stop the field view. The
				// top pair itself stays exempt; its program keeps the existing
				// mixed-representation NotYet before the view is read.
				program, err := load.Load([]string{writeAgreementSource(t, "types"+extension, `import type { Weak } from 'adamic';
interface A { readonly kind: 'a'; readonly label: string }
interface B { readonly kind: 'b'; readonly label: string }
interface C { readonly kind: 'c'; readonly label: string }
type From = `+row.from+`;
type To = `+row.to+`;
declare const from: From;
declare const to: To;`)})
				if err != nil {
					t.Fatal(err)
				}
				checked, release := program.Checker(t.Context(), program.Files()[0])
				defer release()
				var from, to *checker.Type
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if ast.IsIdentifier(node) {
						switch node.Text() {
						case "from":
							from = checked.GetTypeAtLocation(node)
						case "to":
							to = checked.GetTypeAtLocation(node)
						}
					}
					return node.ForEachChild(visit)
				}
				program.Files()[0].Node.ForEachChild(visit)
				if from == nil || to == nil {
					t.Fatal("missing member view types")
				}
				l := &lowering{program: program, checker: checked}
				wantSame := row.name == "top"
				if got := l.sameKeepingWalk(from, to, map[[2]*checker.Type]bool{}); got != wantSame {
					t.Fatalf("%s -> %s: keeping walk answered same=%v, want %v", row.from, row.to, got, wantSame)
				}
			})
		}
	}
}

// Language's programs are kept whole; each location names the view initializer.
func TestUnionKeepingLanguageStops(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, where string }{
			{"r01_generic_alias", `import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type W<T extends object> = Weak<T>;
type Strong = { readonly kind: "v"; readonly v: Box } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: W<Box> } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: box() };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? viewed.v?.label ?? "none" : "w");
`, ":8:24"},
			{"r05_two_deep_union", `import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Inner = { readonly kind: "a"; readonly b: Box } | { readonly kind: "c"; readonly c: number };
type InnerWeak = { readonly kind: "a"; readonly b: Weak<Box> } | { readonly kind: "c"; readonly c: number };
type Strong = { readonly kind: "v"; readonly v: Inner } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: InnerWeak } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: { kind: "a", b: box() } };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" && viewed.v.kind === "a" ? viewed.v.b?.label ?? "none" : "w");
`, ":9:24"},
			{"r07_class_member", `import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
class Holder {
    readonly v: Box;
    constructor(v: Box) {
        this.v = v;
    }
}
interface HolderWeak { readonly v: Weak<Box> }
type Strong = { readonly kind: "v"; readonly h: Holder } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly h: HolderWeak } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", h: new Holder(box()) };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? viewed.h.v?.label ?? "none" : "w");
`, ":14:24"},
			{"r08_function_result", `import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Strong = { readonly kind: "v"; readonly v: Box } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: Weak<Box> } | { readonly kind: "w"; readonly w: number };
const make = (): Strong => ({ kind: "v", v: box() });
const viewed: () => Viewed = make;
const got = viewed();
console.log(got.kind === "v" ? got.v?.label ?? "none" : "w");
`, ":7:30"},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				_, err := lowerSourceAs(t, row.source, extension)
				var ny *NotYet
				want := "main" + extension + row.where
				if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) || !strings.Contains(ny.What, "one keeps something weakly that the other keeps strongly") {
					t.Fatalf("got %v, want retention NotYet at %s", err, want)
				}
			})
		}
	}
}

func TestUnionKeepingCallableFieldsStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, types string }{
			{"a1_parameter", `type A = { readonly kind: 'v'; readonly f: (x: number | string) => string } | { readonly kind: 'w'; readonly w: number };
type Viewed = { readonly kind: 'v'; readonly f: (x: number) => string } | { readonly kind: 'w'; readonly w: number };
function read(x: number | string): string { const kind = typeof x; console.log(kind); return kind; }
function make(): A { return { kind: 'v', f: read }; }
const original: A = make();
const viewed: Viewed = original;
if (viewed.kind === 'v') viewed.f(1);
`, "(x: number | string) => string -> (x: number) => string"},
			// The top-level result-width view stops the same way, so this is consistent, not a new price.
			{"a2_result", `type A = { readonly kind: 'v'; readonly f: () => number } | { readonly kind: 'w'; readonly w: number };
type Viewed = { readonly kind: 'v'; readonly f: () => number | string } | { readonly kind: 'w'; readonly w: number };
function read(): number { return 1; }
function make(): A { return { kind: 'v', f: read }; }
const original: A = make();
const viewed: Viewed = original;
if (viewed.kind === 'v') console.log(typeof viewed.f());
`, "() => number -> () => number | string"},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				_, err := lowerSourceAs(t, row.source, extension)
				var ny *NotYet
				want := "main" + extension + ":6:24"
				if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) {
					t.Fatalf("%s: got %v, want NotYet at %s", row.types, err, want)
				}
			})
		}
	}
}

func TestUnionKeepingCallableFieldsAgree(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source string }{

			{"identical_function_type", `type Callback = (x: number) => string;
type A = { readonly kind: 'v'; readonly f: Callback } | { readonly kind: 'w'; readonly w: number };
type Viewed = { readonly kind: 'v'; readonly f: Callback } | { readonly kind: 'w'; readonly w: number };
function read(x: number): string { return typeof x; }
function make(): A { return { kind: 'v', f: read }; }
const original: A = make();
const viewed: Viewed = original;
if (viewed.kind === 'v') console.log(viewed.f(1));
`},
			{"unchanged_function_field", `type Callback = () => number;
type A = { readonly kind: 'v'; readonly f: Callback; readonly extra: string } | { readonly kind: 'w'; readonly w: number };
type Viewed = { readonly kind: 'v'; readonly f: Callback } | { readonly kind: 'w'; readonly w: number };
function read(): number { return 1; }
function make(): A { return { kind: 'v', f: read, extra: 'extra' }; }
const original: A = make();
const viewed: Viewed = original;
if (viewed.kind === 'v') console.log(typeof viewed.f());
`},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				lowersAndAgreesWithNodeAs(t, row.source, extension)
			})
		}
	}
}

func TestUnionKeepingCallableGrow(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			// Load the growing types without asking the checker to relate them: this
			// pins the keeping walk's bound, independently of the view pre-pass.
			program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, `interface G<T> { readonly kind: 'g'; readonly f: (x: G<G<T>> | { readonly kind: 'n' }) => void }
interface Twin<T> { readonly kind: 'g'; readonly f: (x: Twin<Twin<T>> | { readonly kind: 'n' }) => void }
type From = G<number> | { readonly kind: 'n' };
type To = Twin<number> | { readonly kind: 'n' };
declare const from: From;
declare const to: To;`)})
			if err != nil {
				t.Fatal(err)
			}
			checked, release := program.Checker(t.Context(), program.Files()[0])
			defer release()
			var from, to *checker.Type
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if ast.IsIdentifier(node) {
					switch node.Text() {
					case "from":
						from = checked.GetTypeAtLocation(node)
					case "to":
						to = checked.GetTypeAtLocation(node)
					}
				}
				return node.ForEachChild(visit)
			}
			program.Files()[0].Node.ForEachChild(visit)
			if from == nil || to == nil {
				t.Fatal("missing callable Grow union types")
			}
			l := &lowering{program: program, checker: checked}
			// Either keeping verdict is allowed; this descent must finish under the shared nesting bound.
			t.Logf("G<number> union -> Twin<number> union: same=%v", l.sameKeepingWalk(from, to, map[[2]*checker.Type]bool{}))
		})
	}
}

func TestUnionKeepingFunctionMemberStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			_, err := lowerSourceAs(t, `type Box = { readonly kind: 'box'; readonly label: string };
type A = ((x: number | string) => string) | Box;
type Viewed = ((x: number) => string) | Box;
function read(x: number | string): string { const kind = typeof x; console.log(kind); return kind; }
function make(): A { return read; }
const original: A = make();
const v: Viewed = original;
if (typeof v === 'function') v(1);
`, extension)
			var ny *NotYet
			want := "main" + extension + ":7:19"
			if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) {
				t.Fatalf("((x: number | string) => string) | Box -> ((x: number) => string) | Box: got %v, want NotYet at %s", err, want)
			}
		})
	}
}

func TestUnionKeepingFunctionMemberResultStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			// The top-level scalar-result view also stops at 1:61, so this follows the same rule.
			_, err := lowerSourceAs(t, `type Box = { readonly kind: 'box'; readonly label: string };
type A = (() => number) | Box;
type Viewed = (() => number | string) | Box;
function read(): number { return 1; }
function make(): A { return read; }
const original: A = make();
const v: Viewed = original;
if (typeof v === 'function') console.log(typeof v());
`, extension)
			var ny *NotYet
			want := "main" + extension + ":7:19"
			if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) {
				t.Fatalf("(() => number) | Box -> (() => number | string) | Box: got %v, want NotYet at %s", err, want)
			}
		})
	}
}

// TestUnionKeepingRepeatingLapFinishes pins the open-pair path set (#ttq1599): a union whose member
// is, or holds, a function taking that same union comes back to one storage pair at every lap
// through callableKeeping's ordinary walk, whose union block starts a fresh visited, and its types
// never grow, so the nesting bound never says no. Lower's F and G (on #xmnp0x8) are the member road,
// and K and Twin the field road. Each walk runs under a deadline, so a regression fails this test
// rather than the package.
func TestUnionKeepingRepeatingLapFinishes(t *testing.T) {
	t.Parallel()
	shapes := map[string]string{
		"member": `interface Box { readonly v: number }
type F = ((x: F | number) => string) | Box;
type G = ((x: G | number) => string) | Box;
declare const from: F;
declare const to: G;`,
		"field": `type K = { readonly kind: 'v'; readonly f: (x: K | number) => string } | { readonly kind: 'w' };
type Twin = { readonly kind: 'v'; readonly f: (x: Twin | number) => string } | { readonly kind: 'w' };
declare const from: K;
declare const to: Twin;`,
	}
	for name, source := range shapes {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(name+extension, func(t *testing.T) {
				program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, source)})
				if err != nil {
					t.Fatal(err)
				}
				checked, release := program.Checker(t.Context(), program.Files()[0])
				defer release()
				var from, to *checker.Type
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if ast.IsIdentifier(node) {
						switch node.Text() {
						case "from":
							from = checked.GetTypeAtLocation(node)
						case "to":
							to = checked.GetTypeAtLocation(node)
						}
					}
					return node.ForEachChild(visit)
				}
				program.Files()[0].Node.ForEachChild(visit)
				if from == nil || to == nil {
					t.Fatalf("missing the %s lap's union types", name)
				}
				l := &lowering{program: program, checker: checked}
				done := make(chan bool, 1)
				go func() { done <- l.sameKeepingWalk(from, to, map[[2]*checker.Type]bool{}) }()
				select {
				case same := <-done:
					// The twins keep every slot alike, so the coinductive answer is same.
					if !same {
						t.Fatalf("%s lap: got not same, want same for twins that keep every slot alike", name)
					}
				case <-time.After(60 * time.Second):
					t.Fatalf("%s lap: the keeping walk didn't finish in 60 s", name)
				}
			})
		}
	}
}

func TestUnionKeepingMethodPropertyStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			_, err := lowerSourceAs(t, `import type { Weak } from 'adamic';
class Box {
  readonly label: string;
  constructor(label: string) { this.label = label; }
}
const owner = new Box('c');
function run(value: { readonly kind: 'v'; f(x: number | string): string } | { readonly kind: 'w'; readonly w: number }): void {
  const x: { readonly kind: 'v'; readonly f: (x: number) => string } | { readonly kind: 'w'; readonly w: number } = value;
  if (x.kind === 'v') console.log(x.f(1));
}
run({ kind: 'v', f(x: number | string): string { return typeof x === 'number' ? `+"`n${x}`"+` : x; } });
`, extension)
			var ny *NotYet
			want := "main" + extension + ":8:117"
			wantWhat := `a { readonly kind: "v"; f(x: string | number): string; } | { readonly kind: "w"; readonly w: number; } seen as a { readonly kind: "v"; readonly f: (x: number) => string; } | { readonly kind: "w"; readonly w: number; }`
			if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) || !strings.HasPrefix(ny.What, wantWhat) {
				t.Fatalf("got %v, want %s: stage 0 can't lower %s yet", err, want, wantWhat)
			}
		})
	}
}

func TestUnionKeepingClassMethodPropertyStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			_, err := lowerSourceAs(t, `import type { Weak } from 'adamic';
class Box {
  readonly label: string;
  constructor(label: string) { this.label = label; }
}
const owner = new Box('c');
class M { readonly kind = 'v'; f(x: number | string): string { return typeof x === 'number' ? `+"`n${x}`"+` : x; } }
function run(value: { readonly kind: 'v'; f(x: number | string): string } | { readonly kind: 'w'; readonly w: number }): void {
  const x: { readonly kind: 'v'; readonly f: (x: number) => string } | { readonly kind: 'w'; readonly w: number } = value;
  if (x.kind === 'v') console.log(x.f(1));
}
run(new M());
`, extension)
			var ny *NotYet
			// The view stops first now because the method parameter is held otherwise; the prototype-origin stop at 10:35 caught it later before.
			want := "main" + extension + ":9:117"
			wantWhat := `a { readonly kind: "v"; f(x: string | number): string; } | { readonly kind: "w"; readonly w: number; } seen as a { readonly kind: "v"; readonly f: (x: number) => string; } | { readonly kind: "w"; readonly w: number; }`
			if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) || !strings.HasPrefix(ny.What, wantWhat) {
				t.Fatalf("got %v, want %s: stage 0 can't lower %s yet", err, want, wantWhat)
			}
		})
	}
}

func TestUnionKeepingMethodPropertyAgree(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			lowersAndAgreesWithNodeAs(t, `import type { Weak } from 'adamic';
class Box {
  readonly label: string;
  constructor(label: string) { this.label = label; }
}
const owner = new Box('c');
function run(value: { readonly kind: 'v'; f(x: number): string } | { readonly kind: 'w'; readonly w: number }): void {
  const x: { readonly kind: 'v'; readonly f: (x: number) => string } | { readonly kind: 'w'; readonly w: number } = value;
  if (x.kind === 'v') console.log(x.f(1));
}
run({ kind: 'v', f(x: number): string { return `+"`n${x}`"+`; } });
`, extension)
		})
	}
}

// TestUnionKeepingMethodViewPropertySourceStop pins #3h04df3's narrow stop: a function stored in a
// union member's field, seen through a member that declares it as a method, is a stored closure
// called through its own slots. Its Weakly parameter seen as Strong lowered and, natively, the read
// adapter retained a handle as a target (adamic_retain, measured Nov 16). The reverse keeps the
// method-bivariance refusal it already had.
func TestUnionKeepingMethodViewPropertySourceStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			_, err := lowerSourceAs(t, `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function make(): Strong { return { kind: 'v', v: { label: 'label' } }; }
type Held = { readonly kind: 'read'; readonly read: (x: Weakly) => string } | { readonly kind: 'none' };
type Viewed = { readonly kind: 'read'; read(x: Strong): string } | { readonly kind: 'none' };
function read(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
function hold(): Held { return { kind: 'read', read }; }
const viewed: Viewed = hold();
if (viewed.kind === 'read') console.log(viewed.read(make()));
`, extension)
			var ny *NotYet
			want := "main" + extension + ":10:24"
			if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) {
				t.Fatalf("Held's read: (x: Weakly) => string seen as read(x: Strong): got %v, want NotYet at %s", err, want)
			}
		})
		t.Run("reverse"+extension, func(t *testing.T) {
			_, err := lowerSourceAs(t, `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
const owner: Box = { label: 'label' };
function make(): Weakly { return { kind: 'v', v: owner }; }
type Held = { readonly kind: 'read'; readonly read: (x: Strong) => string } | { readonly kind: 'none' };
type Viewed = { readonly kind: 'read'; read(x: Weakly): string } | { readonly kind: 'none' };
function read(x: Strong): string { return x.kind === 'v' ? x.v.label : 'w'; }
function hold(): Held { return { kind: 'read', read }; }
const viewed: Viewed = hold();
if (viewed.kind === 'read') console.log(viewed.read(make()));
`, extension)
			if err == nil || !strings.Contains(err.Error(), "refuses a function taking Strong seen as one taking Weakly") {
				t.Fatalf("reverse: got %v, want the method-bivariance refusal", err)
			}
		})
	}
}

// t2 differed from Node in the checked runner and in leak-checking mode (#3h04df3).
func TestUnionKeepingPromiseTypeArgumentsStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			_, err := lowerSourceAs(t, `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
function hold(): Promise<Strong> | number { return produce(); }
async function main(): Promise<void> {
  const held: Promise<Strong> | number = hold();
  const viewed: Promise<Weakly> | number = held;
  if (typeof viewed !== 'number') {
    const x = await viewed;
    console.log(show(x));
  }
}
await main();
`, extension)
			var ny *NotYet
			want := "main" + extension + ":10:44"
			if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) {
				t.Fatalf("t2: got %v, want NotYet at %s", err, want)
			}
		})
	}
}

// Language's b01 (no union, the ordinary walk) and b04 (a function's return, reached through
// sameCallablePair and callableKeeping) differed from Node natively on main cf468df0 (#3h04df3).
// Each stops at its view's line.
func TestKeepingPromiseTypeArgumentsOrdinaryStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, line string }{
			{"b01", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
async function main(): Promise<void> {
  const held: Promise<Strong> = produce();
  const viewed: Promise<Weakly> = held;
  console.log(show(await viewed));
}
await main();
`, ":9:"},
			{"b04", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
function hold(): (() => Promise<Strong>) | number { return produce; }
async function main(): Promise<void> {
  const held: (() => Promise<Strong>) | number = hold();
  const viewed: (() => Promise<Weakly>) | number = held;
  if (typeof viewed !== "number") console.log(show(await viewed()));
}
await main();
`, ":10:"},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				_, err := lowerSourceAs(t, row.source, extension)
				var ny *NotYet
				want := "main" + extension + row.line
				if !errors.As(err, &ny) || !strings.Contains(ny.Where, want) {
					t.Fatalf("%s: got %v, want NotYet on line %s", row.name, err, want)
				}
			})
		}
	}
}

func TestUnionKeepingPromiseTypeArgumentsAgree(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			lowersAndAgreesWithNodeAs(t, `interface Box { readonly label: string }
async function produce(): Promise<Box> { return { label: 'label' }; }
function hold(): Promise<Box> | number { return produce(); }
async function main(): Promise<void> {
  const held: Promise<Box> | number = hold();
  const viewed: Promise<Box | undefined> | number = held;
  if (typeof viewed !== 'number') {
    const x = await viewed;
    console.log(x?.label ?? 'none');
  }
}
await main();
`, extension)
		})
	}
}

// Different Promise and PromiseLike targets stay covered by this existing stop (#3h04df3).
func TestUnionKeepingPromiseLikeMemberAnyStop(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source, where string }{
			{"t1", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
function hold(): Promise<Strong> | number { return produce(); }
async function main(): Promise<void> {
  const held: Promise<Strong> | number = hold();
  const viewed: PromiseLike<Weakly> | number = held;
  if (typeof viewed !== 'number') {
    const x = await viewed;
    console.log(show(x));
  }
}
await main();
`, ":10:48"},
			{"t7", `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
function hold(): Promise<Strong> { return produce(); }
async function main(): Promise<void> {
  const held: Promise<Strong> = hold();
  const viewed: PromiseLike<Weakly> = held;
  const x = await viewed;
  console.log(show(x));
}
await main();
`, ":10:39"},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				_, err := lowerSourceAs(t, row.source, extension)
				var ny *NotYet
				want := "main" + extension + row.where
				if !errors.As(err, &ny) || !strings.HasSuffix(ny.Where, want) || !strings.Contains(ny.What, "a member typed any (then()) related to PromiseLike") {
					t.Fatalf("%s: got %v, want the PromiseLike member-typed-any NotYet at %s", row.name, err, want)
				}
			})
		}
	}
}
