package lower

import (
	"errors"
	"testing"
)

// These barriers protect representations and receiver origins, not new language
// refusal rules. An ordinary call or a different NotYet must not mask them.
func TestStep18OptionalCallableBoundaries(t *testing.T) {
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"two guards", `interface Host { readonly run?: () => number } function run(host: Host | undefined): number { return host?.run?.() ?? -1; }`, "a call through ?. (an optional call)"},
		{"observed void", "function run(callback: (() => void) | undefined): void { console.log(`${callback?.()}`); } run((): void => {});", "observing an optional void call's runtime result"},
		{"scalar without optional representation", "const callback = (): number => 1; console.log(`${callback?.()}`);", "an optional call whose result is narrowed to present"},
		{"reference narrowed to present", "const callback = (): string => `value${1}`; console.log(callback?.());", "an optional call whose result is narrowed to present"},
		{"overloaded callable", `interface Callable { (n: number): number; (s: string): number } function run(callback: Callable | undefined): void { callback?.(1); }`, "an optional call without one represented callable signature"},
		{"erased marker", `function run(callback: ((...items: never[]) => void) | undefined): void { callback?.(); }`, "a call through an erased never-rest callable marker"},
		{"literal receiver", `const value = { n: 1, run(): number { return this.n; } }; value.run?.();`, "an optional method call requiring a saved receiver"},
		{"literal receiver through view", `interface View { readonly run?: () => number } const original = { n: 1, run(): number { return this.n; } }; const value: View = original; value.run?.();`, "an optional method call requiring a saved receiver"},
		{"erased prototype view", `class Counter { count = 0; add(n: number): number { this.count += n; return this.count; } } interface View { readonly add: (n: number) => number } const view: View = new Counter(); view.add?.(1);`, "an optional method call requiring a saved receiver"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.want {
				t.Fatalf("got %v, want NotYet %q", err, probe.want)
			}
		})
	}
}
