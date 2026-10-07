package load

import (
	"os/exec"
	"strings"
	"testing"
)

// typeScriptCommit is the typescript-go stage 0 was last verified against: the TypeScript submodule
// inside the cohere submodule.
//
// The shims reach the compiler's internals, which upstream can reshape on any commit, so a bump is a
// gate, not a chore: when cohere moves its TypeScript, this test fails, the suite runs against the
// new checker, and the commit that updates this constant is the record that it passed.
const typeScriptCommit = "8d550c837c90bd1805b047b7eeccc2baac2d5e7a"

func TestTypeScriptIsThePinnedCommit(t *testing.T) {
	t.Parallel()
	output, err := exec.Command("git", "-C", "../../cohere/TypeScript", "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("reading the TypeScript submodule's commit: %v (run git submodule update --init --recursive)", err)
	}
	if got := strings.TrimSpace(string(output)); got != typeScriptCommit {
		t.Errorf("TypeScript is at %s, and stage 0 was verified against %s. Run the suite against it, then update typeScriptCommit", got, typeScriptCommit)
	}
}
