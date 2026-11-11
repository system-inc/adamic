package lower

import (
	"errors"
	"strings"
	"testing"
)

// mapObjectKeysAgree lowers source as .ts, agrees with Node on the JavaScript backend, and runs it as
// sanitized native, whose stdout must be want, what Node prints. lowersAndAgreesWithNodeAs runs Node and
// JavaScript only, and a Map's identity lives in native's map.c, so native is run here too.
func mapObjectKeysAgree(t *testing.T, source, want string) {
	t.Helper()
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != want {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want %q, as Node prints", code, stdout, stderr, want)
	}
}

// tsc's tracing.ts keys a Map<object, number> by getRecursionIdentity's answer (recursionIdentityMap,
// tracing.ts:224, 309 and 312 adapted). object is held as a Union, and a key no primitive fits is keyed by
// identity (keyableType, #ff5eddz). Two objects with equal fields are two keys, as SameValueZero says;
// get, has, set over an existing key, delete, size and iteration in insertion order all agree with Node.
// Mutant: map.c's same_key answering equal for any two references, and the equal-fields row reds.
func TestMapObjectKeysByIdentity(t *testing.T) {
	t.Parallel()
	mapObjectKeysAgree(t, `class Point {
	constructor(readonly x: number, readonly y: number) {}
}
function identity(point: Point): object {
	return point;
}
const tokens = new Map<object, number>();
const a = new Point(1, 2);
const b = new Point(1, 2);
for (const point of [a, b, a]) {
	const key = identity(point);
	let token = tokens.get(key);
	if (token === undefined) {
		token = tokens.size;
		tokens.set(key, token);
	}
	console.log(String(token));
}
console.log(String(tokens.size));
console.log(String(tokens.has(a)) + " " + String(tokens.has(new Point(1, 2))));
tokens.set(b, 7);
for (const [key, value] of tokens) {
	console.log(String(key === a) + " " + String(value));
}
console.log(String(tokens.delete(a)) + " " + String(tokens.delete(a)) + " " + String(tokens.size));
`, "0\n1\n0\n2\ntrue false\ntrue 0\nfalse 7\ntrue false 1\n")
}

// The probe Library and Compiler asked for first: the same value put in through one representation and
// looked up through another finds its entry, because converting an Object or an Array into a Union keeps
// its pointer. Each is set through its own type, read back through object, and the reverse: set through
// object and read through its own type. A Map and a function can't be seen as object yet ("an opaque host
// or collection value viewed as unknown or object", "a function viewed as unknown or object"), so neither
// reaches an object key, and each would need this probe when it does.
func TestMapObjectKeysAcrossRepresentations(t *testing.T) {
	t.Parallel()
	mapObjectKeysAgree(t, `interface Named {
	readonly name: string;
}
const named: Named = { name: "n" };
const list: number[] = [1, 2];
const keys = new Map<object, string>();
keys.set(named, "named");
keys.set(list, "list");
const held: object[] = [named, list];
for (const key of held) {
	console.log(keys.get(key) ?? "missing");
}
const back = new Map<object, number>();
for (let index = 0; index < held.length; index++) {
	back.set(held[index]!, index);
}
console.log(String(back.get(named)) + " " + String(back.get(list)));
console.log(String(back.get({ name: "n" })));
`, "named\nlist\n0 1\nundefined\n")
}

// A Set of object holds by identity the same way, since a Set's elements are its map's keys.
func TestSetObjectElementsByIdentity(t *testing.T) {
	t.Parallel()
	mapObjectKeysAgree(t, `const first = { n: 1 };
const second = { n: 1 };
const seen = new Set<object>();
seen.add(first);
seen.add(second);
seen.add(first);
console.log(String(seen.size) + " " + String(seen.has(first)) + " " + String(seen.has({ n: 1 })));
`, "2 true false\n")
}

// A key type a primitive fits stays NotYet, since a Union holding a string or a number would be keyed by
// its box, not its value. That's the wide primitives (object | string, unknown, {}, and { length: number },
// which a string satisfies; its row sets no key, since stage 0 stops a string seen as that type first),
// and every narrower one the wide primitive isn't assignable to: a literal, a boolean literal, an enum
// member, a template literal and a branded string such as tsc's Path, its brand phantom as stage 3 adapts
// tsc's (Compiler's, Lower's and Library's must on 63979123). A type parameter is judged at each
// instance, as what it's instantiated to. Mutants: keyableType admitting any Union reddens every row; onlyObjects dropped, today's
// assignability test alone, reddens the narrow rows.
func TestMapKeysAPrimitiveFitsStayNotYet(t *testing.T) {
	t.Parallel()
	const stop = "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"
	for _, probe := range []struct{ name, source string }{
		{"empty object type", "const keys = new Map<{}, number>();\nkeys.set(5, 1);\nconsole.log(String(keys.size));\n"},
		{"unknown", "const keys = new Map<unknown, number>();\nkeys.set(5, 1);\nconsole.log(String(keys.size));\n"},
		{"object or string", "const keys = new Map<object | string, number>();\nkeys.set(\"k\", 1);\nconsole.log(String(keys.size));\n"},
		{"a structural interface a string fits", "const keys = new Map<{ length: number }, number>();\nconsole.log(String(keys.size));\n"},
		{"object or a string literal", "const keys = new Map<object | \"a\", number>();\nkeys.set(\"a\", 1);\nconsole.log(String(keys.size));\n"},
		{"object or a number literal", "const keys = new Map<object | 1, number>();\nkeys.set(1, 1);\nconsole.log(String(keys.size));\n"},
		{"object or true", "const keys = new Map<object | true, number>();\nkeys.set(true, 1);\nconsole.log(String(keys.size));\n"},
		{"object or an enum member", "enum Kind { A = 1, B = 2 }\nconst keys = new Map<object | Kind, number>();\nkeys.set(Kind.A, 1);\nconsole.log(String(keys.size));\n"},
		{"object or a template literal", "const keys = new Map<object | `a${string}`, number>();\nkeys.set(\"ab\", 1);\nconsole.log(String(keys.size));\n"},
		{"object or a branded string", "type Path = string & { __brand: undefined };\nconst keys = new Map<object | Path, number>();\nconsole.log(String(keys.size));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.Error(), stop) {
				t.Fatalf("want %q, got %v", stop, err)
			}
		})
	}
}

// A union of two interfaces holds only objects, so it's keyed by identity and agrees with Node: the check
// for primitive members must not over-refuse.
func TestMapUnionOfInterfacesKeysByIdentity(t *testing.T) {
	t.Parallel()
	mapObjectKeysAgree(t, `interface Left {
	readonly left: number;
}
interface Right {
	readonly right: string;
}
const left: Left = { left: 1 };
const right: Right = { right: "r" };
const keys = new Map<Left | Right, number>();
keys.set(left, 1);
keys.set(right, 2);
keys.set({ left: 1 }, 3);
console.log(String(keys.get(left)) + " " + String(keys.get(right)) + " " + String(keys.size));
`, "1 2 3\n")
}
