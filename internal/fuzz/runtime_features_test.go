package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeFeaturesMatchNode(t *testing.T) {
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("../oracle/testdata/arguments_length_value_count.a")
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(directory, "program")
	outcome := checkout.Try(string(source), work)
	emitted, err := os.ReadFile(filepath.Join(work, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(emitted), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		t.Fatal("fixture does not enable closure convention")
	}
	if outcome.Verdict != Agreed {
		t.Fatalf("%s: %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
	t.Logf("Node/native output: %q", outcome.Native.Stdout)
}
