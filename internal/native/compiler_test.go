package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: PATH selects two wrapper compilers with identical version strings.
func TestRuntimeCompilerChoice(t *testing.T) {
	real, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	for index, name := range []string{"clang", "gcc"} {
		script := "#!/bin/sh\nif [ \"$1\" = --version ]; then echo same-version; exit 0; fi\nexec '" + strings.ReplaceAll(real, "'", "'\\''") + "' -DCHOICE=" + string(rune('1'+index)) + " \"$@\"\n"
		if err := os.WriteFile(filepath.Join(directory, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	sources := t.TempDir()
	if err := os.WriteFile(filepath.Join(sources, "choice.c"), []byte("int choice(void) { return CHOICE; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	libraries := []string{}
	for _, name := range []string{"clang", "gcc", "clang"} {
		library, err := RuntimeLibrary(sources, Options{Compiler: name})
		if err != nil {
			t.Fatal(err)
		}
		libraries = append(libraries, library)
		source := filepath.Join(t.TempDir(), "main.c")
		binary := source + ".out"
		if err := os.WriteFile(source, []byte("int choice(void); int main(void) { return choice(); }\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command(real, source, library, "-o", binary).CombinedOutput(); err != nil {
			t.Fatalf("link: %v: %s", err, out)
		}
		err = exec.Command(binary).Run()
		status, ok := err.(*exec.ExitError)
		expected := 1
		if name == "gcc" {
			expected = 2
		}
		if !ok || status.ExitCode() != expected {
			t.Fatalf("%s runtime: got %v, want exit %d", name, err, expected)
		}
	}
	if libraries[0] == libraries[1] || libraries[0] != libraries[2] {
		t.Fatalf("compiler cache identities: %v", libraries)
	}
}

func TestCompilerOptions(t *testing.T) {
	if compiler, err := selectedCompiler(Options{}); err != nil || compiler != "clang" {
		t.Fatalf("default %q: %v", compiler, err)
	}
	if err := Build("int main(void) { return 0; }", filepath.Join(t.TempDir(), "bad"), Options{Compiler: "unknown"}); err == nil {
		t.Fatal("unknown compiler accepted")
	}
	for _, compiler := range []string{"clang", "gcc"} {
		flags := strings.Join(Flags(Options{Compiler: compiler}), " ")
		if strings.Contains(flags, "-fno-strict-aliasing") {
			t.Fatal(flags)
		}
		if strings.Contains(flags, "-Wno-self-assign") != (compiler == "clang") {
			t.Fatal(flags)
		}
	}
}
