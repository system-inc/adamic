package oracle

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Build a private snapshot so runtime mutants never change the checkout or
// another test's embedded runtime. The ordinary source remains unchanged.
func arrayRuntimeMutant(t *testing.T, source, file, before, after string) run {
	t.Helper()
	directory := t.TempDir()
	runtimeDirectory := filepath.Join(directory, "runtime")
	if err := os.Mkdir(runtimeDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(original)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, entry := range entries {
		if entry.IsDir() || !(strings.HasSuffix(entry.Name(), ".c") || strings.HasSuffix(entry.Name(), ".h")) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(original, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			if !strings.Contains(string(content), before) {
				t.Fatal("mutant check not found")
			}
			content = []byte(strings.Replace(string(content), before, after, 1))
			changed = true
		}
		if err := os.WriteFile(filepath.Join(runtimeDirectory, entry.Name()), content, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("mutant runtime file not found")
	}
	library, err := native.RuntimeLibraryForSource(runtimeDirectory, source, native.Options{})
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "main.c")
	binary := filepath.Join(directory, "mutant")
	if err := os.WriteFile(main, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	flags := append(native.Flags(native.Options{}), "-I", filepath.Dir(library), main, "-o", binary)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if output, err := exec.Command("clang", flags...).CombinedOutput(); err != nil {
		t.Fatalf("mutant compile: %v\n%s", err, output)
	}
	return execute(t, binary)
}

func arrayJavaScriptMutant(t *testing.T, source, before, after string) run {
	t.Helper()
	if !strings.Contains(source, before) {
		t.Fatal("JavaScript mutant check not found")
	}
	source = strings.Replace(source, before, after, 1)
	path := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

func TestArrayElementKindRuntimeMutant(t *testing.T) {
	program, _ := interfaceFixture(t, "lane2/array-boolean")
	control := releasedUncached(t, program)
	if control.exitCode != 70 || !strings.Contains(string(control.stderr), "expected number, found boolean") {
		t.Fatalf("control: %#v", control)
	}
	nativeMutant := arrayRuntimeMutant(t, native.C(program), "view_arrays.c", "if (actual != wanted &&", "if (false && actual != wanted &&")
	jsMutant := arrayJavaScriptMutant(t, javascript.JavaScript(program), "if (!valid) panic(\"element read failed:", "if (false && !valid) panic(\"element read failed:")
	for _, got := range []run{nativeMutant, jsMutant} {
		if got.exitCode != 0 || disagreement(control, got) == "" {
			t.Fatalf("element kind omission escaped: %#v", got)
		}
	}
	t.Log("element-kind mismatch check removed; array-boolean rejects the mutant in C and JavaScript")
}
