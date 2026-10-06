package estree

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryAgreement(t *testing.T) {
	directory := os.Getenv("ADAMIC_ESTREE_CORPUS")
	if directory == "" {
		t.Skip("set ADAMIC_ESTREE_CORPUS to a completed Go/Node corpus audit directory")
	}
	data, err := os.ReadFile(filepath.Join(directory, "port-records.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Path, Answer string
		PortStatus   string `json:"portStatus"`
		Bytes        int
	}
	var records []record
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var item record
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Fatal(err)
		}
		if item.PortStatus == "identical" {
			records = append(records, item)
		}
	}
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := build(t, main, true)
	total := 0
	// Bound memory by batching whole files, with all accepted records accounted for.
	for start := 0; start < len(records); start += 100 {
		end := start + 100
		if end > len(records) {
			end = len(records)
		}
		var list strings.Builder
		var want bytes.Buffer
		for _, item := range records[start:end] {
			answer, err := os.ReadFile(item.Answer)
			if err != nil {
				t.Fatal(err)
			}
			if len(answer) != item.Bytes {
				t.Fatal("Go answer byte count changed")
			}
			want.Write(answer)
			list.WriteString(item.Path + "\n")
			total += len(answer)
		}
		manifest := filepath.Join(t.TempDir(), "manifest")
		if err := os.WriteFile(manifest, []byte(list.String()), 0644); err != nil {
			t.Fatal(err)
		}
		for name, got := range map[string][]byte{"source Node": onNode(t, main, "--manifest", manifest), "sanitized native": execute(t, "", binary, "--manifest", manifest), "emitted JS": onNode(t, script, "--manifest", manifest)} {
			if diff := firstDifference(want.Bytes(), got); diff != "" {
				t.Fatalf("records %d..%d %s: %s", start, end, name, diff)
			}
		}
	}
	t.Logf("%d Go-accepted/source-Node-identical repository snapshot files, %d bytes checked on uninstrumented source Node, sanitized native and emitted JS; refusals and mismatches remain excluded and separately reported", len(records), total)
}
