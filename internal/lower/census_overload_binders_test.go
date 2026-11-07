package lower

import (
	"strings"
	"testing"
)

func TestCensusOverloadBinderGuards(test *testing.T) {
	test.Parallel()
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
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), probe.stop) {
				test.Fatalf("want %q, got %v", probe.stop, lowerSourceError)
			}
		})
	}
}
