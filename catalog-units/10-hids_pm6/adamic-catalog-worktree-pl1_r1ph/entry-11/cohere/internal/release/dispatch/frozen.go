package dispatch

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FrozenBinary is a cached binary selected without checking whether it matches the rules on disk.
//
// The hash is carried alongside the path because the caller must be able to say which binary it
// ran. A frozen run that could not name what it ran would be indistinguishable from a normal one,
// which is the thing this type exists to prevent.
type FrozenBinary struct {
	// Path is the binary that will be executed.
	Path string

	// Hash is the input hash the binary was built from, read from its filename.
	Hash string

	// ModifiedAt is when it was built, so a reader can judge how old "frozen" is.
	ModifiedAt string
}

// ResolveFrozen returns the most recently built cached binary, without recomputing the input hash.
//
// This is the escape hatch for a tree that cannot compile: during a dependency migration the source
// does not build, so the ordinary path — hash the inputs, rebuild on a miss — fails even though a
// working binary is sitting in the cache. Freezing lets the instrument keep measuring while the
// floor is being replaced under it.
//
// It is deliberately narrow, and every constraint here follows from one invariant: **a run that did
// not cohere the rules match must never look like one that did.**
//
//   - It is reached only by an explicit request, never automatically.
//   - It is never a fallback from a failed rebuild. That is the failure that would actually happen,
//     because it arrives disguised as helpfulness: a rebuild fails, something quietly runs the last
//     good binary, and a red tree reports green against rules nobody checked.
//   - It returns an error rather than a guess when the cache is empty, because "frozen with nothing
//     to freeze" is a missing binary, and a missing binary is loud here like everywhere else.
//
// The staleness is not the lie. The silence would be. A caller that announces the hash and says it
// may not match the rules on disk has stated a limitation, which is the same discipline as naming
// the files a run did not check.
func ResolveFrozen(paths Paths) (FrozenBinary, error) {
	entries, err := os.ReadDir(paths.BinaryDirectory())
	if err != nil {
		return FrozenBinary{}, fmt.Errorf(
			"--frozen was requested but the binary cache at %s cannot be read, so there is nothing to freeze: %w",
			paths.BinaryDirectory(), err,
		)
	}

	type candidate struct {
		name       string
		hash       string
		modifiedAt int64
		modified   string
	}
	candidates := []candidate{}

	for _, entry := range entries {
		if entry.IsDir() || !isCohereBinaryName(entry.Name()) {
			continue
		}

		information, err := entry.Info()
		if err != nil {
			continue
		}
		// A binary the operating system will refuse to execute is not a candidate, and skipping it
		// here produces a clearer error than an exec failure would.
		if information.Mode()&0o111 == 0 {
			continue
		}

		candidates = append(candidates, candidate{
			name:       entry.Name(),
			hash:       hashFromBinaryName(entry.Name()),
			modifiedAt: information.ModTime().Unix(),
			modified:   information.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	if len(candidates) == 0 {
		return FrozenBinary{}, fmt.Errorf(
			"--frozen was requested but no cached binary exists in %s, so nothing was run",
			paths.BinaryDirectory(),
		)
	}

	// Newest wins: the most recent successful build is the closest thing to the rules on disk that
	// the cache can offer. Ties break by name so the choice is deterministic rather than dependent
	// on directory order.
	sort.Slice(candidates, func(first int, second int) bool {
		if candidates[first].modifiedAt != candidates[second].modifiedAt {
			return candidates[first].modifiedAt > candidates[second].modifiedAt
		}
		return candidates[first].name > candidates[second].name
	})

	newest := candidates[0]
	return FrozenBinary{
		Path:       filepath.Join(paths.BinaryDirectory(), newest.name),
		Hash:       newest.hash,
		ModifiedAt: newest.modified,
	}, nil
}

// isCohereBinaryName reports whether a file in the binary cache is a cohere binary.
//
// Matched against the exact names cohere binaries are given rather than a shared prefix, because the
// cache holds other executables. A bare `cohere-` prefix also matched the Swift engine
// (`cohere-swift-*`) and the launcher itself (`cohere-dispatch`). Measured 2026-10-02: the newest file
// in the cache was `cohere-swift-current`, so `--frozen` would have exec'd the Swift engine as cohere.
// A half-written build carries `.partial-` until it is proven and renamed, and the development
// binary's recorded hash ends in `.hash`; neither is something to run.
func isCohereBinaryName(name string) bool {
	if strings.Contains(name, ".partial-") || strings.HasSuffix(name, ".hash") {
		return false
	}
	return name == developmentBinaryName || strings.HasPrefix(name, platformBinaryPrefix())
}

// hashFromBinaryName reads the input hash back out of a cached binary's filename.
//
// The name is `cohere-<goos>-<goarch>-<hash>`, so the hash is the last segment. A name that does
// not carry one — the stable development binary, for instance — reports that rather than an empty
// string, because the caller prints this and a blank would read as a missing value rather than as
// a binary that never had a hash in its name.
func hashFromBinaryName(name string) string {
	segments := strings.Split(strings.TrimSuffix(name, executableSuffix), "-")
	if len(segments) < 4 {
		return "unnamed (a development build)"
	}
	return segments[len(segments)-1]
}
