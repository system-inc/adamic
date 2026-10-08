package lower

import (
	"strings"
	"testing"
)

func TestLibraryArrayPredicateRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"void view", "const hidden: () => void = () => 1; console.log(`${[1].some(hidden)}`);", "void can hide a returned value"},
		{"empty object view", "const hidden = (): {} => 0; console.log(`${[1].some(hidden)}`);", "hide primitive truth values"},
		{"unknown result", "const hidden = (): unknown => 0; console.log(`${[1].some(hidden)}`);", "function value returning union of differently held members"},
		{"this argument", "console.log(`${[1].some((value) => value, {})}`);", "other than one callback"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want refusal containing %q, got %v", probe.reason, err)
			}
		})
	}
}
