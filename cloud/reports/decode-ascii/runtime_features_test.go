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

// Not parallel: main reads process-wide os.Args.
func TestRuntimeFeaturesMatchNode(test *testing.T) {
	fixture := "../../../internal/oracle/testdata/closure_convention_nested.a"
	program, loadError := load.Load([]string{fixture})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	if !strings.Contains(native.C(lowered), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		test.Fatal("fixture does not enable closure convention")
	}
	binary := filepath.Join(test.TempDir(), "program")
	previous := os.Args
	defer func() { os.Args = previous }()
	os.Args = []string{"decode-ascii", "../../../internal/native/runtime", fixture, "native", binary}
	main()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	actual, combinedOutputError := exec.CommandContext(timeoutContext, binary).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("native: %v: %s", combinedOutputError, actual)
	}
	expected, combinedOutputError := exec.CommandContext(
		timeoutContext,
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../../oracle/node.mjs",
		fixture,
	).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v: %s", combinedOutputError, expected)
	}
	if string(actual) != string(expected) {
		test.Fatalf("native %q != Node %q", actual, expected)
	}
	test.Logf("Node/native output: %q", actual)
}
