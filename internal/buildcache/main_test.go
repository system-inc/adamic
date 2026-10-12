package buildcache

import (
	"fmt"
	"os"
	"testing"
)

// TestMain gives the tests a tools directory of their own, empty, when the environment names none: a tool's place is
// only ever ADAMIC_TOOLS (toolsRoot), never a home's, and a test that asks what happens without it sets it empty itself.
func TestMain(m *testing.M) {
	if os.Getenv("ADAMIC_TOOLS") != "" {
		os.Exit(m.Run())
	}
	directory, err := os.MkdirTemp("", "adamic-tools-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Setenv("ADAMIC_TOOLS", directory)
	code := m.Run()
	os.RemoveAll(directory)
	os.Exit(code)
}
