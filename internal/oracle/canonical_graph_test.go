package oracle

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/canonical_graph.a", "internal/oracle/testdata/canonical_graph_counted.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// LeakSanitizer and the allocator ledger independently hold the source witness.
func TestCanonicalGraphOwnership(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"canonical_graph.a", "canonical_graph_counted.a"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, args := pinnedStack(binary)
			result := execute(t, name, args...)
			oracle := onNode(t, path)
			if result.exitCode != oracle.exitCode || !bytes.Equal(result.stdout, oracle.stdout) {
				t.Fatalf("counted execution disagrees with Node: %d %s %s", result.exitCode, result.stdout, result.stderr)
			}
			if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
				t.Fatal(report)
			}
			if !graphRegionLine.Match(result.stderr) {
				t.Fatalf("canonical fixture did not destroy a graph region: %s", result.stderr)
			}
			t.Logf("balanced graph ownership:\n%s", result.stderr)
		})
	}
}
