package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsc's four shapes and a mint over a union element from a mixed literal, run on Node and through the JavaScript backend, which
// must print the same: a branded array's length, elements and methods are its base array's (#gb0fqaa).
// The union row prints by iteration, since join over a union element is a stop on main for any array.
// The native backend is held by the oracle, which has no fixture for these yet (unbranded.mutants.md).
// Mutant: unbranded returning t; each row stops NotYet at the value.
func TestUnbrandedArraysAgreeWithNode(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, brandMintPrelude+`function empty<T>(): SortedReadonlyArray<T> { return emptyArray as unknown as SortedReadonlyArray<T>; }
function deduplicated<T>(values: readonly T[]): SortedReadonlyArray<T> { const kept: T[] = []; for (const value of values) { if (kept.length === 0 || value !== kept[kept.length - 1]) kept.push(value); } return kept as unknown as SortedReadonlyArray<T>; }
function createSortedArray<T>(): SortedArray<T> { return [] as unknown as SortedArray<T>; }
function toSorted<T>(array: readonly T[], compare: (a: T, b: T) => number): SortedReadonlyArray<T> { return array.slice().sort(compare) as unknown as SortedReadonlyArray<T>; }
const none = empty<number>();
console.log(String(none.length));
const unique = deduplicated([1, 1, 2, 3, 3]);
console.log(String(unique.length) + ' ' + unique.join(','));
const created = createSortedArray<number>();
created.push(4);
created.push(9);
console.log(String(created.length) + ' ' + created.join(','));
const sorted = toSorted([3, 1, 2], (a, b) => a - b);
console.log(String(sorted.length) + ' ' + sorted.join(',') + ' ' + String(sorted[0]));
const mixed = [0, 'one'] as unknown as SortedArray<number | string>;
mixed.push('two');
let parts = '';
for (const value of mixed) {
	parts += String(value) + ';';
}
console.log(String(mixed.length) + ' ' + parts);
let total = 0;
for (const value of sorted) {
	total += value;
}
console.log(String(total));
`, ".ts")
}

// A literal narrower than the brand's union element, `[0, 1] as unknown as SortedArray<number | string>`,
// is refused as `const xs: (number | string)[] = [0, 1]` is, now the brand's slot is seen as its base
// array: before #gb0fqaa it was admitted at the cast and stopped NotYet at its value.
func TestUnbrandedNarrowLiteralMintIsRefused(t *testing.T) {
	t.Parallel()
	const want = "union-typed array built from a narrower literal"
	_, err := lowerSourceAs(t, brandMintPrelude+"const mixed = [0, 1] as unknown as SortedArray<number | string>;\nconsole.log(String(mixed.length));\n", ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Error(), want) {
		t.Fatalf("got %v, want a refusal containing %q", err, want)
	}
}

// A mint is the value it mints from: `xs as unknown as SortedArray<number>` declares sorted with a read of
// xs, held as the array it is, never boxed or copied (brandMintValue, and the brand's base as the slot's
// representation). Behavior can't see this: a Box would print the same. Mutant: unbranded returning t;
// the program stops NotYet at sorted's value.
func TestUnbrandedMintLowersToItsOperand(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, brandMintPrelude+"const xs: number[] = [1, 2];\nconst sorted = xs as unknown as SortedArray<number>;\nconsole.log(String(sorted.length));\n", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	xs, sorted := -1, -1
	for index, local := range program.Locals {
		switch local.Name {
		case "xs":
			xs = index
		case "sorted":
			sorted = index
		}
	}
	if xs < 0 || sorted < 0 {
		t.Fatalf("no locals xs and sorted in %v", program.Locals)
	}
	for _, statement := range program.Main {
		var value ir.Expression
		switch statement := statement.(type) {
		case ir.Declare:
			if statement.Local != sorted {
				continue
			}
			value = statement.Value
		case ir.Assign:
			if statement.Local != sorted {
				continue
			}
			value = statement.Value
		default:
			continue
		}
		if read, isRead := value.(ir.Read); !isRead || read.Local != xs || read.Of != ir.Array {
			t.Fatalf("sorted is given %#v, want a read of xs held as an Array", value)
		}
		return
	}
	t.Fatalf("sorted is never given a value in %v", program.Main)
}

// The brand stays the checker's: a number[] where a SortedArray<number> is wanted, with no mint, is a type
// error, so a plain array never passes for a sorted one.
func TestUnbrandedPlainArrayIsNoBrand(t *testing.T) {
	t.Parallel()
	source := brandMintPrelude + "function first(xs: SortedArray<number>): number { return xs.length; }\nconst plain: number[] = [2, 1];\nconsole.log(String(first(plain)));\n"
	_, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err == nil || !strings.Contains(err.Error(), "__sortedArrayBrand") {
		t.Fatalf("got %v, want the checker's error naming the missing brand member", err)
	}
}

