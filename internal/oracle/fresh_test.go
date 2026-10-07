package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strconv"
	"testing"
)

// Former refusal probes must now agree three ways and free their cyclic region.
// Losing the fresh proof cannot silently turn an accepted program into a leak.
func TestFreshWriteProbesUseRegions(t *testing.T) {
	t.Parallel()
	probes, err := filepath.Glob(filepath.Join(repository, "internal", "oracle", "testdata", "fresh_refused", "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(probes) < 20 {
		t.Fatalf("found only %d probes", len(probes))
	}
	for _, path := range probes {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatal("cycle accepted without graph types")
			}
			oracle, backend := onNode(t, path), onJavaScriptBackend(t, program)
			actual, sanitized := natively(t, program)
			for _, pair := range []struct {
				name   string
				result run
			}{{"JavaScript", backend}, {"sanitized", actual}, {"release", released(t, program)}} {
				if difference := disagreement(oracle, pair.result); difference != "" {
					t.Errorf("%s: %s; Node %q, actual %q; stderr %q", pair.name, difference, oracle.stdout, pair.result.stdout, pair.result.stderr)
				}
			}
			if leak := leaks(t, program, sanitized); leak != "" {
				t.Errorf("leaks: %s", leak)
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, args := pinnedStack(binary)
			result := execute(t, name, args...)
			counts := countsLine.FindSubmatch(result.stderr)
			if result.exitCode != 0 || counts == nil {
				t.Fatalf("counted run failed: %d %s", result.exitCode, result.stderr)
			}
			allocations, _ := strconv.Atoi(string(counts[1]))
			frees, _ := strconv.Atoi(string(counts[2]))
			arenas, _ := strconv.Atoi(string(counts[6]))
			if allocations != frees+arenas {
				t.Errorf("allocations %s, frees %s", counts[1], counts[2])
			}
			if !graphRegionLine.Match(result.stderr) {
				t.Fatalf("no region free evidence: %s", result.stderr)
			}
			t.Logf("%s", fmt.Sprintf("allocations %s frees %s retains %s releases %s peak %s; %s", counts[1], counts[2], counts[3], counts[4], counts[5], graphCountsLine.Find(result.stderr)))
		})
	}
}
