package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: t.Setenv changes the process environment variable WASI_SYSROOT.
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

// Not parallel: t.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASIRefusesUnsupportedOptions(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	// Each refusal by its own words: past them a WASI build fails anyway on the empty sysroot's missing SDK clang, so
	// an error alone can't tell a dropped refusal from a kept one.
	for _, refusal := range []struct {
		options Options
		want    string
	}{
		{Options{Request: true}, "request exports require wasm32-wasi"},
		{Options{Target: "unknown"}, `unsupported target "unknown"`},
		{Options{Target: "wasm32-wasi", Sanitize: true}, "sanitizers are not supported for wasm32-wasi"},
		{Options{Target: "wasm32-wasi", Debug: true}, "debug builds are not supported for wasm32-wasi"},
		{Options{Target: "wasm32-wasi", cpu: "native"}, "native CPU selection is not supported for wasm32-wasi"},
	} {
		if err := ValidateOptions(refusal.options); err == nil || !strings.Contains(err.Error(), refusal.want) {
			t.Fatalf("%+v: got %v, want %q", refusal.options, err, refusal.want)
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
