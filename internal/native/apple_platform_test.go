package native

import (
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/platform"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: PATH is process-wide; these fake-tool controls run serially.
func TestBuildAppleMissingXcrun(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := BuildApple("int main(void) { return 0; }", filepath.Join(t.TempDir(), "program"), Options{})
	var missing *platform.Missing
	if !errors.As(err, &missing) || missing.Platform != "darwin" {
		t.Fatalf("BuildApple: expected PlatformMissing, got %v", err)
	}
}

// Not parallel: PATH is process-wide; discovery runs but finds no SDK.
func TestBuildAppleMissingSDK(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "xcrun"), []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	err := BuildApple("", filepath.Join(dir, "program"), Options{})
	var missing *platform.Missing
	if !errors.As(err, &missing) {
		t.Fatalf("SDK discovery: expected PlatformMissing, got %v", err)
	}
}

// Not parallel: PATH and the cache directory are process-wide; these controls run serially.
func TestSwiftCompilerFailureIsNotPlatformMissing(t *testing.T) {
	for _, phase := range []string{"version", "compile"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			script := "#!/bin/sh\necho swiftc-rejected >&2\nexit 1\n"
			if phase == "compile" {
				script = "#!/bin/sh\nif [ \"$2\" = --version ]; then echo fake-swift; exit 0; fi\necho swiftc-rejected >&2\nexit 1\n"
			}
			if err := os.WriteFile(filepath.Join(dir, "xcrun"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)
			t.Setenv("XDG_CACHE_HOME", dir)
			_, err := swiftUIShim()
			var missing *platform.Missing
			if err == nil || errors.As(err, &missing) || !strings.Contains(err.Error(), "swiftc-rejected") {
				t.Fatalf("compiler failure: %v", err)
			}
		})
	}
}

// Not parallel: PATH is process-wide; SDK discovery succeeds, but clang runs and rejects the build.
func TestClangFailureIsNotPlatformMissing(t *testing.T) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	archiver, err := exec.LookPath("ar")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Symlink(archiver, filepath.Join(dir, "ar")); err != nil {
		t.Fatal(err)
	}
	for name, script := range map[string]string{
		"xcrun": "#!/bin/sh\necho /fake/sdk\n",
		"clang": fmt.Sprintf("#!/bin/sh\ncase \" $* \" in *\" -framework \"*) echo clang-rejected >&2; exit 1;; esac\nexec %q \"$@\"\n", compiler),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("ADAMIC_SANDBOX", "")
	err = BuildApple("int main(void) { return 0; }", filepath.Join(dir, "program"), Options{})
	var missing *platform.Missing
	if err == nil || errors.As(err, &missing) || !strings.Contains(err.Error(), "clang-rejected") {
		t.Fatalf("compiler failure: %v", err)
	}
}

// Not parallel: PATH is process-wide; this also covers direct shim discovery.
func TestSwiftUIShimMissingXcrun(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := swiftUIShim()
	var missing *platform.Missing
	if !errors.As(err, &missing) {
		t.Fatalf("shim: expected PlatformMissing, got %v", err)
	}
}
