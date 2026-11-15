package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// looseSlotOverStopProbes are the six oracle programs #kaz4pk8's loose-slot head newly stops at the entry
// arm (J-we853r1's census). A scratch probe, never merged: it says whether each balances on the counted
// build here, which is what tells an over-stop from a live leak the sanitizer's leak check can't see.
var looseSlotOverStopProbes = []string{
	"generators/map_defined_iterator.a",
	"generators/map_iterator.a",
	"generators/map_iterator_entries.a",
	"generators/new_map_plain.a",
	"to_primitive/arrays.a",
	"to_primitive/probes.a",
}

// TestLooseSlotOverStopProbe lowers each program and runs its counted build, one PROBE line each. It
// fails only when a probe can't be measured.
func TestLooseSlotOverStopProbe(t *testing.T) {
	t.Parallel()
	for _, name := range looseSlotOverStopProbes {
		t.Run(filepath.Base(name), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s stop %v", name, err)
				return
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if buildErr := native.Build(native.C(program), binary, native.Options{Count: true}); buildErr != nil {
				t.Fatalf("PROBE %s lowered, the counted build failed: %v", name, buildErr)
			}
			got := execute(t, binary)
			match := leakcheck.CountsLine.FindSubmatch(got.stderr)
			if match == nil {
				t.Fatalf("PROBE %s lowered, no counts line: exit %d, stderr %q", name, got.exitCode, got.stderr)
			}
			verdict := leakcheck.Unbalanced(leakcheck.Run{Stdout: got.stdout, Stderr: got.stderr, ExitCode: got.exitCode})
			if verdict == "" {
				verdict = "balanced"
			}
			t.Logf("PROBE %s lowered allocations %s frees %s regions %s graph types %d exit %d: %s", name, match[1], match[2], match[6], len(program.GraphTypes), got.exitCode, verdict)
		})
	}
}
