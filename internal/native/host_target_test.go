package native

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: test.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASIHostTargetRefusals(test *testing.T) {
	// This is a lexer/public-boundary unit: every Build must refuse before clang.
	// Isolate it from the SDK version probe, tested separately and by the real WASI gate.
	sdk := test.TempDir()
	sysroot := filepath.Join(sdk, "share", "wasi-sysroot")
	if createDirectoryError := os.MkdirAll(sysroot, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	test.Setenv("WASI_SYSROOT", sysroot)
	compiler := filepath.Join(sdk, "bin", "clang")
	wasiClangChecks.Store(compiler, wasiClangCheck{})
	test.Cleanup(func() { wasiClangChecks.Delete(compiler) })
	for name, want := range wasiHostRefusals {
		test.Run(name, func(test *testing.T) {
			source := "void example(void) { " + name + "(); }"
			var refused *TargetRefused
			if validateBuildError := validateBuild(
				source,
				Options{Target: "wasm32-wasi"},
			); !errors.As(validateBuildError, &refused) || *refused != want {
				test.Fatalf("want %v, got %v", want, validateBuildError)
			}
			output := filepath.Join(test.TempDir(), "refused.wasm")
			for _, err := range []error{
				Build(source, output, Options{Target: "wasm32-wasi"}),
				BuildFlags(source, output, Options{Target: "wasm32-wasi"}, nil),
			} {
				if !errors.As(err, &refused) || *refused != want {
					test.Fatalf("public build boundary lost refusal: %v", err)
				}
			}

			if validateBuildError := validateBuild(source, Options{}); validateBuildError != nil {
				test.Fatal(validateBuildError)
			}
			for _, data := range []string{
				"/* " + name + " */",
				"// " + name + "\n",
				`const char *text = "` + name + `\?";`,
			} {
				if validateBuildError := validateBuild(
					data,
					Options{Target: "wasm32-wasi"},
				); validateBuildError != nil {
					test.Fatalf("data was refused: %v", validateBuildError)
				}
			}
			if !strings.Contains(refused.Error(), want.Reason) {
				test.Fatal(refused)
			}
		})
	}
}
