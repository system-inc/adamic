package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"AppKit"}, {"AppKit", "-o"}, {"AppKit", "-unknown"}} {
		if err := run(args); err == nil {
			t.Errorf("invalid command accepted: %v", args)
		}
	}
	directory := t.TempDir()
	root := filepath.Join("..", "..", "internal", "apple", "generate", "testdata")
	if err := run([]string{"Foundation", "AppKit", "-o", directory, "-headers", filepath.Join(root, "SDK"), "-umbrella", filepath.Join(root, "umbrella.h")}); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(filepath.Join(directory, "appkit", "fixture-panel.d.ts"))
	if err != nil {
		t.Fatal(err)
	}
	// Flipped is bit 63, read unsigned: the case a signed parse refused.
	if !strings.Contains(string(output), "1.style:options(Plain=0,Bright=2,Quiet=8,Flipped=9223372036854775808)") {
		t.Fatal("command did not emit the generated option binding")
	}
	if _, err := os.Stat(filepath.Join(directory, "bindings-check.m")); err != nil {
		t.Fatal("command did not emit the header witness", err)
	}
}
