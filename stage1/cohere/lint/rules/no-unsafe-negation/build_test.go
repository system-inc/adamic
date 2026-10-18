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

// Compile the standalone owned profile in isolated scratch space on every gate, one test for each of its three builds,
// each named for its rule, since a test's tail is what the 60-second bar reads (#mntyb57: the three in one test
// reached 65.7 s on a loaded tree). Timed on a drained Pupa, each phase alone: load and lower about 1 s, the sanitized
// build 18 to 19 s, the release build 8 to 9 s, the JavaScript 0.02 s.

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestNoUnsafeNegationProfileBuildsSanitized(t *testing.T) {
	lowered := compileProfilesLowered(t)
	if err := native.Build(native.C(lowered), filepath.Join(t.TempDir(), "native"), native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: sanitized native compiled")
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestNoUnsafeNegationProfileBuildsRelease(t *testing.T) {
	lowered := compileProfilesLowered(t)
	if err := native.Build(native.C(lowered), filepath.Join(t.TempDir(), "release"), native.Options{}); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: release native compiled")
}

func TestNoUnsafeNegationProfileEmitsJavaScript(t *testing.T) {
	t.Parallel()
	lowered := compileProfilesLowered(t)
	if err := os.WriteFile(filepath.Join(t.TempDir(), "emitted.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: emitted JavaScript compiled")
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
