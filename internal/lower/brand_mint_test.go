package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// brandMintPrelude is tsc's sorted-array brands as adaptation 40 types them: a required member typed
// undefined, which phantomBrandMember calls phantom.
const brandMintPrelude = `interface SortedReadonlyArray<T> extends ReadonlyArray<T> { " __sortedArrayBrand": undefined; }
interface SortedArray<T> extends Array<T> { " __sortedArrayBrand": undefined; }
const emptyArray: never[] = [];
`

// tsc's four brand constructions (core.ts:736, 759, 764 and 1039), each spelled as Adamic ruled on
// #b5w3ycg (Nov 5): `x as unknown as Brand<T>`, the one spelling tsc accepts for a required undefined
// brand member. Each lowers in .ts, judged on the innermost value against the brand's base, and each
// is listed once in Program.BrandMints for the checked report. Mutant: brandMintOf returning nil; every
// row reads "a cast the runtime can't check".
func TestBrandMintsOfTscsSortedArrays(t *testing.T) {
	t.Parallel()
	source := brandMintPrelude + `function deduplicateSorted<T>(array: SortedReadonlyArray<T>): SortedReadonlyArray<T> {
    if (array.length === 0) return emptyArray as unknown as SortedReadonlyArray<T>;
    const deduplicated: T[] = [];
    for (const value of array) {
        if (deduplicated.length === 0 || value !== deduplicated[deduplicated.length - 1]) deduplicated.push(value);
    }
    return deduplicated as unknown as SortedReadonlyArray<T>;
}
function createSortedArray<T>(): SortedArray<T> { return [] as unknown as SortedArray<T>; }
function toSorted<T>(array: readonly T[], compare: (a: T, b: T) => number): SortedReadonlyArray<T> { return array.slice().sort(compare) as unknown as SortedReadonlyArray<T>; }
const sorted = toSorted([3, 1, 2, 1], (a, b) => a - b);
const fresh = createSortedArray<number>();
fresh.push(4);
console.log(deduplicateSorted(sorted).join(',') + ' ' + String(deduplicateSorted(createSortedArray<number>()).length) + ' ' + fresh.join(','));
`
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	targets := map[string]bool{}
	for _, mint := range program.BrandMints {
		targets[mint.Target] = true
		if len(mint.Members) != 1 || mint.Members[0] != " __sortedArrayBrand" {
			t.Errorf("%s: members %q, want the sorted-array brand alone", mint.Where, mint.Members)
		}
	}
	for _, want := range []string{"SortedReadonlyArray<number>", "SortedArray<number>"} {
		if !targets[want] {
			t.Errorf("no brand mint listed for %s: %+v", want, program.BrandMints)
		}
	}
}

// A mint lowers to the value it mints from: the `as unknown` hop makes no value, so nothing is boxed
// (Lower's condition on #b5w3ycg). Mutant: the hop lowered as its own cast; the unknown box appears.
func TestBrandMintIsTheOperandsOwnValue(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, brandMintPrelude+"function brand(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconsole.log(String(brand([1, 2]).length));\n", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	boxed := false
	for _, function := range program.Functions {
		if function.Name != "brand" {
			continue
		}
		walk(function.Body, func(node any) bool {
			if _, ok := node.(ir.Box); ok {
				boxed = true
			}
			return true
		})
	}
	if boxed {
		t.Fatal("the mint boxed its value; the `as unknown` hop must make none")
	}
	if len(program.BrandMints) != 1 {
		t.Fatalf("brand mints %+v, want one", program.BrandMints)
	}
}

// What isn't a pure brand of its operand's base stays as it is on main, refused with today's reason,
// and the mint itself stays refused in .a. Mutants: brandMintOf skipping its every-added-member-is-phantom
// test, so the real-member row lowers; and the declared-member walk replaced by the base's bare
// `continue`, so the redeclaration rows (n01, n07, n08) lower.
func TestBrandMintRecognitionIsExact(t *testing.T) {
	t.Parallel()
	const animals = "class Animal { name = 'animal'; }\nclass Dog extends Animal { bark(): string { return 'woof'; } }\n"
	for _, row := range []struct{ name, source, extension, want string }{
		{"a real member through unknown", brandMintPrelude + "interface Tagged<T> extends Array<T> { tag: string; }\nfunction tag(xs: number[]): Tagged<number> { return xs as unknown as Tagged<number>; }\nconsole.log(String(tag([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"an element narrowing", brandMintPrelude + animals + "function sortDogs(xs: Animal[]): SortedArray<Dog> { return xs as unknown as SortedArray<Dog>; }\nconsole.log(String(sortDogs([new Dog()]).length));\n", ".ts", "a cast the runtime can't check"},
		{"a dropped readonly", brandMintPrelude + "function unsort(xs: readonly number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconsole.log(String(unsort([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"through any", brandMintPrelude + "function viaAny(xs: number[]): SortedArray<number> { return xs as any as SortedArray<number>; }\nconsole.log(String(viaAny([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"an unknown innermost value", brandMintPrelude + "function fromUnknown(xs: unknown): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconsole.log(String(fromUnknown([1]).length));\n", ".ts", "a cast the runtime can't check"},
		// Language's must (#b5w3ycg, Nov 5): a member the base has, redeclared narrower, isn't a mint.
		{"a redeclared base member (n01)", "interface Base { v: number | string; }\ninterface NumV extends Base { v: number; \" __numBrand\": undefined; }\nfunction mint(value: Base): NumV { return value as unknown as NumV; }\nconsole.log(typeof mint({ v: 'hello' }).v);\n", ".ts", "a cast the runtime can't check"},
		{"a redeclared optional member (n07)", "interface Base { v?: number | string; }\ninterface NumV extends Base { v?: number; \" __numBrand\": undefined; }\nfunction mint(value: Base): NumV { return value as unknown as NumV; }\nconsole.log(String(mint({ v: 'x' }).v));\n", ".ts", "a cast the runtime can't check"},
		{"a redeclared generic member (n08)", "interface Box<T> { value: T; }\ninterface NumBox extends Box<number | string> { value: number; \" __numBrand\": undefined; }\nfunction mint(value: Box<number | string>): NumBox { return value as unknown as NumBox; }\nconsole.log(mint({ value: 'x' }).value.toFixed(1));\n", ".ts", "a cast the runtime can't check"},
		{"an own index entry", "interface Base { v: number; }\ninterface Keyed extends Base { [key: string]: number; \" __keyBrand\": undefined; }\nfunction mint(value: Base): Keyed { return value as unknown as Keyed; }\nconsole.log(String(mint({ v: 1 }).v));\n", ".ts", "a cast the runtime can't check"},
		{"an own method", "interface Base { v: number; }\ninterface Shown extends Base { show(): string; \" __showBrand\": undefined; }\nfunction mint(value: Base): Shown { return value as unknown as Shown; }\nconsole.log(String(mint({ v: 1 }).v));\n", ".ts", "a cast the runtime can't check"},
		{"the .a twin", brandMintPrelude + "function createSortedArray(): SortedArray<number> { return [] as unknown as SortedArray<number>; }\nconsole.log(String(createSortedArray().length));\n", ".a", "a brand mint in .a"},
	} {
		_, err := lowerSourceAs(t, row.source, row.extension)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), row.want) {
			t.Errorf("%s: got %v, want a refusal containing %q", row.name, err, row.want)
		}
	}
}
