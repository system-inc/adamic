package lower

import (
	"errors"
	"testing"
)

// These barriers protect representations and receiver origins, not new language
// refusal rules. An ordinary call or a different NotYet must not mask them.
func TestStep18OptionalCallableBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source string
		want   string
	}{
		{"observed void", "function run(callback: (() => void) | undefined): void { console.log(`${callback?.()}`); } run((): void => {});", "observing an optional void call's runtime result"},
		{"scalar without optional representation", "const callback = (): number => 1; console.log(`${callback?.()}`);", "an optional call whose result is narrowed to present"},
		{"reference narrowed to present", "const callback = (): string => `value${1}`; console.log(callback?.());", "an optional call whose result is narrowed to present"},
		{"overloaded callable", `interface Callable { (n: number): number; (s: string): number } function run(callback: Callable | undefined): void { callback?.(1); }`, "an optional call without one represented callable signature"},
		{"chain getter selection", `class Host { get run(): () => number { return (): number => 1; } } function probe(host: Host | undefined): number { return host?.run?.() ?? -1; } const host = new Host();`, "an optional chain callable accessor without represented selection"},
		{"chain weak callable storage", `import type { Weak } from 'adamic'; const callback = (): number => 1; const host: { readonly run: Weak<() => number> } = { run: callback }; function probe(value: typeof host | undefined): number { return value?.run?.() ?? -1; }`, "an optional chain method without closure member storage"},
		{"erased marker", `function run(callback: ((...items: never[]) => void) | undefined): void { callback?.(); }`, "a call through an erased never-rest callable marker"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.want {
				t.Fatalf("got %v, want NotYet %q", err, probe.want)
			}
		})
	}
}
