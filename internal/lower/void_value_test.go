package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestVoidValueErasedResultsStayNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"closure", `function answer(): number { return 7; }
const run: () => void = answer;
console.log(String(run() === undefined));`, "a void call used as a value"},
		{"contextual arrow", `const run: () => void = () => 7;
console.log(String(run() === undefined));`, "a void call used as a value"},
		{"unknown parameter", `function observe(run: () => void): boolean { return run() === undefined; }
function answer(): number { return 7; }
console.log(String(observe(answer)));`, "a void call used as a value"},
		{"override", `class Base { run(): void {} }
class Derived extends Base { run(): number { return 7; } }
function observe(value: Base): boolean { return value.run() === undefined; }
console.log(String(observe(new Derived())));`, "an override with a different native result representation"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q, got %v", probe.reason, err)
			}
		})
	}
}
