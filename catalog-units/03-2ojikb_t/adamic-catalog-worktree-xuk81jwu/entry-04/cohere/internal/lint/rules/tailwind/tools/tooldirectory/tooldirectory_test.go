package tooldirectory

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// The directory is found from either form runtime.Caller gives, and both forms name the same place.
//
// The trimmed form is written out rather than taken from runtime.Caller, so the test covers it on a
// machine that builds without -trimpath too, and the caller's own form covers whichever this build has.
func TestOfFindsTheSourceDirectoryFromEitherForm(t *testing.T) {
	t.Parallel()
	want, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	_, callerPath, _, _ := runtime.Caller(0)
	for _, form := range []string{
		callerPath,
		filepath.Join(want, "tooldirectory_test.go"),
		"github.com/system-inc/cohere/internal/lint/rules/tailwind/tools/tooldirectory/tooldirectory_test.go",
	} {
		got, err := Of(form)
		if err != nil {
			t.Fatalf("%s: %v", form, err)
		}
		if got != want {
			t.Errorf("%s resolved to %s, want %s", form, got, want)
		}
	}
}

// A path naming a file that is not there is refused rather than answered with a directory nobody
// chose, which is what filepath.Dir of a trimmed path used to be.
func TestOfRefusesADirectoryWithoutTheFile(t *testing.T) {
	t.Parallel()
	if directory, err := Of(filepath.Join(t.TempDir(), "caller.go")); err == nil {
		t.Errorf("a directory without the caller's file was accepted: %s", directory)
	}
	if directory, err := Of("github.com/system-inc/cohere/no/such/package/caller.go"); err == nil {
		t.Errorf("an import path that resolves nowhere was accepted: %s", directory)
	}
}
