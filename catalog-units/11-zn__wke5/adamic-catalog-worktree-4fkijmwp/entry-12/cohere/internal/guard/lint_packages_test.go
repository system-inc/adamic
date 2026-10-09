package guard

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"golang.org/x/tools/go/packages"
)

// lintPackages is ./internal/lint/... loaded with its syntax and types, and with its dependencies' too,
// read with overlay in place of the files it names. The creating-getter and TypeReach guards and their
// probes all read it, and each loaded it for itself: six loads of the whole lint tree from source, about
// 21s of CPU in a package run, where three distinct inputs ever occur (the working tree, and one plant for
// each guard's probe) (#nxgt2ca). So each distinct overlay is loaded once per test binary and shared. The
// guards only read what they are handed.
//
// The load is a superset of what the creating-getter scan used to ask for (it read its dependencies from
// export data), and the packages it returns are the same roots, so its scan sees the same calls.
func lintPackages(t *testing.T, overlay map[string][]byte) []*packages.Package {
	t.Helper()

	key := overlayKey(overlay)
	lintLoads.Lock()
	load := lintLoads.byOverlay[key]
	if load == nil {
		load = &lintLoad{}
		lintLoads.byOverlay[key] = load
	}
	lintLoads.Unlock()

	load.once.Do(func() { load.packages, load.failure = loadLintPackages(overlay) })
	if load.failure != "" {
		t.Fatal(load.failure)
	}
	return load.packages
}

var lintLoads = struct {
	sync.Mutex
	byOverlay map[string]*lintLoad
}{byOverlay: map[string]*lintLoad{}}

type lintLoad struct {
	once     sync.Once
	packages []*packages.Package

	// failure, when set, is what every caller fails with: one test cannot fail another, so the load
	// records it and each caller reports it.
	failure string
}

// overlayKey names an overlay by its paths and their contents.
func overlayKey(overlay map[string][]byte) string {
	paths := make([]string, 0, len(overlay))
	for path := range overlay {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, path := range paths {
		fmt.Fprintf(hash, "%d:%s:%d:", len(path), path, len(overlay[path]))
		hash.Write(overlay[path])
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func loadLintPackages(overlay map[string][]byte) ([]*packages.Package, string) {
	root, err := filepath.Abs("../..")
	if err != nil {
		return nil, err.Error()
	}
	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		Dir:     root,
		Overlay: overlay,
	}, "./internal/lint/...")
	if err != nil {
		return nil, fmt.Sprintf("loading the rule packages: %v", err)
	}
	for _, pkg := range loaded {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Sprintf("loading %s: %v", pkg.PkgPath, pkg.Errors[0])
		}
	}
	return loaded, ""
}
