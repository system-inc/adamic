package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Tools come from cloud/setup.sh's environment. Checkout-local declarations
// must instead be this checkout's own pinned install: the one already laid in
// the invoking repository root when it digests to this checkout's pin, else
// npm's install from this checkout's own package lock (#zvaxp1x).
func prepareCheckout(root, tree string) error {
	return provisionCheckout(root, tree, toolchain.Path, copyInstall)
}

// provisionCheckout is prepareCheckout with the tool lookup and the copy given, so a test can take npm away on a box
// that has one, and lose an entry in the copy.
func provisionCheckout(root, tree string, tool func(name string) (string, error), copyTree func(source, target string) error) error {
	api := filepath.Join(tree, "stage3", "api")
	if _, err := os.Stat(filepath.Join(api, "package-lock.json")); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	laid := filepath.Join(root, "stage3", "api", "node_modules")
	pin, unused, err := laidInstall(root, api, tool)
	if err != nil {
		return err
	}
	if unused == "" {
		// A copy, never a link to root's: correctPreparedNodeTypes rewrites @types/node in place, and root's install
		// is the proof job's own. The copy is deliberately unbounded, since node_modules is small, while npm keeps
		// its 90 s limit.
		if err := copyTree(laid, filepath.Join(api, "node_modules")); err != nil {
			return fmt.Errorf("provision stage3/api: copy %s: %w", laid, err)
		}
		// The checkout's own apply.py reads the copy before anything corrects it, against the checkout's own
		// lockfile and pin, so a copy that lost an entry, a mode or a link fails here rather than checking types.
		copied, err := packagesDigest(tree, tool)
		if err != nil {
			return fmt.Errorf("provision stage3/api: the copy of %s: %w", laid, err)
		}
		if copied != pin {
			return fmt.Errorf("provision stage3/api: the copy of %s hashes to %s, not the pin %s", laid, copied, pin)
		}
	} else {
		// npm is the gates' own, found as a build's tools are (ADAMIC_TOOLS, then PATH), since a house box's default
		// PATH has none (Pupa, #sa5ptyj); execute runs it under toolchain.Env, whose PATH also finds the node it needs.
		npm, err := tool("npm")
		if err != nil {
			return fmt.Errorf("provision stage3/api: %w; the install laid at %s was not used: %s", err, laid, unused)
		}
		install := execute(api, 90*time.Second, npm, "ci", "--ignore-scripts")
		if install.Exit != 0 || install.Error != "" {
			return fmt.Errorf("provision stage3/api: %s %s", install.Error, install.Stderr)
		}
	}
	if err := verifyNodeTypes(api); err != nil {
		return err
	}
	return correctPreparedNodeTypes(api)
}

// laidInstall reads api's pin and says why the install laid at <root>/stage3/api/node_modules is not the one it pins,
// or "" when root's own apply.py digests it to that pin. A reason sends the checkout to npm; only a failed read of the
// checkout's own files is an error.
func laidInstall(root, api string, tool func(name string) (string, error)) (pin, unused string, err error) {
	pinned, err := os.ReadFile(filepath.Join(api, "node_modules.sha256"))
	if os.IsNotExist(err) {
		return "", "this checkout pins no install (it has no stage3/api/node_modules.sha256)", nil
	} else if err != nil {
		return "", "", err
	}
	pin = strings.TrimSpace(string(pinned))
	laid := filepath.Join(root, "stage3", "api")
	if _, err := os.Stat(filepath.Join(laid, "node_modules")); os.IsNotExist(err) {
		return pin, "it is absent", nil
	} else if err != nil {
		return pin, err.Error(), nil
	}
	// A fast precheck only, saving a digest that can't match: equal lockfiles still leave the pin to the digest.
	lockfile, err := os.ReadFile(filepath.Join(api, "package-lock.json"))
	if err != nil {
		return "", "", err
	}
	if laidLockfile, err := os.ReadFile(filepath.Join(laid, "package-lock.json")); err != nil || !bytes.Equal(laidLockfile, lockfile) {
		return pin, "its package-lock.json is not this checkout's", nil
	}
	digest, err := packagesDigest(root, tool)
	if err != nil {
		return pin, err.Error(), nil
	}
	if digest != pin {
		return pin, fmt.Sprintf("digest %s is not the pin %s", digest, pin), nil
	}
	return pin, "", nil
}

