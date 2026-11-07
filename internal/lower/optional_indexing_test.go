package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalIndexingKeepsReadonlyIndexSignaturesNotYet(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		"interface MapLike { readonly [key: string]: number; }\nfunction pick(record: MapLike | undefined, key: string): number | undefined { return record?.[key]; }\n",
	)
	var refused *NotYet
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "non-mutable unrestricted string signature") {
		test.Fatalf("got %v, want the readonly dictionary refusal", err)
	}
}

func TestOptionalIndexingKeepsUnsupportedStorageNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{
			"typed ordinary property continuation",
			"function pick(values: readonly [Uint8Array] | undefined): number | undefined { return values?.[0].length; }\n",
			"an optional chain longer than one step",
		},
		{
			"string numeric key",
			"function pick(values: readonly number[] | undefined): number | undefined { return values?.[\"0\"]; }\n",
			"an optional index that isn't a number",
		},
		{
			"sparse literal",
			"const values: (number | undefined)[] = [, 7];\nconsole.log(`${values?.[0]}`);\n",
			"in an array literal",
		},
		{
			"open record",
			"function pick(record: Record<string, number> | undefined, key: string): number | undefined { return record?.[key]; }\n",
			"?.[] on a value",
		},
		{
			"index after ordinary continuation",
			"interface Holder { readonly values: readonly (readonly string[])[]; }\nfunction pick(holder: Holder | undefined): string | undefined { return holder?.values[0]?.[0]; }\n",
			"",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			_, err := lowerSource(test, probe.source)
			if probe.reason == "" {
				if err != nil {
					test.Fatal(err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.reason) {
				test.Fatalf("got %v, want NotYet containing %q", err, probe.reason)
			}
		})
	}
}
