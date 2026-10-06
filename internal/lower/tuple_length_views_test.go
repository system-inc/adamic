package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestTupleLengthViewsWithoutRuntimeLengthAreNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function size(entry: [string, number?] | undefined): number { return entry?.length ?? -1; }`,
		`function size(entry: [string, ...number[]] | undefined): number { return entry?.length ?? -1; }`,
		`function size(entry: readonly [string] | readonly [string, number?]): number { return entry.length; }`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(err.Error(), "runtime length is not stored") {
				t.Fatalf("got %v, want NotYet explaining the missing runtime length", err)
			}
		})
	}
}
