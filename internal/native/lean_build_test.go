package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// BuildLean refuses a target before it reads anything, so the WASI_SYSROOT runtimeproducts' leanEnvironment lets its
// reach name, read by Flags for wasm32-wasi alone, is never read on its path: the census follows every call and can't
// tell Flags is called only after the refusal, so this holds it. Handed wasm32-wasi while WASI_SYSROOT names a directory
// that isn't there, with a compiler and a source that aren't there either, it fails with its own refusal, not
// ValidateOptions' (WASI_SYSROOT is not a directory) or a read's, and makes nothing beside its output, not even its build
// directory. Mutants: BuildLean validating its options first; the refusal moved after the source's read.
// Not parallel: t.Setenv changes the process environment variable WASI_SYSROOT.
func TestBuildLeanRefusesATargetFirst(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("WASI_SYSROOT", filepath.Join(directory, "absent-sysroot"))
	output := filepath.Join(directory, "output")
	if err := os.Mkdir(output, 0o755); err != nil {
		t.Fatal(err)
	}
	err := BuildLean(filepath.Join(directory, "absent-clang"), filepath.Join(directory, "absent.c"), filepath.Join(directory, "absent-archive"),
		filepath.Join(output, "native"), Options{Target: "wasm32-wasi"}, nil)
	if err == nil || !strings.HasPrefix(err.Error(), "native: a lean build is one native translation unit, with no target") {
		t.Fatalf("a wasm32-wasi lean build: %v, want BuildLean's refusal of a target", err)
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 0 {
		t.Fatalf("a refused lean build made %v beside its output", entries)
	}
}

// BuildLean runs only the compiler it's handed by an absolute path: a bare name would be found on `PATH` or under the
// tools directories, which no key names. Refused before it reads anything or makes its build directory.
func TestBuildLeanRefusesACompilerByName(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	output := filepath.Join(directory, "output")
	if err := os.Mkdir(output, 0o755); err != nil {
		t.Fatal(err)
	}
	err := BuildLean("clang", filepath.Join(directory, "absent.c"), filepath.Join(directory, "absent-archive"), filepath.Join(output, "native"), Options{}, nil)
	if err == nil || !strings.Contains(err.Error(), "runs the compiler it's handed by its absolute path") {
		t.Fatalf("a lean build handed clang by name: %v, want its refusal", err)
	}
	if entries, err := os.ReadDir(output); err != nil || len(entries) > 0 {
		t.Fatalf("a refused lean build made %v beside its output: %v", entries, err)
	}
}
