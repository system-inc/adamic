package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeDeclarationsArePinned(t *testing.T) {
	t.Parallel()
	api := t.TempDir()
	dir := filepath.Join(api, "node_modules/@types/node")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.d.ts"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(dir, "package.json")
	if err := os.WriteFile(metadata, []byte(`{"name":"@types/node","version":"25.3.3"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNodeTypes(api); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadata, []byte(`{"name":"@types/node","version":"24.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyNodeTypes(api); err == nil || !strings.Contains(err.Error(), "want @types/node") {
		t.Fatal("wrong declarations accepted", err)
	}
}
