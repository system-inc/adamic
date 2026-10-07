package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// These are accepted known leaks, deliberately outside the leak-clean fixture
// registry. A compiler fix should make this test fail so the probes can graduate.
func TestGraphAllocationClassificationGapIsLeakOnly(t *testing.T) {
	for _, name := range []string{"return", "conditional", "mixed"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/graph_regions/classification_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatal("probe did not reach graph classification")
			}
			code := native.C(program)
			binary := filepath.Join(t.TempDir(), "sanitized")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: %d %s", node.exitCode, node.stderr)
			}
			if difference := disagreement(node, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatal(difference)
			}
			normal := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if difference := disagreement(node, normal); difference != "" {
				t.Fatalf("ASan/UBSan: %s: %d %s", difference, normal.exitCode, normal.stderr)
			}
			report := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			diagnostics := string(report.stderr)
			if report.exitCode == 0 || !strings.Contains(diagnostics, "LeakSanitizer: detected memory leaks") || strings.Contains(diagnostics, "heap-use-after-free") || strings.Contains(diagnostics, "runtime error:") || string(report.stdout) != string(node.stdout) {
				t.Fatalf("want leak only, got %d %s", report.exitCode, diagnostics)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			counts := execute(t, counted)
			match := countsLine.FindSubmatch(counts.stderr)
			if counts.exitCode != 0 || match == nil || string(match[1]) == string(match[2]) || graphRegionLine.Match(counts.stderr) {
				t.Fatalf("gap unexpectedly freed: %d %s", counts.exitCode, counts.stderr)
			}
			t.Logf("Node and ASan/UBSan output %q; counted: %s; leak report: %s", node.stdout, counts.stderr, diagnostics)
		})
	}
}
