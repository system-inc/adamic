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
// rebuild and the build cache's audit called it poisoned). Two builds of one program, each with its own home (so its
// own runtime cache), temporary directory and output directory, of different lengths, must be byte for byte the same.
// Sanitized, so the debug info is in them.
// Not parallel: points HOME, XDG_CACHE_HOME and TMPDIR at this test through t.Setenv.
func TestBuildIsTheSameFromAnyDirectory(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	var binaries [][]byte
	for _, name := range []string{"a", "a-much-longer-second-root"} {
		root := filepath.Join(t.TempDir(), name)
		for _, directory := range []string{"home", "cache", "tmp", "out"} {
			if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("HOME", filepath.Join(root, "home"))
		t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
		t.Setenv("TMPDIR", filepath.Join(root, "tmp"))
		binary := filepath.Join(root, "out", "program")
		if err := Build(source, binary, Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		binaries = append(binaries, content)
	}
	if !bytes.Equal(binaries[0], binaries[1]) {
		offset := 0
		for offset < len(binaries[0]) && offset < len(binaries[1]) && binaries[0][offset] == binaries[1][offset] {
			offset++
		}
		t.Fatalf("two builds differ: %d and %d bytes, first difference at byte %d", len(binaries[0]), len(binaries[1]), offset)
	}
}
