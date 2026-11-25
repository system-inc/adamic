package oracle

import (
	"path/filepath"
	"testing"
)

// Like regexp_served_values_test.go, compare both backends with the independent Node oracle.
func TestOptionalRegExpReceiverAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"121-condition.ts",
		"121-corpus.ts",
		"121-global.ts",
		"121-nested-argument.ts",
		"121-nullish.ts",
		"121-once.ts",
		"121-optional-regexp-test.ts",
		"121-parenthesized-receiver.ts",
		"121-result.ts",
		"121-ungrouped-cleared.ts",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_regexp_receiver", name))
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
