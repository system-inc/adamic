package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// These source programs exercise the delivered placeholder machinery through both backends.
func TestParserConstructionSourceFactoryUnset(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "factory-unset")
}

func TestParserConstructionSourceOwnKeyOrder(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "own-key-order")
}

func TestParserConstructionSourceNodeArrayKeys(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "node-array-keys")
}

func TestParserConstructionSourceNodeArrayJson(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "node-array-json")
}

func TestParserConstructionSourceNodeArrayLength(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "node-array-length")
}

func TestParserConstructionSourceNodeArrayUnset(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "node-array-unset")
}

// Sixteen constant records in a node array, enough for constantRecords' run of eight, still build a
// node array: the metadata makes fieldTypes and dynamicProperties true, which keep constantRecords
// out whichever check comes first. Mutant: both guards removed, and native stops with a TypeError.
func TestParserConstructionSourceNodeArrayRecords(test *testing.T) {
	test.Parallel()
	assertParserConstructionSource(test, "node-array-records")
}

func assertParserConstructionSource(test *testing.T, name string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/rulings", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	actual, binary := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if d := disagreement(truth, got); d != "" {
			test.Fatalf("%s: %s", backend, d)
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log("source Node, native sanitizers, leak check and JavaScript agree; delivered placeholder machinery")
}

func TestParserConstructionUnsetUse(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/rulings/factory-use.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onSliceNode(test, path)
	if truth.exitCode != 70 ||
		string(truth.stdout) != "true\n" ||
		!strings.Contains(string(truth.stderr), "TypeError") {
		test.Fatalf("Node %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	message := "unset use failed: field 'text', factory 'createBaseNode', use 'node.text' at " +
		parserDisplayedPath(test, path) +
		":5:13; expected string"
	want := run{stdout: []byte("true\n"), stderr: []byte("adamic: panic: " + message + "\n"), exitCode: 70}
	actual, _ := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if d := disagreement(want, got); d != "" {
			test.Fatalf("%s: %s", backend, d)
		}
	}
	test.Log("undefined presence observation is accepted; declared string use stops naming field, factory and use")
}

// The delivered placeholder machinery retains null independently of factory unset.
func TestParserConstructionNullPlaceholderDelivered(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "stage3/parser-ahead/rulings/null-placeholder-pending.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "true\ntrue\n" || len(truth.stderr) != 0 {
		test.Fatalf("Node %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	actual, binary := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if d := disagreement(truth, got); d != "" {
			test.Fatalf("%s: %s", backend, d)
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

func TestParserProgramRegionLifetime(test *testing.T) {
	test.Parallel()
	test.Skip("awaits compiler/program-region-lowering: parser trees in the Program region")
}
