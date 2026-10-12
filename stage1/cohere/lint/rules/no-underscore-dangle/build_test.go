package assertions

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
)

// Compile the standalone owned profile in isolated scratch space on every gate, as two tests: the sanitized build is
// most of the work (about 20 s of 30 alone on 4 cores, #s8sd87z), so each half fits a unit's share on its own.

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestCompileProfiles_Sanitized(t *testing.T) {
	lowered := compileProfilesLowered(t)
	if err := native.Build(native.C(lowered), filepath.Join(t.TempDir(), "native"), native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: sanitized native compiled")
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestCompileProfiles_Release(t *testing.T) {
	lowered := compileProfilesLowered(t)
	directory := t.TempDir()
	if err := native.Build(native.C(lowered), filepath.Join(directory, "release"), native.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "emitted.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: release native and emitted JavaScript compiled")
}

// compileProfilesLowered loads and lowers profile.a, about a second, so each test lowers its own.
func compileProfilesLowered(t *testing.T) *ir.Program {
	t.Helper()
	entry, err := filepath.Abs("profile.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	return lowered
}
