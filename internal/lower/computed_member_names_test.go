package lower

import (
	"fmt"
	"strings"
	"testing"
)

// lowerWithoutPanic lowers source and turns a panic into a test failure, so a walk that asks a computed
// member name for its text fails by name, not by killing the package's tests.
func lowerWithoutPanic(t *testing.T, source string) (err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("lowering panicked: %v", recovered)
		}
	}()
	_, err = lowerSourceAs(t, source, ".ts")
	return err
}

// A member whose computed name spells no text (a well-known symbol, a non-constant key) or spells one
// through a computed name (["x"], [1]) is passed over by every walk over the whole program that asks
// which strings name a getter or a method, and refused where it lowers (#ppgk12h: a Go panic in
// noteAccessorNames on tsc's harness, SortedMap's get [Symbol.toStringTag]()).
func TestComputedMemberNamesStopWithoutPanic(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, want string }{
		{"the harness's SortedMap", `class SortedMap {
    private _keys: string[] = [];
    public get size(): number { return this._keys.length; }
    public get [Symbol.toStringTag]() { return "SortedMap"; }
}
console.log(String(new SortedMap().size));
`, "a method with a computed name"},
		{"a string literal getter in a class", "class A { get [\"x\"]() { return 1; } }\nconsole.log(String(new A().x));\n", "a method with a computed name"},
		{"a numeric literal getter in a class", "class A { get [1]() { return 2; } }\nconsole.log(String(new A()[1]));\n", "a method with a computed name"},
		{"a non-constant getter in a class", "const k = \"x\" as string;\nclass A { get [k]() { return 1; } }\nconsole.log(String(new A() !== undefined));\n", "a method with a computed name"},
		{"a non-constant getter in a literal", "const k = \"x\" as string;\nconst o = { get [k]() { return 1; } };\nconsole.log(String(o !== undefined));\n", "a record literal member other than a data entry"}, // A non-constant key types the literal as a record.
		{"a symbol getter in a literal", "const o = { n: 1, get [Symbol.toStringTag]() { return \"o\"; } };\nconsole.log(String(o.n));\n", "a computed accessor literal name"},
		{"a bind beside a symbol getter", `const o = { m() { return 1; } };
const f = o.m.bind(o);
console.log(String(f()));
const p = { n: 1, get [Symbol.toStringTag]() { return "p"; } };
console.log(String(p.n));
`, "a computed accessor literal name"},
		// unknown.go's member scans: in on an unknown asks presenceHazard, and a read of the narrowed
		// unknown asks dynamicReadHazard, each walking every class's members, the iterable one's included.
		{"in and a read of an unknown beside an iterable class", `class C { [Symbol.iterator]() { return [1][Symbol.iterator](); } }
function show(u: unknown): string {
    if (typeof u === "object" && u !== null && "a" in u) {
        return String(u.a);
    }
    return "none";
}
console.log(show({ a: 1 }));
`, ""},
		{"a read of an unknown field beside an iterable class", `class C { [Symbol.iterator]() { return [1][Symbol.iterator](); } }
interface Holder { value: unknown }
function read(holder: Holder): string {
    const value = holder.value;
    if (typeof value === "object" && value !== null && "a" in value) {
        return String(value.a);
    }
    return "none";
}
console.log(read({ value: { a: 1 } }));
`, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := lowerWithoutPanic(t, test.source)
			t.Logf("lowering stops at: %v", err)
			if test.want == "" {
				return // Reaching the end without a panic is the pin; lowering may stop or not.
			}
			if err == nil || !strings.Contains(fmt.Sprint(err), test.want) {
				t.Fatalf("got %v, want a stop naming %q", err, test.want)
			}
		})
	}
}
