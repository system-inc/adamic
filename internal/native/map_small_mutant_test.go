package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func mapSmallMutantLibrary(t *testing.T, directory, changed string) string {
	t.Helper()
	library, err := RuntimeLibrary("", Options{})
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "map.c")
	object := filepath.Join(directory, "map.o")
	archive := filepath.Join(directory, "runtime.a")
	if err := os.WriteFile(source, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	args := append(Flags(Options{}), "-I", filepath.Dir(library), "-c", source, "-o", object)
	if out, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, out)
	}
	contents, err := os.ReadFile(library)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archive, contents, 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("ar", "rcs", archive, object).CombinedOutput(); err != nil {
		t.Fatalf("mutant archive: %v\n%s", err, out)
	}
	headers, err := filepath.Glob(filepath.Join(filepath.Dir(library), "*.h"))
	if err != nil {
		t.Fatal(err)
	}
	for _, header := range headers {
		contents, err := os.ReadFile(header)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(header)), contents, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return archive
}
