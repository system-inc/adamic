package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// graphFlowLoopSource is Memory's loop witness for #y7wexf5: a long-lived cyclic graph, and a loop of
// %d iterations each making a copy of its array by %s, reading it and dropping it.
const graphFlowLoopSource = `interface Node { name: string; next: Node[] }
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
let total = 0;
for (let i = 0; i < %d; i++) {
 const copy = a.next.%s;
 total += copy.length;
}
console.log(` + "`${total} ${a.next.length}`" + `);
`

// TestGraphFlowSortedCopyPeak builds the loop witness counted at N = 10 and N = 1000, with slice() and
// with toSorted, and holds toSorted's peak to slice()'s shape. A planned copy of a graph's array holds
// the graph's members, so adamic_graph_adopt_owned joins it to their region as it's made (runtime
// graph_regions.c:291), and a dropped copy is held until the graph dies: slice()'s peak grows with N at
// main (Memory measured 18 at N = 10 and 1008 at N = 1000 on 0816e185), which #5nrs3q2 holds. toSorted's
// copy went unplanned at main, so its peak was flat there; planned now, it may grow as slice()'s does,
// and no faster. Every run frees everything at exit and prints what Node prints.
func TestGraphFlowSortedCopyPeak(t *testing.T) {
	t.Parallel()
	copies := []struct{ name, call string }{
		{"slice", "slice()"},
		{"toSorted", "toSorted((x, y) => (x.name < y.name ? -1 : 1))"},
	}
	peaks := map[string]map[int]int{}
	for _, each := range copies {
		peaks[each.name] = map[int]int{}
		for _, iterations := range []int{10, 1000} {
			path := filepath.Join(t.TempDir(), fmt.Sprintf("loop_%s_%d.a", each.name, iterations))
			if err := os.WriteFile(path, []byte(fmt.Sprintf(graphFlowLoopSource, iterations, each.call)), 0o644); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("%s at N = %d on Node: %d %s", each.name, iterations, truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("%s at N = %d: %v", each.name, iterations, err)
			}
			if len(program.GraphTypes) == 0 {
				t.Fatalf("%s at N = %d: no graph types selected, so the loop measures nothing of graph regions", each.name, iterations)
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			name, arguments := pinnedStack(binary)
			result := execute(t, name, arguments...)
			if result.exitCode != truth.exitCode || string(result.stdout) != string(truth.stdout) {
				t.Fatalf("%s at N = %d: counted disagrees with Node: %d %q %s", each.name, iterations, result.exitCode, result.stdout, result.stderr)
			}
			if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
				t.Fatalf("%s at N = %d: %s", each.name, iterations, report)
			}
			match := countsLine.FindSubmatch(result.stderr)
			if match == nil {
				t.Fatalf("%s at N = %d: no counts at the end of stderr: %q", each.name, iterations, result.stderr)
			}
			peak, err := strconv.Atoi(string(match[5]))
			if err != nil {
				t.Fatal(err)
			}
			peaks[each.name][iterations] = peak
		}
	}
	slice, sorted := peaks["slice"], peaks["toSorted"]
	t.Logf("counted peak: slice() %d at N = 10, %d at N = 1000; toSorted %d at N = 10, %d at N = 1000", slice[10], slice[1000], sorted[10], sorted[1000])
	if sorted[1000]-sorted[10] > slice[1000]-slice[10] {
		t.Errorf("toSorted's peak grows by %d from N = 10 to N = 1000, more than slice()'s %d (#5nrs3q2's growth, which this pass must not worsen)", sorted[1000]-sorted[10], slice[1000]-slice[10])
	}
}
