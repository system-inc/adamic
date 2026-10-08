package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The checker-archive build includes tsgo_runtime.h, and with it adamic.h, before main.c's own
// #defines, so the features it turns on have to arrive as flags. Without them the
// closure-convention method tables meet adamic.h's older entry type and clang refuses the C.
func TestTSGoBuildSeesTheProgramsFeatures(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_extended.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	directory := t.TempDir()
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		contents, err := runtime.ReadFile("runtime/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte("#include \"tsgo_runtime.h\"\n"+source), 0o644); err != nil {
		t.Fatal(err)
	}
	compile := func(flags []string) error {
		arguments := append(append([]string{"-fsyntax-only", "-I" + directory}, flags...), main)
		output, err := exec.Command("clang", arguments...).CombinedOutput()
		if err != nil {
			return &compileError{output}
		}
		return nil
	}
	if len(featureFlags(source)) == 0 {
		t.Fatal("arguments_length_extended.a no longer turns on a runtime feature; pick a program that does")
	}
	if err := compile(tsgoFlags(source, Options{})); err != nil {
		t.Fatalf("the checker-archive build's main.c doesn't compile with the program's features: %v", err)
	}
	// The test can fail: without the flags, the same C is the gate's red.
	if err := compile(append(Flags(Options{}), "-DADAMIC_TSGO")); err == nil {
		t.Fatal("main.c compiled without its feature flags, so this test no longer covers the include order")
	}
}

type compileError struct{ output []byte }

func (e *compileError) Error() string { return string(e.output) }
