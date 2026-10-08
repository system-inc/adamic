package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalCallsLower(t *testing.T) {
	for _, source := range []string{
		"const words: string[] = []; const text: string | undefined = words[0]; console.log(text?.toUpperCase() ?? 'none');",
		"const steps: (() => void)[] = []; const step = steps[0]; step?.();",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOptionalCallStopsStayNamed(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{"function observed(run: (() => void) | undefined): void { return run?.(); }", "a void optional call used as a value"},
		{"interface Runner { run?(): number; } const runner: Runner = { run(): number { return 1; } }; console.log(`${runner.run?.() ?? 0}`);", "a literal method through a view that erases its receiver"},
		{"function spread(run: ((...values: number[]) => number) | undefined): number | undefined { return run?.(1, 2); }", "a rest parameter in an optional callable"},
	} {
		_, err := lowerSource(t, probe.source)
		var stop *NotYet
		if !errors.As(err, &stop) || !strings.Contains(err.Error(), probe.reason) {
			t.Fatalf("got %v, want %s", err, probe.reason)
		}
	}
}
