package main

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: sets the CLI opt-in with t.Setenv and restores it.
func TestProgramRegionCLIHandsMembershipAndRuntimeOneMode(t *testing.T) {
	t.Setenv("ADAMIC_PROGRAM_REGION", "1")
	path := filepath.Join(t.TempDir(), "main.a")
	source := `interface Link {next:Link|undefined} const root:Link={next:undefined}; root.next=root; console.log('ok');`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	called := false
	status := buildWithNative(path, filepath.Join(t.TempDir(), "out"), nil, func(code, output string, options native.Options) error {
		called = true
		if !options.ProgramRegion {
			t.Fatal("Program lowering must hand the same mode to every native builder, including the checker-archive paths")
		}
		if !strings.Contains(code, "adamic_program_adopt_owned") {
			t.Fatal("CLI did not infer the cyclic member")
		}
		return nil
	})
	if status != 0 || !called {
		t.Fatalf("CLI status=%d builder-called=%t", status, called)
	}
}
