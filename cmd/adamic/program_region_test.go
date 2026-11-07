package main

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: sets the CLI opt-in with test.Setenv and restores it.
func TestProgramRegionCLIHandsMembershipAndRuntimeOneMode(test *testing.T) {
	test.Setenv("ADAMIC_PROGRAM_REGION", "1")
	path := filepath.Join(test.TempDir(), "main.a")
	source := `interface Link {next:Link|undefined} const root:Link={next:undefined}; root.next=root; console.log('ok');`
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	called := false
	status := buildWithNative(path, filepath.Join(
		test.TempDir(),
		"out",
	), nil, func(code, output string, options native.Options) error {
		called = true
		if !options.ProgramRegion {
			test.Fatal(
				"Program lowering must hand the same mode to every native builder, including the checker-archive paths",
			)
		}
		if !strings.Contains(code, "adamic_program_adopt_owned") {
			test.Fatal("CLI did not infer the cyclic member")
		}
		return nil
	})
	if status != 0 || !called {
		test.Fatalf("CLI status=%d builder-called=%t", status, called)
	}
}
