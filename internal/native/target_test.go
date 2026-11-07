package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: test.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASITargetFlags(test *testing.T) {
	test.Setenv("WASI_SYSROOT", test.TempDir())
	flags := strings.Join(Flags(Options{Target: "wasm32-wasi"}), " ")
	for _, flag := range []string{
		"--target=wasm32-wasi",
		"--sysroot=" + os.Getenv("WASI_SYSROOT"),
		"-DADAMIC_TARGET_WASI=1",
		"-Oz",
		"-mno-atomics",
		"-ffp-contract=off",
	} {
		if !strings.Contains(flags, flag) {
			test.Errorf("missing %s in %s", flag, flags)
		}
	}
	if strings.Contains(strings.Join(Flags(Options{}), " "), "WASI") {
		test.Fatal("native flags select WASI")
	}
}

// Not parallel: test.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASIRefusesUnsupportedOptions(test *testing.T) {
	test.Setenv("WASI_SYSROOT", test.TempDir())
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
			test.Fatalf("%+v: got %v, want %q", refusal.options, err, refusal.want)
		}
	}
	test.Setenv("WASI_SYSROOT", "")
	if err := ValidateOptions(
		Options{Target: "wasm32-wasi"},
	); err == nil || !strings.Contains(err.Error(), "requires WASI_SYSROOT") {
		test.Fatalf("missing sysroot: %v", err)
	}
	test.Setenv("WASI_SYSROOT", filepath.Join(test.TempDir(), "absent"))
	if err := ValidateOptions(Options{Target: "wasm32-wasi"}); err == nil {
		test.Fatal("accepted nonexistent sysroot")
	}
}
