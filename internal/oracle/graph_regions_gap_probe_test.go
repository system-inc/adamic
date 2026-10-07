package oracle

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// The former leaks must match Node and free their graph, independently of
// merely seeing graph types in the lowered program.
func TestGraphAllocationFlowIsLeakClean(t *testing.T) {
	for _, name := range []string{"return", "conditional", "mixed", "override"} {
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
			report, err := leakcheck.Check(leakcheck.Program{
				C: code, Sanitized: binary,
				Counted: filepath.Join(t.TempDir(), "counted-leak"),
				Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
					// Exit-time stack/register words can conservatively hide a
					// missed allocation. Keep these roots out of this leak probe.
					environment = append([]string{"UBSAN_OPTIONS=halt_on_error=1", "LSAN_OPTIONS=use_stacks=0:use_registers=0"}, environment...)
					return leakRun(executeWith(t, environment, name, arguments...))
				},
			})
			if err != nil || report != "" {
				t.Fatalf("shared leak check: %v %s", err, report)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			counts := execute(t, counted)
			match := countsLine.FindSubmatch(counts.stderr)
			if counts.exitCode != 0 || match == nil || !graphRegionLine.Match(counts.stderr) {
				t.Fatalf("region free missing: %d %s", counts.exitCode, counts.stderr)
			}
			allocations, _ := strconv.Atoi(string(match[1]))
			frees, _ := strconv.Atoi(string(match[2]))
			arenas, _ := strconv.Atoi(string(match[6]))
			if allocations != frees+arenas {
				t.Fatalf("allocation flow leaked: %s", counts.stderr)
			}
			t.Logf("Node, ASan/UBSan and shared leak check output %q; counted: %s", node.stdout, counts.stderr)
		})
	}
}
