package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

const unionKeepingPrelude = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function make(): Strong { return { kind: 'v', v: { label: 'label' } }; }
`

func TestUnionKeepingPairs(t *testing.T) {
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
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct{ name, source string }{
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

func TestUnionKeepingParameterDirections(t *testing.T) {
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct {
			name, source, types string
			cast                bool
		}{
			{"weak_array_like_function_strong_array_view", `type WeakElements = {readonly kind:'v';readonly v:ArrayLike<Weak<Box>>} | {readonly kind:'w';readonly w:number};
type StrongElements = {readonly kind:'v';readonly v:readonly Box[]} | {readonly kind:'w';readonly w:number};
function read(x: WeakElements): string { return 'label'; }
const viewed: (x: StrongElements) => string = read;`, "(x: WeakElements) => string -> (x: StrongElements) => string", false},
			{"union_weak_array_like_function_strong_array_view", `type WeakElements = {readonly kind:'v';readonly v:ArrayLike<Weak<Box>>} | {readonly kind:'w';readonly w:number};
type StrongElements = {readonly kind:'v';readonly v:readonly Box[]} | {readonly kind:'w';readonly w:number};
type Held = {readonly kind:'read';readonly read:(x:WeakElements)=>string} | {readonly kind:'none'};
type Viewed = {readonly kind:'read';readonly read:(x:StrongElements)=>string} | {readonly kind:'none'};
function read(x: WeakElements): string { return 'label'; }
function hold(): Held { return {kind:'read',read}; }
const viewed: Viewed = hold();`, "Held with (x: WeakElements) => string -> Viewed with (x: StrongElements) => string", false},
			{"weak_function_strong_view", `function read(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
const viewed: (x: Strong) => string = read;
console.log(viewed(make()));`, "(x: Weakly) => string -> (x: Strong) => string", false},
			{"strong_function_weak_view", `function read(x: Strong): string { return x.kind === 'v' ? x.v.label : 'w'; }
const viewed = read as (x: Weakly) => string;
console.log(viewed(make()));`, "(x: Strong) => string -> (x: Weakly) => string", true},
			{"union_weak_function_strong_view", `type Held = {readonly kind:'read';readonly read:(x:Weakly)=>string} | {readonly kind:'none'};
type Viewed = {readonly kind:'read';readonly read:(x:Strong)=>string} | {readonly kind:'none'};
function read(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
function hold(): Held { return {kind:'read',read}; }
const viewed: Viewed = hold();
if (viewed.kind === 'read') console.log(viewed.read(make()));`, "Held with (x: Weakly) => string -> Viewed with (x: Strong) => string", false},
			{"union_strong_function_weak_view", `type Held = {readonly kind:'read';readonly read:(x:Strong)=>string} | {readonly kind:'none'};
type Viewed = {readonly kind:'read';readonly read:(x:Weakly)=>string} | {readonly kind:'none'};
function read(x: Strong): string { return x.kind === 'v' ? x.v.label : 'w'; }
function hold(): Held { return {kind:'read',read}; }
const viewed = hold() as Viewed;
if (viewed.kind === 'read') console.log(viewed.read(make()));`, "Held with (x: Strong) => string -> Viewed with (x: Weakly) => string", true},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				// The reverse function assignment is contravariantly invalid in TypeScript;
				// its explicit cast must stop too, rather than create an unchecked callable view.
				_, err := lowerSourceAs(t, unionKeepingPrelude+row.source, extension)
				var ny *NotYet
				var refused *Refused
				if errors.As(err, &ny) {
					return
				}
				if row.cast && errors.As(err, &refused) && strings.Contains(err.Error(), "adamic/no-unchecked-cast") {
					return
				}
				t.Fatalf("%s: got %v, want the callable view to stop", row.types, err)
			})
		}
	}
}
