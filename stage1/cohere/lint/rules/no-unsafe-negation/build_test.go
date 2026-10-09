package assertions

import (
	"context"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
)

// Compile the standalone owned profile in isolated scratch space on every gate.
// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestCompileProfiles(t *testing.T) {
	entry, err := filepath.Abs("profile.a")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	if err := native.Build(native.C(lowered), filepath.Join(directory, "native"), native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(native.C(lowered), filepath.Join(directory, "release"), native.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "emitted.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: sanitized native, release native and emitted JavaScript compiled")
}
