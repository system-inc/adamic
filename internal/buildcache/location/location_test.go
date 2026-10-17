package location

import (
	"path/filepath"
	"slices"
	"testing"
)

// Unset or empty, ADAMIC_BUILD_OUTPUT_DIR is today: the output directory is the build cache, a lookup reads one root, and
// every path is read and written where it is. Mutants: Output ignoring the variable's absence (an empty output
// directory); Roots listing the cache twice.
// Not parallel: t.Setenv.
func TestAnUnsetOutputDirectoryIsTheBuildCache(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	t.Setenv("ADAMIC_SANDBOX", "")
	for _, value := range []string{"", cache} {
		t.Setenv(OutputDirectoryVariable, value)
		output, err := Output()
		if err != nil || output != cache {
			t.Fatalf("with %s=%q the output directory is %q (%v), want the build cache %s", OutputDirectoryVariable, value, output, err, cache)
		}
		roots, err := Roots()
		if err != nil || !slices.Equal(roots, []string{cache}) {
			t.Fatalf("with %s=%q the roots are %q (%v), want the build cache alone", OutputDirectoryVariable, value, roots, err)
		}
		product := filepath.Join(cache, "local", "key")
		if places := ReadPlaces(product); !slices.Equal(places, []string{product}) || WritePlace(product) != product {
			t.Fatalf("with %s=%q a product in the cache is read at %q and written at %s", OutputDirectoryVariable, value, places, WritePlace(product))
		}
	}
}

// Set, the output directory is made absolute and read before the build cache, at the same relative place: a path under
// either root is read in both and written in the output directory alone, and any other path stays where it is. In a
// sandbox it is allowed while the cache keeps its rule. Mutants: ReadPlaces putting the cache first; WritePlace leaving a
// cache path in the cache; Output refused in a sandbox.
// Not parallel: t.Setenv and t.Chdir.
func TestASetOutputDirectoryIsReadFirstAndWrittenAlone(t *testing.T) {
	cache, scratch := t.TempDir(), t.TempDir()
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	t.Setenv("ADAMIC_SANDBOX", "")
	t.Chdir(scratch)
	t.Setenv(OutputDirectoryVariable, "output")
	output, err := Output()
	if want := filepath.Join(scratch, "output"); err != nil || output != want {
		t.Fatalf("the output directory is %q (%v), want %s, made absolute", output, err, want)
	}
	if roots, err := Roots(); err != nil || !slices.Equal(roots, []string{output, cache}) {
		t.Fatalf("the roots are %q (%v), want the output directory, then the cache", roots, err)
	}
	for _, path := range []string{filepath.Join(cache, "local", "key"), filepath.Join(output, "local", "key")} {
		want := []string{filepath.Join(output, "local", "key"), filepath.Join(cache, "local", "key")}
		if places := ReadPlaces(path); !slices.Equal(places, want) {
			t.Errorf("%s is read at %q, want %q", path, places, want)
		}
		if written := WritePlace(path); written != want[0] {
			t.Errorf("%s is written at %s, want %s", path, written, want[0])
		}
	}
	if places := ReadPlaces(cache); !slices.Equal(places, []string{output, cache}) {
		t.Errorf("the cache itself is read at %q", places)
	}
	elsewhere := filepath.Join(scratch, "elsewhere", "file")
	if places := ReadPlaces(elsewhere); !slices.Equal(places, []string{elsewhere}) || WritePlace(elsewhere) != elsewhere {
		t.Errorf("a path outside both roots is read at %q and written at %s", places, WritePlace(elsewhere))
	}
	t.Setenv("ADAMIC_SANDBOX", "1")
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", "")
	if got, err := Output(); err != nil || got != output {
		t.Errorf("in a sandbox the output directory is %q (%v), want %s", got, err, output)
	}
	if _, err := BuildCache(); err == nil {
		t.Error("in a sandbox an unset ADAMIC_BUILD_CACHE_DIR names a build cache")
	}
}
