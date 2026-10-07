package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareLocalProvenance(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	log := filepath.Join(root, "plain.jsonl")
	data := "{\"Action\":\"run\",\"Package\":\"p\",\"Test\":\"TestPass\"}\n{\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestPass\"}\n"
	if err := os.WriteFile(log, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	results, _, _, err := readLog(log)
	if err != nil {
		t.Fatal(err)
	}
	commit := strings.Repeat("a", 40)
	m := merged{Green: true, Plan: plan{Commit: commit, GoVersion: "go version go1.27.1 linux/amd64", NodeVersion: "v24.19.0"}, Results: results, Pass: 1, TestEvents: 1}
	path := filepath.Join(root, "merged.json")
	if err := saveJSON(path, m); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ name, commit, goVersion, node, want string }{
		{"matching", commit, m.Plan.GoVersion, m.Plan.NodeVersion, ""},
		{"other-commit", strings.Repeat("b", 40), m.Plan.GoVersion, m.Plan.NodeVersion, "plain commit"},
		{"other-node", commit, m.Plan.GoVersion, "v22.0.0", "plain Node version"},
		{"other-go", commit, "go version go1.26.0 linux/amd64", m.Plan.NodeVersion, "plain Go version"},
		{"missing-node", commit, m.Plan.GoVersion, "", "plain Node version"},
		{"missing-go", commit, "", m.Plan.NodeVersion, "plain Go version"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := saveJSON(log+".provenance.json", plainProvenance{test.commit, test.goVersion, test.node}); err != nil {
				t.Fatal(err)
			}
			err := compare(path, log)
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("want %q, got %v", test.want, err)
			}
		})
	}
	if err := os.Remove(log + ".provenance.json"); err != nil {
		t.Fatal(err)
	}
	if err := compare(path, log); err == nil {
		t.Fatal("missing provenance accepted")
	}
	notes := "commit=" + commit + "\ngo=" + m.Plan.GoVersion + "\nnode=" + m.Plan.NodeVersion + "\n"
	if err := os.WriteFile(filepath.Join(root, "run-notes.txt"), []byte(notes), 0600); err != nil {
		t.Fatal(err)
	}
	if err := compare(path, log); err != nil {
		t.Fatal(err)
	}
	m.BuildFlags = []string{`go="go version go1.27.1 linux/amd64" node="v24.19.0"`}
	m.Plan.GoVersion, m.Plan.NodeVersion = "", ""
	if err := saveJSON(path, m); err != nil {
		t.Fatal(err)
	}
	if err := compare(path, log); err != nil {
		t.Fatal(err)
	}
	m.BuildFlags = append(m.BuildFlags, `go="go version go1.26.0 linux/amd64" node="v24.19.0"`)
	if err := saveJSON(path, m); err != nil {
		t.Fatal(err)
	}
	if err := compare(path, log); err == nil || !strings.Contains(err.Error(), "versions differ") {
		t.Fatal("inconsistent shard provenance accepted", err)
	}
}
