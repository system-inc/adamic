package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/platform"
)

// The test fails with the change removed, or with any ordinary compiler error
// mapped to 5. Text resembling the platform stop never reserves its exit code.
func TestPlatformMissingExitReserved(t *testing.T) {
	t.Parallel()
	missing := &platform.Missing{Platform: "darwin", Requirement: "Xcode SDK"}
	for _, err := range []error{missing, fmt.Errorf("load: %w", missing), errors.Join(errors.New("context"), missing)} {
		if got := compilerErrorExit(err); got != platform.MissingExitCode {
			t.Fatalf("typed platform error: exit %d", got)
		}
	}
	for _, err := range []error{nil, errors.New("compile failed"), errors.New(missing.Error()), errors.New("error TS2307"), errors.New("stage 0 can't lower"), &load.CheckError{}, os.ErrNotExist, &os.PathError{Op: "open", Path: "source.ts", Err: os.ErrPermission}, &exec.ExitError{}, fmt.Errorf("wrapped: %w", os.ErrInvalid)} {
		if got := compilerErrorExit(err); got == platform.MissingExitCode {
			t.Fatalf("ordinary error %T reserved exit 5", err)
		}
		_, code := checkWithLoader([]string{"source.a"}, func([]string) (*load.Program, error) { return nil, err })
		if code == platform.MissingExitCode {
			t.Fatalf("ordinary load error %T reserved exit 5", err)
		}

	}
	if platform.MissingExitCode != 5 {
		t.Fatal("platform stop changed its reserved exit")
	}
}

func TestPlatformMissingCheckExit(t *testing.T) {
	t.Parallel()
	program, code := checkWithLoader([]string{"source.a"}, func([]string) (*load.Program, error) {
		return nil, fmt.Errorf("load: %w", &platform.Missing{Platform: "darwin", Requirement: "SDK"})
	})
	if program != nil || code != platform.MissingExitCode {
		t.Fatalf("check: program %v, exit %d", program, code)
	}
}

func TestPlatformMissingBuildExit(t *testing.T) {
	t.Parallel()
	code := buildWithNative("testdata/build_target.a", filepath.Join(t.TempDir(), "program"), nil, func(string, string, native.Options) error {
		return fmt.Errorf("native: %w", &platform.Missing{Platform: "darwin", Requirement: "xcrun"})
	})
	if code != platform.MissingExitCode {
		t.Fatalf("native build: exit %d", code)
	}
}

// Not parallel: Process environment is serial: use the real loader rather than an injected error.
func TestAppleImportExitsPlatformMissing(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Linux missing-platform control")
	}
	t.Setenv(load.AppleBindingsVariable, "")
	source := filepath.Join(t.TempDir(), "apple.a")
	if err := os.WriteFile(source, []byte("import 'apple/foundation';\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, code := check([]string{source})
	if code != platform.MissingExitCode {
		t.Fatalf("Apple import: exit %d, want 5", code)
	}
}
