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
		Name: "admission-compiler",
		Flags: []string{
			sha,
			fmt.Sprintf("adapter=%x", adapter),
			"-trimpath",
			"-buildvcs=false",
			"GOFLAGS=" + compilerGoFlags(),
			"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
			"GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"),
		},
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
	directory, getError := buildcache.Get(compilerInputs(sha), func(product string) error {
		return buildCompiler(root, scratch, sha, name, product)
	})
	if getError != nil {
		return "", getError
	}
	return filepath.Join(directory, "adamic"), nil
}

func buildCompiler(root, scratch, sha, name, product string) error {
	tree := filepath.Join(scratch, name)
	if _, gitError := git(root, "worktree", "add", "--detach", tree, sha); gitError != nil {
		return gitError
	}
	defer git(root, "worktree", "remove", "--force", tree)
	// Local worktrees reuse pinned objects without cloning the checker history.
	cohereSHA, gitError := git(root, "rev-parse", sha+":cohere")
	if gitError == nil {
		cohereRoot := filepath.Join(root, "cohere")
		cohereTree := filepath.Join(tree, "cohere")
		if _, gitError = git(cohereRoot, "worktree", "add", "--detach", cohereTree, cohereSHA); gitError != nil {
			return gitError
		}
		defer git(cohereRoot, "worktree", "remove", "--force", cohereTree)
		typescriptSHA, gitError := git(cohereTree, "rev-parse", "HEAD:TypeScript")
		if gitError != nil {
			return gitError
		}
		typescriptRoot := filepath.Join(cohereRoot, "TypeScript")
		typescriptTree := filepath.Join(cohereTree, "TypeScript")
		if _, gitError = git(
			typescriptRoot,
			"worktree",
			"add",
			"--detach",
			typescriptTree,
			typescriptSHA,
		); gitError != nil {
			return gitError
		}
		defer git(typescriptRoot, "worktree", "remove", "--force", typescriptTree)
	}
	if writeError := os.WriteFile(
		filepath.Join(tree, "cmd", "adamic", "admission_adapter.go"),
		[]byte(loweringAdapter),
		0600,
	); writeError != nil {
		return writeError
	}
	built := execute(
		tree,
		5*time.Minute,
		"go",
		"build",
		"-trimpath",
		"-buildvcs=false",
		"-o",
		filepath.Join(product, "adamic"),
		"./cmd/adamic",
	)
	if built.Exit != 0 || built.Error != "" {
		return fmt.Errorf("build %s: %s %s", sha, built.Error, built.Stderr)
	}
	return nil
}
