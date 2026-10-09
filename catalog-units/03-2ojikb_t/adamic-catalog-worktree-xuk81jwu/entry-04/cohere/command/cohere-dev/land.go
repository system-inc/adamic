package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// cohere-dev land: a merge queue, so a green gate can't be starved by main moving under it (#fz6xejy).
//
// Landing used to be: gate the branch, then fast-forward main if it hadn't moved. A whole-module gate took
// 7 to 10 minutes on a loaded machine, and others fast-forwarded main during it, so a branch could gate
// green eight times in an hour and never land (#sycrdr6), and every one of those gates compiled a merge
// that never existed again, much of what churned the Go build cache.
//
// Now a landing takes the machine's one land lock, in arrival order, and holds it from the merge to the
// fast-forward: it merges current main into the worktree being landed, gates the result through a pool
// token, and fast-forwards main to it. Nothing that lands this way can move main during another's gate. If
// main moved anyway, by a fast-forward outside land, the fast-forward is refused by git and land merges and
// gates again.

// landAttempts is how many times land merges and gates before giving up on a main that keeps moving
// outside it.
const landAttempts = 3

// land lands the commit the working directory's worktree is at onto main, and returns its exit code.
func land(arguments []string) int {
	if len(arguments) > 0 {
		fmt.Fprintln(os.Stderr, "cohere-dev: land takes no arguments: it lands the commit this worktree is at onto main")
		return 2
	}
	worktree, err := gitOutput("", "rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: land runs inside the worktree to land: %v\n", err)
		return 2
	}
	// Tracked changes would be left behind by a landing of commits. Untracked files are the worktree's own.
	// Submodules are checked apart (checkSubmodules): most worktrees are made with TypeScript a symlink to
	// another checkout's, and git status refuses that outright, which used to read here as a dirty tree
	// with no change named. A git that fails says so, as itself.
	changes, err := gitOutput(worktree, "status", "--porcelain", "--untracked-files=no", "--ignore-submodules=all")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: reading %s's status: %v\n", worktree, err)
		return 1
	}
	if changes != "" {
		fmt.Fprintf(os.Stderr, "cohere-dev: land lands commits, and %s has changes that are not committed:\n%s\n", worktree, changes)
		return 2
	}
	if err := checkSubmodules(worktree); err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 2
	}
	mainCheckout, err := checkoutOfMain(worktree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	if mainCheckout == worktree {
		fmt.Fprintln(os.Stderr, "cohere-dev: land runs in the worktree being landed, not in the checkout of main itself")
		return 2
	}

	directory, err := landDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	head, _ := gitOutput(worktree, "rev-parse", "--short", "HEAD")
	held, err := waitInLine(directory, 1, "landing "+head+" from "+worktree, "land lock", func() {
		fmt.Fprintln(os.Stderr, "cohere-dev: waiting to land; one landing runs at a time, merge to fast-forward, and this one holds it:")
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	defer held.release()

	for attempt := 1; attempt <= landAttempts; attempt++ {
		mainCommit, err := gitOutput(worktree, "rev-parse", "main")
		if err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: reading main: %v\n", err)
			return 1
		}
		if err := mergeMain(worktree, mainCommit); err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
			return 1
		}
		// The merge may have moved a pin, and the gate must build against what it lands.
		if err := checkSubmodules(worktree); err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: after merging main at %s: %v\n", shortCommit(mainCommit), err)
			return 2
		}
		landing, _ := gitOutput(worktree, "rev-parse", "HEAD")
		if code := gate(worktree, landing); code != 0 {
			fmt.Fprintf(os.Stderr, "cohere-dev: the gate failed, so nothing landed; %s is merged with main at %s for you to fix\n",
				worktree, shortCommit(mainCommit))
			return code
		}
		// git refuses a fast-forward when main has moved since the merge, which is the test that nothing
		// landed outside this lock during the gate.
		if output, err := gitCombined(mainCheckout, "merge", "--ff-only", landing); err != nil {
			moved, _ := gitOutput(worktree, "rev-parse", "main")
			if moved != mainCommit {
				fmt.Fprintf(os.Stderr, "cohere-dev: main moved outside land during the gate, %s to %s; merging and gating again\n",
					shortCommit(mainCommit), shortCommit(moved))
				continue
			}
			fmt.Fprintf(os.Stderr, "cohere-dev: fast-forwarding main in %s: %v\n%s\n", mainCheckout, err, output)
			return 1
		}
		fmt.Fprintf(os.Stderr, "cohere-dev: landed %s on main\n", shortCommit(landing))
		return 0
	}
	fmt.Fprintf(os.Stderr, "cohere-dev: main moved outside land during each of %d gates, so nothing landed\n", landAttempts)
	return 1
}

