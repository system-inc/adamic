package helpers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestSlot02Batch9(t *testing.T) {
	// Not parallel: baseline and compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch9"
	file, err := os.Open(base + "/testdata/sources.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zipped, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer zipped.Close()
	var inputs []map[string]string
	observed := map[string]bool{}
	scan := bufio.NewScanner(zipped)
	scan.Buffer(make([]byte, 65536), 16<<20)
	for scan.Scan() {
		var row map[string]string
		if err = json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, row)
		observed[row["rule"]] = true
	}
	if err = scan.Err(); err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	symbols := []string{"control_flow_graph.*Builder[E].setCycleBarrier", "control_flow_graph.*Builder[E].linkWithCycleBarrier", "control_flow_graph.*Builder[E].link"}
	for _, symbol := range symbols {
		count := 0
		for _, row := range ledger.Remaining {
			for _, h := range row.Helpers {
				if strings.HasSuffix(h, "/"+symbol) {
					count++
					if !observed[row.Rule] {
						t.Fatalf("missing captured consumer %s", row.Rule)
					}
				}
			}
		}
		if count != 4 {
			t.Fatalf("consumer count drift: %s: %d", symbol, count)
		}
		t.Logf("%s: %d consumers with actual runtime capture", symbol, count)
	}
	config := smallFixture(t, inputs)
	cohere, _ := filepath.Abs("../../../../cohere")
	oracleSource, _ := filepath.Abs(base + "/testdata/oracle.go")
	export, _ := filepath.Abs(base + "/testdata/cfg_export.go")
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch9.go"): oracleSource, filepath.Join(cohere, "internal/lint/ecmascript/control_flow_graph/adamic_slot02_batch9.go"): export}
	cfgFile := filepath.Join(cohere, "internal/lint/ecmascript/control_flow_graph/cfg.go")
	cfgData, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	cfgText := string(cfgData)
	for _, hook := range []struct{ anchor, body string }{
		{"link(from, to *Block[E]) {", `slot02Capture("link",from,to,0,false)`},
		{"linkWithCycleBarrier(from, to *Block[E], barrier bool) int {", `slot02Capture("with",from,to,0,barrier)
slot02TraceCall("with",from,to,0,barrier)`},
		{"setCycleBarrier(from *Block[E], index int, barrier bool) {", `slot02Capture("set",from,nil,index,barrier)
slot02TraceCall("barrier",from,nil,index,barrier)`},
		{"appendSuccessor(from, to *Block[E]) {", `slot02TraceCall("append",from,to,0,false)`},
	} {
		if strings.Count(cfgText, hook.anchor) != 1 {
			t.Fatal("Go instrumentation anchor drift")
		}
		cfgText = strings.Replace(cfgText, hook.anchor, hook.anchor+"\n"+hook.body, 1)
	}
	instrumented := filepath.Join(t.TempDir(), "cfg.go")
	if err = os.WriteFile(instrumented, []byte(cfgText), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[cfgFile] = instrumented
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch9.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                     string
		Sources, Roots, Observed int
		Cases                    []json.RawMessage
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	var input map[string]json.RawMessage
	if err = json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	delete(input, "Want")
	path := smallFixture(t, input)
	want := []byte(corpus.Want)
	t.Logf("%d captures; %d parsed sources; %d graph roots; %d observed states; %d total cases", len(inputs), corpus.Sources, corpus.Roots, corpus.Observed, len(corpus.Cases))
	entry, _ := filepath.Abs(base + "/main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	check := func(t *testing.T, entry string) []byte {
		t.Helper()
		gotNode := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path)
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			t.Fatal(err)
		}
		nativePath := filepath.Join(t.TempDir(), "program")
		if err = native.Build(native.C(ir), nativePath, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := run(t, "", nativePath, path)
		jsPath := filepath.Join(t.TempDir(), "program.mjs")
		if err = os.WriteFile(jsPath, []byte(javascript.JavaScript(ir)), 0644); err != nil {
			t.Fatal(err)
		}
		gotJS := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, jsPath, path)
		compare(t, got, gotNode)
		compare(t, gotJS, gotNode)
		return got
	}
	compare(t, check(t, entry), want)
	t.Log("Actual Go, Node source, emitted JavaScript and sanitized native agree")
	mutants := []struct{ name, file, anchor, replacement string }{
		{"nil barrier false stays nil", "set_cycle_barrier.a", "if (!barrier) { return; }", "if (!barrier && index >= block.successors.length) { return; }"},
		{"barrier padding false", "set_cycle_barrier.a", "while (block.barriers.length < block.successors.length) { block.barriers.push(false); }", "while (block.barriers.length < block.successors.length) { block.barriers.push(true); }"},
		{"barrier value", "set_cycle_barrier.a", "block.barriers[index] = barrier;", "block.barriers[index] = !barrier;"},
		{"negative index guard", "set_cycle_barrier.a", "index < 0", "index <= 0"},
		{"edge index before append", "link_with_cycle_barrier.a", "return index;", "return index + 1;"},
		{"reachable propagation", "link_with_cycle_barrier.a", "if (source.reachable)", "if (!source.reachable)"},
		{"incoming retains true", "link_with_cycle_barrier.a", "target.incoming = true;", "target.incoming = false;"},
		{"barrier forwarding", "link_with_cycle_barrier.a", "dependencies.setCycleBarrier(from, index, barrier);", "dependencies.setCycleBarrier(from, index, !barrier);"},
		{"plain link barrier false", "link.a", "linkWithCycleBarrier(from, to, false);", "linkWithCycleBarrier(from, to, true);"},
		{"nil endpoint return", "link_with_cycle_barrier.a", "return -1;", "return -2;"},
		{"plain link exactly once", "link.a", "linkWithCycleBarrier(from, to, false);", "linkWithCycleBarrier(from, to, false); linkWithCycleBarrier(from, to, false);"},
		{"append before barrier", "link_with_cycle_barrier.a", "dependencies.appendSuccessor(from, to);\n    dependencies.setCycleBarrier(from, index, barrier);", "if (!barrier && !source.barriersPresent) { dependencies.setCycleBarrier(from, index, barrier); dependencies.appendSuccessor(from, to); } else { dependencies.appendSuccessor(from, to); dependencies.setCycleBarrier(from, index, barrier); }"},
		{"plain link target identity", "link.a", "linkWithCycleBarrier(from, to, false);", "linkWithCycleBarrier(from, from, false);"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/state.a", base + "/set_cycle_barrier.a", base + "/link_with_cycle_barrier.a", base + "/link.a"}
			for _, name := range files {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == base+"/"+mutant.file {
					if strings.Count(string(data), mutant.anchor) != 1 {
						t.Fatal("mutant anchor drift")
					}
					data = []byte(strings.Replace(string(data), mutant.anchor, mutant.replacement, 1))
				}
				targetName := name
				if name == "options_json.ts" {
					targetName = "options_json.a"
				}
				if name == base+"/main.a" {
					data = []byte(strings.Replace(string(data), "../../options_json.ts", "../../options_json.a", 1))
				}
				target := filepath.Join(dir, targetName)
				if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(target, data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := check(t, filepath.Join(dir, base, "main.a"))
			if bytes.Equal(got, want) {
				t.Fatal("compiled semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i, line := range a {
				if i < len(b) && line != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, line, b[i])
					break
				}
			}
		})
	}
}
