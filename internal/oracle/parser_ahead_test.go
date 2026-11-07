package oracle

import (
	"context"
	"os"
	"reflect"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

// The original scout's base factory preserves one allocation through its checked
// Identifier view; reservation permits later writes without synthesizing own keys.
func TestParserAheadFactoryBoundariesComplete(test *testing.T) {
	test.Parallel()
	assertParserAheadFactoryBoundary(test, "complete", "ready\n")
}

func TestParserAheadFactoryBoundariesEscaped(test *testing.T) {
	test.Parallel()
	assertParserAheadFactoryBoundary(test, "escaped", "undefined\n")
}

func assertParserAheadFactoryBoundary(test *testing.T, name, stdout string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
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
	test.Log("base factory source matches both backends; delivered placeholder machinery")
}

func TestParserAheadFactoryCompleteControl(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "stage3/parser-ahead/factories/already-complete.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\n" || len(truth.stderr) != 0 {
		test.Fatalf("source Node %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if loweredParserError != nil {
		test.Fatal(loweredParserError)
	}
	actual, binary := nativelyUncached(test, program)
	for _, got := range []run{actual, onSliceJavaScript(test, program)} {
		if d := disagreement(truth, got); d != "" {
			test.Fatal(d)
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	changed := 0
	for index, text := range program.Strings {
		if text == "ready" {
			program.Strings[index] = "wrong"
			changed++
		}
	}
	if changed != 1 {
		test.Fatalf("payload mutant changed %d constants", changed)
	}
	mutant, _ := nativelyUncached(test, program)
	for _, got := range []run{mutant, onSliceJavaScript(test, program)} {
		if got.exitCode != 0 || string(got.stdout) != "wrong\n" {
			test.Fatalf("mutant must execute normally: %+v", got)
		}
		if disagreement(truth, got) == "" {
			test.Fatal("factory payload mutant survived")
		}
	}
	test.Log("factory payload mutant caught in native and JavaScript by source Node stdout")
}

func TestParserAheadNodeArraySourceFields(test *testing.T) {
	test.Parallel()
	assertParserAheadNodeArraySource(test, "fields", "-1|7|true|1|1|1\ntrue\n", "a cast the runtime can't check")
}

func TestParserAheadNodeArraySourcePresence(test *testing.T) {
	test.Parallel()
	assertParserAheadNodeArraySource(test, "presence", "false|true\ntrue|true\ntrue|ready\n", "an unproven relation")
}

func assertParserAheadNodeArraySource(test *testing.T, name, stdout, reason string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/node-array", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
		test.Fatalf("source Node %+v", truth)
	}
	program, loweredParserError := loweredParser(test, path)
	if name == "fields" && loweredParserError == nil {
		actual, binary := nativelyUncached(test, program)
		for _, got := range []run{actual, onSliceJavaScript(test, program)} {
			if d := disagreement(truth, got); d != "" {
				test.Fatal(d)
			}
		}
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatal(report)
		}
		test.Log("all four NodeArray metadata fields match source Node; delivered placeholder machinery")
		return
	}
	if loweredParserError == nil || !strings.Contains(loweredParserError.Error(), reason) {
		test.Fatalf("pinned optional-only array boundary changed: %v", loweredParserError)
	}
	test.Log("optional-only array relation remains refused; this boundary test does not certify acceptance")
}

// Display source positions relative to the repository, so diagnostics keep no machine path.
func parserDisplayedPath(test *testing.T, path string) string {
	test.Helper()
	if !filepath.IsAbs(path) {
		return path
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	relative, relativePathError := filepath.Rel(root, path)
	if relativePathError != nil {
		test.Fatal(relativePathError)
	}
	return filepath.ToSlash(relative)
}

func loweredParser(test *testing.T, path string) (*ir.Program, error) {
	test.Helper()
	program, loadDisplayedError := load.LoadDisplayed(
		[]string{path},
		func(name string) string { return parserDisplayedPath(test, name) },
	)
	if loadDisplayedError != nil {
		test.Fatal(loadDisplayedError)
	}
	return lower.Lower(context.Background(), program)
}

func TestParserPlaceholderProductsUseDisplayedPaths(test *testing.T) {
	test.Parallel()
	assertParserProductsUseDisplayedPaths(test, "internal/oracle/testdata/placeholder_nonnull_saved_leak.a")
}

func TestParserFactoryProductsUseDisplayedPaths(test *testing.T) {
	test.Parallel()
	assertParserProductsUseDisplayedPaths(test, "stage3/parser-ahead/rulings/factory-use.a")
}

func TestParserSpeculationProductsUseDisplayedPaths(test *testing.T) {
	test.Parallel()
	assertParserProductsUseDisplayedPaths(test, "stage3/parser-next/speculation/misfit.a")
}

func assertParserProductsUseDisplayedPaths(test *testing.T, fixture string) {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join(repository, fixture))
	if readError != nil {
		test.Fatal(readError)
	}
	var previous *ir.Program
	for range 2 {
		path := filepath.Join(test.TempDir(), "main.a")
		if writeError := os.WriteFile(path, source, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		checked, loadDisplayedError := load.LoadDisplayed([]string{path}, filepath.Base)
		if loadDisplayedError != nil {
			test.Fatal(loadDisplayedError)
		}
		program, lowerError := lower.Lower(context.Background(), checked)
		if lowerError != nil {
			test.Fatal(lowerError)
		}
		if previous != nil {
			if native.C(previous) != native.C(program) ||
				javascript.JavaScript(previous) != javascript.JavaScript(program) ||
				!reflect.DeepEqual(previous.PlaceholderSources, program.PlaceholderSources) ||
				!reflect.DeepEqual(previous.PlaceholderViews, program.PlaceholderViews) ||
				!reflect.DeepEqual(previous.PlaceholderChecks, program.PlaceholderChecks) {
				test.Fatal("parser product changed with its source directory")
			}
		}
		previous = program
	}
}

// Dedicated slice checks observe Node directly without preparing the result cache's three runtimes.
func onSliceNode(test *testing.T, path string) run {
	test.Helper()
	return execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "node.mjs"),
		path,
	)
}

func onSliceJavaScript(test *testing.T, program *ir.Program) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if writeError := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return onSliceNode(test, path)
}
