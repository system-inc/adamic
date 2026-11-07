package native

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// A native binary's bytes are a function of its source and options, wherever it's built (Oct 9: a stored
// typescript-parser-native carried ~/.cache/adamic/runtime/.build-<random> in its debug info, so it never matched a
// rebuild and the build cache's audit called it poisoned). Two builds of one program, each with its own home and build
// cache (so its own runtime cache, which lives under the build cache), temporary directory and output directory, of
// different lengths, must be byte for byte the same. Sanitized, so the debug info is in them.
// Not parallel: points HOME, XDG_CACHE_HOME, ADAMIC_BUILD_CACHE_DIR and TMPDIR at this test through test.Setenv.
func TestBuildIsTheSameFromAnyDirectory(test *testing.T) {
	checked, loadError := load.Load(
		[]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")},
	)
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	var binaries [][]byte
	for _, name := range []string{"a", "a-much-longer-second-root"} {
		root := filepath.Join(test.TempDir(), name)
		for _, directory := range []string{"home", "cache", "tmp", "out"} {
			if createDirectoryError := os.MkdirAll(filepath.Join(root, directory), 0o755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
		}
		test.Setenv("HOME", filepath.Join(root, "home"))
		test.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
		test.Setenv("ADAMIC_BUILD_CACHE_DIR", filepath.Join(root, "cache", "adamic-build"))
		test.Setenv("ADAMIC_BUILD_OUTPUT_DIR", "") // its writes too: an inherited output directory would take them
		test.Setenv("TMPDIR", filepath.Join(root, "tmp"))
		binary := filepath.Join(root, "out", "program")
		if buildError := Build(source, binary, Options{Sanitize: true}); buildError != nil {
			test.Fatal(buildError)
		}
		content, readError := os.ReadFile(binary)
		if readError != nil {
			test.Fatal(readError)
		}
		binaries = append(binaries, content)
	}
	if !bytes.Equal(binaries[0], binaries[1]) {
		offset := 0
		for offset < len(binaries[0]) && offset < len(binaries[1]) && binaries[0][offset] == binaries[1][offset] {
			offset++
		}
		test.Fatalf(
			"two builds differ: %d and %d bytes, first difference at byte %d",
			len(binaries[0]),
			len(binaries[1]),
			offset,
		)
	}
}
