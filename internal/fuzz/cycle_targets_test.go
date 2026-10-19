package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The generated programs combine these unrelated producers with a closure array.
// Unknown callable arrays must remain conservative, but these arrays cannot
// supply direct function elements to the closure array's producer pool.
func TestClosureArraysWithUnrelatedProducersAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := PrepareBuilt("../..", buildcache.Adamic(t), filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("testdata/cycle_targets_arrays.a")
	if err != nil {
		t.Fatal(err)
	}
	outcome := checkout.Try(string(source), filepath.Join(directory, "program"))
	if outcome.Verdict != Agreed {
		t.Fatalf("%s: %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
	if got := string(outcome.Node.Stdout); got != "7\n1\n2,3\n1,2\n" {
		t.Fatalf("Node output %q", got)
	}
	t.Logf("Node, JavaScript and sanitized native agree: %q", outcome.Node.Stdout)
}
