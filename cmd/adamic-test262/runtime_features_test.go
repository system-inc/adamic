package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeFeaturesMatchNode(t *testing.T) {
	e, err := prepareMode("../..", "testdata/mini", t.TempDir(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	e.cache = nil
	e.compiler = &compilerWorker{}
	defer e.compiler.close()
	source, err := os.ReadFile("../../internal/oracle/testdata/closure_convention_plain.a")
	if err != nil {
		t.Fatal(err)
	}
	got := e.attempt(classified{Path: "closure_convention_plain.a", Program: string(source)})
	emitted, err := os.ReadFile(filepath.Join(e.work, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(emitted), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		t.Fatal("fixture does not enable closure convention")
	}
	if got.Kind != outcomePass {
		t.Fatalf("Node/native comparison: %+v", got)
	}
	t.Log("Node/native output agrees for closure_convention_plain.a")
}
