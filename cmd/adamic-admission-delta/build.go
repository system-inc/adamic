package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The adapter calls the revision's own compile function, stopping after lowering.
// Keeping it in cmd/adamic preserves that revision's loading and diagnostics.
const loweringAdapter = `package main
import "os"
func init() {
 if len(os.Args) == 3 && os.Args[1] == "admission-lower" {
  _, code := compile(os.Args[2])
  os.Exit(code)
 }
}
`

func compilerInputs(sha string) buildcache.Inputs {
	adapter := sha256.Sum256([]byte(loweringAdapter))
	return buildcache.Inputs{
		Name:      "admission-compiler",
		Flags:     []string{sha, fmt.Sprintf("adapter=%x", adapter), "-trimpath", "-buildvcs=false", "GOFLAGS=" + compilerGoFlags(), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT")},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
}

// These options are explicit build arguments; their environment copies and the toolchain's
// selection policy do not change the bytes. The selected Go release is already keyed.
func compilerGoFlags() string {
	var flags []string
	for _, flag := range strings.Fields(os.Getenv("GOFLAGS")) {
		if flag == "-trimpath" || strings.HasPrefix(flag, "-buildvcs=") {
			continue
		}
		flags = append(flags, flag)
	}
	return strings.Join(flags, " ")
}

func cachedCompiler(root, scratch, sha, name string) (string, error) {
	directory, err := buildcache.Get(compilerInputs(sha), func(product string) error {
		return buildCompiler(root, scratch, sha, name, product)
	})
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "adamic"), nil
}

func buildCompiler(root, scratch, sha, name, product string) error {
	tree := filepath.Join(scratch, name)
	if _, err := git(root, "worktree", "add", "--detach", tree, sha); err != nil {
		return err
	}
	defer git(root, "worktree", "remove", "--force", tree)
	// The revision's pinned cohere and its TypeScript, each from the first repository that holds the pin: the
	// invoking repository's own submodule checkout, which reuses its objects without cloning the checker's history,
	// or else the box's mirrors, which is where a Loom door job's are, since its lay kit lays the submodules as plain
	// files with no repository of their own (#t7b148j, J-b54e1fs).
	cohereSHA, err := git(root, "rev-parse", sha+":cohere")
	if err == nil {
		cohereRoot := filepath.Join(root, "cohere")
		cohereTree := filepath.Join(tree, "cohere")
		cleanup, err := checkoutPinned("cohere", cohereSHA, cohereRoot, cohereTree)
		if err != nil {
			return err
		}
		defer cleanup()
		typescriptSHA, err := git(cohereTree, "rev-parse", "HEAD:TypeScript")
		if err != nil {
			return err
		}
		cleanup, err = checkoutPinned("TypeScript", typescriptSHA, filepath.Join(cohereRoot, "TypeScript"), filepath.Join(cohereTree, "TypeScript"))
		if err != nil {
			return err
		}
		defer cleanup()
	}
	if err := os.WriteFile(filepath.Join(tree, "cmd", "adamic", "admission_adapter.go"), []byte(loweringAdapter), 0600); err != nil {
		return err
	}
	built := execute(tree, 5*time.Minute, "go", "build", "-trimpath", "-buildvcs=false", "-o", filepath.Join(product, "adamic"), "./cmd/adamic")
	if built.Exit != 0 || built.Error != "" {
		return fmt.Errorf("build %s: %s %s", sha, built.Error, built.Stderr)
	}
	return nil
}

// submoduleMirrors are the directories that hold the box's mirrors, each as <name>.git: ADAMIC_ADMISSION_MIRRORS's
// colon-separated list when it's set, else the two the lay kit lays from (contract/proof/lay.sh): the mirrors loom's
// proof kits clone from, and where ahra loom push puts them.
func submoduleMirrors() []string {
	if listed, set := os.LookupEnv("ADAMIC_ADMISSION_MIRRORS"); set {
		return filepath.SplitList(listed)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{filepath.Join(home, "scratch", "system_loom", "modules"), filepath.Join(home, "loom-mirrors")}
}

// checkoutPinned checks name out at sha into target, from own when that's a repository of its own holding sha (a
// worktree of it, as before), else from the first of submoduleMirrors' <name>.git that holds it (a clone that shares
// the mirror's objects, so the mirror itself is only read). It hands back what undoes it. A pin no repository holds
// is named with every place looked in, never left to git's "invalid reference".
func checkoutPinned(name, sha, own, target string) (func(), error) {
	return checkoutPinnedFrom(name, sha, own, target, submoduleMirrors())
}

// checkoutPinnedFrom is checkoutPinned with the mirrors' directories given, so a test needs no environment.
func checkoutPinnedFrom(name, sha, own, target string, mirrors []string) (func(), error) {
	var looked []string
	if _, err := os.Stat(filepath.Join(own, ".git")); err == nil {
		if _, err := git(own, "cat-file", "-e", sha+"^{commit}"); err == nil {
			if _, err := git(own, "worktree", "add", "--detach", target, sha); err != nil {
				return nil, err
			}
			return func() { git(own, "worktree", "remove", "--force", target) }, nil
		}
		looked = append(looked, own+" (lacks it)")
	} else {
		looked = append(looked, own+" (not a repository of its own)")
	}
	for _, directory := range mirrors {
		mirror := filepath.Join(directory, name+".git")
		if _, err := os.Stat(mirror); err != nil {
			looked = append(looked, mirror+" (absent)")
			continue
		}
		if _, err := git(mirror, "cat-file", "-e", sha+"^{commit}"); err != nil {
			looked = append(looked, mirror+" (lacks it)")
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			return nil, err
		}
		if _, err := git(filepath.Dir(target), "clone", "--quiet", "--shared", "--no-checkout", mirror, target); err != nil {
			return nil, err
		}
		if _, err := git(target, "checkout", "--quiet", "--detach", sha); err != nil {
			os.RemoveAll(target)
			return nil, err
		}
		return func() { os.RemoveAll(target) }, nil
	}
	return nil, fmt.Errorf("admission delta: %s at %s is in no repository this box holds (looked in %s); lay it from a mirror, or set ADAMIC_ADMISSION_MIRRORS", name, sha, strings.Join(looked, ", "))
}
