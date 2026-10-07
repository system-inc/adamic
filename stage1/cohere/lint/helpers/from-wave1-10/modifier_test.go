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

func modifierInputs(t *testing.T) string {
	t.Helper()
	source, err := os.Open("testdata/modifier-cases.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var cases []string
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	var coverage []struct {
		Rule     string
		Files    []string
		Literals int
	}
	ledger, err := os.ReadFile("testdata/modifier-coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(ledger, &coverage); err != nil {
		t.Fatal(err)
	}
	var readiness struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	ledger, err = os.ReadFile("../readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(ledger, &readiness); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	for _, row := range readiness.Remaining {
		for _, symbol := range row.Helpers {
			if symbol == "github.com/system-inc/cohere/internal/lint/ecmascript/regexp.modifierGroup" {
				want[row.Rule] = true
			}
		}
	}
	for _, row := range coverage {
		if !want[row.Rule] || row.Literals == 0 {
			t.Fatalf("bad consumer %+v", row)
		}
		delete(want, row.Rule)
	}
	if len(want) != 0 || len(coverage) != 4 {
		t.Fatalf("missing consumers %v", want)
	}
	t.Logf("%d sources / %d queries, all four consumer fixture families", len(cases), len(cases)*16)
	path := filepath.Join(t.TempDir(), "cases.json")
	write(t, path, data)
	return path
}
func modifierOracle(t *testing.T) string {
	t.Helper()
	root := absolute(t, "../../../../../cohere")
	virtual := filepath.Join(root, "adamic_modifier_group.go")
	replacement := filepath.Join(root, "internal/lint/ecmascript/regexp/adamic_modifier_group.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: absolute(t, "testdata/modifier_oracle.go.txt"), replacement: absolute(t, "testdata/modifier_exports.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(directory, "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func TestModifierGroup(t *testing.T) {
	path := modifierInputs(t)
	want := run(t, "", modifierOracle(t), path)
	for i, command := range commands(t, absolute(t, "modifier_main.a"), path) {
		got := run(t, "", command[0], command[1:]...)
		if !bytes.Equal(got, want) {
			t.Fatalf("runtime %d differs: got %d bytes Go %d", i, len(got), len(want))
		}
		t.Logf("runtime %d matched %d bytes", i, len(got))
	}
}
func TestModifierGroupMutant(t *testing.T) {
	path := modifierInputs(t)
	want := run(t, "", modifierOracle(t), path)
	directory := t.TempDir()
	main, err := os.ReadFile("modifier_main.a")
	if err != nil {
		t.Fatal(err)
	}
	main = []byte(strings.ReplaceAll(string(main), "../options_json.ts", filepath.ToSlash(absolute(t, "../options_json.ts"))))
	write(t, filepath.Join(directory, "modifier_main.a"), main)
	source, err := os.ReadFile("regexp_modifier_group.a")
	if err != nil {
		t.Fatal(err)
	}
	anchor := "if((named & bit) !== 0)"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("mutant anchor drift")
	}
	write(t, filepath.Join(directory, "regexp_modifier_group.a"), []byte(strings.Replace(string(source), anchor, "if(false)", 1)))
	for i, command := range commands(t, filepath.Join(directory, "modifier_main.a"), path) {
		got := run(t, "", command[0], command[1:]...)
		if bytes.Equal(got, want) {
			t.Fatalf("runtime %d duplicate_guard_disabled survived", i)
		}
		t.Logf("runtime %d duplicate_guard_disabled compiled, exited with empty stderr, caught only by Go byte comparison", i)
	}
}
