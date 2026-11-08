package main

import (
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeDeclarationsArePinned(t *testing.T) {
	t.Parallel()
	api := t.TempDir()
	dir := filepath.Join(api, "node_modules/@types/node")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.d.ts"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(dir, "package.json")
	if err := os.WriteFile(metadata, []byte(`{"name":"@types/node","version":"25.3.3"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNodeTypes(api); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadata, []byte(`{"name":"@types/node","version":"24.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNodeTypes(api); err == nil || !strings.Contains(err.Error(), "want @types/node") {
		t.Fatal("wrong declarations accepted", err)
	}
}

// stage3/api is provisioned by the gates' own npm, found as a build's tools are, and a box without one is told so
// by name. Mutant: execute runs plain "npm" from the caller's PATH again (the sandbox's refusal never appears).
//
// Not parallel: t.Setenv sets the process-wide ADAMIC_SANDBOX and ADAMIC_TOOLS.
func TestProvisioningFindsNpmAsBuildsFindTools(t *testing.T) {
	tree := t.TempDir()
	api := filepath.Join(tree, "stage3", "api")
	if err := os.MkdirAll(api, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(api, "package-lock.json"), []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_SANDBOX", "1")
	t.Setenv("ADAMIC_TOOLS", t.TempDir())
	err := prepareCheckout(t.TempDir(), tree)
	if err == nil || !strings.Contains(err.Error(), "provision stage3/api: npm: ADAMIC_SANDBOX=1 runs only tools under ADAMIC_TOOLS") ||
		!strings.Contains(err.Error(), "was not used: this checkout pins no install") {
		t.Fatalf("provisioning without the gates' npm: %v, want the missing npm and the unused laid install named", err)
	}
}

func TestPreparedNodeTypesUseHostAxioms(t *testing.T) {
	t.Parallel()
	api := t.TempDir()
	files, err := load.NodeHostAxiomFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		published, err := os.ReadFile(filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(api, "node_modules", "@types", "node", file)
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, published, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := correctPreparedNodeTypes(api); err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		published, err := os.ReadFile(filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		want, err := load.CorrectNodeHostAxioms(file, string(published))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(api, "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s: prepared checker differs from bundled checker", file)
		}
	}
}

// fakeStage3 lays a tree's stage3 as provisioning reads it: the repository's own apply.py and caches.py, beside an api
// whose lockfile names one fake package and @types/node, in the shape apply.py's installed_packages accepts.
func fakeStage3(t *testing.T) string {
	t.Helper()
	tree := t.TempDir()
	api := filepath.Join(tree, "stage3", "api")
	if err := os.MkdirAll(api, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"apply.py", "caches.py"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "stage3", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tree, "stage3", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"package.json":      `{"name":"fake","version":"1.0.0","dependencies":{"@types/node":"25.3.3","fake":"1.0.0"}}`,
		"package-lock.json": `{"name":"fake","lockfileVersion":3,"packages":{"":{"name":"fake"},"node_modules/@types/node":{"version":"25.3.3"},"node_modules/fake":{"version":"1.0.0"}}}`,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(api, name), []byte(text+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return tree
}

// layInstall lays the install fakeStage3's lockfile names under tree's stage3/api/node_modules: npm's hidden lockfile,
// a file, an executable, a nested directory, node_modules/.bin's relative link, and @types/node as published, so
// verifyNodeTypes and correctPreparedNodeTypes accept it.
func layInstall(t *testing.T, tree string) {
	t.Helper()
	modules := filepath.Join(tree, "stage3", "api", "node_modules")
	files := map[string]string{
		".package-lock.json":       `{"packages":{"node_modules/@types/node":{"version":"25.3.3"},"node_modules/fake":{"version":"1.0.0"}}}`,
		"fake/package.json":        `{"version":"1.0.0"}`,
		"fake/bin/run.js":          "#!/usr/bin/env node\n",
		"fake/lib/deep/index.js":   "module.exports = 1\n",
		"@types/node/package.json": `{"name":"@types/node","version":"25.3.3"}`,
		"@types/node/index.d.ts":   "",
	}
	hostFiles, err := load.NodeHostAxiomFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range hostFiles {
		published, err := os.ReadFile(filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		files["@types/node/"+file] = string(published)
	}
	for name, text := range files {
		path := filepath.Join(modules, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(filepath.Join(modules, "fake", "bin", "run.js"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(modules, ".bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../fake/bin/run.js", filepath.Join(modules, ".bin", "fake")); err != nil {
		t.Fatal(err)
	}
}

// pinnedPair is a root with a laid install and a checkout with the same lockfile, both pinning that install's digest
// as the root's own apply.py computes it.
func pinnedPair(t *testing.T) (root, tree, pin string) {
	t.Helper()
	root, tree = fakeStage3(t), fakeStage3(t)
	layInstall(t, root)
	pin, err := packagesDigest(root, toolchain.Path)
	if err != nil {
		t.Fatal(err)
	}
	for _, each := range []string{root, tree} {
		if err := os.WriteFile(filepath.Join(each, "stage3", "api", "node_modules.sha256"), []byte(pin+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root, tree, pin
}

// withoutNpm finds tools as a build does, except npm, which no house box has, so a box with npm still takes the roads
// a house box does.
func withoutNpm(name string) (string, error) {
	if name == "npm" {
		return "", errors.New("npm: not on this box")
	}
	return toolchain.Path(name)
}

// A checkout whose pin is the laid install's digest gets a real copy of it, with its executable bit and its relative
// link, and the host corrections land in the checkout's @types/node, never in the proof job's own. Mutant: a symlink
// instead of a copy (the checkout's node_modules is a link, and the root's @types/node is rewritten).
func TestProvisioningCopiesThePinnedLaidInstall(t *testing.T) {
	t.Parallel()
	root, tree, _ := pinnedPair(t)
	if err := provisionCheckout(root, tree, withoutNpm, copyInstall); err != nil {
		t.Fatal(err)
	}
	modules := filepath.Join(tree, "stage3", "api", "node_modules")
	if info, err := os.Lstat(modules); err != nil || !info.IsDir() {
		t.Fatalf("the checkout's node_modules is not a directory of its own: %v", err)
	}
	if link, err := os.Readlink(filepath.Join(modules, ".bin", "fake")); err != nil || link != "../fake/bin/run.js" {
		t.Fatalf("node_modules/.bin/fake: %q %v, want the relative link kept", link, err)
	}
	if info, err := os.Stat(filepath.Join(modules, "fake", "bin", "run.js")); err != nil || info.Mode().Perm()&0100 == 0 {
		t.Fatalf("fake/bin/run.js lost its executable bit: %v", err)
	}
	files, err := load.NodeHostAxiomFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		published, err := os.ReadFile(filepath.Join("..", "..", "internal", "load", "node_types", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		laid, err := os.ReadFile(filepath.Join(root, "stage3", "api", "node_modules", "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		if string(laid) != string(published) {
			t.Fatalf("%s: the laid install was rewritten", file)
		}
		want, err := load.CorrectNodeHostAxioms(file, string(published))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(modules, "@types", "node", file))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s: the checkout's copy is not corrected", file)
		}
	}
}

// A laid install that doesn't digest to the checkout's pin is never copied, and a box without npm is told both why npm
// can't install and why the laid install wasn't used. Mutant: copy before checking the digest (the checkout holds a
// node_modules its pin never named).
func TestProvisioningWithoutNpmNamesBothRoads(t *testing.T) {
	t.Parallel()
	root, tree, pin := pinnedPair(t)
	// The laid install moves on after the pin was taken, as a stale or corrupt one does.
	if err := os.WriteFile(filepath.Join(root, "stage3", "api", "node_modules", "fake", "lib", "deep", "index.js"), []byte("module.exports = 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	laid, err := packagesDigest(root, toolchain.Path)
	if err != nil {
		t.Fatal(err)
	}
	err = provisionCheckout(root, tree, withoutNpm, copyInstall)
	if err == nil || !strings.Contains(err.Error(), "provision stage3/api: npm: not on this box") ||
		!strings.Contains(err.Error(), "digest "+laid+" is not the pin "+pin) {
		t.Fatalf("provisioning with a mismatched laid install and no npm: %v, want both roads named", err)
	}
	if _, err := os.Lstat(filepath.Join(tree, "stage3", "api", "node_modules")); !os.IsNotExist(err) {
		t.Fatalf("a laid install the pin doesn't name was copied: %v", err)
	}
}

// The checkout's own apply.py reads the copy before anything corrects it, so a copy that loses an entry fails loud
// rather than having its types checked. Mutant: skip the copied tree's digest check (provisioning succeeds on a copy
// that lost node_modules/.bin's link).
func TestProvisioningVerifiesTheCopy(t *testing.T) {
	t.Parallel()
	root, tree, pin := pinnedPair(t)
	dropLink := func(source, target string) error {
		if err := copyInstall(source, target); err != nil {
			return err
		}
		return os.Remove(filepath.Join(target, ".bin", "fake"))
	}
	err := provisionCheckout(root, tree, withoutNpm, dropLink)
	if err == nil || !strings.Contains(err.Error(), "hashes to") || !strings.Contains(err.Error(), "not the pin "+pin) {
		t.Fatalf("a copy without node_modules/.bin's link: %v, want the copy's digest refused", err)
	}
}
