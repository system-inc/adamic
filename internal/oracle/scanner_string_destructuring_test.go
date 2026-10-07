package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"scanner_string_length_destructuring.a", "scanner_string_destructuring.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
}

// A valid length mutant must fail solely at the external Node comparison.
func TestScannerStringDestructuringMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scanner_string_length_destructuring.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	mutant := strings.ReplaceAll(source, "adamic_string_length(", "scanner_mutant_length(")
	if mutant == source {
		t.Fatal("mutant changed nothing")
	}
	mutant = `#include "adamic.h"
static double scanner_mutant_length(const adamic_string *text) { return adamic_string_length(text) + 1; }
` + mutant
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if difference := disagreement(onNode(t, path), got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Logf("Node prints 10; length + 1 mutant prints %s; caught by stdout comparison", strings.TrimSpace(string(got.stdout)))
}
