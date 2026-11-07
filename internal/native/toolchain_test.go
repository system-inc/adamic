package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The wasm32 toolchain is pinned: a system LLVM clang beside the sysroot must be refused by name, and
// wasi-sdk's own clang accepted. Fake compilers stand in for both, so this runs on every gate.
//
// Not parallel: it writes executables and runs them, and a parallel sibling that forks between a write and its exec
// holds the script open for writing, so the exec fails with `ETXTBSY`, which checkWASIClang reads as an empty version
// (#nf1x9fp, at load 80 on Loom). Serial tests run before the package's parallel ones resume.
func TestWASIClangIsPinned(test *testing.T) {
	fake := func(test *testing.T, version string) string {
		test.Helper()
		path := filepath.Join(test.TempDir(), "clang")
		if writeError := os.WriteFile(
			path,
			[]byte("#!/bin/sh\necho '"+version+"'\necho 'Target: wasm32-unknown-wasi'\n"),
			0o755,
		); writeError != nil {
			test.Fatal(writeError)
		}
		return path
	}
	if checkWasiClangError := checkWASIClang(fake(
		test,
		"clang version 20.1.8-wasi-sdk (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)",
	)); checkWasiClangError != nil {
		test.Fatalf("wasi-sdk %s's clang refused: %v", WASISDKVersion, checkWasiClangError)
	}
	for _, version := range []string{
		"Ubuntu clang version 18.1.3 (1ubuntu1)",
		"clang version 21.1.0",
		"clang version 20.1.8",
		"Apple clang version 21.0.0 (clang-2100.3.34.2)",
	} {
		checkWasiClangError := checkWASIClang(fake(test, version))
		if checkWasiClangError == nil {
			test.Errorf("accepted %q as wasi-sdk %s's clang", version, WASISDKVersion)
			continue
		}
		if !strings.Contains(checkWasiClangError.Error(), version) ||
			!strings.Contains(checkWasiClangError.Error(), "wasi-sdk "+WASISDKVersion) {
			test.Errorf("refusal must name the clang found and the pinned release: %v", checkWasiClangError)
		}
	}
	// A clang that can't be run names its failure, rather than reading as one that printed an empty version.
	broken := filepath.Join(test.TempDir(), "clang")
	if writeError := os.WriteFile(
		broken,
		[]byte("#!/bin/sh\necho 'clang version 20.1.8-wasi-sdk'\nexit 3\n"),
		0o755,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if checkWasiClangError := checkWASIClang(
		broken,
	); checkWasiClangError == nil || !strings.Contains(checkWasiClangError.Error(), "--version failed: exit status 3") {
		test.Errorf("a clang that fails to run must be refused with its failure: %v", checkWasiClangError)
	}
	if checkWasiClangError := checkWASIClang(
		filepath.Join(test.TempDir(), "clang"),
	); checkWasiClangError == nil || !strings.Contains(checkWasiClangError.Error(), "cloud/setup.sh --wasi-sdk") {
		test.Errorf("a missing clang must say how to install the pinned one: %v", checkWasiClangError)
	}
}
