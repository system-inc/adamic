package lower

import (
	"errors"
	"strings"
	"testing"
)

// Language's attack on S1c's void exemption (#qk8rztp, 22:36Z), each held as a pin. The exemption
// trusts a generator's yielded type, so a void call's value must never reach a yield whose type
// says unknown: no-void-read refuses it at the yield's operand, under spread (v1), Array.from (v2)
// and a yield* into an unknown-yielding generator (v4), before any exemption is asked. A return of a
// void call in a void-TReturn generator (v3) refuses as a void value too, an extra refusal Language
// notes for whoever extends H2's discard rule. A user object named Array (v5) is no Array.from and
// isn't exempt. v6, the control, lowers.
func TestGeneratorVoidYieldsAreRefusedAtTheYield(test *testing.T) {
	test.Parallel()
	const make = "const make: () => void = () => 42;\n"
	for _, probe := range []struct{ name, source string }{
		{"v1 spread", make +
			"function* g(): Generator<unknown, void, unknown> { yield make(); }\nconsole.log(String([...g()].length));"},
		{"v2 Array.from", make +
			"function* g(): Generator<unknown, void, unknown> { yield make(); }\nconsole.log(String(Array.from(g()).length));"},
		{"v3 return", make +
			"function* g(): Generator<number, void, unknown> { yield 1; return make(); }\nconsole.log(String([...g()].length));"},
		{"v4 yield*", make +
			"function* inner() { yield make(); }\nfunction* g(): Generator<unknown, void, unknown> { yield* inner(); }\nconsole.log(String([...g()].length));"},
		{
			"v5 shadowed Array",
			"function* numbers() { yield 1; yield 2; }\nconst Array = { from(values: Iterable<number>): number[] { const out: number[] = []; for (const value of values) out.push(value); return out; } };\nconsole.log(Array.from(numbers()).join(','));",
		},
	} {
		_, lowerSourceError := lowerSource(test, probe.source)
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) || !strings.Contains(refused.Error(), "adamic/no-void") {
			test.Errorf("%s: got %v, want a void rule's refusal", probe.name, lowerSourceError)
		}
	}
	control := "function* numbers() { yield 1; yield 2; }\nconsole.log([...numbers()].join(',') + ' ' + Array.from(numbers()).join(','));"
	if _, lowerSourceError := lowerSource(test, control); lowerSourceError != nil {
		test.Errorf("v6 control: %v", lowerSourceError)
	}
}
