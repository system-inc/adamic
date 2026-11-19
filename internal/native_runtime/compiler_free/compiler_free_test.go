// Package compiler_free holds the guard that internal/native_runtime imports no package of the compiler, so a runtime
// archive's key, which holds the runtime's import closure, never moves with the compiler (#tb722h2 H2). It sits apart
// from package native_runtime, so its test is in no runtime archive's key, as source_paths' is.
package compiler_free

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// compilerPackages are the compiler's packages the runtime may never import, each with every package under it.
// internal/native_runtime isn't under internal/native: a package path is under another only past a slash.
var compilerPackages = []string{
	"github.com/system-inc/adamic/internal/native",
	"github.com/system-inc/adamic/internal/ir",
	"github.com/system-inc/adamic/internal/flow",
	"github.com/system-inc/adamic/internal/lower",
	"github.com/system-inc/adamic/internal/javascript",
}

// internal/native_runtime, its own files without its tests, imports no package of the compiler, directly or through
// another: go list's dependencies of it name none of compilerPackages, nor any package under one. The archive's build
// moved here from internal/native so its key would hold toolchain, buildcache/location and the runtime alone, and one
// import of the compiler would put every compiler edit back in the key of every archive. Mutant
// (compiler_free.mutant.md): a blank import of internal/ir in options.go, which this names.
func TestTheRuntimeImportsNoCompilerPackage(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	listed, err := toolchain.CommandIn(root, "go", "list", "-deps", "-f", "{{.ImportPath}}", "./internal/native_runtime").Output()
	if err != nil {
		t.Fatalf("go list -deps ./internal/native_runtime: %v", err)
	}
	dependencies := strings.Fields(string(listed))
	if !slices.Contains(dependencies, "github.com/system-inc/adamic/internal/native_runtime") {
		t.Fatalf("go list's dependencies of internal/native_runtime don't hold it, so they aren't its: %q", dependencies)
	}
	var found []string
	for _, dependency := range dependencies {
		for _, compiler := range compilerPackages {
			if dependency == compiler || strings.HasPrefix(dependency, compiler+"/") {
				found = append(found, dependency)
			}
		}
	}
	if len(found) > 0 {
		t.Fatalf("internal/native_runtime imports the compiler, so every compiler edit moves every runtime archive's key: %s", strings.Join(found, ", "))
	}
}
