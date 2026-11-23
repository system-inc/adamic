//go:build arrayarmprobe

package lower

import (
	"testing"
)

// Lower's review probe on #9z2ksv7 head 2: what the object side's both-ways relation lets through.
func TestArrayArmProbe(t *testing.T) {
	for _, row := range []struct{ name, source, want string }{
		{"readonly made writable", `interface R { readonly a: number }
interface W { a: number }
function f(x: R | readonly string[]): W { return x as W; }
const r: R = { a: 1 };
const w = f(r);
w.a = 2;
console.log(String(r.a));
`, "2\n"},
		{"readonly control", `interface R { readonly a: number }
function f(x: R | readonly string[]): R { return x as R; }
const r: R = { a: 1 };
console.log(String(f(r).a));
`, "1\n"},
		{"extra optional written", `interface O { a: number }
interface P { a: number; b?: string }
function f(x: O | readonly string[]): P { return x as P; }
const o: O = { a: 1 };
const p = f(o);
p.b = "s";
console.log(p.b === undefined ? "none" : p.b);
`, "s\n"},
		{"extra optional read", `interface O { a: number }
interface P { a: number; b?: string }
function f(x: O | readonly string[]): P { return x as P; }
const o = { a: 1, b: 7 };
const p = f(o);
console.log(p.b === undefined ? "none" : typeof p.b);
`, "number\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			_, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				t.Logf("ROW %s: STOPS %v", row.name, err)
				return
			}
			program := lowersAndAgreesWithNodeAs(t, row.source, ".ts")
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("ROW %s: LOWERS native exit %d stdout %q stderr %.200q want %q", row.name, code, stdout, stderr, row.want)
		})
	}
}
