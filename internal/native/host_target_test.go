package native

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Not parallel: t.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASIHostTargetRefusals(t *testing.T) {
	// This is a lexer/public-boundary unit: every Build must refuse before clang.
	// Isolate it from the SDK version probe, tested separately and by the real WASI gate.
	sdk := t.TempDir()
	sysroot := filepath.Join(sdk, "share", "wasi-sysroot")
	if err := os.MkdirAll(sysroot, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WASI_SYSROOT", sysroot)
	compiler := filepath.Join(sdk, "bin", "clang")
	wasiClangChecks.Store(compiler, wasiClangCheck{})
	t.Cleanup(func() { wasiClangChecks.Delete(compiler) })
	for name, want := range wasiHostRefusals {
		t.Run(name, func(t *testing.T) {
			source := "void example(void) { " + name + "(); }"
			var refused *TargetRefused
			if err := validateBuild(source, Options{Target: "wasm32-wasi"}); !errors.As(err, &refused) || *refused != want {
				t.Fatalf("want %v, got %v", want, err)
			}
			output := filepath.Join(t.TempDir(), "refused.wasm")
			for _, err := range []error{Build(source, output, Options{Target: "wasm32-wasi"}), BuildFlags(source, output, Options{Target: "wasm32-wasi"}, nil)} {
				if !errors.As(err, &refused) || *refused != want {
					t.Fatalf("public build boundary lost refusal: %v", err)
				}
			}

			if err := validateBuild(source, Options{}); err != nil {
				t.Fatal(err)
			}
			for _, data := range []string{"/* " + name + " */", "// " + name + "\n", `const char *text = "` + name + `\?";`} {
				if err := validateBuild(data, Options{Target: "wasm32-wasi"}); err != nil {
					t.Fatalf("data was refused: %v", err)
				}
			}
			if !strings.Contains(refused.Error(), want.Reason) {
				t.Fatal(refused)
			}
		})
	}
}

// __filename and __dirname, bound in a CommonJS build's .ts files, are refused on wasm32-wasi by name, as
// process.argv is (Adamic's ruling, Nov 6, #redxes2): a read's C names its runtime function, and that function's row
// refuses it as the binding the program wrote. TestWASIHostTargetRefusals holds every row to the build boundary.
// Mutants: either row taken out of wasiHostRefusals, or nodeProcessCall mapping its operation to another function.
func TestWASIRefusesCommonJSModuleValuesByName(t *testing.T) {
	t.Parallel()
	for _, one := range []struct {
		operation, function string
		want                TargetRefused
	}{
		{"filename", "adamic_node_filename", TargetRefused{"__filename", "WASI has no executable paths"}},
		{"dirname", "adamic_node_dirname", TargetRefused{"__dirname", "WASI has no executable paths"}},
	} {
		t.Run(one.operation, func(t *testing.T) {
			t.Parallel()
			program := &ir.Program{Source: "commonjs.ts", Main: []ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.ProcessCall{Operation: one.operation, Of: ir.String}}}}
			if code := C(program); !strings.Contains(code, one.function+"()") {
				t.Fatalf("a read of %s doesn't call %s", one.want.Member, one.function)
			}
			if refusal, found := wasiHostRefusals[one.function]; !found || refusal != one.want {
				t.Fatalf("want %v, got %v (found %t)", one.want, refusal, found)
			}
		})
	}
}
