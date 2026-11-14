package generate

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/platform"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: PATH is process-wide; no installed Apple tools are used by these controls.
func TestSDKDiscoveryPlatformMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	for _, installed := range []bool{false, true} {
		if installed {
			if err := os.WriteFile(filepath.Join(dir, "xcrun"), []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
				t.Fatal(err)
			}
		}
		_, err := SDKBuild("macosx")
		var missing *platform.Missing
		if !errors.As(err, &missing) {
			t.Fatalf("SDKBuild installed=%v: expected PlatformMissing, got %v", installed, err)
		}
		_, err = FromSDK(context.Background(), "macosx", "macos", nil)
		if !errors.As(err, &missing) {
			t.Fatalf("FromSDK installed=%v: expected PlatformMissing, got %v", installed, err)
		}
	}
}

// Not parallel: PATH is process-wide; xcrun starts clang, whose rejection is not a missing SDK.
func TestSDKCheckCompilerFailureIsOrdinary(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "xcrun"), []byte("#!/bin/sh\necho check-rejected >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	err := CheckSDK(context.Background(), "macosx", dir)
	var missing *platform.Missing
	if err == nil || errors.As(err, &missing) || !strings.Contains(err.Error(), "check-rejected") {
		t.Fatalf("check failure: %v", err)
	}
}

// Not parallel: PATH is process-wide; discovery succeeds and the generator's clang fails.
func TestSDKGeneratorFailureIsOrdinary(t *testing.T) {
	dir := t.TempDir()
	headers, err := filepath.Abs("testdata/SDK")
	if err != nil {
		t.Fatal(err)
	}
	frameworks := filepath.Join(dir, "System", "Library")
	if err := os.MkdirAll(frameworks, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(headers, filepath.Join(frameworks, "Frameworks")); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\ncase \"$3\" in\n--show-sdk-path) echo '" + dir + "';;\n--show-sdk-version) echo 15.0;;\n--find) echo '" + filepath.Join(dir, "clang") + "';;\n*) exit 2;;\nesac\n"
	if err := os.WriteFile(filepath.Join(dir, "xcrun"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "clang"), []byte("#!/bin/sh\necho generator-rejected >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	_, err = FromSDK(context.Background(), "macosx", "macos", []string{"Foundation"})
	var missing *platform.Missing
	if err == nil || errors.As(err, &missing) || !strings.Contains(err.Error(), "generator-rejected") {
		t.Fatalf("generator failure: %v", err)
	}
}
