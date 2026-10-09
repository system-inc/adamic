package oracletest

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeGolden(t *testing.T, contents record) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "golden.json.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zipped := gzip.NewWriter(file)
	if err := json.NewEncoder(zipped).Encode(contents); err != nil {
		t.Fatal(err)
	}
	zipped.Close()
	file.Close()
	return path
}

// TestAGoldenFromOtherBundlesIsRefused is the property the golden exists to keep: answers recorded from
// bundles this build does not embed are never compared against.
func TestAGoldenFromOtherBundlesIsRefused(t *testing.T) {
	t.Parallel()
	path := writeGolden(t, record{Digest: "recorded", Answers: map[string]Answer{"k": {Output: "x"}}})
	if _, err := read(path, "embedded"); err == nil || !strings.Contains(err.Error(), "recorded from Prettier bundles recorded") {
		t.Fatalf("a golden from other bundles was read: %v", err)
	}
	// The control: the same file, with the digest it was recorded under, reads.
	answers, err := read(path, "recorded")
	if err != nil || answers["k"].Output != "x" {
		t.Fatalf("a golden from these bundles was refused: %v, %v", answers, err)
	}
}

// TestAMissingGoldenIsRefused: no golden is a failure that says how to record one, never an empty pass.
func TestAMissingGoldenIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := read(filepath.Join(t.TempDir(), "absent.json.gz"), "embedded"); err == nil {
		t.Fatal("a missing golden read as empty")
	}
}

// TestKeysDoNotRunTogether: inputs that concatenate to the same text are different keys.
func TestKeysDoNotRunTogether(t *testing.T) {
	t.Parallel()
	if Key("ab", "c") == Key("a", "bc") {
		t.Fatal("two different inputs share a key")
	}
	if Key("a", "b") != Key("a", "b") {
		t.Fatal("one input has two keys")
	}
}

// TestTheBundlesHaveADigest: the digest every golden is stamped with is readable from this build.
func TestTheBundlesHaveADigest(t *testing.T) {
	t.Parallel()
	digest, err := BundlesDigest()
	if err != nil || len(digest) != 64 {
		t.Fatalf("no digest for the embedded bundles: %q, %v", digest, err)
	}
}
