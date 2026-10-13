package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ValidateOptions refuses unsupported combinations before invoking clang.
func ValidateOptions(options Options) error {
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
	_, err := WASIClang()
	return err
}

// WASISDKVersion is the one wasi-sdk release every wasm32 build and test uses (bash cloud/setup.sh
// --wasi-sdk installs it). Another clang can emit different Wasm for the same sha, and then two boxes
// give two verdicts, so any other clang is refused by name.
const WASISDKVersion = "27"

// wasiClangVersion is how that release's clang begins its --version output.
const wasiClangVersion = "clang version 20.1.8-wasi-sdk"

type wasiClangCheck struct{ err error }

// Each clang path is checked once per process; the oracle builds hundreds of modules.
var wasiClangChecks sync.Map

// WASIClang is the pinned wasi-sdk clang beside WASI_SYSROOT, or an error naming what it found instead.
func WASIClang() (string, error) {
	candidate := filepath.Join(filepath.Dir(filepath.Dir(os.Getenv("WASI_SYSROOT"))), "bin", "clang")
	if checked, ok := wasiClangChecks.Load(candidate); ok {
		return candidate, checked.(wasiClangCheck).err
	}
	err := checkWASIClang(candidate)
	wasiClangChecks.Store(candidate, wasiClangCheck{err})
	return candidate, err
}

func checkWASIClang(path string) error {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("native: wasm32-wasi is pinned to wasi-sdk %s, whose clang belongs at %s beside WASI_SYSROOT, and none is there; run bash cloud/setup.sh --wasi-sdk", WASISDKVersion, path)
	}
	output, err := exec.Command(path, "--version").Output()
	first, _, _ := strings.Cut(string(output), "\n")
	if err != nil || !strings.HasPrefix(first, wasiClangVersion) {
		return fmt.Errorf("native: wasm32-wasi is pinned to wasi-sdk %s (%s); %s is %q", WASISDKVersion, wasiClangVersion, path, strings.TrimSpace(first))
	}
	return nil
}

// The pinned SDK compiler for wasm32, and the native compiler otherwise. ValidateOptions has refused any
// other wasm32 clang before this is reached.
func compilerName(options Options) string {
	if options.Target == "wasm32-wasi" {
		return filepath.Join(filepath.Dir(filepath.Dir(os.Getenv("WASI_SYSROOT"))), "bin", "clang")
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
	flags := []string{"-Wl,-z,stack-size=131072", "-Wl,--strip-debug"}
	if !options.Request {
		return append(flags, "-mexec-model=command")
	}
	flags = append(flags, "-mexec-model=reactor", "-Wl,--export-memory")
	for _, name := range []string{"adamic_request", "adamic_response_bytes", "adamic_response_length", "malloc", "free", "adamic_release"} {
		flags = append(flags, "-Wl,--export-if-defined="+name)
	}
	if options.Count {
		flags = append(flags, "-Wl,--export=adamic_live", "-Wl,--export=adamic_regions")
	}
	return flags
}
