package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWASITargetFlags(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	flags := strings.Join(Flags(Options{Target: "wasm32-wasi"}), " ")
	for _, flag := range []string{"--target=wasm32-wasi", "--sysroot=" + os.Getenv("WASI_SYSROOT"), "-DADAMIC_TARGET_WASI=1", "-Oz", "-mno-atomics", "-ffp-contract=off"} {
		if !strings.Contains(flags, flag) {
			t.Errorf("missing %s in %s", flag, flags)
		}
	}
	if strings.Contains(strings.Join(Flags(Options{}), " "), "WASI") {
		t.Fatal("native flags select WASI")
	}
}

func TestWASIRefusesUnsupportedOptions(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	for _, options := range []Options{{Request: true}, {Target: "unknown"}, {Target: "wasm32-wasi", Sanitize: true}, {Target: "wasm32-wasi", cpu: "native"}} {
		if err := ValidateOptions(options); err == nil {
			t.Fatalf("accepted %+v", options)
		}
	}
	t.Setenv("WASI_SYSROOT", "")
	if err := ValidateOptions(Options{Target: "wasm32-wasi"}); err == nil || !strings.Contains(err.Error(), "requires WASI_SYSROOT") {
		t.Fatalf("missing sysroot: %v", err)
	}
	t.Setenv("WASI_SYSROOT", filepath.Join(t.TempDir(), "absent"))
	if err := ValidateOptions(Options{Target: "wasm32-wasi"}); err == nil {
		t.Fatal("accepted nonexistent sysroot")
	}
}
