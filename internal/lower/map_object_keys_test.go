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
// looked up through another finds its entry, because converting an Object, an Array or a Closure into a
// Union keeps its pointer. Each is set through its own type, read back through object, and the reverse:
// set through object and read through its own type. A Map can't be seen as object yet ("an opaque host or
// collection value viewed as unknown or object"), so no Map reaches an object key.
func TestMapObjectKeysAcrossRepresentations(t *testing.T) {
	t.Parallel()
	mapObjectKeysAgree(t, `interface Named {
	readonly name: string;
}
const named: Named = { name: "n" };
const list: number[] = [1, 2];
const callback = (value: number): number => value + 1;
const keys = new Map<object, string>();
keys.set(named, "named");
keys.set(list, "list");
keys.set(callback, "callback");
const held: object[] = [named, list, callback];
for (const key of held) {
	console.log(keys.get(key) ?? "missing");
}
const back = new Map<object, number>();
for (let index = 0; index < held.length; index++) {
	back.set(held[index]!, index);
}
console.log(String(back.get(named)) + " " + String(back.get(list)) + " " + String(back.get(callback)));
console.log(String(back.get({ name: "n" })));
`, "named\nlist\ncallback\n0 1 2\nundefined\n")
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
// its box, not its value: {}, unknown and object | string. A type parameter is judged at each instance, as what it's instantiated to. Mutant: keyableType
// admitting any Union, and these lower.
func TestMapKeysAPrimitiveFitsStayNotYet(t *testing.T) {
	t.Parallel()
	const stop = "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"
	for _, probe := range []struct{ name, source string }{
		{"empty object type", "const keys = new Map<{}, number>();\nkeys.set(5, 1);\nconsole.log(String(keys.size));\n"},
		{"unknown", "const keys = new Map<unknown, number>();\nkeys.set(5, 1);\nconsole.log(String(keys.size));\n"},
		{"object or string", "const keys = new Map<object | string, number>();\nkeys.set(\"k\", 1);\nconsole.log(String(keys.size));\n"},
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
