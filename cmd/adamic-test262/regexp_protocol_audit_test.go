package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A read-only adaptation export for the independent stock TypeScript audit.
// It does not change classification, verdicts, or the runner's guard.
func TestRegExpProtocolTypeScriptAudit(t *testing.T) {
	corpus, output := os.Getenv("ADAMIC_REGEXP_TSC_CORPUS"), os.Getenv("ADAMIC_REGEXP_TSC_SOURCES")
	rejections := os.Getenv("ADAMIC_REGEXP_TSC_REJECTIONS")
	if corpus == "" || output == "" || rejections == "" {
		t.Skip("explicit corpus, external rejection manifest, and scratch output required")
	}
	data, err := os.ReadFile(rejections)
	if err != nil {
		t.Fatal(err)
	}
	type row struct{ Path, Adamic, Program string }
	rows := []row{}
	for _, line := range strings.Split(string(data), "\n")[1:] {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		source, err := os.ReadFile(filepath.Join(corpus, "test", parts[0]))
		if err != nil {
			t.Fatal(err)
		}
		one := classify(parts[0], string(source), true)
		if one.Skip != "" {
			t.Fatalf("baseline diagnostic unexpectedly skipped: %s: %s", parts[0], one.Skip)
		}
		rows = append(rows, row{parts[0], parts[1], one.Program})
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("exported %d adapted diagnostic programs", len(rows))
}
