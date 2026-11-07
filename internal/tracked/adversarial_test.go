package tracked

// The independent review's adversarial cases for a source with no .git (#cyasrr4), kept as tests: hostile names
// survive -z parsing, and each change git sees is seen here.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

func gg(test *testing.T, dir string, args ...string) string {
	test.Helper()
	c := toolchain.GitCommand(append(
		[]string{"-C", dir, "-c", "user.name=x", "-c", "user.email=x@x", "-c", "commit.gpgsign=false"},
		args...,
	)...)
	out, err := c.CombinedOutput()
	if err != nil {
		test.Fatalf("git %v: %v %s", args, err, out)
	}
	return string(out)
}

func TestAdversarialEntries(test *testing.T) {
	test.Parallel()
	repo := test.TempDir()
	gg(test, repo, "init", "-q")
	names := []string{
		"tab\there.md",
		"new\nline.md",
		" lead space.md",
		"trail space .md",
		"\"quoted\".md",
		"unié.md",
		"Case.md",
	}
	for _, n := range names {
		os.WriteFile(filepath.Join(repo, n), []byte(n), 0o644)
	}
	os.WriteFile(filepath.Join(repo, "run.sh"), []byte("#!/bin/sh\n"), 0o755)
	os.Symlink("Case.md", filepath.Join(repo, "link"))
	gg(test, repo, "add", ".")
	gg(test, repo, "commit", "-qm", "1")
	source := test.TempDir()
	if err := Write(repo, source); err != nil {
		test.Fatal(err)
	}
	// unpack via git archive
	arch := exec.Command("sh", "-c", "git -C "+repo+" archive HEAD | tar -x -C "+source)
	arch.Env = toolchain.GitEnv()
	if out, err := arch.CombinedOutput(); err != nil {
		test.Fatal(err, string(out))
	}
	entries, err := Files(source)
	if err != nil {
		test.Fatal(err)
	}
	var paths []string
	for _, e := range entries {
		paths = append(paths, e.Mode+" "+e.Path)
	}
	test.Logf("entries: %q", paths)
	missing, changed, err := Changed(source, entries)
	test.Logf("clean: missing=%q changed=%q err=%v", missing, changed, err)
	if len(missing)+len(changed) != 0 || err != nil {
		test.Errorf("clean tree not clean")
	}
	if len(entries) != len(names)+2 {
		test.Errorf("%d entries for %d files", len(entries), len(names)+2)
	}
	check := func(name string, mutate func(), want ...string) {
		test.Helper()
		mutate()
		m, c, err := Changed(source, entries)
		test.Logf("%s: missing=%q changed=%q err=%v", name, m, c, err)
		if err != nil || len(m) != 0 || strings.Join(c, "|") != strings.Join(want, "|") {
			test.Errorf("%s: changed %q, git would say %q", name, c, want)
		}
	}
	src := func(n string) string { return filepath.Join(source, n) }
	check("group-only x on 100644 file (git: unchanged)", func() { os.Chmod(src("Case.md"), 0o654) })
	os.Chmod(src("Case.md"), 0o644)
	check("other-x only on 100755 file (git: mode change)", func() { os.Chmod(src("run.sh"), 0o645) }, "run.sh")
	os.Chmod(src("run.sh"), 0o755)
	check("symlink retargeted", func() { os.Remove(src("link")); os.Symlink("tab\there.md", src("link")) }, "link")
	check(
		"symlink replaced by regular file with target text",
		func() { os.Remove(src("link")); os.WriteFile(src("link"), []byte("Case.md"), 0o644) },
		"link",
	)
	os.Remove(src("link"))
	os.Symlink("Case.md", src("link"))
	check("tab-name content changed", func() { os.WriteFile(src("tab\there.md"), []byte("zz"), 0o644) }, "tab\there.md")
}
