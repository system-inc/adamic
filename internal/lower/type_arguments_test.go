package lower

import (
	"errors"
	"strings"
	"testing"
)

// Only a type reference has type arguments, and the checker dereferences nil asking anything else for
// them. Each site here asked a type a program writes, a union or a string, and panicked (#egxccg8, the
// sweep after the rest parameter's): a destructuring declaration judged for nominal ancestry (rule
// 3400, now member by member through a union, a primitive passing and anything else NotYet), and
// Object.entries's pairs written into a wider array.
func TestTypeArgumentsAreReadOnlyFromReferences(test *testing.T) {
	test.Parallel()
	const pick = "function pick(flag: boolean): [string, number] | [number, string] {\n    return flag ? [\"a\", 1] : [1, \"a\"];\n}\n"
	for _, probe := range []struct{ name, source, where, what string }{
		// Rule 3400 lets these by, and lowering, which destructures only a tuple, stops at the pattern.
		{
			"destructured union of tuples",
			"class Box {}\n" +
				pick +
				"const [a, b]: [string, number] | [number, string] = pick(true);\nconsole.log(`${a}${b}`);\n",
			":5:7",
			"NotYet destructuring anything but a tuple into [names]",
		},
		{
			"destructured union whose members all pass",
			"class Box {\n    value = 1;\n}\nfunction pick(flag: boolean): [Box, number] | [Box, string] {\n    return flag ? [new Box(), 1] : [new Box(), \"a\"];\n}\nconst [box, n]: [Box, number] | [Box, string] = pick(true);\nconsole.log(`${box.value}${n}`);\n",
			":7:7",
			"NotYet destructuring anything but a tuple into [names]",
		},
		{
			"destructured string",
			"class Box {}\nconst [c, d]: string = \"xy\";\nconsole.log(`${c}${d}`);\n",
			":2:7",
			"NotYet destructuring anything but a tuple into [names]",
		},
		// A member that isn't a reference or a primitive has elements rule 3400 can't read, so it's
		// NotYet at the pattern, never passed unjudged.
		{
			"destructured intersection",
			"class Box {}\nfunction first(tagged: string[] & { tag: string }): string {\n    const [x]: string[] & { tag: string } = tagged;\n    return x ?? \"none\";\n}\nconsole.log(\"one\");\n",
			":3:11",
			"NotYet a destructured string[] & { tag: string; }, whose elements can't be judged for nominal ancestry",
		},
		{
			"destructured intersection of tuples",
			"class Animal {\n    name = \"a\";\n}\nfunction split(pair: [Animal, Animal] & { tag: 1 }): string {\n    const [a, b]: [Animal, Animal] & { tag: 1 } = pair;\n    return a.name + b.name;\n}\nconsole.log(\"one\");\n",
			":5:11",
			"NotYet a destructured [Animal, Animal] & { tag: 1; }, whose elements can't be judged for nominal ancestry",
		},
		// A union of tuples is still judged for nominal ancestry, member by member, as one tuple is.
		{
			"destructured union seen as a class",
			"class Box {\n    value = 1;\n}\nfunction plain(flag: boolean): [{ value: number }] | [{ value: number }, number] {\n    return flag ? [{ value: 2 }] : [{ value: 3 }, 4];\n}\nconst [box]: [Box] | [Box, number] = plain(true);\nconsole.log(`${box.value}`);\n",
			":7:8",
			"Refused a destructured value without nominal ancestry seen as Box",
		},
		// An optional tuple destructured after its check is the tuple arm, judged as one tuple is: a
		// plain object seen as a Dog refuses, and Dogs lower.
		{
			"optional tuple seen as a class",
			"class Dog {\n    bark = 1;\n}\nfunction find(flag: boolean): [{ bark: number }, { bark: number }] | undefined {\n    return flag ? [{ bark: 1 }, { bark: 2 }] : undefined;\n}\nconst found = find(true);\nif (found !== undefined) {\n    const [first, second]: [Dog, Dog] = found;\n    console.log(`${first.bark}${second.bark}`);\n}\n",
			":9:12",
			"Refused a destructured value without nominal ancestry seen as Dog",
		},
		{
			"optional tuple of the class",
			"class Dog {\n    bark = 1;\n}\nfunction find(flag: boolean): [Dog, Dog] | undefined {\n    return flag ? [new Dog(), new Dog()] : undefined;\n}\nconst found = find(true);\nif (found !== undefined) {\n    const [first, second]: [Dog, Dog] = found;\n    console.log(`${first.bark}${second.bark}`);\n}\n",
			"",
			"lowers",
		},
		{
			"entries seen as unknown",
			"const shape = { a: 1, b: 2 };\nconst pairs: readonly unknown[] = Object.entries(shape);\nconsole.log(`${pairs.length}`);\n",
			":2:35",
			"Refused Object.entries: tsc's entries must have a proven value type",
		},
		{
			"entries seen as a union of tuples",
			"const shape = { a: 1, b: 2 };\nconst pairs: readonly ([string, number] | [string, string])[] = Object.entries(shape);\nconsole.log(`${pairs.length}`);\n",
			":2:65",
			"Refused Object.entries: tsc's entries must have a proven value type",
		},
		// tsc lets entries' strong pairs into an element of Weak<[string, V]>, so the pairs are never read
		// through a Weak: Weak<T> is a union with undefined, and the brand's intersection isn't a
		// reference either. Read through the Weak, the branded pairs went on to the relation's NotYet.
		{
			"entries seen as weak",
			"import type { Weak } from 'adamic';\nconst shape = { a: 1, b: 2 };\nconst pairs: readonly Weak<[string, number]>[] = Object.entries(shape);\nconsole.log(`${pairs.length}`);\n",
			":3:50",
			"Refused Object.entries: tsc's entries must have a proven value type",
		},
		{
			"entries seen as weakly branded",
			"import type { WeakBrand } from 'adamic';\nconst shape = { a: 1, b: 2 };\nconst pairs: readonly ([string, number] & WeakBrand)[] = Object.entries(shape);\nconsole.log(`${pairs.length}`);\n",
			":3:58",
			"Refused Object.entries: tsc's entries must have a proven value type",
		},
	} {
		for _, extension := range []string{".ts", ".a"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				defer func() {
					if failure := recover(); failure != nil {
						test.Fatalf("panicked: %v", failure)
					}
				}()
				_, lowerSourceAsError := lowerSourceAs(test, probe.source, extension)
				var notYet *NotYet
				var refused *Refused
				got, where := "lowers", ""
				switch {
				case errors.As(lowerSourceAsError, &notYet):
					got, where = "NotYet "+notYet.What, notYet.Where
				case errors.As(lowerSourceAsError, &refused) && strings.HasPrefix(refused.What, "Object."):
					got, where = "Refused "+refused.What+": "+refused.Fix, refused.Where
				case errors.As(lowerSourceAsError, &refused):
					got, where = "Refused "+refused.What, refused.Where
				case lowerSourceAsError != nil:
					got = lowerSourceAsError.Error()
				}
				if got != probe.what || !strings.HasSuffix(where, probe.where) {
					test.Fatalf("want %s at %s, got %v", probe.what, probe.where, lowerSourceAsError)
				}
			})
		}
	}
}
