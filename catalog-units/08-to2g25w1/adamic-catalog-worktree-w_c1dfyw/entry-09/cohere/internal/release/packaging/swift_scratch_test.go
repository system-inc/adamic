package release

import (
	"path/filepath"
	"strings"
	"testing"
)

// The Swift engine's build directory is never inside the output, which the release packs and uploads
// whole, and a directory that only shares the output's name as a prefix is not inside it.
func TestTheSwiftScratchIsNeverInsideTheOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	output := filepath.Join(root, "dist")

	scratch, err := swiftScratchDirectory(Options{OutputDirectory: output})
	if err != nil {
		t.Fatalf("the default was refused: %v", err)
	}
	if strings.HasPrefix(scratch, output+string(filepath.Separator)) {
		t.Fatalf("the default %s is inside the output %s", scratch, output)
	}

	for _, testCase := range []struct {
		scratch string
		allowed bool
	}{
		{filepath.Join(output, ".swift-build"), false},
		{filepath.Join(output, "a", "b"), false},
		{output, false},
		{filepath.Join(root, "dist-scratch"), true},
		{filepath.Join(root, "elsewhere"), true},
		{root, true},
	} {
		_, err := swiftScratchDirectory(Options{OutputDirectory: output, SwiftScratchDirectory: testCase.scratch})
		if (err == nil) != testCase.allowed {
			t.Errorf("%s: allowed %v, want %v (%v)", testCase.scratch, err == nil, testCase.allowed, err)
		}
	}
}

// Build holds the rule, before anything is compiled: a scratch inside the output is refused by name.
func TestBuildRefusesASwiftScratchInsideTheOutput(t *testing.T) {
	t.Parallel()

	output := t.TempDir()
	_, err := Build(Options{
		Version:               "1.0.0",
		ModuleDirectory:       t.TempDir(),
		OutputDirectory:       output,
		SwiftScratchDirectory: filepath.Join(output, ".swift-build"),
	})
	if err == nil || !strings.Contains(err.Error(), "is inside the output") {
		t.Fatalf("Build accepted a Swift scratch inside its output: %v", err)
	}
}
