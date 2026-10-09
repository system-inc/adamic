package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestRuntimeFeaturesMatchNode(t *testing.T) {
	fixture := "../../../internal/oracle/testdata/closure_convention_nested.a"
	program, err := load.Load([]string{fixture})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(native.C(lowered), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		t.Fatal("fixture does not enable closure convention")
	}
	binary := filepath.Join(t.TempDir(), "program")
	previous := os.Args
	defer func() { os.Args = previous }()
	os.Args = []string{"decode-ascii", "../../../internal/native/runtime", fixture, "native", binary}
	main()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	actual, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v: %s", err, actual)
	}
	expected, err := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", "../../../oracle/node.mjs", fixture).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v: %s", err, expected)
	}
	if string(actual) != string(expected) {
		t.Fatalf("native %q != Node %q", actual, expected)
	}
	t.Logf("Node/native output: %q", actual)
}
