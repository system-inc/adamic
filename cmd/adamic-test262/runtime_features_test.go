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
	engine, prepareModeError := prepareMode("../..", "testdata/mini", test.TempDir(), nil, nil, true)
	if prepareModeError != nil {
		test.Fatal(prepareModeError)
	}
	engine.cache = nil
	engine.compiler = &compilerWorker{}
	defer engine.compiler.close()
	source, readError := os.ReadFile("../../internal/oracle/testdata/arguments_length_value_count.a")
	if readError != nil {
		test.Fatal(readError)
	}
	got := engine.attempt(classified{Path: "arguments_length_value_count.a", Program: string(source)})
	emitted, readError := os.ReadFile(filepath.Join(engine.work, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.Contains(string(emitted), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		test.Fatal("fixture does not enable closure convention")
	}
	if got.Kind != outcomePass {
		test.Fatalf("Node/native comparison: %+v", got)
	}
	test.Log("Node/native output agrees for arguments_length_value_count.a")
}
