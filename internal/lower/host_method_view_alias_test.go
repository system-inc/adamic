package lower

import (
	"errors"
	"testing"
)

// The host method view rule exempts only a routed binding's own declaration admission, whose reads lower
// through the model's path. A value of the binding's type seen again, an alias, an inferred alias or a
// NonNullable one, is judged like any view: NotYet, where Node prints what nextTick is (Compiler's road on
// #sn788gx, Codex 01a1c25d: each stopped at main too, on browser's unregistered member). The direct read
// still lowers. Mutant (d) in node-host-method-view.mutants.md turns the direct read red.
func TestHostBindingAliasesStayJudged(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		body  string
		lower bool
	}{
		{"console.log(String(!!process.nextTick && !process.browser));", true},
		{"const p: typeof process = process; console.log(String(!!p?.nextTick));", false},
		{"const p = process; console.log(String(!!p.nextTick && !p.browser));", false},
		{"const p: NonNullable<typeof process> = process; console.log(typeof p.nextTick);", false},
	} {
		t.Run(probe.body, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, processModules(probe.body))
			if probe.lower {
				if err != nil {
					t.Fatalf("want it to lower, got %v", err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a NotYet, got %v", err)
			}
		})
	}
}
