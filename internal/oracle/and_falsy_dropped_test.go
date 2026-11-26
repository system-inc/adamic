package oracle

import (
	"path/filepath"
	"testing"
)

// An && whose consumer drops a falsy result hands no slot its left, so only its right is judged (andFalsyDropped,
// Adamic's ruling of Nov 19 on #snypn8c, rule 3, c1): tsc's compareSignaturesRelated under || with a boolean left
// (boolean_or, checker.ts:21987), two && lefts (double_and_or, :22020) and a number left assigned through &=
// (number_and_assign, :22066), a ! over a flag read (element_flags, :32547) and an if condition (if_condition). A
// left that can only be truthy, 1, never reaches the result (truthy_literal). Each lowers and agrees with
// Node on the JavaScript backend and natively under the leak check, every other left falsy in some call and truthy
// in another. The stops are pinned in internal/lower/and_falsy_dropped_test.go, and the mutants in
// testdata/and_falsy_dropped/mutants.md.
func TestAndFalsyDroppedAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, stdout := range map[string]string{
		"boolean_or":        "-1, 0, 3, 3\n",
		"double_and_or":     "-1, 0, 0, 3\n",
		"number_and_assign": "-1, 3, 3, 3, 0\n",
		"element_flags":     "optional, optional, required, optional\n",
		"if_condition":      "unrelated, related -1, unrelated, related 3\n",
		"truthy_literal":    "-1, -1\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/and_falsy_dropped", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != stdout {
				t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", stdout, truth.exitCode, truth.stdout, truth.stderr)
			}
			backend := onJavaScriptBackend(t, program)
			if diff := disagreement(truth, backend); diff != "" {
				t.Fatalf("javascript: %s; backend stdout %q stderr %q exit %d", diff, backend.stdout, backend.stderr, backend.exitCode)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if comparable {
				if diff := disagreement(truth, actual); diff != "" {
					t.Fatalf("native: %s; stdout %q stderr %q exit %d", diff, actual.stdout, actual.stderr, actual.exitCode)
				}
			}
		})
	}
}
