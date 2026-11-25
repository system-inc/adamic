package oracle

import (
	"path/filepath"
	"testing"
)

// Like optional_intrinsic_test.go, compare both backends with the independent Node oracle.
func TestOptionalArraySpreadAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"92-corpus.ts",
		"92-identity.ts",
		"92-mixed.ts",
		"92-nullish.ts",
		"92-optional-array-append.ts",
		"92-order.ts",
		"92-result.ts",
		"92-spread-count.ts",
		"92-unshift.ts",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_array_spread", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node exit %d: %s", truth.exitCode, truth.stderr)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
			if !comparable {
				t.Fatal("native observation unavailable")
			}
			for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s; Node stdout=%q; backend stdout=%q stderr=%q", backend, difference, truth.stdout, got.stdout, got.stderr)
				}
			}
		})
	}
}