// Only an interface whose own members are all phantom is held as its base. One with a real member, a
// method or an own index entry keeps today's stop at its value, as does a brand whose type arguments
// aren't its base's, whose elements lowering would read off the brand's own arguments (unbranded.go).
// Mutant: phantomExtension admitting a real member (phantomBrandMember's test dropped); the real-member
// row lowers.
func TestUnbrandedOnlyPhantomExtensions(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, declaration, typed string }{
		{"a real member", "interface Tagged<T> extends Array<T> { tag: string; }", "Tagged<number>"},
		{"a method", "interface Shown<T> extends Array<T> { show(): string; }", "Shown<number>"},
		{"an own index entry", "interface Keyed<T> extends Array<T> { [key: string]: unknown; \" __keyedBrand\": undefined; }", "Keyed<number>"},
		{"other type arguments", "interface NumberBrand extends Array<number> { \" __numberBrand\": undefined; }", "NumberBrand"},
	} {
		source := row.declaration + "\nfunction count(xs: " + row.typed + " | undefined): number { return xs === undefined ? 0 : xs.length; }\nconsole.log(String(count(undefined)));\n"
		_, err := lowerSourceAs(t, source, ".ts")
		var stopped *NotYet
		if !errors.As(err, &stopped) {
			t.Errorf("%s: got %v, want NotYet at the value", row.name, err)
		}
	}
}

// sortedValue is a minted SortedArray<number>, and markedValue a brand whose member is a plain name.
const sortedValue = brandMintPrelude + "const sorted = [1, 2, 3] as unknown as SortedArray<number>;\n"
const markedValue = "interface Marked<T> extends Array<T> { __marked: undefined; }\nconst marked = [1, 2] as unknown as Marked<number>;\n"