// packagesDigest is <tree>/stage3/apply.py --packages-digest: the digest of <tree>/stage3/api/node_modules, which
// also refuses an install that is not <tree>'s own lockfile's. The digest lives in apply.py alone, never in a Go
// copy that could drift from it.
func packagesDigest(tree string, tool func(name string) (string, error)) (string, error) {
	python, err := tool("python3")
	if err != nil {
		return "", err
	}
	script := filepath.Join(tree, "stage3", "apply.py")
	digest := execute(tree, 60*time.Second, python, script, "--packages-digest")
	if digest.Exit != 0 || digest.Error != "" {
		return "", fmt.Errorf("%s --packages-digest: %s %s", script, digest.Error, strings.TrimSpace(digest.Stderr))
	}
	return strings.TrimSpace(digest.Stdout), nil
}

// copyInstall copies the install at source to target, which must not exist yet: directories, regular files with
// their modes, and symlinks as symlinks with their own target text, so node_modules/.bin's relative links resolve
// inside target. A source that is itself a link (a runner's linked node_modules) is copied as the tree it names,
// as apply.py's os.walk reads it. Every copied file is owner-writable whatever the source's mode, since
// correctPreparedNodeTypes rewrites the copy's @types/node: Loom's door lays the install as read-only hard links to
// the box's one install (#t7b148j, J-2yma76w), and the digest reads only a file's executable bit, so the pin holds.
func copyInstall(source, target string) error {
	resolved, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	return filepath.WalkDir(resolved, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(resolved, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.Type()&fs.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, destination)
		case entry.IsDir():
			// Owner-writable while it fills; a directory's mode is in no digest.
			return os.Mkdir(destination, info.Mode().Perm()|0700)
		case entry.Type().IsRegular():
			return copyFile(path, destination, info.Mode().Perm()|0200)
		}
		return fmt.Errorf("%s is neither a file, a directory nor a symlink", path)
	})
}

func copyFile(source, target string, mode fs.FileMode) error {
	from, err := os.Open(source)
	if err != nil {
		return err
	}
	defer from.Close()
	to, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(to, from); err != nil {
		to.Close()
		return err
	}
	if err := to.Close(); err != nil {
		return err
	}
	// OpenFile's mode passes through the umask, and the executable bit is in the pin's digest.
	return os.Chmod(target, mode)
}
func verifyNodeTypes(api string) error {
	path := filepath.Join(api, "node_modules", "@types", "node", "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pin struct {
		Name    string
		Version string
	}
	if err = json.Unmarshal(data, &pin); err != nil {
		return err
	}
	if pin.Name != "@types/node" || pin.Version != "25.3.3" {
		return fmt.Errorf("want @types/node 25.3.3, found %s %s", pin.Name, pin.Version)
	}
	_, err = os.Stat(filepath.Join(filepath.Dir(path), "index.d.ts"))
	return err
}

// Either road lays the untouched publication first. The stock checker then sees the
// same host corrections as Adamic's bundled checker, never a second row list.
func correctPreparedNodeTypes(api string) error {
	files, err := load.NodeHostAxiomFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		name := filepath.Join(api, "node_modules", "@types", "node", filepath.FromSlash(file))
		published, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		corrected, err := load.CorrectNodeHostAxioms(file, string(published))
		if err != nil {
			return err
		}
		if err := os.WriteFile(name, []byte(corrected), 0600); err != nil {
			return err
		}
	}
	return nil
}
