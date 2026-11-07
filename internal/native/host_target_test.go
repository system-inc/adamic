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
	if err := os.MkdirAll(sysroot, 0755); err != nil {
		test.Fatal(err)
	}
	test.Setenv("WASI_SYSROOT", sysroot)
	compiler := filepath.Join(sdk, "bin", "clang")
	wasiClangChecks.Store(compiler, wasiClangCheck{})
	test.Cleanup(func() { wasiClangChecks.Delete(compiler) })
	for name, want := range wasiHostRefusals {
		test.Run(name, func(test *testing.T) {
			source := "void example(void) { " + name + "(); }"
			var refused *TargetRefused
			if err := validateBuild(
				source,
				Options{Target: "wasm32-wasi"},
			); !errors.As(err, &refused) || *refused != want {
				test.Fatalf("want %v, got %v", want, err)
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

			if err := validateBuild(source, Options{}); err != nil {
				test.Fatal(err)
			}
			for _, data := range []string{
				"/* " + name + " */",
				"// " + name + "\n",
				`const char *text = "` + name + `\?";`,
			} {
				if err := validateBuild(data, Options{Target: "wasm32-wasi"}); err != nil {
					test.Fatalf("data was refused: %v", err)
				}
			}
			if !strings.Contains(refused.Error(), want.Reason) {
				test.Fatal(refused)
			}
		})
	}
}
