package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Every discard position lowers and agrees with Node natively and on the JavaScript backend: a
// value-returning callback for a void one, a statement, void's operand, a comma's left, a for
// statement's initializer and incrementor, a return from an exactly void function, an arrow body
// whose arrow is void, a conditional's branch and the right of && in a statement.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/void_reads_admitted.a", true, false})
}

// A void call's result is discarded natively and on the JavaScript backend, so a read of it printed
// undefined where Node printed what the function returned, and the two backends agreed (#x2gt39p).
// Each program reads a void value one way and runs clean on Node; its whole diagnostic is pinned in
// its .refused file. return_union is Language's finding, which the void widening refusal's return
// check stops first. Mutant: voidDiscarded admitting a variable's initializer lowers binding.a,
// which then prints true natively where Node prints false.
func TestVoidReadsAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/void_reads_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			if truth := onNode(t, absolute); truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want a void read refusal, got %v", err)
			}
			expected, err := os.ReadFile(absolute + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 15 {
		t.Fatalf("found %d refused programs, want 15", programs)
	}
}
