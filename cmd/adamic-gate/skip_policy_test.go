package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkipCensusHookRefusesLandedUnwiredChecker(t *testing.T) {
	root := t.TempDir()
	summary, err := checkSkipCensus(root)
	if err != nil || summary.Status != "not-landed" {
		t.Fatal(summary, err)
	}
	path := filepath.Join(root, "internal", "skipcensus", "census.go")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package skipcensus\n"), 0600); err != nil {
		t.Fatal(err)
	}
	summary, err = checkSkipCensus(root)
	if err == nil || summary.Status != "not-wired" || !strings.Contains(err.Error(), "CheckLog") {
		t.Fatal("landed checker silently bypassed", summary, err)
	}
}
