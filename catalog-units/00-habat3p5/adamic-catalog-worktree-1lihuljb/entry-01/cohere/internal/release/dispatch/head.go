package dispatch

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readHead returns the commit a checkout's HEAD names, read from the repository's own files.
//
// The launcher asks this on every run, and asking git cost a process each time for an answer that sits
// in two small files. A cache hit now execs the cached binary without starting any git: the commit names
// the binary, and git is needed only to build one (`git archive` is the committed tree, and nothing
// short of git reads a packfile).
//
// It follows what a clone, a worktree and a submodule each lay out: `.git` as a directory or as a
// `gitdir:` file, a worktree's `commondir`, HEAD as a commit or as a symbolic ref, and the ref as a loose
// file or a line of `packed-refs`. Anything else, such as the reftable ref store, is an error that names
// it, never a guess: a wrong commit builds and runs the wrong rules while looking like the right ones.
func readHead(repository string) (string, error) {
	gitDirectory, err := gitDirectoryOf(repository)
	if err != nil {
		return "", err
	}
	commonDirectory := gitDirectory
	if contents, err := os.ReadFile(filepath.Join(gitDirectory, "commondir")); err == nil {
		commonDirectory = resolveFrom(gitDirectory, strings.TrimSpace(string(contents)))
	}
	if config, err := os.ReadFile(filepath.Join(commonDirectory, "config")); err == nil && strings.Contains(string(config), "reftable") {
		return "", fmt.Errorf("%s keeps its refs in the reftable store, which cohere does not read", repository)
	}

	head, err := os.ReadFile(filepath.Join(gitDirectory, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("reading HEAD in %s: %w", repository, err)
	}
	value := strings.TrimSpace(string(head))
	// A symbolic ref can name another, so the chain is followed, and bounded so a loop is an error.
	for range 8 {
		reference, symbolic := strings.CutPrefix(value, "ref: ")
		if !symbolic {
			if !isObjectName(value) {
				return "", fmt.Errorf("HEAD in %s names %q, which is not a commit", repository, value)
			}
			return value, nil
		}
		value, err = readReference(gitDirectory, commonDirectory, strings.TrimSpace(reference))
		if err != nil {
			return "", fmt.Errorf("reading HEAD in %s: %w", repository, err)
		}
	}
	return "", fmt.Errorf("HEAD in %s is a chain of symbolic refs that does not end", repository)
}

// gitDirectoryOf finds a checkout's git directory: `.git` itself, or where a `.git` file points.
func gitDirectoryOf(repository string) (string, error) {
	dotGit := filepath.Join(repository, ".git")
	information, err := os.Stat(dotGit)
	if err != nil {
		return "", fmt.Errorf("%s is not a git checkout: %w", repository, err)
	}
	if information.IsDir() {
		return dotGit, nil
	}
	contents, err := os.ReadFile(dotGit)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", dotGit, err)
	}
	target, found := strings.CutPrefix(strings.TrimSpace(string(contents)), "gitdir: ")
	if !found {
		return "", fmt.Errorf("%s is a file that does not name a git directory", dotGit)
	}
	return resolveFrom(repository, strings.TrimSpace(target)), nil
}

// readReference reads one ref's value: its loose file, the worktree's first and then the shared one, or
// its line in `packed-refs`. A loose file wins over a packed line, as it does for git.
func readReference(gitDirectory string, commonDirectory string, reference string) (string, error) {
	if reference == "" || strings.Contains(reference, "..") || filepath.IsAbs(reference) {
		return "", fmt.Errorf("%q is not a ref name", reference)
	}
	for _, directory := range []string{gitDirectory, commonDirectory} {
		contents, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(reference)))
		if err == nil {
			return strings.TrimSpace(string(contents)), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("reading %s: %w", reference, err)
		}
	}

	packed, err := os.Open(filepath.Join(commonDirectory, "packed-refs"))
	if err != nil {
		return "", fmt.Errorf("%s is neither a loose ref nor packed (%v)", reference, err)
	}
	defer packed.Close()
	lines := bufio.NewScanner(packed)
	for lines.Scan() {
		// `<object> <ref>`, with `#` headers and `^<object>` peel lines that name no ref.
		object, name, found := strings.Cut(lines.Text(), " ")
		if found && name == reference {
			return object, nil
		}
	}
	if err := lines.Err(); err != nil {
		return "", fmt.Errorf("reading packed-refs: %w", err)
	}
	return "", fmt.Errorf("%s is neither a loose ref nor packed", reference)
}

// isObjectName reports whether a value is a full object name: 40 hex digits under SHA-1, 64 under
// SHA-256. An abbreviation is not one, because it can name a different commit tomorrow.
func isObjectName(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func resolveFrom(base string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, path)
}
