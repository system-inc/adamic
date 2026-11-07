package location

import (
	"path/filepath"
	"slices"
	"testing"
)

// Unset or empty, ADAMIC_BUILD_OUTPUT_DIR is today: the output directory is the build cache, a lookup reads one root,
// and every path is read and written where it is. Mutants: Output ignoring the variable's absence (an empty output
// directory); Roots listing the cache twice.
// Not parallel: test.Setenv.
func TestAnUnsetOutputDirectoryIsTheBuildCache(test *testing.T) {
	cache := test.TempDir()
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	test.Setenv("ADAMIC_SANDBOX", "")
	for _, value := range []string{"", cache} {
		test.Setenv(OutputDirectoryVariable, value)
		output, err := Output()
		if err != nil || output != cache {
			test.Fatalf(
				"with %s=%q the output directory is %q (%v), want the build cache %s",
				OutputDirectoryVariable,
				value,
				output,
				err,
				cache,
			)
		}
		roots, err := Roots()
		if err != nil || !slices.Equal(roots, []string{cache}) {
			test.Fatalf(
				"with %s=%q the roots are %q (%v), want the build cache alone",
				OutputDirectoryVariable,
				value,
				roots,
				err,
			)
		}
		product := filepath.Join(cache, "local", "key")
		if places := ReadPlaces(product); !slices.Equal(places, []string{product}) || WritePlace(product) != product {
			test.Fatalf(
				"with %s=%q a product in the cache is read at %q and written at %s",
				OutputDirectoryVariable,
				value,
				places,
				WritePlace(product),
			)
		}
	}
}

// Set, the output directory is made absolute and read before the build cache, at the same relative place: a path under
// either root is read in both and written in the output directory alone, and any other path stays where it is. In a
// sandbox it is allowed while the cache keeps its rule. Mutants: ReadPlaces putting the cache first; WritePlace leaving
// a cache path in the cache; Output refused in a sandbox.
// Not parallel: test.Setenv and test.Chdir.
func TestASetOutputDirectoryIsReadFirstAndWrittenAlone(test *testing.T) {
	cache, scratch := test.TempDir(), test.TempDir()
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", cache)
	test.Setenv("ADAMIC_SANDBOX", "")
	test.Chdir(scratch)
	test.Setenv(OutputDirectoryVariable, "output")
	output, err := Output()
	if want := filepath.Join(scratch, "output"); err != nil || output != want {
		test.Fatalf("the output directory is %q (%v), want %s, made absolute", output, err, want)
	}
	if roots, err := Roots(); err != nil || !slices.Equal(roots, []string{output, cache}) {
		test.Fatalf("the roots are %q (%v), want the output directory, then the cache", roots, err)
	}
	for _, path := range []string{filepath.Join(cache, "local", "key"), filepath.Join(output, "local", "key")} {
		want := []string{filepath.Join(output, "local", "key"), filepath.Join(cache, "local", "key")}
		if places := ReadPlaces(path); !slices.Equal(places, want) {
			test.Errorf("%s is read at %q, want %q", path, places, want)
		}
		if written := WritePlace(path); written != want[0] {
			test.Errorf("%s is written at %s, want %s", path, written, want[0])
		}
	}
	if places := ReadPlaces(cache); !slices.Equal(places, []string{output, cache}) {
		test.Errorf("the cache itself is read at %q", places)
	}
	elsewhere := filepath.Join(scratch, "elsewhere", "file")
	if places := ReadPlaces(
		elsewhere,
	); !slices.Equal(places, []string{elsewhere}) || WritePlace(elsewhere) != elsewhere {
		test.Errorf("a path outside both roots is read at %q and written at %s", places, WritePlace(elsewhere))
	}
	test.Setenv("ADAMIC_SANDBOX", "1")
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", "")
	if got, err := Output(); err != nil || got != output {
		test.Errorf("in a sandbox the output directory is %q (%v), want %s", got, err, output)
	}
	if _, err := BuildCache(); err == nil {
		test.Error("in a sandbox an unset ADAMIC_BUILD_CACHE_DIR names a build cache")
	}
}
