package lower

import (
	"errors"
	"strings"
	"testing"
)

// brandMintPrelude is tsc's sorted-array brands as adaptation 40 types them: a required member typed
// undefined, which phantomBrandMember calls phantom.
const brandMintPrelude = `interface SortedReadonlyArray<T> extends ReadonlyArray<T> { " __sortedArrayBrand": undefined; }
interface SortedArray<T> extends Array<T> { " __sortedArrayBrand": undefined; }
const emptyArray: never[] = [];
`

// tsc's four brand constructions (core.ts:736, 759, 764 and 1039), each spelled as Adamic ruled on
// #b5w3ycg (Nov 5): `x as unknown as Brand<T>`, the one spelling tsc accepts for a required undefined
// brand member. Each mint is admitted at its cast, and then the value stops: a branded array has no
// representation yet, which is the next piece of #yth65gz (the normalization, Compiler's (b) on
// #b5w3ycg). So this head moves no tsc line on its own, and every row must stop NotYet at the value,
// never Refused at the cast. Mutant: brandMintOf returning nil; every row reads "a cast the runtime
// can't check".
func TestBrandMintsOfTscsSortedArraysAreAdmitted(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		brandMintPrelude + "function empty<T>(): SortedReadonlyArray<T> { return emptyArray as unknown as SortedReadonlyArray<T>; }\nconsole.log(String(empty<number>().length));\n",
		brandMintPrelude + "function deduplicated<T>(values: readonly T[]): SortedReadonlyArray<T> { const kept: T[] = []; for (const value of values) { if (kept.length === 0 || value !== kept[kept.length - 1]) kept.push(value); } return kept as unknown as SortedReadonlyArray<T>; }\nconsole.log(String(deduplicated([1, 1, 2]).length));\n",
		brandMintPrelude + "function createSortedArray<T>(): SortedArray<T> { return [] as unknown as SortedArray<T>; }\nconsole.log(String(createSortedArray<number>().length));\n",
		brandMintPrelude + "function toSorted<T>(array: readonly T[], compare: (a: T, b: T) => number): SortedReadonlyArray<T> { return array.slice().sort(compare) as unknown as SortedReadonlyArray<T>; }\nconsole.log(String(toSorted([3, 1], (a, b) => a - b).length));\n",
	} {
		_, err := lowerSourceAs(t, source, ".ts")
		var stopped *NotYet
		if !errors.As(err, &stopped) {
			t.Errorf("%s: got %v, want the cast admitted and the branded value NotYet", source, err)
		}
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
		// Language's probes on the fix: a merged declaration adds the narrowing outside the first one,
		// and an intermediate interface narrows before the brand.
		{"a merged declaration", "interface Base { v: number | string; }\ninterface NumV extends Base { \" __numBrand\": undefined; }\ninterface NumV { v: number; }\nfunction mint(value: Base): NumV { return value as unknown as NumV; }\nconsole.log(typeof mint({ v: 'hello' }).v);\n", ".ts", "a cast the runtime can't check"},
		{"an intermediate narrowing", "interface Base { v: number | string; }\ninterface Mid extends Base { v: number; }\ninterface NumV extends Mid { \" __numBrand\": undefined; }\nfunction mint(value: Base): NumV { return value as unknown as NumV; }\nconsole.log(typeof mint({ v: 'hello' }).v);\n", ".ts", "a cast the runtime can't check"},
		{"an own index entry", "interface Base { v: number; }\ninterface Keyed extends Base { [key: string]: number | undefined; \" __keyBrand\": undefined; }\nfunction mint(value: Base): Keyed { return value as unknown as Keyed; }\nconsole.log(String(mint({ v: 1 }).v));\n", ".ts", "a cast the runtime can't check"},
		{"an own method", "interface Base { v: number; }\ninterface Shown extends Base { show(): string; \" __showBrand\": undefined; }\nfunction mint(value: Base): Shown { return value as unknown as Shown; }\nconsole.log(String(mint({ v: 1 }).v));\n", ".ts", "a cast the runtime can't check"},
		// Language's must on the literal exemption (#b5w3ycg): a named element or a spread can hide an
		// optional field the brand's base reads, so only what the literal itself makes is exempt.
		{"a named element hiding a field", "const wide = { a: 1, b: 2 };\nconst narrow: { a: number } = wide;\ninterface Brand extends Array<{ a: number; b?: string }> { \" __brand\": undefined; }\nconst minted = [narrow] as unknown as Brand;\nconsole.log(typeof minted[0]?.b);\n", ".ts", "a cast the runtime can't check"},
		// Lower's row on 9b16af30: the named twin. The array isn't a literal, so the hidden field sits in a
		// named container's element type, and the array skip of castOptionalFields rests on
		// optionalRelationFailure following it there.
		{"a named array whose element hides a field", "const wide = { a: 1, b: 2 };\nconst xs: { a: number }[] = [wide];\ninterface Brand extends Array<{ a: number; b?: string }> { \" __brand\": undefined; }\nconst minted = xs as unknown as Brand;\nconsole.log(typeof minted[0]?.b);\n", ".ts", "a cast the runtime can't check"},
		{"a spread hiding a field", "const wide = { a: 1, b: 2 };\nconst narrow: { a: number } = wide;\ninterface Shape { a: number; b?: string; }\ninterface ObjBrand extends Shape { \" __brand\": undefined; }\nconst minted = { ...narrow } as unknown as ObjBrand;\nconsole.log(typeof minted.b);\n", ".ts", "a cast the runtime can't check"},
		{"the .a twin", brandMintPrelude + "function createSortedArray(): SortedArray<number> { return [] as unknown as SortedArray<number>; }\nconsole.log(String(createSortedArray().length));\n", ".a", "a brand mint in .a"},
	} {
		_, err := lowerSourceAs(t, row.source, row.extension)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), row.want) {
			t.Errorf("%s: got %v, want a refusal containing %q", row.name, err, row.want)
		}
	}
}

// A value at a cast is judged by optionalRelationFailure with its own origin: an object literal written
// there is exact, so `{ a: 1 }` mints an object brand whose base has an optional b, and lowers whole (an
// object brand is held as an object today). An array literal of objects is judged by its elements and
// mints, then stops at its value as every array brand does until #gb0fqaa. Mutants: the operand not
// passed as origin; the object row refused. castOptionalFields kept for object literals; the object row
// refused.
func TestBrandMintOfALiteralIsExact(t *testing.T) {
	t.Parallel()
	object := "interface Shape { a: number; b?: string; }\ninterface ObjBrand extends Shape { \" __brand\": undefined; }\nconst minted = { a: 1 } as unknown as ObjBrand;\nconsole.log(String(minted.a) + ' ' + typeof minted.b);\n"
	if _, err := lowerSourceAs(t, object, ".ts"); err != nil {
		t.Errorf("an object literal: %v", err)
	}
	array := brandMintPrelude + "const minted = [{ a: 1 }] as unknown as SortedArray<{ a: number; b?: string }>;\nconsole.log(String(minted.length));\n"
	_, err := lowerSourceAs(t, array, ".ts")
	var stopped *NotYet
	if !errors.As(err, &stopped) {
		t.Errorf("an array literal of objects: got %v, want the cast admitted and the branded value NotYet", err)
	}
}
