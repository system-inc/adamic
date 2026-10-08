package native

import (
	"errors"
	"strings"
	"testing"
)

func TestWASIHostTargetRefusals(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	for name, want := range wasiHostRefusals {
		t.Run(name, func(t *testing.T) {
			source := "void example(void) { " + name + "(); }"
			var refused *TargetRefused
			if err := validateBuild(source, Options{Target: "wasm32-wasi"}); !errors.As(err, &refused) || *refused != want {
				t.Fatalf("want %v, got %v", want, err)
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
