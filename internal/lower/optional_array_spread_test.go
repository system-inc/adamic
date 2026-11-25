package lower

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOptionalArraySpreadLowering(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ name, stop string }{
		{"92-corpus.ts", ""},
		{"92-identity.ts", ""},
		{"92-kept-join.ts", "a SpreadElement"},
		{"92-mixed.ts", ""},
		{"92-nullish.ts", ""},
		{"92-optional-array-append.ts", ""},
		{"92-order.ts", ""},
		{"92-result.ts", ""},
		{"92-spread-count.ts", ""},
		{"92-unshift.ts", ""},
		// A parenthesized callee ends the chain, so these stop at #492fbpd's stop above the spread gate,
		// with a spread (push, and concat, which the gate never admits) or without one.
		{"92-kept-grouped.ts", "a call of a parenthesized optional chain, whose parentheses end the chain"},
		{"92-grouped-plain.ts", "a call of a parenthesized optional chain, whose parentheses end the chain"},
		{"92-grouped-concat.ts", "a call of a parenthesized optional chain, whose parentheses end the chain"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("..", "oracle", "testdata", "optional_array_spread", fixture.name)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			if fixture.stop == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var pending *NotYet
			if !errors.As(err, &pending) || pending.What != fixture.stop {
				t.Fatalf("want NotYet %q, got %v", fixture.stop, err)
			}
		})
	}
}
