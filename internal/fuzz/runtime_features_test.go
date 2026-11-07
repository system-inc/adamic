package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestRuntimeFeaturesMatchNode(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	checkout, prepareBuiltError := PrepareBuilt("../..", buildcache.Adamic(test), filepath.Join(directory, "checkout"))
	if prepareBuiltError != nil {
		test.Fatal(prepareBuiltError)
	}
	source, readError := os.ReadFile("../oracle/testdata/arguments_length_value_count.a")
	if readError != nil {
		test.Fatal(readError)
	}
	work := filepath.Join(directory, "program")
	outcome := checkout.Try(string(source), work)
	emitted, readError := os.ReadFile(filepath.Join(work, "main.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.Contains(string(emitted), "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		test.Fatal("fixture does not enable closure convention")
	}
	if outcome.Verdict != Agreed {
		test.Fatalf("%s: %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
	test.Logf("Node/native output: %q", outcome.Native.Stdout)
}
