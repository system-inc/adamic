package wave10

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func foldInputs(t *testing.T) string {
	t.Helper()
	file, err := os.Open("testdata/fold-cases.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := os.ReadFile("testdata/fold-coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var coverage struct {
		Go, Unicode string
		Transitions int
		Consumers   []struct {
			Rule            string
			Literals, Runes int
		}
	}
	if err = json.Unmarshal(metadata, &coverage); err != nil {
		t.Fatal(err)
	}
	ledger, err := os.ReadFile("../readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	var readiness struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if err = json.Unmarshal(ledger, &readiness); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	for _, row := range readiness.Remaining {
		for _, symbol := range row.Helpers {
			if symbol == "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.simpleFold" {
				want[row.Rule] = true
			}
		}
	}
	runes := 0
	for _, row := range coverage.Consumers {
		if !want[row.Rule] || row.Literals == 0 || row.Runes == 0 {
			t.Fatalf("bad consumer %+v", row)
		}
		delete(want, row.Rule)
		runes += row.Runes
	}
	if len(want) != 0 || len(coverage.Consumers) != 4 {
		t.Fatalf("missing consumers %v", want)
	}
	t.Logf("%d queries, four consumer families, %s Unicode %s, %d transitions", 1114113+6+runes, coverage.Go, coverage.Unicode, coverage.Transitions)
	path := filepath.Join(t.TempDir(), "cases.json")
	write(t, path, data)
	return path
}
func foldOracle(t *testing.T) string {
	t.Helper()
	root := absolute(t, "../../../../../cohere")
	virtual := filepath.Join(root, "adamic_simple_fold.go")
	export := filepath.Join(root, "internal/lint/ecmascript/regexp/adamic_simple_fold.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: absolute(t, "testdata/fold_oracle.go.txt"), export: absolute(t, "testdata/fold_exports.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	overlay := filepath.Join(directory, "overlay.json")
	write(t, overlay, data)
	binary := filepath.Join(directory, "go-oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func TestSimpleFold(t *testing.T) {
	path := foldInputs(t)
	want := run(t, "", foldOracle(t), path)
	for i, command := range commands(t, absolute(t, "fold_main.a"), path) {
		got := run(t, "", command[0], command[1:]...)
		if !bytes.Equal(got, want) {
			t.Fatalf("runtime %d differs: got %d bytes Go %d", i, len(got), len(want))
		}
		t.Logf("runtime %d matched %d bytes", i, len(got))
	}
}
func TestSimpleFoldMutant(t *testing.T) {
	path := foldInputs(t)
	want := run(t, "", foldOracle(t), path)
	directory := t.TempDir()
	main, err := os.ReadFile("fold_main.a")
	if err != nil {
		t.Fatal(err)
	}
	main = []byte(strings.ReplaceAll(string(main), "../options_json.ts", filepath.ToSlash(absolute(t, "../options_json.ts"))))
	write(t, filepath.Join(directory, "fold_main.a"), main)
	source, err := os.ReadFile("regexp_simple_fold.a")
	if err != nil {
		t.Fatal(err)
	}
	anchor := "if(folded < least)"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("mutant anchor drift")
	}
	write(t, filepath.Join(directory, "regexp_simple_fold.a"), []byte(strings.Replace(string(source), anchor, "if(folded > least)", 1)))
	for i, command := range commands(t, filepath.Join(directory, "fold_main.a"), path) {
		got := run(t, "", command[0], command[1:]...)
		if bytes.Equal(got, want) {
			t.Fatalf("runtime %d maximum_fold_selected survived", i)
		}
		t.Logf("runtime %d maximum_fold_selected compiled, exited with empty stderr, caught only by Go byte comparison", i)
	}
}
