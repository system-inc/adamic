package lower

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOptionalRegExpReceiverLowering(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ name, stop string }{
		{"121-condition.ts", ""},
		{"121-corpus.ts", ""},
		{"121-exec.ts", "an optional RegExp call"},
		{"121-global.ts", ""},
		{"121-kept-chain.ts", "an optional RegExp call"},
		{"121-kept-dynamic.ts", "RegExp with a nonconstant pattern"},
		{"121-kept-grouped.ts", "a call of a parenthesized optional chain, whose parentheses end the chain"},
		{"121-kept-spread.ts", "a call through ?. (an optional call)"},
		{"121-nested-argument.ts", ""},
		{"121-nullish.ts", ""},
		{"121-once.ts", ""},
		{"121-optional-regexp-test.ts", ""},
		{"121-parenthesized-receiver.ts", ""},
		{"121-result.ts", ""},
		{"121-ungrouped-cleared.ts", ""},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("..", "oracle", "testdata", "optional_regexp_receiver", fixture.name)
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
