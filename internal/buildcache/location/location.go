// Package location is where the build cache lives on this machine. It is a package of its own so internal/native,
// which every Adamic compiler links, keeps its caches under the same directory buildcache keeps products in without
// linking buildcache itself (testing, net/http and the tracer come with it).
//
// A process may write somewhere else than it reads (OutputDirectoryVariable): Loom's tree builder runs a candidate's
// product tests as a user that can't write the build cache, and copies what they wrote into it afterwards. Then every
// write goes to the output directory, and every lookup reads the output directory first and the build cache after it,
// each holding what it holds at the same relative path.
package location

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// OutputDirectoryVariable names the directory a process writes every product, read set and native cache entry to, when
// that isn't the build cache. Unset or empty, the process writes to the build cache, as it always has.
const OutputDirectoryVariable = "ADAMIC_BUILD_OUTPUT_DIR"

// BuildCache is ADAMIC_BUILD_CACHE_DIR, or the user cache directory's adamic-build. In a sandbox (ADAMIC_SANDBOX=1) it
// is ADAMIC_BUILD_CACHE_DIR or an error: products are never written under a home the sandbox didn't declare.
func BuildCache() (string, error) {
	if directory := os.Getenv("ADAMIC_BUILD_CACHE_DIR"); directory != "" {
		return filepath.Abs(directory)
	}
	if toolchain.Sandboxed() {
		return "", errors.New("ADAMIC_SANDBOX=1 keeps products only where ADAMIC_BUILD_CACHE_DIR says, and ADAMIC_BUILD_CACHE_DIR is unset")
	}
	user, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(user, "adamic-build"), nil
}

// Output is where this process writes: ADAMIC_BUILD_OUTPUT_DIR made absolute, or the build cache (BuildCache) when it is
// unset. A sandbox allows it whenever it is set, since then the sandbox named it; the build cache keeps its own rule.
// Making the directory is the writer's, as making the build cache is.
func Output() (string, error) {
	if directory := os.Getenv(OutputDirectoryVariable); directory != "" {
		return filepath.Abs(directory)
	}
	return BuildCache()
}

// Roots are the directories a lookup reads, in order: the output directory, then the build cache, or one directory when
// they are the same, as they are while ADAMIC_BUILD_OUTPUT_DIR is unset.
func Roots() ([]string, error) {
	output, err := Output()
	if err != nil {
		return nil, err
	}
	cache, err := BuildCache()
	if err != nil {
		return nil, err
	}
	if cache == output {
		return []string{output}, nil
	}
	return []string{output, cache}, nil
}

// ReadPlaces is where a lookup reads path: for a path at or under one of Roots, the same place under each root, in their
// order, so what the output directory holds is read before what the build cache holds; any other path, and every path
// while ADAMIC_BUILD_OUTPUT_DIR is unset, alone. A root is matched as Roots spells it.
func ReadPlaces(path string) []string {
	if os.Getenv(OutputDirectoryVariable) == "" {
		return []string{path}
	}
	roots, err := Roots()
	if err != nil || len(roots) == 1 {
		return []string{path}
	}
	for _, root := range roots {
		relative, ok := within(root, path)
		if !ok {
			continue
		}
		places := make([]string, len(roots))
		for index, each := range roots {
			places[index] = filepath.Join(each, relative)
		}
		return places
	}
	return []string{path}
}

// WritePlace is where path is written: the same place under the output directory, for a path at or under one of Roots,
// so nothing is ever written in a build cache this process only reads; any other path, where it is.
func WritePlace(path string) string {
	return ReadPlaces(path)[0]
}

// within is path relative to root, when it is root or under it.
func within(root, path string) (string, bool) {
	path = filepath.Clean(path)
	if path == root {
		return ".", true
	}
	if rest, ok := strings.CutPrefix(path, strings.TrimSuffix(root, string(filepath.Separator))+string(filepath.Separator)); ok {
		return rest, true
	}
	return "", false
}
