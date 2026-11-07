package main

import "testing"

func TestAdaptSingletonArrayCallback(test *testing.T) {
	test.Parallel()
	got := adaptSource("function cb(x){return x;} [1].forEach(cb);")
	want := "function cb(x: number){return x;} [1].forEach(cb);"
	if got.Source != want || got.Counts[adaptCallbackParam] != 1 || len(got.Counts) != 1 {
		test.Fatalf("singleton array callback: got %#v, want source %q and one callback-param rewrite", got, want)
	}
}

func TestAdaptPositiveExponentVarToLet(test *testing.T) {
	test.Parallel()
	got := adaptSource("var x=1e+2;")
	want := "let x=1e+2;"
	if got.Source != want || got.Counts[adaptVarToLet] != 1 || len(got.Counts) != 1 {
		test.Fatalf("positive exponent: got %#v, want source %q and one var-to-let rewrite", got, want)
	}
}

func TestAdaptClassKeepsVars(test *testing.T) {
	test.Parallel()
	// The scan skips class bodies, so it cannot prove var-to-let safe in this file.
	source := "var x=1; class Box { value: number; }"
	got := adaptSource(source)
	if got.Source != source || len(got.Counts) != 0 {
		test.Fatalf("class must preserve vars: got %#v, want source %q and no rewrites", got, source)
	}
}
