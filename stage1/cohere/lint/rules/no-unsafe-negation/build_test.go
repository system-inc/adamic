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

// Not parallel: writes artifacts to the explicitly selected scratch directory.
func TestCompileProfiles(t *testing.T) {
	entry := os.Getenv("WAVE07_PROFILE")
	directory := os.Getenv("WAVE07_ARTIFACTS")
	if entry == "" || directory == "" {
		t.Skip("owned profile compilation requires scratch paths")
	}
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
}
