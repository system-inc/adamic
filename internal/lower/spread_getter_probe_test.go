//go:build spreadgetterprobe

package lower

import "testing"

// Lower's probe: does a spread of a class instance run its prototype getters natively, where Node never does?
func TestSpreadGetterProbe(t *testing.T) {
	for _, row := range []struct{ name, source string }{
		{"throwing getter", `class A { n = 1; get x(): number { throw new Error("ran"); } }
const a = new A();
try {
    const o = { ...a };
    console.log("ok " + String(o.n));
} catch (e) {
    console.log("caught");
}
`},
		{"counting getter", `let count = 0;
class A { n = 1; get x(): number { count++; return 2; } }
const a = new A();
const o = { ...a };
console.log(String(count) + " " + String(o.n));
`},
		{"literal getter control", `let count = 0;
const a = { n: 1, get x(): number { count++; return 2; } };
const o = { ...a };
console.log(String(count) + " " + String(o.n) + " " + String(o.x));
`},
	} {
		t.Run(row.name, func(t *testing.T) {
			_, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				t.Logf("ROW %s: STOPS %v", row.name, err)
				return
			}
			program := lowersAndAgreesWithNodeAs(t, row.source, ".ts")
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("ROW %s: LOWERS native exit %d stdout %q stderr %.200q", row.name, code, stdout, stderr)
		})
	}
}
