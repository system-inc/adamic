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
	if changeModeError := os.Chmod(filepath.Join(directory, "node_modules/a/cli.js"), 0o755); changeModeError != nil {
		test.Fatal(changeModeError)
	}
	if createDirectoryError := os.MkdirAll(
		filepath.Join(directory, "node_modules/.bin"),
		0o755,
	); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if symlinkError := os.Symlink("../a/cli.js", filepath.Join(directory, "node_modules/.bin/a")); symlinkError != nil {
		test.Fatal(symlinkError)
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
	rig := newRig(test)
	write(test, rig.root, "web/package-lock.json", "{}\n")
	write(test, rig.root, "vendored/package-lock.json", "{}\n")
	write(test, rig.root, "vendored/node_modules/c.js", "c\n")
	gitIn(test, rig.root, "add", "web/package-lock.json", "vendored/package-lock.json", "vendored/node_modules/c.js")
	gitIn(test, rig.root, "commit", "-q", "-m", "a project's lockfile, and a node_modules git tracks")
	forgetTracked()
	writeInstall(test, filepath.Join(rig.root, "web"))
	write(test, rig.root, "web/node_modules/a/node_modules/b/index.js", "b\n")
	Product(test, port, building("built"))
	rig.window(rig.built(test)["port"].Build, func() {
		rig.open(rig.pid, "source/a.c")
		rig.open(rig.pid, "web/node_modules/a/index.js")
		rig.open(rig.pid, "web/node_modules/a/node_modules/b/index.js")
		rig.missing(rig.pid, "web/node_modules/a/index.json")
		rig.open(rig.pid, "vendored/node_modules/c.js")
	})
	rig.add(rig.pid, "+++ exited with 0 +++")
	if settlement := rig.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	entries := rig.entries(test, NameKey(rig.root, port))
	if !slices.Contains(entries, "node install web") ||
		!slices.Contains(entries, "content vendored/node_modules/c.js") ||
		slices.ContainsFunc(entries, func(entry string) bool { return strings.Contains(entry, "web/node_modules") }) {
		test.Fatalf("entries %q: want the install as one entry and the tracked node_modules as a content read", entries)
	}
	if _, readingError := reading(test, port); readingError != nil {
		test.Fatalf("an unchanged install: %v", readingError)
	}
	write(test, rig.root, "web/node_modules/a/index.js", "module.exports = 2\n")
	if _, readingError := reading(test, port); !errors.Is(readingError, ErrNotBuilt) {
		test.Fatalf("a changed file in the install: %v", readingError)
	}
}

// A read through a link inside an installed node_modules, npm's .bin, is a read of the install, never an untracked link
// refused (#4gp2ytz: stage3/api's .bin/tsc refused Compiler's adapted tree), and a lookup misses when the link points
// elsewhere. A link that leads out of the install is the install and what it leads to, keyed in its place. Mutants that
// fail it: the link walk asking for the install after the link (refused untracked), and the target outside left out
// (the tree file it reads unkeyed).
// Not parallel: newRig.
func TestAReadThroughALinkInAnInstalledNodeModulesKeysAsItsInstall(test *testing.T) {
	rig := newRig(test)
	write(test, rig.root, "web/package-lock.json", "{}\n")
	gitIn(test, rig.root, "add", "web/package-lock.json")
	gitIn(test, rig.root, "commit", "-q", "-m", "a project's lockfile")
	forgetTracked()
	writeInstall(test, filepath.Join(rig.root, "web"))
	if symlinkError := os.Symlink(
		"../../../source/a.c",
		filepath.Join(rig.root, "web/node_modules/.bin/out"),
	); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	Product(test, port, building("built"))
	rig.window(rig.built(test)["port"].Build, func() {
		rig.open(rig.pid, "web/node_modules/.bin/a")
		rig.open(rig.pid, "web/node_modules/.bin/out")
	})
	rig.add(rig.pid, "+++ exited with 0 +++")
	if settlement := rig.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	entries := rig.entries(test, NameKey(rig.root, port))
	if !slices.Contains(entries, "node install web") || !slices.Contains(entries, "content source/a.c") ||
		slices.ContainsFunc(entries, func(entry string) bool { return strings.Contains(entry, "web/node_modules") }) {
		test.Fatalf("entries %q: want the install and the tree file the outward link reads", entries)
	}
	if _, readingError := reading(test, port); readingError != nil {
		test.Fatalf("an unchanged install: %v", readingError)
	}
	link := filepath.Join(rig.root, "web/node_modules/.bin/a")
	os.Remove(link)
	if symlinkError := os.Symlink("../a/index.js", link); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	if _, readingError := reading(test, port); !errors.Is(readingError, ErrNotBuilt) {
		test.Fatalf("a .bin link pointed elsewhere: %v", readingError)
	}
}
