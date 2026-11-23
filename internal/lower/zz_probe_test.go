package lower

import "testing"

// A probe, never merged: what each candidate for the cast stop table's removes-nullish row gives at this head.
func TestZZProbeDropsNullishRows(t *testing.T) {
	for _, source := range []string{
		"interface S { readonly n: number } function f(x: S | void): S { return x as S; }",
		"interface S { readonly n: number } function f(x: S | undefined | void): S { return x as S; }",
		"interface Base { readonly n: number } interface Sub extends Base { readonly m: number } function f(x: Sub | undefined): Base { return x as Base; }",
		"interface S { readonly n: number } interface T { readonly n: number; readonly k?: string } function f(x: S | undefined): T { return x as T; }",
	} {
		_, err := lowerSourceAs(t, source, ".ts")
		t.Logf("PROBE %q => %v", source, err)
	}
}
