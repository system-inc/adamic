package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"

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
			{"weak_member_control", `import type { Weak } from 'adamic';
interface A { readonly kind: 'a'; readonly label: string }
interface C { readonly kind: 'c'; readonly label: string }
interface Held { readonly item: Weak<A> | C }
interface Twin { readonly item: Weak<A> | { readonly kind: 'c'; readonly label: string } }
function view(held: Held): Twin { const same: Held = held; const twin: Twin = same; return twin; }
console.log('control');`},
			{"promise_union_view", `const p: string | PromiseLike<string> = Promise.resolve('x'); console.log(typeof p);`},
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
