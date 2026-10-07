package native

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateOptions refuses unsupported combinations before invoking clang.
func ValidateOptions(options Options) error {
	if options.Target == "" {
		return nil
	}
	if options.Target != "wasm32-wasi" {
		return fmt.Errorf("native: unsupported target %q", options.Target)
	}
	if options.Sanitize {
		return fmt.Errorf("native: sanitizers are not supported for wasm32-wasi")
	}
	if options.cpu != "" {
		return fmt.Errorf("native: native CPU selection is not supported for wasm32-wasi")
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		return fmt.Errorf("native: wasm32-wasi requires WASI_SYSROOT")
	}
	if info, err := os.Stat(sysroot); err != nil || !info.IsDir() {
		return fmt.Errorf("native: WASI_SYSROOT is not a directory: %s", sysroot)
	}
	return nil
}

// Prefer the SDK compiler beside its sysroot without changing the native compiler.
func compilerName(options Options) string {
	if options.Target == "wasm32-wasi" {
		candidate := filepath.Join(filepath.Dir(filepath.Dir(os.Getenv("WASI_SYSROOT"))), "bin", "clang")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
			return candidate
		}
	}
	return "clang"
}

func archiverName(compiler string) string {
	candidate := filepath.Join(filepath.Dir(compiler), "llvm-ar")
	if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
		return candidate
	}
	return "ar"
}
