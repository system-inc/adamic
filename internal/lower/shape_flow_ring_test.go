package lower

import (
	"testing"
	"time"
)

// A ring of stores on one field, two per receiver, used to make every projected query reopen the
// same receivers on every path before the depth cap: fan-out^32 frames and follows, so lowering
// never finished (#tzy22kd, Performance's binder.ts:1311). The getter named n makes finishAccessors
// ask the graph where every `.n` read's receiver can come from (class_accessors.go). Lowering must
// finish well inside the budget and the program agree with Node, natively and on JavaScript. Mutant:
// the frame memo and the follow dedupe both dropped, so the budget runs out.
func TestShapeFlowStoreRingFinishes(t *testing.T) {
	t.Parallel()
	const source = `class Getter { get n(): number { return 7; } }
interface Box { n: Box | null; v: number }
const a: Box = { n: null, v: 1 };
const b: Box = { n: null, v: 2 };
const c: Box = { n: a, v: 3 };
a.n = b.n;
a.n = c.n;
b.n = a.n;
b.n = c.n;
c.n = a.n;
c.n = b.n;
const x = a.n;
if (x !== null) {
	const y = x.n;
	if (y !== null) {
		console.log(String(y.v + x.v));
	}
}
console.log(String(new Getter().n));
`
	done := make(chan error, 1)
	go func() {
		_, err := lowerSourceAs(t, source, ".ts")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("want it lowered, got %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("lowering the store ring took longer than 20 seconds")
	}
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n7\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "2\n7\n")
	}
}
