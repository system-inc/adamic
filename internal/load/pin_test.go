package load

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/submodules"
	"github.com/system-inc/adamic/internal/tracked"
)

// typeScriptCommit is the typescript-go stage 0 was last verified against: the TypeScript submodule
// inside the cohere submodule.
//
// The shims reach the compiler's internals, which upstream can reshape on any commit, so a bump is a
// gate, not a chore: when cohere moves its TypeScript, this test fails, the suite runs against the
// new checker, and the commit that updates this constant is the record that it passed.
const typeScriptCommit = "d92d9bfee114c80be2c375d72edae966176e3a4f"

func TestTypeScriptIsThePinnedCommit(t *testing.T) {
	t.Parallel()
	// The path is read from the .gitmodules files by the submodule's name, never spelled here, since cohere moves
	// TypeScript to libraries/typescript and keeps its name. A refusal stops the test in the reader's words: a
	// guessed path would answer for a checkout that isn't the one cohere declares.
	path, err := submodules.Path("../..", "TypeScript")
	if err != nil {
		t.Fatal(err)
	}
	// git where the submodule has .git, else the tree's manifest, as a Loom runner's unpacked source has.
	got, err := tracked.Head(filepath.Join("../..", filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("reading the TypeScript submodule's commit: %v (run git submodule update --init --recursive)", err)
	}
	if got != typeScriptCommit {
		t.Errorf("TypeScript is at %s, and stage 0 was verified against %s. Run the suite against it, then update typeScriptCommit", got, typeScriptCommit)
	}
}
