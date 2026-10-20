package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The wasm32 toolchain is pinned: a system LLVM clang beside the sysroot must be refused by name, and
// wasi-sdk's own clang accepted. Fake compilers stand in for both, so this runs on every gate.
func TestWASIClangIsPinned(t *testing.T) {
	t.Parallel()
	fake := func(t *testing.T, version string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "clang")
		if err := os.WriteFile(path, []byte("#!/bin/sh\necho '"+version+"'\necho 'Target: wasm32-unknown-wasi'\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		return path
	}
	if err := checkWASIClang(fake(t, "clang version 20.1.8-wasi-sdk (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)")); err != nil {
		t.Fatalf("wasi-sdk %s's clang refused: %v", WASISDKVersion, err)
	}
	for _, version := range []string{
		"Ubuntu clang version 18.1.3 (1ubuntu1)",
		"clang version 21.1.0",
		"clang version 20.1.8",
		"Apple clang version 21.0.0 (clang-2100.3.34.2)",
	} {
		err := checkWASIClang(fake(t, version))
		if err == nil {
			t.Errorf("accepted %q as wasi-sdk %s's clang", version, WASISDKVersion)
			continue
		}
		if !strings.Contains(err.Error(), version) || !strings.Contains(err.Error(), "wasi-sdk "+WASISDKVersion) {
			t.Errorf("refusal must name the clang found and the pinned release: %v", err)
		}
	}
	if err := checkWASIClang(filepath.Join(t.TempDir(), "clang")); err == nil || !strings.Contains(err.Error(), "cloud/setup.sh --wasi-sdk") {
		t.Errorf("a missing clang must say how to install the pinned one: %v", err)
	}
}
