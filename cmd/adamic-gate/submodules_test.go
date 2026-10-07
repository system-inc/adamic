package main

import (
	"fmt"

	"github.com/system-inc/adamic/internal/boundedrun"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: shard and merge use the process working directory.
func TestShardRefusesStaleRecursiveSubmodules(t *testing.T) {
	root, cohere, typescript, old := submoduleFixture(t)
	t.Chdir(root)
	initial, err := inspectSubmodules(".", "HEAD")
	if err != nil || len(initial) != 2 || requirePinnedSubmodules(initial) != nil {
		t.Fatalf("clean recursive provenance: %v, %v", initial, err)
	}
	if initial[0].Path != "cohere" || initial[1].Path != "cohere/typescript-go" {
		t.Fatal(initial)
	}
	// A local ignore setting must not hide a stale checkout from this audit.
	gitFixture(t, root, "config", "submodule.cohere.ignore", "all")
	gitFixture(t, cohere, "config", "submodule.typescript-go.ignore", "all")
	gitFixture(t, cohere, "checkout", "--detach", old)
	err = shard(0, 2, filepath.Join(t.TempDir(), "out"), t.TempDir(), false)
	if err == nil || !strings.Contains(err.Error(), "submodule cohere: checked-out commit") {
		t.Fatal("stale cohere checkout was not refused before discovery", err)
	}
	t.Log(err)
	// The old checkout has a different nested pin. Expected pins still come
	// from the tested root's gitlink, not from the old cohere HEAD.
	stale, err := inspectSubmodules(".", "HEAD")
	if err != nil || stale[1].Pinned != initial[1].Pinned {
		t.Fatalf("stale checkout redefined nested pin: %v %v", stale, err)
	}
	gitFixture(t, cohere, "checkout", "--detach", initial[0].Pinned)
	gitFixture(t, typescript, "checkout", "--detach", oldNestedPin(t, cohere, old))
	err = shard(0, 2, filepath.Join(t.TempDir(), "out"), t.TempDir(), true)
	if err == nil || !strings.Contains(err.Error(), "submodule cohere/typescript-go: checked-out commit") {
		t.Fatal("stale nested checkout accepted on resume", err)
	}
	gitFixture(t, typescript, "checkout", "--detach", initial[1].Pinned)
	for _, path := range []string{"tracked", "untracked"} {
		if err := os.WriteFile(filepath.Join(typescript, path), []byte("dirty\n"), 0600); err != nil {
			t.Fatal(err)
		}
		err = shard(0, 2, filepath.Join(t.TempDir(), "out"), t.TempDir(), false)
		if err == nil || !strings.Contains(err.Error(), "submodule cohere/typescript-go: working tree is dirty") {
			t.Fatal("dirty nested submodule accepted", err)
		}
		if path == "tracked" {
			gitFixture(t, typescript, "checkout", "--", path)
		}
	}
}

func TestUninitializedSubmoduleIsNamed(t *testing.T) {
	t.Parallel()
	root, _, nested, _ := submoduleFixture(t)
	initial, err := inspectSubmodules(root, "HEAD")
	if err != nil || len(initial) != 2 || requirePinnedSubmodules(initial) != nil {
		t.Fatalf("initialized fixture: %v %v", initial, err)
	}
	if err := os.Rename(filepath.Join(nested, ".git"), filepath.Join(t.TempDir(), "gitfile")); err != nil {
		t.Fatal(err)
	}
	_, err = inspectSubmodules(root, "HEAD")
	if err == nil || !strings.Contains(err.Error(), "cohere/typescript-go is not initialized") {
		t.Fatal("uninitialized module was mistaken for the parent repository", err)
	}
}

func TestSubmoduleSummaryDisagreement(t *testing.T) {
	t.Parallel()
	expected := []submoduleRecord{{Path: "cohere", Pinned: "a", Checked: "a"}, {Path: "cohere/typescript-go", Pinned: "b", Checked: "b"}}
	baseline := []summary{{Index: 0, Submodules: append([]submoduleRecord{}, expected...)}, {Index: 7, Submodules: append([]submoduleRecord{}, expected...)}}
	if problems := compareSubmodules(expected, baseline); len(problems) != 0 {
		t.Fatal(problems)
	}
	for _, field := range []string{"pin", "checkout", "nested", "missing", "dirty", "duplicate"} {
		t.Run(field, func(t *testing.T) {
			mutant := append([]summary{}, baseline...)
			mutant[1].Submodules = append([]submoduleRecord{}, expected...)
			switch field {
			case "pin":
				mutant[1].Submodules[0].Pinned = "other"
			case "checkout":
				mutant[1].Submodules[0].Checked = "other"
			case "nested":
				mutant[1].Submodules[1].Checked = "other"
			case "missing":
				mutant[1].Submodules = nil
			case "dirty":
				mutant[1].Submodules[1].Dirty = true
			case "duplicate":
				mutant[1].Submodules = append(mutant[1].Submodules, expected[0])
			}
			problems := compareSubmodules(expected, mutant)
			if len(problems) == 0 {
				t.Fatal("mutant accepted", field)
			}
			message := strings.Join(problems, "\n")
			if !strings.Contains(message, "cohere") || !strings.Contains(message, "7") {
				t.Fatal("missing shard and submodule names", message)
			}
			if field == "pin" || field == "checkout" || field == "nested" {
				if !strings.Contains(message, "shards 0 and 7 disagree") {
					t.Fatal(message)
				}
			}
			t.Log(message)
		})
	}
}

// Not parallel: the real shard and merge entry points use process-wide cwd/env.
func TestMergeRejectsDifferentCohereCommits(t *testing.T) {
	root, _, _, old := submoduleFixture(t)
	t.Chdir(root)
	t.Setenv("GOWORK", "off")
	// The tiny fixture uses no external corpora or SDK. Avoid hashing an entire
	// fleet tools directory as a fixture input; executable hashes still apply.
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("ADAMIC_MARKDOWNWIDTH_DEPS", "")
	t.Setenv("WASI_SYSROOT", "")
	originalTemporary := os.Getenv("TMPDIR")
	t.Cleanup(func() { os.Setenv("TMPDIR", originalTemporary) })
	for _, directory := range []string{"cmd/adamic-gate", "internal"} {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for name, contents := range map[string]string{
		"go.mod":        "module provenanceprobe\n\ngo 1.27\n",
		"probe_test.go": "package provenanceprobe\nimport \"testing\"\nfunc TestProbe(t *testing.T) {}\n",
		timingPath:      "{}\n",
	} {
		if err := os.WriteFile(name, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	gitFixture(t, root, "add", ".")
	gitFixture(t, root, "commit", "-qm", "Small gate fixture")
	evidence := t.TempDir()
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	scratch, err := os.MkdirTemp(cache, "gate-submodule-scratch-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(scratch) })
	dirs := []string{filepath.Join(evidence, "shard-0"), filepath.Join(evidence, "shard-1")}
	for index, dir := range dirs {
		if err := shard(index, 2, dir, scratch, false); err != nil {
			t.Fatal(err)
		}
		var s summary
		if err := loadJSON(filepath.Join(dir, "summary.json"), &s); err != nil {
			t.Fatal(err)
		}
		if len(s.Submodules) != 2 || requirePinnedSubmodules(s.Submodules) != nil {
			t.Fatal("shard did not record recursive provenance", s.Submodules)
		}
	}
	if err := merge(dirs, filepath.Join(evidence, "green")); err != nil {
		t.Fatal("baseline merge", err)
	}
	var s summary
	if err := loadJSON(filepath.Join(dirs[1], "summary.json"), &s); err != nil {
		t.Fatal(err)
	}
	s.Submodules[0].Checked = old
	if err := saveJSON(filepath.Join(dirs[1], "summary.json"), s); err != nil {
		t.Fatal(err)
	}
	red := filepath.Join(evidence, "red")
	if err := merge(dirs, red); err == nil {
		t.Fatal("conflicting cohere summaries merged green")
	}
	var m merged
	if err := loadJSON(filepath.Join(red, "merged.json"), &m); err != nil {
		t.Fatal(err)
	}
	if m.Green || !strings.Contains(strings.Join(m.Errors, "\n"), "shards 0 and 1 disagree on submodule cohere") {
		t.Fatal("red merge did not name both shards and cohere", m.Errors)
	}
	t.Log(strings.Join(m.Errors, "\n"))
}

// Fixture repositories must not inherit a caller's object store, template,
// index, hooks or signing configuration. In particular GIT_OBJECT_DIRECTORY
// can send commit writes to a read-only store even when t.TempDir is writable.
func fixtureGitOutput(directory string, environment []string, args ...string) (string, error) {
	cmd, release := boundedrun.Command(boundedrun.Probe, "git", append([]string{
		"-c", "init.templateDir=", "-c", "core.hooksPath=", "-c", "commit.gpgSign=false", "-C", directory,
	}, args...)...)
	defer release()
	cmd.Env = []string{}
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "GIT_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_TERMINAL_PROMPT=0")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git -C %s %v: %w\n%s", directory, args, err, b)
	}
	return strings.TrimSpace(string(b)), nil
}

func gitFixture(t *testing.T, directory string, args ...string) string {
	t.Helper()
	value, err := fixtureGitOutput(directory, os.Environ(), args...)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestGitFixtureOwnsItsObjectDirectory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	blocked := filepath.Join(t.TempDir(), "not-an-object-directory")
	if err := os.WriteFile(blocked, []byte("outside fixture\n"), 0444); err != nil {
		t.Fatal(err)
	}
	template := t.TempDir()
	if err := os.Symlink(blocked, filepath.Join(template, "objects")); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(config, []byte("[init]\n\ttemplateDir = "+filepath.ToSlash(template)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	environment := append(os.Environ(), "GIT_OBJECT_DIRECTORY="+blocked,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES="+blocked, "GIT_TEMPLATE_DIR="+template,
		"GIT_CONFIG_GLOBAL="+config, "GIT_INDEX_FILE="+blocked)
	for _, args := range [][]string{
		{"init", "-q"}, {"config", "user.name", "Gate fixture"},
		{"config", "user.email", "gate@example.invalid"},
		{"commit", "--allow-empty", "-qm", "Isolated fixture"},
	} {
		if _, err := fixtureGitOutput(root, environment, args...); err != nil {
			t.Fatal(err)
		}
	}
	objects, err := os.Lstat(filepath.Join(root, ".git", "objects"))
	if err != nil || !objects.IsDir() {
		t.Fatalf("object store is not a private directory: %v %v", objects, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".git", "objects", "info", "alternates")); !os.IsNotExist(err) {
		t.Fatalf("fixture unexpectedly uses alternates: %v", err)
	}
}

func oldNestedPin(t *testing.T, cohere, commit string) string {
	t.Helper()
	row := gitFixture(t, cohere, "ls-tree", commit, "typescript-go")
	return strings.Fields(row)[2]
}

func submoduleFixture(t *testing.T) (root, cohere, typescript, old string) {
	t.Helper()
	root = t.TempDir()
	initRepo := func(directory string) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		gitFixture(t, directory, "init", "-q")
		gitFixture(t, directory, "config", "user.name", "Gate fixture")
		gitFixture(t, directory, "config", "user.email", "gate@example.invalid")
	}
	write := func(directory, name, value string) {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
		gitFixture(t, directory, "add", name)
		gitFixture(t, directory, "commit", "-qm", "Fixture commit")
	}
	nestedSource := t.TempDir()
	initRepo(nestedSource)
	write(nestedSource, "tracked", "first\n")
	parentSource := t.TempDir()
	initRepo(parentSource)
	gitFixture(t, parentSource, "-c", "protocol.file.allow=always", "submodule", "add", nestedSource, "typescript-go")
	gitFixture(t, parentSource, "commit", "-qm", "First nested pin")
	old = gitFixture(t, parentSource, "rev-parse", "HEAD")
	write(nestedSource, "tracked", "second\n")
	gitFixture(t, filepath.Join(parentSource, "typescript-go"), "fetch", "origin")
	gitFixture(t, filepath.Join(parentSource, "typescript-go"), "checkout", "--detach", gitFixture(t, nestedSource, "rev-parse", "HEAD"))
	gitFixture(t, parentSource, "add", "typescript-go")
	gitFixture(t, parentSource, "commit", "-qm", "Second nested pin")
	initRepo(root)
	gitFixture(t, root, "-c", "protocol.file.allow=always", "submodule", "add", parentSource, "cohere")
	gitFixture(t, root, "commit", "-qm", "Pin cohere")
	gitFixture(t, root, "-c", "protocol.file.allow=always", "submodule", "update", "--init", "--recursive")
	return root, filepath.Join(root, "cohere"), filepath.Join(root, "cohere", "typescript-go"), old
}
