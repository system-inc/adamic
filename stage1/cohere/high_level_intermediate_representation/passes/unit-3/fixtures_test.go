package unit3

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type fixture struct {
	Key, Pass, Before, After string
	Probe, Flow              bool
}

var passes = []string{"outline_functions", "drop_manual_memoization", "inline_iife", "inline_iife_including_memo_callbacks", "inline_remap", "invoked_functions", "dead_code_elimination", "merge_consecutive_blocks"}

func unit3Root(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func unit3Command(t *testing.T, root string, env []string, args ...string) []byte {
	t.Helper()
	c := exec.Command(args[0], args[1:]...)
	c.Dir = root
	c.Env = append(os.Environ(), env...)
	out, err := c.CombinedOutput()
	if err != nil {
		if len(out) > 8192 {
			out = out[len(out)-8192:]
		}
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return out
}
func observeCalls(data []byte) []byte {
	var scan scanner.Scanner
	files := token.NewFileSet()
	file := files.AddFile("test.go", files.Base(), len(data))
	scan.Init(file, data, nil, 0)
	type edit struct {
		start, end int
		text       string
	}
	var edits []edit
	for {
		pos, tok, text := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.IDENT && (text == "Lower" || text == "ForFunction" || text == "ForFunctionWithoutManualMemoization") {
			at := file.Offset(pos)
			edits = append(edits, edit{at, at + len(text), "stage1Observed" + text})
		}
	}
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		data = append(append(append([]byte{}, data[:e.start]...), []byte(e.text)...), data[e.end:]...)
	}
	return data
}
func TestGenerateUnit3Fixtures(t *testing.T) {
	t.Parallel()
	destination := os.Getenv("HIR_UNIT3_GENERATE")
	if destination == "" {
		t.Skip("set HIR_UNIT3_GENERATE to regenerate complete Go fixtures")
	}
	root := unit3Root(t)
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	own := filepath.Join(lane, "passes/unit-3")
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	temp := t.TempDir()
	replacements := map[string]string{}
	for name, source := range map[string]string{"inputs": "replay/inputs_test.go", "checkpoint": "replay/oracle_test.go", "oracle": "testdata/oracle_test.go", "census": "testdata/census_test.go"} {
		replacements[filepath.Join(upstream, "stage1_hir_"+name+"_test.go")] = filepath.Join(lane, source)
	}
	for _, name := range []string{"checkpoint_adapter_test.go", "census_adapter_test.go"} {
		replacements[filepath.Join(upstream, "stage1_unit3_"+name)] = filepath.Join(own, "testdata", name)
	}
	files, err := filepath.Glob(filepath.Join(upstream, "*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		copy := filepath.Join(temp, filepath.Base(file))
		if err := os.WriteFile(copy, observeCalls(data), 0600); err != nil {
			t.Fatal(err)
		}
		replacements[file] = copy
	}
	provider, err := os.ReadFile(filepath.Join(root, "bridge/tsgo/checker/symbol_graph.go"))
	if err != nil {
		t.Fatal(err)
	}
	providerPath := filepath.Join(temp, "symbol_graph_test.go")
	if err := os.WriteFile(providerPath, []byte(strings.Replace(string(provider), "package checker", "package high_level_intermediate_representation", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	replacements[filepath.Join(upstream, "stage1_hir_symbol_graph_test.go")] = providerPath
	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("go", "test", "-v", "-count=1", "-timeout=3h", "-tags=lintoracle", "-overlay", overlay, "./internal/lint/ecmascript/high_level_intermediate_representation")
	c.Dir = filepath.Join(root, "cohere")
	c.Env = append(os.Environ(), "GOWORK="+filepath.Join(root, "cohere/go.work"), "HIR_CENSUS="+destination, "HIR_UNIT3_EXPORT="+filepath.Join(destination, "passes"), "HIR_CORPUS="+filepath.Join(lane, "testdata/corpus.txt"), "HIR_OUTPUT="+filepath.Join(temp, "small.dump"))
	output, runErr := c.CombinedOutput()
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "go-tests.log"), output, 0600); err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("Go census: %v; %s", runErr, filepath.Join(destination, "go-tests.log"))
	}
	manifest, err := os.ReadFile(filepath.Join(destination, "checkpoint-manifest.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	records := []fixture{}
	originals, probes, flow := 0, 0, 0
	for _, line := range strings.Split(string(manifest), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		probe := fields[3] == "true"
		source, err := os.ReadFile(filepath.Join(destination, fields[0]+".ts"))
		if err != nil {
			for _, ext := range []string{".tsx", ".js", ".jsx"} {
				source, err = os.ReadFile(filepath.Join(destination, fields[0]+ext))
				if err == nil {
					break
				}
			}
		}
		if err != nil {
			t.Fatal(err)
		}
		// Flow admission is recorded by the owner's manifest, not reclassified here.
		isFlow := false
		ownerManifest, err := os.ReadFile(filepath.Join(destination, "manifest.tsv"))
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range strings.Split(string(ownerManifest), "\n") {
			f := strings.Split(row, "\t")
			if len(f) > 9 && f[0] == fields[0] {
				isFlow = f[9] == "Flow"
				break
			}
		}
		_ = source
		if probe {
			probes++
		} else {
			originals++
			if isFlow {
				flow++
			}
		}
		labels := append([]string{}, passes...)
		children, err := os.ReadDir(filepath.Join(destination, "passes", fields[0]))
		if err != nil {
			t.Fatal(err)
		}
		for _, child := range children {
			if strings.HasPrefix(child.Name(), "inline_remap:") {
				labels = append(labels, child.Name())
			}
		}
		for _, pass := range labels {
			folder := filepath.Join(destination, "passes", fields[0], pass)
			before, err := os.ReadFile(filepath.Join(folder, "before.checkpoint"))
			if err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(filepath.Join(folder, "after.checkpoint"))
			if err != nil {
				t.Fatal(err)
			}
			records = append(records, fixture{fields[0], pass, string(before), string(after), probe, isFlow})
		}
	}
	if originals != 1465 || probes != 72 || flow != 23 {
		t.Fatalf("changed census: originals=%d probes=%d Flow=%d", originals, probes, flow)
	}
	file, err := os.Create(filepath.Join(own, "testdata", "fixtures.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	zip := gzip.NewWriter(file)
	if err := json.NewEncoder(zip).Encode(records); err != nil {
		t.Fatal(err)
	}
	if err := zip.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("exported %d original and %d probe records, each with %d independent subpass checkpoints; %d Flow graphs retained", originals, probes, len(passes), flow)
}
func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	f, err := os.Open("testdata/fixtures.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zip, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer zip.Close()
	var records []fixture
	if err := json.NewDecoder(zip).Decode(&records); err != nil {
		t.Fatal(err)
	}
	return records
}
func prepareFixtures(t *testing.T, records []fixture, onlyPass string) (string, map[string]string) {
	t.Helper()
	temp := t.TempDir()
	var manifest strings.Builder
	expected := map[string]string{}
	for i, r := range records {
		if onlyPass != "" && r.Pass != onlyPass {
			continue
		}
		path := filepath.Join(temp, fmt.Sprintf("%d.checkpoint", i))
		if err := os.WriteFile(path, []byte(r.Before), 0600); err != nil {
			t.Fatal(err)
		}
		key := r.Key + "/" + r.Pass
		fmt.Fprintf(&manifest, "%s\t%s\n", key, path)
		expected[key] = r.After
	}
	path := filepath.Join(temp, "manifest.tsv")
	if err := os.WriteFile(path, []byte(manifest.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return path, expected
}
func compareFixtures(t *testing.T, output []byte, expected map[string]string) {
	t.Helper()
	seen := map[string]bool{}
	for _, part := range strings.Split(string(output), "checkpoint\t")[1:] {
		key, body, ok := strings.Cut(part, "\n")
		if !ok || seen[key] {
			t.Fatal("malformed/duplicate output")
		}
		seen[key] = true
		want, exists := expected[key]
		if !exists {
			t.Fatal("extra case " + key)
		}
		if body != want {
			gotLines, wantLines := strings.Split(body, "\n"), strings.Split(want, "\n")
			for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
				if gotLines[i] != wantLines[i] {
					t.Fatalf("%s line %d\ngot %s\nwant %s", key, i+1, gotLines[i], wantLines[i])
				}
			}
			t.Fatalf("%s differs in length: %d/%d", key, len(body), len(want))
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("missing outputs: %d/%d", len(seen), len(expected))
	}
}

// Reports all raw byte comparisons. The two inliner certificates remain stopped
// on independently proven Go allocation-order nondeterminism; no IDs are renamed.
func TestNodeSubpasses(t *testing.T) {
	t.Parallel()
	root := unit3Root(t)
	records := loadFixtures(t)
	filter := os.Getenv("HIR_UNIT3_PASS")
	manifest, expected := prepareFixtures(t, records, filter)
	entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-3/main.ts")
	output := unit3Command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", entry, "--census", manifest)
	actual := map[string]string{}
	for _, part := range strings.Split(string(output), "checkpoint\t")[1:] {
		key, body, ok := strings.Cut(part, "\n")
		if !ok {
			t.Fatal("malformed output")
		}
		if _, yes := actual[key]; yes {
			t.Fatal("duplicate output")
		}
		actual[key] = body
	}
	if len(actual) != len(expected) {
		t.Fatalf("missing/extra outputs: %d/%d", len(actual), len(expected))
	}
	type counts struct {
		OriginalTotal, OriginalMatched, ProbeTotal, ProbeMatched, FlowTotal, FlowMatched int
		NativeMatched, JavaScriptMatched                                                 int
		FirstMismatch                                                                    string
	}
	receipt := map[string]*counts{}
	for _, r := range records {
		if filter != "" && r.Pass != filter {
			continue
		}
		pass := strings.Split(r.Pass, ":")[0]
		c := receipt[pass]
		if c == nil {
			c = &counts{}
			receipt[pass] = c
		}
		key := r.Key + "/" + r.Pass
		body, exists := actual[key]
		if !exists {
			t.Fatal("missing " + key)
		}
		same := body == r.After
		if r.Probe {
			c.ProbeTotal++
			if same {
				c.ProbeMatched++
			}
		} else {
			c.OriginalTotal++
			if same {
				c.OriginalMatched++
			}
			if r.Flow {
				c.FlowTotal++
				if same {
					c.FlowMatched++
				}
			}
		}
		if !same && c.FirstMismatch == "" {
			c.FirstMismatch = key
		}
	}
	for _, pass := range passes {
		c := receipt[pass]
		if c == nil {
			continue
		}
		t.Logf("%s: Node originals %d/%d, probes %d/%d, Flow %d/%d; native 0; emitted JS 0; first mismatch %s", pass, c.OriginalMatched, c.OriginalTotal, c.ProbeMatched, c.ProbeTotal, c.FlowMatched, c.FlowTotal, c.FirstMismatch)
		if pass != "inline_iife" && pass != "inline_iife_including_memo_callbacks" && (c.OriginalMatched != c.OriginalTotal || c.ProbeMatched != c.ProbeTotal) {
			for _, r := range records {
				if r.Key+"/"+r.Pass == c.FirstMismatch {
					got := actual[c.FirstMismatch]
					compareFixtures(t, []byte("checkpoint\t"+c.FirstMismatch+"\n"+got), map[string]string{c.FirstMismatch: r.After})
				}
			}
		}
	}
	if path := os.Getenv("HIR_UNIT3_RECEIPT"); path != "" {
		data, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if os.Getenv("HIR_UNIT3_REQUIRE_CERTIFICATE") == "1" {
		for pass, c := range receipt {
			if c.OriginalMatched != c.OriginalTotal || c.ProbeMatched != c.ProbeTotal {
				t.Errorf("%s raw-byte certificate is stopped", pass)
			}
		}
	}
}

func TestNativeNominalCallbackGap(t *testing.T) {
	t.Parallel()
	root := unit3Root(t)
	entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-3/testdata/nominal_callback_gap.a")
	node := unit3Command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", entry)
	if !bytes.Equal(node, []byte("1\n")) {
		t.Fatalf("Node repro: %s", node)
	}
	c := exec.Command("go", "run", "./cmd/adamic", "build", entry, "-o", filepath.Join(t.TempDir(), "gap"))
	c.Dir = root
	output, err := c.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "value without nominal ancestry seen as Block") || !strings.Contains(string(output), "adamic/nominal-class") {
		t.Fatalf("gap changed: %v\n%s", err, output)
	}
	t.Log("Node succeeds; native refuses constructed nominal return in generic graph callback")
}

func TestGoInlineOracleOrderGap(t *testing.T) {
	t.Parallel()
	root := unit3Root(t)
	own := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-3")
	target := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	replacements := map[string]string{filepath.Join(target, "stage1_unit3_order_test.go"): filepath.Join(own, "testdata/oracle_order_gap_test.go"), filepath.Join(target, "stage1_unit3_dump_test.go"): filepath.Join(own, "../../testdata/oracle_test.go")}
	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	output := unit3Command(t, filepath.Join(root, "cohere"), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work")}, "go", "test", "-v", "-count=1", "-timeout=3h", "-tags=lintoracle", "-overlay", overlay, "-run=^TestUnit3GoInlineOrderGap$", "./internal/lint/ecmascript/high_level_intermediate_representation")
	t.Logf("%s", output)
}

func TestNativeRecursiveInitializerGap(t *testing.T) {
	t.Parallel()
	root := unit3Root(t)
	entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-3/testdata/recursive_initializer_gap.a")
	node := unit3Command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", entry)
	if !bytes.Equal(node, []byte("0\n")) {
		t.Fatalf("Node repro: %s", node)
	}
	c := exec.Command("go", "run", "./cmd/adamic", "build", entry, "-o", filepath.Join(t.TempDir(), "gap"))
	c.Dir = root
	output, err := c.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "stage 0 can't lower a function value that captures the variable its own initializer declares yet") {
		t.Fatalf("gap changed: %v\n%s", err, output)
	}
	t.Log("Node succeeds; native refuses a local recursive closure initializer")
}
