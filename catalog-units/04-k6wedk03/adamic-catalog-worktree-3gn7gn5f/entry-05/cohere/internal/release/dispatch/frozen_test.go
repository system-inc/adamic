package dispatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Freezing runs a binary whose inputs were never compared against the rules on disk. The only thing
// separating that from the stale-binary defect is that the operator is told, so every test here
// asserts some form of "it refuses to be silent."

func TestResolveFrozenRefusesWhenTheCacheIsEmpty(t *testing.T) {
	t.Parallel()

	// "Frozen with nothing to freeze" is a missing binary, and a missing binary is loud here like
	// everywhere else. Returning a path that does not exist, or an empty one, would hand the caller
	// something to exec that fails later and further from the cause.
	paths := Paths{ModuleDirectory: t.TempDir(), CacheDirectory: t.TempDir()}
	if err := os.MkdirAll(paths.BinaryDirectory(), 0o755); err != nil {
		t.Fatalf("creating the binary directory: %v", err)
	}

	frozen, err := ResolveFrozen(paths)
	if err == nil {
		t.Fatalf("freezing an empty cache succeeded and selected %q", frozen.Path)
	}
	if !strings.Contains(err.Error(), "nothing was run") {
		t.Errorf("the error does not say that nothing ran, so a reader could take it for a warning: %v", err)
	}
}

func TestResolveFrozenRefusesWhenTheCacheIsMissing(t *testing.T) {
	t.Parallel()

	// A cache directory that was never created is a different failure from an empty one — it means
	// no build has ever run here — and it must be just as loud rather than treated as "no results".
	paths := Paths{ModuleDirectory: t.TempDir(), CacheDirectory: filepath.Join(t.TempDir(), "absent")}

	if _, err := ResolveFrozen(paths); err == nil {
		t.Fatal("freezing against a cache directory that does not exist succeeded")
	}
}

func TestResolveFrozenNamesWhatItSelected(t *testing.T) {
	t.Parallel()

	// The caller prints the hash and the build time, and it can only do that if this reports them.
	// A frozen run that could not say which binary it ran would be indistinguishable from an
	// ordinary one, which is the whole thing being guarded against.
	paths := newCacheWithBinaries(t, map[string]string{platformBinaryPrefix() + "abc123def4560000": "binary"})

	frozen, err := ResolveFrozen(paths)
	if err != nil {
		t.Fatalf("freezing a cache with one binary failed: %v", err)
	}

	if frozen.Hash != "abc123def4560000" {
		t.Errorf("hash is %q, so the announcement would not name the binary that ran", frozen.Hash)
	}
	if frozen.ModifiedAt == "" {
		t.Error("no build time, so a reader cannot judge how old the frozen binary is")
	}
	if !strings.HasSuffix(frozen.Path, platformBinaryPrefix()+"abc123def4560000") {
		t.Errorf("selected %q, which is not the binary in the cache", frozen.Path)
	}
}

func TestResolveFrozenPicksTheNewestBinary(t *testing.T) {
	t.Parallel()

	// The most recent successful build is the closest thing to the rules on disk the cache can
	// offer. Picking any other one would be arbitrarily staler for no reason a reader could predict.
	paths := newCacheWithBinaries(t, map[string]string{
		platformBinaryPrefix() + "1111111111111111": "old",
		platformBinaryPrefix() + "2222222222222222": "newer",
		platformBinaryPrefix() + "3333333333333333": "newest",
	})

	setModificationTimes(t, paths, []string{
		platformBinaryPrefix() + "1111111111111111",
		platformBinaryPrefix() + "2222222222222222",
		platformBinaryPrefix() + "3333333333333333",
	})

	frozen, err := ResolveFrozen(paths)
	if err != nil {
		t.Fatalf("freezing failed: %v", err)
	}
	if frozen.Hash != "3333333333333333" {
		t.Errorf("selected %q rather than the newest binary", frozen.Hash)
	}
}

func TestResolveFrozenIgnoresNonExecutableEntries(t *testing.T) {
	t.Parallel()

	// The sidecar recording the development build's hash sits in the same directory and is not
	// something to exec. Selecting it would produce an exec failure whose message is about file
	// formats rather than about the cache.
	paths := newCacheWithBinaries(t, map[string]string{platformBinaryPrefix() + "abc123def4560000": "binary"})

	hashSidecar := filepath.Join(paths.BinaryDirectory(), "cohere-dev.hash")
	if err := os.WriteFile(hashSidecar, []byte("abc123def4560000\n"), 0o644); err != nil {
		t.Fatalf("writing the sidecar: %v", err)
	}

	frozen, err := ResolveFrozen(paths)
	if err != nil {
		t.Fatalf("freezing failed: %v", err)
	}
	if strings.HasSuffix(frozen.Path, ".hash") {
		t.Errorf("selected the hash sidecar %q, which is metadata rather than a binary", frozen.Path)
	}
}

// newCacheWithBinaries builds a cache directory holding the named executables.
func newCacheWithBinaries(t *testing.T, binaries map[string]string) Paths {
	t.Helper()

	paths := Paths{ModuleDirectory: t.TempDir(), CacheDirectory: t.TempDir()}
	if err := os.MkdirAll(paths.BinaryDirectory(), 0o755); err != nil {
		t.Fatalf("creating the binary directory: %v", err)
	}

	for name, contents := range binaries {
		path := filepath.Join(paths.BinaryDirectory(), name)
		if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	return paths
}

// setModificationTimes stamps the named binaries oldest-first, one second apart.
//
// Written rather than relying on creation order because a map iterates in a random order and the
// files would otherwise share a timestamp, which would make the newest-wins assertion pass or fail
// by chance.
func setModificationTimes(t *testing.T, paths Paths, oldestFirst []string) {
	t.Helper()

	base := time.Date(2026, time.August, 23, 1, 0, 0, 0, time.UTC)
	for index, name := range oldestFirst {
		stamp := base.Add(time.Duration(index) * time.Second)
		if err := os.Chtimes(filepath.Join(paths.BinaryDirectory(), name), stamp, stamp); err != nil {
			t.Fatalf("stamping %s: %v", name, err)
		}
	}
}
