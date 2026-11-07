package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"

	"github.com/system-inc/adamic/internal/toolchain"
)

// buildEnvironment is the list of settings buildPlatform reads, this process's environment; a test hands it a list the
// process can't hold, two settings of one name, which Go's own environment keeps one of.
var buildEnvironment = os.Environ

// buildPlatform is the goos and goarch a build is for: the environment's `GOOS` and `GOARCH`, read by toolchain.Platform,
// the one reader cgo's compilers read too, so a build cross-compiled for arm64 on an amd64 box names, keys and links
// arm64 (#5xsd2ha). Either one unset or empty is this process's own, runtime.GOOS or runtime.GOARCH, which is what go
// itself builds for then. That fallback is a person's road, running adamic or go test by hand, never Loom's: every
// unit environment Loom builds carries both.
func buildPlatform() (goos string, goarch string) {
	return toolchain.Platform(buildEnvironment())
}

// darwinBuild says whether the build is for macOS (buildPlatform): its link is ld64's, which loads an archive whole with
// -force_load and records each object it links in a debug map, where an ELF link takes --whole-archive and lld.
func darwinBuild() bool {
	goos, _ := buildPlatform()
	return goos == "darwin"
}

// ResolveTarget is options with an empty Target made the build's own triple (toolchain.Triple of buildPlatform, the one
// table cgo's compilers read too), so every flag and key reads a named target, never "" for native. Any other Target is
// kept as it is, for ValidateOptions to accept or refuse. It fails only when the build's platform has no row. Flags,
// ValidateOptions and every key resolve through it.
func ResolveTarget(options Options) (Options, error) {
	if options.Target != "" {
		return options, nil
	}
	triple, err := toolchain.Triple(buildPlatform())
	if err != nil {
		return options, fmt.Errorf("native: %w", err)
	}
	options.Target = triple
	return options, nil
}

// nativeTargetRefusal says why a native triple can't be built here, or nil when it can: one the table doesn't hold, or a
// triple for another platform than this host's. A darwin cross build would need an SDK for that arch, and none is
// pinned. A linux cross build (`GOARCH=arm64` on an amd64 box) would need that arch's C library and headers, and with
// none pinned clang would compile against this host's, so it's refused by name until #5xsd2ha's second half pins its
// sysroot.
func nativeTargetRefusal(triple string) error {
	goos, goarch, found := toolchain.TriplePlatform(triple)
	if !found {
		return fmt.Errorf("native: unsupported target %q", triple)
	}
	if goos == "darwin" && (goos != goruntime.GOOS || goarch != goruntime.GOARCH) {
		return fmt.Errorf("native: %s: a darwin cross build has no pinned SDK, and this host is %s/%s", triple, goruntime.GOOS, goruntime.GOARCH)
	}
	if goos == "linux" && (goos != goruntime.GOOS || goarch != goruntime.GOARCH) {
		return fmt.Errorf("native: %s: a linux cross build needs its pinned sysroot (#5xsd2ha), and this host is %s/%s", triple, goruntime.GOOS, goruntime.GOARCH)
	}
	return nil
}

// ValidateOptions refuses unsupported combinations before invoking clang.
func ValidateOptions(options Options) error {
	options, err := ResolveTarget(options)
	if err != nil {
		return err
	}
	if (options.ProgramRegion || os.Getenv("ADAMIC_PROGRAM_REGION") == "1") && (options.Request || options.Target == "wasm32-wasi") {
		return fmt.Errorf("native: Program regions require a one-shot native CLI")
	}
	if options.Request && options.Target != "wasm32-wasi" {
		return fmt.Errorf("native: request exports require wasm32-wasi")
	}
	if options.Target != "wasm32-wasi" {
		return nativeTargetRefusal(options.Target)
	}
	if options.Sanitize {
		return fmt.Errorf("native: sanitizers are not supported for wasm32-wasi")
	}
	// The WASI emitter writes no #line directives and WASILinkFlags strips debug sections, so a debug
	// module would be the release one, built slower, and say nothing.
	if options.Debug {
		return fmt.Errorf("native: debug builds are not supported for wasm32-wasi")
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
	_, err = WASIClang()
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
	// A clang that couldn't be run says why, rather than reading as one that printed nothing (#b548dw1).
	if err != nil {
		return fmt.Errorf("native: wasm32-wasi is pinned to wasi-sdk %s (%s); %s --version failed: %w", WASISDKVersion, wasiClangVersion, path, err)
	}
	first, _, _ := strings.Cut(string(output), "\n")
	if !strings.HasPrefix(first, wasiClangVersion) {
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
