package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeInstall makes an installed node_modules under directory, as npm ci leaves one: a package's files, one of them
// executable, and a .bin link to it.
func writeInstall(test *testing.T, directory string) {
	test.Helper()
	write(test, directory, "node_modules/a/index.js", "module.exports = 1\n")
	write(test, directory, "node_modules/a/cli.js", "#!/usr/bin/env node\n")
	if err := os.Chmod(filepath.Join(directory, "node_modules/a/cli.js"), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "node_modules/.bin"), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.Symlink("../a/cli.js", filepath.Join(directory, "node_modules/.bin/a")); err != nil {
		test.Fatal(err)
	}
}

// An installed node_modules is valued by what's there, as Loom's manifest names it (#qx8fkaf): the same install at
// two checkout paths values alike, and a changed byte, executable bit or link target, or a file no lockfile names,
// each moves it. Mutants that each fail it: the value without the executable bit; without links; a file whose path
// no lockfile names left out.
func TestANodeInstallIsValuedByWhatIsThere(test *testing.T) {
	test.Parallel()
	one, other := test.TempDir(), test.TempDir()
	writeInstall(test, one)
	writeInstall(test, other)
	base := nodeInstallValue(filepath.Join(one, "node_modules"))
	if !strings.HasPrefix(base, "install ") || nodeInstallValue(filepath.Join(other, "node_modules")) != base {
		test.Fatalf("one install at two paths: %s and %s", base, nodeInstallValue(filepath.Join(other, "node_modules")))
	}
	for _, change := range []struct {
		name string
		make func(directory string)
	}{
		{
			"a changed byte",
			func(directory string) { write(test, directory, "node_modules/a/index.js", "module.exports = 2\n") },
		},
		{
			"the executable bit",
			func(directory string) { os.Chmod(filepath.Join(directory, "node_modules/a/cli.js"), 0o644) },
		},
		{"a link's target", func(directory string) {
			link := filepath.Join(directory, "node_modules/.bin/a")
			os.Remove(link)
			os.Symlink("../a/index.js", link)
		}},
		{
			"a file no lockfile names",
			func(directory string) { write(test, directory, "node_modules/a/extra.js", "\n") },
		},
	} {
		changed := test.TempDir()
		writeInstall(test, changed)
		change.make(changed)
		if value := nodeInstallValue(filepath.Join(changed, "node_modules")); value == base {
			test.Errorf("%s left the install's value as it was: %s", change.name, value)
		}
	}
	if value := nodeInstallValue(filepath.Join(test.TempDir(), "node_modules")); value != "absent" {
		test.Errorf("no install: %s", value)
	}
}

// A traced build that reads a project's installed node_modules, which git doesn't track beside a package-lock.json it
// does, settles with one entry for the install, never refused as untracked, a nested node_modules the same install as
// the first; a lookup finds the product while the install is what it was and misses it when a file in it changes. A
// node_modules git tracks stays a tree read, lockfile or not (#qx8fkaf). Mutants that each fail it: the install taken
// past the first node_modules, which names no project and is refused; a tracked node_modules taken as an install.
// Not parallel: newRig.
func TestAReadUnderAnInstalledNodeModulesKeysAsItsInstall(test *testing.T) {
	r := newRig(test)
	write(test, r.root, "web/package-lock.json", "{}\n")
	write(test, r.root, "vendored/package-lock.json", "{}\n")
	write(test, r.root, "vendored/node_modules/c.js", "c\n")
	gitIn(test, r.root, "add", "web/package-lock.json", "vendored/package-lock.json", "vendored/node_modules/c.js")
	gitIn(test, r.root, "commit", "-q", "-m", "a project's lockfile, and a node_modules git tracks")
	forgetTracked()
	writeInstall(test, filepath.Join(r.root, "web"))
	write(test, r.root, "web/node_modules/a/node_modules/b/index.js", "b\n")
	Product(test, port, building("built"))
	r.window(r.built(test)["port"].Build, func() {
		r.open(r.pid, "source/a.c")
		r.open(r.pid, "web/node_modules/a/index.js")
		r.open(r.pid, "web/node_modules/a/node_modules/b/index.js")
		r.missing(r.pid, "web/node_modules/a/index.json")
		r.open(r.pid, "vendored/node_modules/c.js")
	})
	r.add(r.pid, "+++ exited with 0 +++")
	if settlement := r.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	entries := r.entries(test, NameKey(r.root, port))
	if !slices.Contains(entries, "node install web") ||
		!slices.Contains(entries, "content vendored/node_modules/c.js") ||
		slices.ContainsFunc(entries, func(entry string) bool { return strings.Contains(entry, "web/node_modules") }) {
		test.Fatalf("entries %q: want the install as one entry and the tracked node_modules as a content read", entries)
	}
	if _, err := reading(test, port); err != nil {
		test.Fatalf("an unchanged install: %v", err)
	}
	write(test, r.root, "web/node_modules/a/index.js", "module.exports = 2\n")
	if _, err := reading(test, port); !errors.Is(err, ErrNotBuilt) {
		test.Fatalf("a changed file in the install: %v", err)
	}
}

// A read through a link inside an installed node_modules, npm's .bin, is a read of the install, never an untracked link
// refused (#4gp2ytz: stage3/api's .bin/tsc refused Compiler's adapted tree), and a lookup misses when the link points
// elsewhere. A link that leads out of the install is the install and what it leads to, keyed in its place. Mutants that
// fail it: the link walk asking for the install after the link (refused untracked), and the target outside left out
// (the tree file it reads unkeyed).
// Not parallel: newRig.
func TestAReadThroughALinkInAnInstalledNodeModulesKeysAsItsInstall(test *testing.T) {
	r := newRig(test)
	write(test, r.root, "web/package-lock.json", "{}\n")
	gitIn(test, r.root, "add", "web/package-lock.json")
	gitIn(test, r.root, "commit", "-q", "-m", "a project's lockfile")
	forgetTracked()
	writeInstall(test, filepath.Join(r.root, "web"))
	if err := os.Symlink("../../../source/a.c", filepath.Join(r.root, "web/node_modules/.bin/out")); err != nil {
		test.Fatal(err)
	}
	Product(test, port, building("built"))
	r.window(r.built(test)["port"].Build, func() {
		r.open(r.pid, "web/node_modules/.bin/a")
		r.open(r.pid, "web/node_modules/.bin/out")
	})
	r.add(r.pid, "+++ exited with 0 +++")
	if settlement := r.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	entries := r.entries(test, NameKey(r.root, port))
	if !slices.Contains(entries, "node install web") || !slices.Contains(entries, "content source/a.c") ||
		slices.ContainsFunc(entries, func(entry string) bool { return strings.Contains(entry, "web/node_modules") }) {
		test.Fatalf("entries %q: want the install and the tree file the outward link reads", entries)
	}
	if _, err := reading(test, port); err != nil {
		test.Fatalf("an unchanged install: %v", err)
	}
	link := filepath.Join(r.root, "web/node_modules/.bin/a")
	os.Remove(link)
	if err := os.Symlink("../a/index.js", link); err != nil {
		test.Fatal(err)
	}
	if _, err := reading(test, port); !errors.Is(err, ErrNotBuilt) {
		test.Fatalf("a .bin link pointed elsewhere: %v", err)
	}
}
