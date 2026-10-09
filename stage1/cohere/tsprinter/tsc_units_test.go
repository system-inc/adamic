package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func tscOracle(t *testing.T, root string) string {
	t.Helper()
	expressionSide, _ := filepath.Abs("testdata/expressions_side_test.go")
	statementSide, _ := filepath.Abs("testdata/statements_side_test.go")
	cohere := root + "/cohere"
	dir := t.TempDir()
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
		cohere + "/internal/format/javascript/adamic_expressions_test.go": expressionSide,
		cohere + "/internal/format/javascript/adamic_statements_test.go":  statementSide,
	}})
	if err != nil {
		t.Fatal(err)
	}
	path := dir + "/overlay.json"
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	// No hand-listed build inputs: overlay Go builds remain private.
	command := bounded(t, "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-o="+dir+"/oracle", "-overlay="+path, "./internal/format/javascript")
	command.Dir = cohere
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go TSC printer oracle: %v\n%s", err, output)
	}
	t.Logf("build Go TSC printer oracle cold wall %.3fs (overlay, uncached)", time.Since(start).Seconds())
	return dir + "/oracle"
}

type tsPrinterProducts struct{ source, backend, sanitized, release string }

func prepareTSPrinterProducts(t *testing.T, path, family string) tsPrinterProducts {
	t.Helper()
	loweredProduct := expressionBuild(t, printerBuildInputs{Name: "lowered TSC " + family, Files: expressionInputFiles(t, repository), Toolchain: runtime.Version()}, func(dir string) error {
		program := lowered(t, path)
		if err := os.WriteFile(dir+"/port.c", []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
	})
	data, err := os.ReadFile(loweredProduct + "/port.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	clang := executeOne(t, nil, "clang", "--version")
	if clang.exitCode != 0 {
		t.Fatalf("clang version: %s", clang.stderr)
	}
	inputs := append([]string{loweredProduct + "/port.c"}, expressionInputFiles(t, filepath.Join(repository, "internal/native"))...)
	sanitized := expressionBuild(t, printerBuildInputs{Name: "sanitized TSC " + family, Files: inputs, Flags: native.Flags(native.Options{Sanitize: true}), Toolchain: string(clang.stdout)}, func(dir string) error {
		return native.Build(source, dir+"/port", native.Options{Sanitize: true})
	}) + "/port"
	release := ""
	if runtime.GOOS == "darwin" {
		release = expressionBuild(t, printerBuildInputs{Name: "release TSC " + family, Files: inputs, Flags: native.Flags(native.Options{}), Toolchain: string(clang.stdout)}, func(dir string) error {
			return native.Build(source, dir+"/port", native.Options{})
		}) + "/port"
	}
	return tsPrinterProducts{source: path, backend: loweredProduct + "/program.mjs", sanitized: sanitized, release: release}
}

func tscPartition(t *testing.T, family string, cases []printerCase, count, offset int) ([]corpusShard, [][]byte, string) {
	t.Helper()
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	weights := make([]int, len(cases))
	var whole strings.Builder
	for i, item := range cases {
		weights[i] = len(item.Source)
		whole.WriteString("ok\t" + escape.Replace(item.Want) + "\n")
	}
	directory := t.TempDir()
	var shards []corpusShard
	var outputs [][]byte
	for number, indices := range assignCorpus(weights, count) {
		var input, want strings.Builder
		ids := make([]int, len(indices))
		for j, index := range indices {
			item := cases[index]
			input.WriteString(">" + escape.Replace(item.Source) + "\n")
			want.WriteString("ok\t" + escape.Replace(item.Want) + "\n")
			ids[j] = offset + index
		}
		path := filepath.Join(directory, fmt.Sprintf("%s-shard-%03d.txt", family, number))
		if err := os.WriteFile(path, []byte(input.String()), 0644); err != nil {
			t.Fatal(err)
		}
		shards = append(shards, corpusShard{indices: ids, text: path})
		outputs = append(outputs, []byte(want.String()))
	}
	return shards, outputs, whole.String()
}

func TestTSCShardPlantedDisagreement(t *testing.T) {
	cases := []printerCase{{Label: "case-0", Source: "1", Want: "1;\n"}, {Label: "case-1", Source: "2", Want: "2;\n"}, {Label: "case-2", Source: "3", Want: "3;\n"}, {Label: "case-3", Source: "4", Want: "4;\n"}}
	shards, outputs, whole := tscPartition(t, "expressions", cases, 2, 0)
	plan := &corpusPlan{labels: []string{"case-0", "case-1", "case-2", "case-3"}, shards: shards}
	if err := expressionUnion(plan, outputs, []byte(whole)); err != nil {
		t.Fatal(err)
	}
	port, _ := filepath.Abs("main.ts")
	results := make([]run, len(shards))
	for i, shard := range shards {
		results[i] = onNode(t, port, "--cases", shard.text, "80")
		if err := expressionDisagreement(plan, i, outputs[i], results[i]); err != nil {
			t.Fatal(err)
		}
	}
	outputs[0] = []byte(strings.Replace(string(outputs[0]), "ok\t1", "ok\tplanted disagreement", 1))
	caught := 0
	for i, result := range results {
		if err := expressionDisagreement(plan, i, outputs[i], result); err != nil {
			caught++
			if i != 0 || !strings.Contains(err.Error(), "shard-000 case-0") {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted disagreement caught: %v", err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}

func TestTSCShardUnionRejectsMissingAndRepeated(t *testing.T) {
	cases := []printerCase{{Source: "1", Want: "1"}, {Source: "2", Want: "2"}, {Source: "3", Want: "3"}, {Source: "4", Want: "4"}}
	shards, outputs, whole := tscPartition(t, "expressions", cases, 2, 0)
	plan := &corpusPlan{labels: []string{"0", "1", "2", "3"}, shards: shards}
	if err := expressionUnion(plan, outputs, []byte(whole)); err != nil {
		t.Fatal(err)
	}
	first := append([]int(nil), plan.shards[0].indices...)
	plan.shards[0].indices = first[:len(first)-1]
	if err := expressionUnion(plan, outputs, []byte(whole)); err == nil {
		t.Fatal("missing id accepted")
	}
	plan.shards[0].indices = append(first, first[0])
	if err := expressionUnion(plan, outputs, []byte(whole)); err == nil {
		t.Fatal("repeated id accepted")
	}
}
