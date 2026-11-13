package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The test fails with the change removed, or with any ordinary compiler error
// mapped to 5. Text resembling the platform stop never reserves its exit code.
func TestPlatformMissingExitReserved(t *testing.T) {
	t.Parallel()
	missing := &toolchain.PlatformMissing{Platform: "darwin", Missing: "Xcode SDK"}
	for _, err := range []error{missing, fmt.Errorf("load: %w", missing), errors.Join(errors.New("context"), missing)} {
		if got := compilerErrorExit(err); got != toolchain.PlatformMissingExitCode {
			t.Fatalf("typed platform error: exit %d", got)
		}
	}
	for _, err := range []error{nil, errors.New("compile failed"), errors.New(missing.Error()), errors.New("error TS2307"), errors.New("stage 0 can't lower"), &load.CheckError{}, os.ErrNotExist, &os.PathError{Op: "open", Path: "source.ts", Err: os.ErrPermission}, &exec.ExitError{}, fmt.Errorf("wrapped: %w", os.ErrInvalid)} {
		if got := compilerErrorExit(err); got == toolchain.PlatformMissingExitCode {
			t.Fatalf("ordinary error %T reserved exit 5", err)
		}
		_, code := checkWithLoader([]string{"source.a"}, func([]string) (*load.Program, error) { return nil, err })
		if code == toolchain.PlatformMissingExitCode {
			t.Fatalf("ordinary load error %T reserved exit 5", err)
		}

	}
	if toolchain.PlatformMissingExitCode != 5 {
		t.Fatal("platform stop changed its reserved exit")
	}
}

func TestPlatformMissingCheckExit(t *testing.T) {
	t.Parallel()
	program, code := checkWithLoader([]string{"source.a"}, func([]string) (*load.Program, error) {
		return nil, fmt.Errorf("load: %w", &toolchain.PlatformMissing{Platform: "darwin", Missing: "SDK"})
	})
	if program != nil || code != toolchain.PlatformMissingExitCode {
		t.Fatalf("check: program %v, exit %d", program, code)
	}
}

func TestPlatformMissingBuildExit(t *testing.T) {
	t.Parallel()
	code := buildWithNative("testdata/build_target.a", filepath.Join(t.TempDir(), "program"), nil, func(string, string, native.Options) error {
		return fmt.Errorf("native: %w", &toolchain.PlatformMissing{Platform: "darwin", Missing: "xcrun"})
	})
	if code != toolchain.PlatformMissingExitCode {
		t.Fatalf("native build: exit %d", code)
	}
}
