package native

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateOptions refuses unsupported combinations before invoking clang.
func ValidateOptions(options Options) error {
	if options.ProfileGenerate && options.Profile != "" {
		return fmt.Errorf("native: profile generation and use are mutually exclusive")
	}
	if options.Request && options.Target != "wasm32-wasi" {
		return fmt.Errorf("native: request exports require wasm32-wasi")
	}
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

// WASILinkFlags selects commands unless the driver emitted a request handler.
func WASILinkFlags(options Options) []string {
	flags := []string{"-Wl,-z,stack-size=131072"}
	if !options.Request {
		return append(flags, "-mexec-model=command")
	}
	flags = append(flags, "-mexec-model=reactor", "-Wl,--export-memory")
	for _, name := range []string{"adamic_request", "adamic_response_bytes", "adamic_response_length", "malloc", "free", "adamic_release"} {
		flags = append(flags, "-Wl,--export="+name)
	}
	if options.Count {
		flags = append(flags, "-Wl,--export=adamic_live", "-Wl,--export=adamic_regions")
	}
	return flags
}
