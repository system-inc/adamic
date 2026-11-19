package javascript

import "testing"

// A checked field read tests its receiver's own kind before it looks the name up, as native's adamic_object_view does
// (internal/native/object_view_receiver_test.go), so both backends stop a receiver that isn't an object in the same
// words. A string has own length and index keys, so without the test a widened .length read through a type a view
// checks read on in JavaScript where native stopped with found string (Emit on #zpzm9ta). An object, a tuple and a
// class read through its constructor stay readable. Mutant: adamicViewField without its receiver line, and every row
// goes red, the Map row stopping as found missing and the rest reading on; the closure test dropped from
// adamicViewReceiver, and the closure row reads its length.
func TestViewFieldStopsAtANonObjectReceiver(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, receiver, found string }{
		{"string", `"abcd"`, "string"},
		{"array", `[1, 2, 3]`, "array"},
		{"Map", `new Map([["length", 3]])`, "Map"},
		{"function", `function f(a, b, c) {}`, "function"},
		// Every closure the backend makes is an AdamicClosure, an object with an own length (Emit on #zpzm9ta).
		{"closure", `(() => { const closure = new AdamicClosure(() => 0); closure.length = 3; return closure; })()`, "function"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			source := viewTestRuntime + fieldReadinessRuntime + "console.log(String(adamicViewField(" + row.receiver + `, "length", "h.length", 1)));` + "\n"
			runViewNode(t, source, "", "adamic: panic: field read failed: h.length is not a number; expected number, found "+row.found+"\n", 70)
		})
	}
}

// The receivers a checked read is made for still read by name: a plain object, a tuple, and a class's static field
// through its constructor and through a subclass's, which JavaScript holds as functions where native holds objects.
func TestViewFieldReadsAnObjectReceiver(t *testing.T) {
	t.Parallel()
	source := viewTestRuntime + fieldReadinessRuntime + `class Base {} Base.count = 4; adamicClassIdentities.set(Base, 1);
class Derived extends Base {} adamicClassIdentities.set(Derived, 2);
console.log(String(adamicViewField({ length: 3 }, "length", "h.length", 1)));
console.log(String(adamicViewField(adamicTuple([7]), "0", "h[0]", 1)));
console.log(String(adamicViewField(Base, "count", "Base.count", 1)));
console.log(String(adamicViewField(Derived, "count", "Derived.count", 1)));
`
	runViewNode(t, source, "3\n7\n4\n4\n", "", 0)
}