// checkSubmodules requires every submodule of the worktree to be checked out at the commit its tree
// records, so the gate builds against what lands. The checkout may be the submodule itself, a linked
// worktree of it, or a symlink to another checkout's, as most of the house's worktrees are made: whichever,
// its head must be the pin.
func checkSubmodules(worktree string) error {
	links, err := gitlinksOf(worktree)
	if err != nil {
		return err
	}
	for _, link := range links {
		pin, path := link.pin, link.path
		directory := filepath.Join(worktree, path)
		notCheckedOut := fmt.Errorf("%s's submodule %s is not checked out, so the gate could not build against its pin %s",
			worktree, path, shortCommit(pin))
		// An empty directory is a submodule never checked out, and git run in it answers for the repository
		// around it, so its top level has to be the submodule's own path, symlinks resolved.
		top, err := gitOutput(directory, "rev-parse", "--show-toplevel")
		if err != nil {
			return notCheckedOut
		}
		resolvedTop, topErr := filepath.EvalSymlinks(top)
		resolvedDirectory, directoryErr := filepath.EvalSymlinks(directory)
		if topErr != nil || directoryErr != nil || resolvedTop != resolvedDirectory {
			return notCheckedOut
		}
		head, err := gitOutput(directory, "rev-parse", "HEAD")
		if err != nil {
			return fmt.Errorf("reading %s's head: %w", directory, err)
		}
		if head != pin {
			return fmt.Errorf("%s's %s is at %s, and the tree being landed pins %s, so the gate would build against "+
				"the wrong one: move it (git -C %s checkout --detach %s), or point it at a checkout that is, and land again",
				worktree, path, shortCommit(head), shortCommit(pin), filepath.Join(worktree, path), pin)
		}
	}
	return nil
}

// landDirectory holds the land lock and its line, beside the pool's.
func landDirectory() (string, error) {
	pool, err := poolDirectory()
	if err != nil {
		return "", err
	}
	directory := filepath.Join(pool, "land")
	return directory, os.MkdirAll(directory, 0o755)
}

// checkoutOfMain is the worktree that has main checked out, whose working tree a fast-forward must move
// along with the branch.
func checkoutOfMain(worktree string) (string, error) {
	listing, err := gitOutput(worktree, "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("listing the repository's worktrees: %w", err)
	}
	path := ""
	for _, line := range strings.Split(listing, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			path = strings.TrimPrefix(line, "worktree ")
		case line == "branch refs/heads/main":
			return path, nil
		}
	}
	return "", errors.New("no worktree of this repository has main checked out, so there is no main to land on")
}

