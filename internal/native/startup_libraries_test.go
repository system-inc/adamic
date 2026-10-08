package native

import (
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"os"
	"path/filepath"
	"testing"
)

func TestStartupDefaultLibraries(t *testing.T) {
	directory := t.TempDir()
	if err := installStartupLibraries(directory); err != nil {
		t.Fatal(err)
	}
	fs := bundled.WrapFS(osvfs.FS())
	for _, name := range bundled.LibNames {
		want, ok := fs.ReadFile(bundled.LibPath().ResolveFile(name))
		if !ok {
			t.Fatal(name)
		}
		got, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("library %s differs from compiler's bundled file", name)
		}
	}
	if err := installStartupLibraries(directory); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "lib.d.ts"), []byte("different"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installStartupLibraries(directory); err == nil {
		t.Fatal("must refuse overwriting a different default library")
	}
}
