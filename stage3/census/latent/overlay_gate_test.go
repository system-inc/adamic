//go:build !census_overlay

package latent_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The tagged test binary is a Loom product. The test unit looks it up
// through the build cache and executes it without generating an overlay.
func latentCensusOverlayBinary(t *testing.T) string {
	t.Helper()
	const binary = "latent-overlay.test"
	flags := append(buildcache.ReproducibleGoFlags(), "-tags=census_overlay")
	// The unoverlaid compiler supplies the Go input closure. The measurements,
	// tests and generator inputs that are outside that closure are keyed below.
	inputs, err := buildcache.GoInputs(binary, "./internal/lower", flags, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("python3", "--version"))
	for _, pattern := range []string{
		"*.go", "*.go.txt", "make_overlay.py", "importclean/*.go",
		"refusalrewrite/*.go", "refusalrewrite/cmd/*.go", "replay/main.go.txt",
		"statecopy/*.go", "statementrewrite/*.go", "timing/timing.go",
	} {
		files, err := filepath.Glob(pattern)
		if err != nil || len(files) == 0 {
			t.Fatalf("overlay inputs %q: %v, found %d files", pattern, err, len(files))
		}
		for _, file := range files {
			inputs.Files = append(inputs.Files, filepath.ToSlash(filepath.Join("stage3/census/latent", file)))
		}
	}
	scratch := t.TempDir()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		generate := toolchain.CommandContextIn(t.Context(), root, "python3", "stage3/census/latent/make_overlay.py", root, scratch)
		if output, err := generate.CombinedOutput(); err != nil {
			return fmt.Errorf("generate census overlay: %w\n%s", err, output)
		}
		args := append([]string{"test", "-c"}, flags...)
		args = append(args, "-overlay="+filepath.Join(scratch, "overlay.json"), "-o", filepath.Join(directory, binary), "./stage3/census/latent")
		compile := toolchain.CommandContextIn(t.Context(), root, "go", args...)
		if output, err := compile.CombinedOutput(); err != nil {
			return fmt.Errorf("compile census overlay tests: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, binary)
}

func TestProduct_LatentCensusOverlay(t *testing.T) {
	t.Parallel()
	latentCensusOverlayBinary(t)
}

func TestLatentCensusOverlayGateHcxsh1k(t *testing.T) {
	t.Parallel()
	binary := latentCensusOverlayBinary(t)

	tests := []string{
		"TestLatentWrittenAnyDeferredCensusStopsZb0q0zd",
		"TestLatentScopedTailEntryFindingLoom03dg",
	}
	run := exec.CommandContext(t.Context(), binary, "-test.count=1", "-test.v", "-test.run=^("+strings.Join(tests, "|")+")$")
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("census overlay tests: %v\n%s", err, output)
	}
	for _, test := range tests {
		if !strings.Contains(string(output), "--- PASS: "+test+" (") {
			t.Fatalf("missing pass for %s\n%s", test, output)
		}
	}
}
