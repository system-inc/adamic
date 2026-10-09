package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalIndexingKeepsReadonlyRecordLookupNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "interface MapLike { readonly [key: string]: number; }\nfunction pick(record: MapLike | undefined, key: string): number | undefined { return record?.[key]; }\n")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "an optional record lookup" {
		t.Fatalf("got %v, want the unsupported optional record lookup", err)
	}
}

func TestOptionalIndexingKeepsUnsupportedStorageNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"typed ordinary property continuation", "function pick(values: readonly [Uint8Array] | undefined): number | undefined { return values?.[0].length; }\n", "an optional chain longer than one step"},
		{"string numeric key", "function pick(values: readonly number[] | undefined): number | undefined { return values?.[\"0\"]; }\n", "an optional index that isn't a number"},
		{"sparse literal", "const values: (number | undefined)[] = [, 7];\nconsole.log(`${values?.[0]}`);\n", "in an array literal"},
		{"open record", "function pick(record: Record<string, number> | undefined, key: string): number | undefined { return record?.[key]; }\n", "a value of type Record<string, number> | undefined"},
		{"index after ordinary continuation", "interface Holder { readonly values: readonly (readonly string[])[]; }\nfunction pick(holder: Holder | undefined): string | undefined { return holder?.values[0]?.[0]; }\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if probe.reason == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.reason) {
				t.Fatalf("got %v, want NotYet containing %q", err, probe.reason)
			}
		})
	}
}
