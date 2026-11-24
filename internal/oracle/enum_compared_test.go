package oracle

import (
	"path/filepath"
	"testing"
)

// The program's own comparison of a const local proves an open numeric enum's value is one of the members
// a slot takes (enumComparedMember, Adamic's ruling of Nov 19 on #snypn8c, c3): a case chain (case_chain,
// parser.ts:4754), an || with an inner && (or_and, parser.ts:8417), an && chain (and_chain,
// parser.ts:8701) and a conditional's true arm either way round (conditional). Each lowers and agrees with
// Node on the JavaScript backend and natively under the leak check, every value a member it proves or
// one it doesn't, 7 among them, which never reaches the slot. negative_zero holds -0 beside 0 through the
// slot, its sign kept (Language's z01). The stops are pinned in internal/lower/enum_compared_test.go, and
// the mutants in testdata/enum_compared/mutants.md.
func TestEnumComparedMemberAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, stdout := range map[string]string{
		"case_chain":    "took 0, took 1, c, other 3, other 7\n",
		"or_and":        "took 0, took 1, skip 1, skip 2, skip 7\n",
		"and_chain":     "took 0, took 1, skip 1, skip 0, skip 2, skip 7\n",
		"conditional":   "took 0 took 0, not a not a, not a not a\n",
		"negative_zero": "-Infinity false, Infinity true, other\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enum_compared", name+".a"))
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