// mergeMain merges main into the worktree, unless it already holds it. A merge that conflicts is undone,
// leaving the worktree as it was, and says which files conflicted.
func mergeMain(worktree string, mainCommit string) error {
	if _, err := gitOutput(worktree, "merge-base", "--is-ancestor", mainCommit, "HEAD"); err == nil {
		return nil
	}
	output, err := gitCombined(worktree, "merge", "--no-edit", "-m", "Merge main into the work being landed", mainCommit)
	if err == nil {
		return nil
	}
	conflicts, _ := gitOutput(worktree, "diff", "--name-only", "--diff-filter=U")
	gitCombined(worktree, "merge", "--abort")
	if conflicts != "" {
		return fmt.Errorf("main at %s conflicts with this work in:\n%s\nmerge main and resolve them, then land again",
			shortCommit(mainCommit), conflicts)
	}
	return fmt.Errorf("merging main at %s: %v\n%s", shortCommit(mainCommit), err, output)
}

// gate runs the landing gate on the worktree through a pool token: go vet and the tests of every module
// go.work names outside a submodule, each in one go invocation (workspace.go), saying which it covered.
func gate(worktree string, commit string) int {
	covered, err := readWorkspace(worktree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: reading the modules to gate: %v\n", err)
		return 2
	}
	covered.report(os.Stderr)
	patterns := covered.patterns()
	return withToken("land gate on "+shortCommit(commit)+" in "+worktree, func(budget []string) int {
		environment := gateEnvironment(budget)
		vet := exec.Command("go", append([]string{"vet"}, patterns...)...)
		vet.Dir, vet.Env = worktree, environment
		vet.Stdout, vet.Stderr = os.Stdout, os.Stderr
		if err := vet.Run(); err != nil {
			return exitCodeOf(err)
		}
		// The same environment cohere-dev test gives the module's tests, corpora included (corpora.go).
		testing, uncovered, err := testEnvironment(environment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
			return 2
		}
		defer printUncovered(os.Stderr, uncovered)
		test := exec.Command("go", append([]string{"test"}, patterns...)...)
		test.Dir, test.Env = worktree, testing
		test.Stdout, test.Stderr = os.Stdout, os.Stderr
		if err := test.Run(); err != nil {
			return exitCodeOf(err)
		}
		return 0
	})
}

// gateEnvironment is the budget's environment with GOFLAGS reduced to the pool's own -p and -buildvcs=false,
// so every landing is gated the same way, whatever its caller's shell has set. A caller's -trimpath failed
// tests that find their fixtures from their own source path, and the gate read that as the landing's fault
// (2026-10-05).
//
// -buildvcs=false because most worktrees are made with TypeScript a symlink to another checkout's, and git
// status refuses that tree, so every go build a test runs failed "error obtaining VCS status": cache's land
// at 04:15 on 2026-10-06, after a full gate. Callers used to pass it themselves, and the reduction to -p
// dropped it. Nothing the gate runs reads its own stamp: a build with none says so (packaging's version.go).
func gateEnvironment(budget []string) []string {
	environment := make([]string, 0, len(budget))
	for _, variable := range budget {
		value, isFlags := strings.CutPrefix(variable, "GOFLAGS=")
		if !isFlags {
			environment = append(environment, variable)
			continue
		}
		kept := []string{}
		for _, flag := range strings.Fields(value) {
			if strings.HasPrefix(flag, "-p=") {
				kept = append(kept, flag)
			}
		}
		environment = append(environment, "GOFLAGS="+strings.Join(append(kept, "-buildvcs=false"), " "))
	}
	return environment
}

func exitCodeOf(err error) int {
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		return exited.ExitCode()
	}
	fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
	return 1
}

// gitOutput runs git in a directory, the working directory when it is empty, and returns its trimmed output.
func gitOutput(directory string, arguments ...string) (string, error) {
	command := exec.Command("git", arguments...)
	command.Dir = directory
	var standardError bytes.Buffer
	command.Stderr = &standardError
	output, err := command.Output()
	if err != nil {
		return strings.TrimSpace(string(output)), fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(standardError.String()))
	}
	return strings.TrimSpace(string(output)), nil
}

// gitCombined runs git in a directory and returns everything it printed.
func gitCombined(directory string, arguments ...string) (string, error) {
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func shortCommit(commit string) string {
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}
