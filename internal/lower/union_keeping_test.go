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
