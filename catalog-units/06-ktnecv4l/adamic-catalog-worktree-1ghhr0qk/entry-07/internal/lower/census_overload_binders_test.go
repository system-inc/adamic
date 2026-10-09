package lower

import (
	"strings"
	"testing"
)

func TestCensusOverloadBinderGuards(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stop string }{
		{"constraint", `function constrain(value: string): string;
function constrain<T extends 'yes'>(value: T): string { return value; }
console.log(constrain('no'));`, "cannot satisfy implementation parameter"},
		{"parameter", `function pair<T>(left: T, right: string): void;
function pair<T, Extra>(left: T, right: T): void {}
pair(42, 'no');`, "cannot be served by implementation parameter"},
		{"result", `function wrong<T>(value: T): T;
function wrong<T, Extra>(value: T): T | string { return 'wrong'; }
const answer = wrong(42); console.log('ran');`, "cannot be served by implementation result"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.stop) {
				t.Fatalf("want %q, got %v", probe.stop, err)
			}
		})
	}
}
