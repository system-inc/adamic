package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeFeaturesMatchNode(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	e, err := prepareMode("../..", "testdata/mini", test.TempDir(), nil, nil, true)
	if err != nil {
		test.Fatal(err)
	}
	e.cache = nil
	e.compiler = &compilerWorker{}
	defer e.compiler.close()
	source, err := os.ReadFile("../../internal/oracle/testdata/arguments_length_value_count.a")
	if err != nil {
		test.Fatal(err)
	}
	got := e.attempt(classified{Path: "arguments_length_value_count.a", Program: string(source)})
	emitted, err := os.ReadFile(filepath.Join(e.work, "program.c"))
	if err != nil {
		test.Fatal(err)
	}
	if !strings.Contains(string(emitted), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		test.Fatal("fixture does not enable closure convention")
	}
	if got.Kind != outcomePass {
		test.Fatalf("Node/native comparison: %+v", got)
	}
	test.Log("Node/native output agrees for arguments_length_value_count.a")
}
