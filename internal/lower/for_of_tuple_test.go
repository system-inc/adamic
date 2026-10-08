package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestForOfTupleStorageChecks(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`const values: [] = []; for (const value of values) { }`, "optional, rest or no elements"},
		{`const values: [string, string?] = ["one"]; for (const value of values) { }`, "optional, rest or no elements"},
		{`const values: [string, ...string[]] = ["one", "two"]; for (const value of values) { console.log(value); }`, "optional, rest or no elements"},
		{`const values: [string, number] = ["one", 2]; for (const value of values) { }`, "homogeneous primitive storage"},
		{`const values: [{ value: number }, { value: number }] = [{ value: 1 }, { value: 2 }]; for (const value of values) { }`, "homogeneous primitive storage"},
		{`const values: [{ value: number }, { value: number }] = [{ value: 1 }, { value: 2 }]; for (const { value } of values) { }`, "destructuring in for...of over a tuple"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var stop *NotYet
			if !errors.As(err, &stop) || !strings.Contains(stop.What, probe.reason) {
				t.Fatalf("want tuple storage stop %q, got %v", probe.reason, err)
			}
		})
	}
}
