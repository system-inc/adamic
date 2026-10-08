package unit4

import (
	"encoding/json"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func unit4ObserveTestCalls(data []byte) []byte {
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
			offset := file.Offset(pos)
			edits = append(edits, edit{offset, offset + len(text), "stage1Observed" + text})
		}
	}
	for index := len(edits) - 1; index >= 0; index-- {
		e := edits[index]
		data = append(append(append([]byte{}, data[:e.start]...), []byte(e.text)...), data[e.end:]...)
	}
	return data
}
func unit4ExportCensus(t *testing.T, destination string) string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	temp := t.TempDir()
	if destination == "" {
		destination = filepath.Join(temp, "census")
	}
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	replacements := map[string]string{
		filepath.Join(upstream, "stage1_hir_inputs_test.go"):     filepath.Join(lane, "replay/inputs_test.go"),
		filepath.Join(upstream, "stage1_hir_checkpoint_test.go"): filepath.Join(lane, "replay/oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_oracle_test.go"):     filepath.Join(lane, "testdata/oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_census_test.go"):     filepath.Join(lane, "testdata/census_test.go"),
	}
	replacements[filepath.Join(upstream, "stage1_hir_unit4_test.go")] = filepath.Join(lane, "passes/unit-4/oracle_test.go")
	files, err := filepath.Glob(filepath.Join(upstream, "*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		observed := unit4ObserveTestCalls(data)
		copy := filepath.Join(temp, filepath.Base(file))
		if err := os.WriteFile(copy, observed, 0600); err != nil {
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

	encoded, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlay, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("go", "test", "-v", "-count=1", "-timeout=15m", "-tags=lintoracle", "-overlay", overlay, "./internal/lint/ecmascript/high_level_intermediate_representation")
	c.Dir = filepath.Join(root, "cohere")
	c.Env = append(os.Environ(), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "HIR_CENSUS=" + destination, "HIR_UNIT4_EXPORT=" + filepath.Join(destination, "unit4"), "HIR_CORPUS=" + filepath.Join(lane, "testdata/corpus.txt"), "HIR_OUTPUT=" + filepath.Join(temp, "small.dump")}...)
	output, runErr := c.CombinedOutput()
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "go-tests.log"), output, 0600); err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("Go construction suite failed: %v; full log: %s", runErr, filepath.Join(destination, "go-tests.log"))
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "HIR construction census:") {
			t.Log(line)
		}
	}
	return destination
}

func unit4Run(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	command := exec.Command(args[0], args[1:]...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, output)
	}
	return output
}
func TestUnit4Census(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	destination := unit4ExportCensus(t, os.Getenv("HIR_UNIT4_CENSUS"))
	source, err := os.ReadFile(filepath.Join(destination, "checkpoint-manifest.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest strings.Builder
	counts := 0
	for _, pass := range []string{"primitive", "reactive", "effects", "effects-nested"} {
		manifest.Reset()
		counts = 0
		for _, line := range strings.Split(string(source), "\n") {
			if line == "" {
				continue
			}
			fields := strings.Split(line, "\t")
			if fields[3] == "true" {
				continue
			}
			counts++
			key := fields[0]
			manifest.WriteString(key + "\t" + filepath.Join(destination, "unit4", key+"."+pass+".before.checkpoint") + "\t" + filepath.Join(destination, "unit4", key+"."+pass+".after.checkpoint") + "\n")
		}
		if counts != 1465 {
			t.Fatalf("census count %d", counts)
		}
		manifestPath := filepath.Join(destination, "unit4-"+pass+"-manifest.tsv")
		if err := os.WriteFile(manifestPath, []byte(manifest.String()), 0600); err != nil {
			t.Fatal(err)
		}
		entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-4/main.a")
		if strings.HasPrefix(pass, "effects") {
			entry = filepath.Join(filepath.Dir(entry), "effects_main.a")
		}
		output := unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", entry, manifestPath)
		unit4Compare(t, output, manifest.String())
		t.Logf("%s Node: %d/1465", pass, counts)
		unit4SemanticMutants(t, root, pass, manifest.String(), manifestPath)
		if pass == "primitive" {
			unit4PrimitiveBackends(t, root, manifest.String(), manifestPath)
		}
	}
}
func unit4Compare(t *testing.T, output []byte, manifest string) {
	t.Helper()
	cases := map[string]string{}
	for _, chunk := range strings.Split(string(output), "checkpoint\t")[1:] {
		key, body, ok := strings.Cut(chunk, "\n")
		if !ok {
			t.Fatal("missing body")
		}
		if _, exists := cases[key]; exists {
			t.Fatal("duplicate case")
		}
		cases[key] = body
	}
	for _, line := range strings.Split(manifest, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		want, err := os.ReadFile(fields[2])
		if err != nil {
			t.Fatal(err)
		}
		if cases[fields[0]] != string(want) {
			t.Fatalf("%s differs from Go: after %s", fields[0], fields[2])
		}
		delete(cases, fields[0])
	}
	if len(cases) != 0 {
		t.Fatal("extra cases")
	}
}
