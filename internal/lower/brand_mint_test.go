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
    const deduplicated: T[] = [array[0]];
    for (let index = 1; index < array.length; index++) {
        if (array[index] !== deduplicated[deduplicated.length - 1]) deduplicated.push(array[index]);
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
// and the mint itself stays refused in .a. Mutant: brandMintOf skipping its every-added-member-is-phantom
// test; the real-member row lowers.
func TestBrandMintRecognitionIsExact(t *testing.T) {
	t.Parallel()
	const animals = "class Animal { name = 'animal'; }\nclass Dog extends Animal { bark(): string { return 'woof'; } }\n"
	for _, row := range []struct{ name, source, extension, want string }{
		{"a real member through unknown", brandMintPrelude + "interface Tagged<T> extends Array<T> { tag: string; }\nfunction tag(xs: number[]): Tagged<number> { return xs as unknown as Tagged<number>; }\nconsole.log(String(tag([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"an element narrowing", brandMintPrelude + animals + "function sortDogs(xs: Animal[]): SortedArray<Dog> { return xs as unknown as SortedArray<Dog>; }\nconsole.log(String(sortDogs([new Dog()]).length));\n", ".ts", "a cast the runtime can't check"},
		{"a dropped readonly", brandMintPrelude + "function unsort(xs: readonly number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconsole.log(String(unsort([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"through any", brandMintPrelude + "function viaAny(xs: number[]): SortedArray<number> { return xs as any as SortedArray<number>; }\nconsole.log(String(viaAny([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"an unknown innermost value", brandMintPrelude + "function fromUnknown(xs: unknown): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconsole.log(String(fromUnknown([1]).length));\n", ".ts", "a cast the runtime can't check"},
		{"the .a twin", brandMintPrelude + "function createSortedArray(): SortedArray<number> { return [] as unknown as SortedArray<number>; }\nconsole.log(String(createSortedArray().length));\n", ".a", "a brand mint in .a"},
	} {
		_, err := lowerSourceAs(t, row.source, row.extension)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), row.want) {
			t.Errorf("%s: got %v, want a refusal containing %q", row.name, err, row.want)
		}
	}
}