// Every road rulings.md:31 names that observes a brand member's presence is refused at its node, on the
// brand, before lowering reads the value as its base (phantom_array_presence.go). b09's union `in` is
// refused by this rule, naming the brand, not by inProperty's union refusal. Mutants, each road's clause
// removed in turn (unbranded.mutants.md): its rows here go red.
func TestUnbrandedPresenceIsRefused(t *testing.T) {
	t.Parallel()
	const sortedMember = "an observation of SortedArray<number>'s phantom member  __sortedArrayBrand"
	for _, row := range []struct{ name, source, want string }{
		{"in, a literal key", sortedValue + "console.log(String(' __sortedArrayBrand' in sorted));\n", sortedMember + " (an `in` test)"},
		{"in, a key not a literal", sortedValue + "const key: string = 'x';\nconsole.log(String(key in sorted));\n", sortedMember + " (an `in` test)"},
		{"in over b09's union", brandMintPrelude + "function size(x: SortedArray<number> | { n: number }): number { if (' __sortedArrayBrand' in x) { return x.length; } return x.n; }\nconsole.log(String(size({ n: 2 })));\n", "an observation of SortedArray<number>'s phantom member  __sortedArrayBrand (an `in` test)"},
		{"Object.hasOwn", sortedValue + "console.log(String(Object.hasOwn(sorted, ' __sortedArrayBrand')));\n", sortedMember + " (Object.hasOwn)"},
		{"hasOwnProperty", sortedValue + "console.log(String(sorted.hasOwnProperty(' __sortedArrayBrand')));\n", sortedMember + " (hasOwnProperty)"},
		{"Object.keys", sortedValue + "console.log(String(Object.keys(sorted).length));\n", sortedMember + " (Object.keys)"},
		{"Object.values", sortedValue + "console.log(String(Object.values(sorted).length));\n", sortedMember + " (Object.values)"},
		{"Object.entries", sortedValue + "console.log(String(Object.entries(sorted).length));\n", sortedMember + " (Object.entries)"},
		{"Object.getOwnPropertyNames", sortedValue + "console.log(String(Object.getOwnPropertyNames(sorted).length));\n", sortedMember + " (Object.getOwnPropertyNames)"},
		{"Object.assign, a source", sortedValue + "console.log(String(Object.assign({}, sorted).length));\n", sortedMember + " (Object.assign)"},
		{"Object.assign, the target", sortedValue + "console.log(String(Object.assign(sorted, { n: 1 }).n));\n", sortedMember + " (Object.assign)"},
		{"JSON.stringify", sortedValue + "console.log(String(JSON.stringify(sorted)));\n", sortedMember + " (JSON.stringify)"},
		{"an object spread", sortedValue + "const copy = { ...sorted };\nconsole.log(String(copy.length));\n", sortedMember + " (an object spread)"},
		{"an element write, a literal key", sortedValue + "sorted[' __sortedArrayBrand'] = undefined;\nconsole.log(String(sorted.length));\n", sortedMember + " (a write)"},
		{"an element write, a key not a literal", sortedValue + "const key: ' __sortedArrayBrand' = ' __sortedArrayBrand';\nsorted[key] = undefined;\nconsole.log(String(sorted.length));\n", sortedMember + " (a write)"},
		{"propertyIsEnumerable, Language's l04", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconsole.log(String(s.propertyIsEnumerable(\" __sortedArrayBrand\")));\n", "an observation of SortedArray<number>'s phantom member  __sortedArrayBrand (propertyIsEnumerable)"},
		{"a program's own Object member, which no list names", "interface Object { probe(key: string): boolean; }\ninterface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconsole.log(String(s.probe(\" __sortedArrayBrand\")));\n", "an observation of SortedArray<number>'s phantom member  __sortedArrayBrand (probe)"},
		{"hasOwnProperty, a key the brand doesn't add", sortedValue + "console.log(String(sorted.hasOwnProperty('length')));\n", sortedMember + " (hasOwnProperty)"},
		{"in through a type parameter's union constraint, Language's l13", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nfunction pick<T extends SortedArray<number> | { n: number }>(x: T): number { if (\" __sortedArrayBrand\" in x) return x.length; return x.n; }\nconsole.log(String(pick(s)));\n", "an observation of SortedArray<number>'s phantom member  __sortedArrayBrand (an `in` test)"},
		{"in through a union with a constrained type parameter, Language's l14", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nfunction pick<T extends SortedArray<number>>(x: T | { n: number }): number { if (\" __sortedArrayBrand\" in x) return x.length; return x.n; }\nconsole.log(String(pick(s)));\n", "an observation of SortedArray<number>'s phantom member  __sortedArrayBrand (an `in` test)"},
		{"a property write", markedValue + "marked.__marked = undefined;\nconsole.log(String(marked.length));\n", "an observation of Marked<number>'s phantom member __marked (a write)"},
	} {
		_, err := lowerSourceAs(t, row.source, ".ts")
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Error(), row.want) {
			t.Errorf("%s: got %v, want a refusal containing %q", row.name, err, row.want)
		}
	}
}

// What observes no brand member is left to lowering as it was: a literal key the brand doesn't add, an
// array spread, which iterates elements, and a write at a number. Mutants: each road's key rule widened
// to refuse every key; its control here goes red.
func TestUnbrandedPresenceControls(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"in, a key the brand doesn't add", sortedValue + "console.log(String('length' in sorted));\n"},
		{"an array spread", sortedValue + "const copy = [...sorted];\nconsole.log(String(copy.length));\n"},
		{"Object.hasOwn, a key the brand doesn't add", sortedValue + "console.log(String(Object.hasOwn(sorted, 'length')));\n"},
		{"an element write at a number", sortedValue + "sorted[0] = 5;\nconsole.log(String(sorted[0]));\n"},
	} {
		_, err := lowerSourceAs(t, row.source, ".ts")
		var refused *Refused
		if errors.As(err, &refused) && strings.Contains(refused.Error(), "phantom member") {
			t.Errorf("%s: refused as an observation of a brand member: %v", row.name, err)
		}
	}
}

// A brand still typed void fails closed under the void-container rule (rulings.md, Oct 29), so unbranded
// keeps it as it is and it never gains its base's layout: by length alone it stops NotYet at its value, as
// every branded array did before #gb0fqaa, and a read of its elements is the void-container rule's
// refusal. Neither lowers whole. Mutant: unbranded's void guard removed; the length row lowers.
func TestUnbrandedVoidBrandStops(t *testing.T) {
	t.Parallel()
	const declaration = "interface VoidSorted<T> extends Array<T> { \" __b\": void; }\n"
	for _, row := range []struct {
		name, body string
		refused    bool
		want       string
	}{
		{"its length", "return xs === undefined ? 0 : xs.length;", false, "a value of type VoidSorted<number>"},
		{"its elements read", "return xs === undefined ? 0 : xs.join(',').length;", true, "a read of a void container's elements"},
	} {
		source := declaration + "function count(xs: VoidSorted<number> | undefined): number { " + row.body + " }\nconsole.log(String(count(undefined)));\nconsole.log(String(count([1, 2] as unknown as VoidSorted<number>)));\n"
		_, err := lowerSourceAs(t, source, ".ts")
		t.Logf("%s: %v", row.name, err)
		var stopped *NotYet
		var refused *Refused
		switch {
		case row.refused && (!errors.As(err, &refused) || !strings.Contains(refused.Error(), row.want)):
			t.Errorf("%s: got %v, want a refusal containing %q", row.name, err, row.want)
		case !row.refused && (!errors.As(err, &stopped) || !strings.Contains(stopped.Error(), row.want)):
			t.Errorf("%s: got %v, want NotYet containing %q", row.name, err, row.want)
		}
	}
}

// A cast is judged on the brand (Compiler's condition on #gb0fqaa), so castAssignableAt's array skip never
// takes a phantom extension for an array, though isLibraryType now unbrands and lowering holds its value
// as its base. No program tells the two apart today, since a mint is judged against its unbranded base,
// so this asks arrayLikeType of each parameter's type directly. Mutant: arrayLikeType's phantomExtension
// guard removed; SortedReadonlyArray reads as an array, through isLibraryType's unbranding.
func TestUnbrandedCastSeesTheBrand(t *testing.T) {
	t.Parallel()
	source := brandMintPrelude + "function kinds(sortedReadonly: SortedReadonlyArray<number>, sorted: SortedArray<number>, readonlyPlain: readonly number[], plain: number[]): number { return sortedReadonly.length + sorted.length + readonlyPlain.length + plain.length; }\nconsole.log(String(kinds([] as unknown as SortedReadonlyArray<number>, [] as unknown as SortedArray<number>, [], [])));\n"
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked}
	want := map[string]bool{"sortedReadonly": false, "sorted": false, "readonlyPlain": true, "plain": true}
	seen := 0
	for _, statement := range program.Files()[0].Statements.Nodes {
		if statement.Kind != ast.KindFunctionDeclaration || statement.Name() == nil || statement.Name().Text() != "kinds" {
			continue
		}
		for _, parameter := range statement.Parameters() {
			name := parameter.Name().Text()
			seen++
			if got := l.arrayLikeType(checked.GetTypeAtLocation(parameter.Name())); got != want[name] {
				t.Errorf("arrayLikeType of %s's type is %v, want %v", name, got, want[name])
			}
		}
	}
	if seen != len(want) {
		t.Fatalf("judged %d parameters, want %d", seen, len(want))
	}
}

// Roads rulings.md:31 names that rule 450 doesn't refuse, because each already stops at a later rule, held
// here so the head that lifts one of those stops meets this test first (Language's attack on #gb0fqaa). Each
// row must not lower; its present stop is pinned by its words.
func TestUnbrandedPresenceHeldElsewhere(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, want string }{
		{"for...in, b08", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nlet k = \"\"; for (const each in s) k += each; console.log(k);\n", "for...in over an array (holes and own enumerable properties are not represented; use for...of for elements)"},
		{"Reflect.has, b13", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconsole.log(String(Reflect.has(s, \" __sortedArrayBrand\")));\n", "a namespace member without a lowered binding"},
		{"Reflect.ownKeys, l06", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconsole.log(String(Reflect.ownKeys(s).length));\n", "a namespace member without a lowered binding"},
		{"Reflect.get, l05", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconsole.log(String(Reflect.get(s, \" __sortedArrayBrand\")));\n", "Reflect.get other than Reflect.get(Object(value), key, value) with one identifier as value"},
		{"destructuring the brand member, b14", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconst { \" __sortedArrayBrand\": mark } = s; console.log(String(mark));\n", "an array going where a plain object type is held"},
		{"destructuring the brand member into a write, l08", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nlet mark: undefined; ({ \" __sortedArrayBrand\": mark } = s); console.log(String(mark));\n", "an array going where a plain object type is held"},
		{"Object.keys through object, l10", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nconst o: object = s;\nconsole.log(Object.keys(o).join(\",\"));\n", "Object.keys on a non-object"},
		{"Object.keys through a generic, l11", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\nfunction keysOf<T extends object>(x: T): string[] { return Object.keys(x); }\nconsole.log(keysOf(s).join(\",\"));\n", "Object.keys on a non-object"},
		{"delete, l07", "interface SortedArray<T> extends Array<T> { \" __sortedArrayBrand\": undefined; }\ninterface SortedReadonlyArray<T> extends ReadonlyArray<T> { readonly \" __sortedArrayBrand\": undefined; }\nfunction mint(xs: number[]): SortedArray<number> { return xs as unknown as SortedArray<number>; }\nconst s = mint([3, 1]);\ndelete (s as { \" __sortedArrayBrand\"?: undefined })[\" __sortedArrayBrand\"]; console.log(\"deleted\");\n", "Adamic 0.1 refuses delete; fixed objects cannot lose fields"},
	} {
		_, err := lowerSourceAs(t, row.source, ".ts")
		if err == nil {
			t.Errorf("%s: lowers, but it observes the brand member and must stop", row.name)
			continue
		}
		t.Logf("%s: %v", row.name, err)
		if !strings.Contains(err.Error(), row.want) {
			t.Errorf("%s: got %v, want a stop containing %q", row.name, err, row.want)
		}
	}
}
