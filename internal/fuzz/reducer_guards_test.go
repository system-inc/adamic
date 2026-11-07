package fuzz

import (
	"strings"
	"sync/atomic"
	"testing"
)

func TestReduceRejectsDifferentRefusal(test *testing.T) {
	test.Parallel()
	// Controlled observations isolate the reducer's signature contract from
	// changes in which syntax the compiler currently admits. Both failures have
	// the same verdict and checker diagnostics; only the refusal text differs.
	const source = "originalFailure();\notherFailure();\n"
	original := Observation{Verdict: NotYet, Lines: map[string]string{"refusal": "original refusal"}}
	signature, err := Derive(original)
	if err != nil {
		test.Fatal(err)
	}
	var sawDifferent atomic.Bool
	observe := func(candidate string, slot int) Observation {
		if strings.Contains(candidate, "originalFailure") {
			return original
		}
		if strings.Contains(candidate, "otherFailure") {
			sawDifferent.Store(true)
			return Observation{Verdict: NotYet, Lines: map[string]string{"refusal": "different refusal"}}
		}
		return Observation{Verdict: Agreed, Lines: map[string]string{}}
	}
	reduced := Reduce("refusals.a", source, signature, original, observe, 1, 30)
	if !sawDifferent.Load() {
		test.Fatal("reducer never tried the smaller program with a different refusal")
	}
	final := observe(reduced.Source, 0)
	if !final.Has(signature) || reduced.Signature != signature {
		test.Fatalf(
			"reducer reported %s for source %q that actually fails as %v",
			reduced.Signature,
			reduced.Source,
			final.Lines,
		)
	}
	if strings.Contains(reduced.Source, "otherFailure") {
		test.Fatalf("unrelated refusal was not reduced away: %q", reduced.Source)
	}
}

func TestExactSignatureRejectsSuffix(test *testing.T) {
	test.Parallel()
	signature := Signature{Kind: "refusal", Text: "abc", Exact: true}
	if !signature.matches("abc") {
		test.Fatal("exact signature does not match itself")
	}
	if signature.matches("abc-suffix") {
		test.Fatal("exact signature abc matched abc-suffix")
	}
}

func TestBytesSharedRejects63Bytes(test *testing.T) {
	test.Parallel()
	if bytesShared(63, 128) {
		test.Fatal("63 bytes must not share a 128-byte owner")
	}
	if !bytesShared(64, 128) {
		test.Fatal("64 bytes must share a 128-byte owner")
	}
}
