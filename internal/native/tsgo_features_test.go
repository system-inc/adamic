package native

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The checker-archive build includes tsgo_runtime.h, and with it adamic.h, before main.c's own
// #defines, so the features it turns on have to arrive as flags. Without them the
// closure-convention method tables meet adamic.h's older entry type and clang refuses the C.
func TestTSGoBuildSeesTheProgramsFeatures(test *testing.T) {
	test.Parallel()
	loaded, loadError := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_extended.a")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	directory := test.TempDir()
	entries, readDirectoryError := runtime.ReadDir("runtime")
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		contents, readFileError := runtime.ReadFile("runtime/" + entry.Name())
		if readFileError != nil {
			test.Fatal(readFileError)
		}
		if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	main := filepath.Join(directory, "main.c")
	if writeError := os.WriteFile(main, []byte("#include \"tsgo_runtime.h\"\n"+source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	compile := func(flags []string) error {
		arguments := append(append([]string{"-fsyntax-only", "-I" + directory}, flags...), main)
		output, combinedOutputError := toolchain.Command("clang", arguments...).CombinedOutput()
		if combinedOutputError != nil {
			return &compileError{output}
		}
		return nil
	}
	if len(featureFlags(source)) == 0 {
		test.Fatal("arguments_length_extended.a no longer turns on a runtime feature; pick a program that does")
	}
	if compileError := compile(tsgoFlags(source, Options{})); compileError != nil {
		test.Fatalf("the checker-archive build's main.c doesn't compile with the program's features: %v", compileError)
	}
	// The test can fail: without the flags, the same C is the gate's red.
	if compileError := compile(append(Flags(Options{}), "-DADAMIC_TSGO")); compileError == nil {
		test.Fatal("main.c compiled without its feature flags, so this test no longer covers the include order")
	}
}

type compileError struct{ output []byte }

func (compileError *compileError) Error() string { return string(compileError.output) }
