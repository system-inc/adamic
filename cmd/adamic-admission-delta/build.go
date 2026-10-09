package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		Flags:     []string{sha, fmt.Sprintf("adapter=%x", adapter), "-trimpath", "-buildvcs=false", "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT")},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
}

func cachedCompiler(root, scratch, sha, name string) (string, error) {
	directory, err := buildcache.Get(compilerInputs(sha), func(product string) error {
		tree := filepath.Join(scratch, name)
		if _, err := git(root, "worktree", "add", "--detach", tree, sha); err != nil {
			return err
		}
		defer git(root, "worktree", "remove", "--force", tree)
		// Local worktrees reuse pinned objects without cloning the checker history.
		cohereSHA, err := git(root, "rev-parse", sha+":cohere")
		if err == nil {
			cohereRoot := filepath.Join(root, "cohere")
			cohereTree := filepath.Join(tree, "cohere")
			if _, err = git(cohereRoot, "worktree", "add", "--detach", cohereTree, cohereSHA); err != nil {
				return err
			}
			defer git(cohereRoot, "worktree", "remove", "--force", cohereTree)
			typescriptSHA, err := git(cohereTree, "rev-parse", "HEAD:TypeScript")
			if err != nil {
				return err
			}
			typescriptRoot := filepath.Join(cohereRoot, "TypeScript")
			typescriptTree := filepath.Join(cohereTree, "TypeScript")
			if _, err = git(typescriptRoot, "worktree", "add", "--detach", typescriptTree, typescriptSHA); err != nil {
				return err
			}
			defer git(typescriptRoot, "worktree", "remove", "--force", typescriptTree)
		}
		if err := os.WriteFile(filepath.Join(tree, "cmd", "adamic", "admission_adapter.go"), []byte(loweringAdapter), 0600); err != nil {
			return err
		}
		built := execute(tree, 5*time.Minute, "go", "build", "-trimpath", "-buildvcs=false", "-o", filepath.Join(product, "adamic"), "./cmd/adamic")
		if built.Exit != 0 || built.Error != "" {
			return fmt.Errorf("build %s: %s %s", sha, built.Error, built.Stderr)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "adamic"), nil
}
