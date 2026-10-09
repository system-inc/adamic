package release

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitIn runs git in a test repository and returns its trimmed output.
func gitIn(t *testing.T, directory string, arguments ...string) string {
	t.Helper()

	output, err := gitRun(directory, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return output
}

// gitRun is gitIn for a fixture built outside any one test, which has no *testing.T to fail.
func gitRun(directory string, arguments ...string) (string, error) {
	command := exec.Command("git", append([]string{"-C", directory, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(arguments, " "), err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

// repositoryWithSideCommit returns a repository checked out at its first commit, the hash of that
// commit, and the hash of a second commit on a branch the checkout does not contain.
func repositoryWithSideCommit(t *testing.T) (directory string, checkedOut string, notContained string) {
	t.Helper()

	directory = t.TempDir()
	gitIn(t, directory, "init", "--quiet")

	writeFile(t, filepath.Join(directory, "first.txt"), "first\n")
	gitIn(t, directory, "add", "first.txt")
	gitIn(t, directory, "commit", "--quiet", "-m", "first")
	checkedOut = gitIn(t, directory, "rev-parse", "HEAD")

	gitIn(t, directory, "checkout", "--quiet", "-b", "side")
	writeFile(t, filepath.Join(directory, "second.txt"), "second\n")
	gitIn(t, directory, "add", "second.txt")
	gitIn(t, directory, "commit", "--quiet", "-m", "second")
	notContained = gitIn(t, directory, "rev-parse", "HEAD")

	gitIn(t, directory, "checkout", "--quiet", checkedOut)
	return directory, checkedOut, notContained
}

// TestReleaseProceedsFromACommitContainingTheMinimum is the positive half. Without it, a check that
// refused everything would pass the controls below.
func TestReleaseProceedsFromACommitContainingTheMinimum(t *testing.T) {
	t.Parallel()

	directory, checkedOut, _ := repositoryWithSideCommit(t)

	if err := requireAncestor(directory, checkedOut, "the reason"); err != nil {
		t.Fatalf("a checkout containing the minimum was refused: %v", err)
	}
}

// TestReleaseRefusesACommitMissingTheMinimum is the known-dirty control: the minimum exists in the
// repository but not in the history being released, and the refusal says why.
func TestReleaseRefusesACommitMissingTheMinimum(t *testing.T) {
	t.Parallel()

	directory, _, notContained := repositoryWithSideCommit(t)

	err := requireAncestor(directory, notContained, "the reason it matters")
	if err == nil {
		t.Fatal("a checkout missing the minimum was allowed to release")
	}
	if !strings.Contains(err.Error(), "the reason it matters") {
		t.Fatalf("the refusal does not carry the reason: %v", err)
	}
}

// TestReleaseRefusesWhenGitCannotTell holds the third exit code. A minimum git has never seen,
// which is what a shallow clone looks like, must not read as a pass.
func TestReleaseRefusesWhenGitCannotTell(t *testing.T) {
	t.Parallel()

	directory, _, _ := repositoryWithSideCommit(t)

	err := requireAncestor(directory, strings.Repeat("0", 40), "the reason")
	if err == nil {
		t.Fatal("a minimum git could not find was treated as present")
	}
	if !strings.Contains(err.Error(), "could not tell") {
		t.Fatalf("an undecidable check was reported as a plain refusal: %v", err)
	}
}

// TestBuildRefusesBeforeTheMinimum proves Build calls the check, which the tests above cannot: they
// drive requireAncestor directly and would pass with the call deleted from Build. A repository that
// never contained MinimumReleaseCommit must be refused before anything is compiled. It pins a compiler
// correctly, because Build reads the pin first and would otherwise refuse on that instead.
func TestBuildRefusesBeforeTheMinimum(t *testing.T) {
	t.Parallel()

	fixture := moduleWithCompiler(t)

	_, err := Build(Options{Version: "1.0.0", ModuleDirectory: fixture.Module, OutputDirectory: t.TempDir()})
	if err == nil {
		t.Fatal("Build released from a repository that never contained the minimum commit")
	}
	// Refused for the minimum, by name. Any earlier refusal also returns an error, and passed this test
	// unnoticed when the copyright check first stopped the fixture on a missing license.
	if !strings.Contains(err.Error(), ShortCommit(MinimumReleaseCommit)) {
		t.Fatalf("Build refused, but not for the minimum commit: %v", err)
	}
	if !strings.Contains(err.Error(), ShortCommit(MinimumReleaseCommit)) {
		t.Fatalf("Build failed, but not on the minimum commit: %v", err)
	}
}
