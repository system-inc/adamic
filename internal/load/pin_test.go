package load

import (
	"testing"

	"github.com/system-inc/adamic/internal/tracked"
)

// typeScriptCommit is the typescript-go stage 0 was last verified against: the TypeScript submodule
// inside the cohere submodule.
//
// The shims reach the compiler's internals, which upstream can reshape on any commit, so a bump is a
// gate, not a chore: when cohere moves its TypeScript, this test fails, the suite runs against the
// new checker, and the commit that updates this constant is the record that it passed.
const typeScriptCommit = "d92d9bfee114c80be2c375d72edae966176e3a4f"

func TestTypeScriptIsThePinnedCommit(test *testing.T) {
	test.Parallel()
	// git where the submodule has .git, else the tree's manifest, as a Loom runner's unpacked source has.
	got, headError := tracked.Head("../../cohere/TypeScript")
	if headError != nil {
		test.Fatalf(
			"reading the TypeScript submodule's commit: %v (run git submodule update --init --recursive)",
			headError,
		)
	}
	if got != typeScriptCommit {
		test.Errorf(
			"TypeScript is at %s, and stage 0 was verified against %s. Run the suite against it, then update typeScriptCommit",
			got,
			typeScriptCommit,
		)
	}
}
