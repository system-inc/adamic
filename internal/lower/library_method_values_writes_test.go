package lower

import (
	"errors"
	"strings"
	"testing"
)

// A delayed Array method writes what call or apply hands it, which lib's any[] Array.prototype never
// checks, so each value it writes into the receiver must be the receiver's element type, as the checker
// judges it (#gw3yrsn, Language's k3 to k10 on main 31fdd16d). k10 is an object pushed into a string[],
// which sanitized native read as a string and died on a null load under UBSan where Node throws. unshift
// has no adapter and stays NotYet. Mutant: delayedArrayWritesFit answering nil, under which push, fill,
// splice and concat lower again, and k10 dies natively under UBSan.
func TestDelayedArrayWritesMustFitTheElements(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"k10 push an object", `const push = Array.prototype.push;
const array = ["a", "b"];
const other = { x: 1 };
push.call(array, other);
const last = array[2] ?? "";
console.log(last.toUpperCase());`, "a delayed push writing a value that isn't the receiver's element type"},
		{"k6 push a number", `const push = Array.prototype.push;
const array = ["a", "b"];
push.call(array, 5);
console.log(String((array[2] ?? "").length));`, "a delayed push writing a value that isn't the receiver's element type"},
		{"k8 push through apply", `const push = Array.prototype.push;
const array = ["a", "b"];
push.apply(array, [5]);
console.log(String((array[2] ?? "").length));`, "a delayed push writing a value that isn't the receiver's element type"},
		{"k5 fill", `const fill = Array.prototype.fill;
const array = ["a", "b"];
fill.call(array, 5);
console.log(String((array[0] ?? "").length));`, "a delayed fill writing a value that isn't the receiver's element type"},
		{"k4 splice", `const splice = Array.prototype.splice;
const array = ["a", "b"];
splice.apply(array, [0, 0, 5]);
console.log(String((array[0] ?? "").length));`, "a delayed splice writing a value that isn't the receiver's element type"},
		{"k3 concat an array", `const concat = Array.prototype.concat;
const array = ["a", "b"];
const combined: string[] = concat.apply(array, [[5]]);
console.log(combined.join("|") + " " + String((combined[2] ?? "").length));`, "a delayed concat of an array whose elements aren't the receiver's element type"},
		{"k7 concat a value", `const concat = Array.prototype.concat;
const array = ["a", "b"];
const combined: string[] = concat.call(array, 5);
console.log(String((combined[2] ?? "").length));`, "a delayed concat of an array whose elements aren't the receiver's element type"},
		{"k9 unshift", `const unshift = Array.prototype.unshift;
const array = ["a", "b"];
unshift.call(array, 5);
console.log(String((array[0] ?? "").length));`, "delayed Array.prototype.unshift (no adapter proves"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, probe.source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// Writes of the receiver's own element type still lower and agree with Node, through call and apply.
func TestDelayedArrayWritesOfTheElementsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"push", `const push = Array.prototype.push; const array = ["a", "b"]; push.call(array, "c"); push.apply(array, ["d"]); console.log(array.join("|"));`},
		{"fill", `const fill = Array.prototype.fill; const array = ["a", "b"]; fill.call(array, "z"); console.log(array.join("|"));`},
		{"splice", `const splice = Array.prototype.splice; const array = ["a", "b"]; splice.apply(array, [1, 0, "c"]); console.log(array.join("|"));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
		})
	}
}
