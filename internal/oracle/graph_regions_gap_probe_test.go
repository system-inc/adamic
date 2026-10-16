package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// The same lowered program with graph metadata disabled is a counterfactual
// counted baseline. Cycles leak there; it is not an accepted compiler mode.
func withoutGraph(program *ir.Program) *ir.Program {
	// Lowered programs are shared immutable setup across parallel oracle tests.
	baseline := *program
	baseline.Functions = append([]ir.Function(nil), program.Functions...)
	baseline.Locals = append([]ir.Local(nil), program.Locals...)
	baseline.Classes = append([]ir.Class(nil), program.Classes...)
	program = &baseline
	program.GraphTypes = nil
	for i := range program.Functions {
		program.Functions[i].GraphClosure = false
	}
	for i := range program.Locals {
		program.Locals[i].GraphCell = false
	}
	for i := range program.Classes {
		program.Classes[i].Graph = false
	}
	return program
}

var graphRSS = regexp.MustCompile(`(?m)^graph_peak_rss_kib (\d+)\n`)
var graphRegionLine = regexp.MustCompile(`(?m)^adamic: graph region: live (\d+) bytes (\d+) reachable (\d+) bytes (\d+) unreachable (\d+) bytes (\d+) metadata (\d+)\n`)

// Run in a fresh shell child: Linux otherwise inherits the Go parent's high
// water mark across exec. RSS includes allocator retention and runtime metadata.
func graphMeasured(t *testing.T, code string) run {
	t.Helper()
	at := strings.LastIndex(code, "\treturn 0;")
	if at < 0 {
		t.Fatal("missing main return")
	}
	code = "#include <sys/resource.h>\n#include <stdio.h>\n" + code[:at] + "\tstruct rusage graph_usage; getrusage(RUSAGE_SELF, &graph_usage); printf(\"graph_peak_rss_kib %ld\\n\", graph_usage.ru_maxrss);\n" + code[at:]
	binary := filepath.Join(t.TempDir(), "measured")
	if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	name, args := pinnedStack("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", binary)
	result := executeWith(t, []string{"ADAMIC_THREADS=1"}, name, args...)
	if result.exitCode != 0 || countsLine.FindSubmatch(result.stderr) == nil || graphRSS.FindSubmatch(result.stdout) == nil {
		t.Fatalf("measurement failed: %d %s %s", result.exitCode, result.stdout, result.stderr)
	}
	return result
}

func graphMeasureFixture(t *testing.T, path string, million bool) {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, absolute)
	program, err := lowered(t, absolute)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) == 0 {
		t.Log("no graph types selected; counted and graph emission are identical")
	}
	code := native.C(program)
	if million {
		at := strings.LastIndex(code, "\treturn 0;")
		code = code[:at] + "\tprintf(\"graph_node_bytes %zu\\n\", adamic_object_size(3));\n" + code[at:]
	}
	after := graphMeasured(t, code)
	if report := leakcheck.Unbalanced(leakRun(after)); report != "" {
		t.Fatal(report)
	}
	original := native.C(program)
	baseline := native.C(withoutGraph(program))
	if native.C(program) != original {
		t.Fatal("counted baseline mutated shared graph IR")
	}
	before := graphMeasured(t, baseline)
	for _, result := range []run{before, after} {
		output := graphRSS.ReplaceAll(result.stdout, nil)
		if million {
			output = regexp.MustCompile(`(?m)^graph_node_bytes \d+\n`).ReplaceAll(output, nil)
		}
		if string(output) != string(truth.stdout) || truth.exitCode != 0 {
			t.Fatalf("measurement changed Node output: %q != %q", output, truth.stdout)
		}
	}
	oldCounts := countsLine.FindSubmatch(before.stderr)
	newCounts := countsLine.FindSubmatch(after.stderr)
	t.Logf("%s counted R/L/peak-live=%s/%s/%s peak-RSS-KiB=%s; graph=%s/%s/%s peak-RSS-KiB=%s; counted leak: %s", path, oldCounts[3], oldCounts[4], oldCounts[5], graphRSS.FindSubmatch(before.stdout)[1], newCounts[3], newCounts[4], newCounts[5], graphRSS.FindSubmatch(after.stdout)[1], leakcheck.Unbalanced(leakRun(before)))
	if million {
		size := regexp.MustCompile(`(?m)^graph_node_bytes (\d+)$`).FindSubmatch(after.stdout)
		if size == nil {
			t.Fatal("missing full object size")
		}
		bytes, err := strconv.ParseUint(string(size[1]), 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		want := fmt.Sprintf("live 1000000 bytes %d reachable 950001 bytes %d unreachable 49999 bytes %d metadata 16000064", 1000000*bytes, 950001*bytes, 49999*bytes)
		if !strings.Contains(string(after.stderr), want) {
			t.Fatalf("layout-derived million bytes missing: %s", after.stderr)
		}
		t.Logf("million node full allocation=%d bytes, graph prefix=16 bytes/member, singleton region record=64 bytes total; amortized total=%g bytes/member", bytes, float64(bytes)+16+64.0/1000000)
	}
}

// Not parallel: compare resident sets without overlapping large witnesses.
func TestGraphMeasurementMillion(t *testing.T) {
	graphMeasureFixture(t, "internal/oracle/testdata/graph_regions/million.a", true)
}
func TestGraphMeasurementLinks(t *testing.T) {
	t.Parallel()
	graphMeasureFixture(t, "internal/oracle/testdata/graph_regions_list.a", false)
}
func TestGraphMeasurementArrays(t *testing.T) {
	t.Parallel()
	graphMeasureFixture(t, "internal/oracle/testdata/graph_regions_parse.a", false)
}
func TestGraphMeasurementMaps(t *testing.T) {
	t.Parallel()
	graphMeasureFixture(t, "internal/oracle/testdata/graph_regions_symbols.a", false)
}
func TestGraphMeasurementCanonical(t *testing.T) {
	t.Parallel()
	graphMeasureFixture(t, "internal/oracle/testdata/canonical_graph.a", false)
}
