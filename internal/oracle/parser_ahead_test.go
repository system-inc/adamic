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
func TestParserAheadFactoryBoundariesComplete(t *testing.T) {
	t.Parallel()
	assertParserAheadFactoryBoundary(t, "complete", "ready\n")
}

func TestParserAheadFactoryBoundariesEscaped(t *testing.T) {
	t.Parallel()
	assertParserAheadFactoryBoundary(t, "escaped", "undefined\n")
}

func assertParserAheadFactoryBoundary(t *testing.T, name, stdout string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onSliceNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
		t.Fatalf("Node %+v", truth)
	}
	program, err := loweredParser(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(t, program)} {
		if d := disagreement(truth, got); d != "" {
			t.Fatalf("%s: %s", backend, d)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("base factory source matches both backends; delivered placeholder machinery")
}

func TestParserAheadFactoryCompleteControl(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories/already-complete.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onSliceNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node %+v", truth)
	}
	program, err := loweredParser(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := nativelyUncached(t, program)
	for _, got := range []run{actual, onSliceJavaScript(t, program)} {
		if d := disagreement(truth, got); d != "" {
			t.Fatal(d)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	changed := 0
	for i, s := range program.Strings {
		if s == "ready" {
			program.Strings[i] = "wrong"
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("payload mutant changed %d constants", changed)
	}
	mutant, _ := nativelyUncached(t, program)
	for _, got := range []run{mutant, onSliceJavaScript(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "wrong\n" {
			t.Fatalf("mutant must execute normally: %+v", got)
		}
		if disagreement(truth, got) == "" {
			t.Fatal("factory payload mutant survived")
		}
	}
	t.Log("factory payload mutant caught in native and JavaScript by source Node stdout")
}

func TestParserAheadNodeArraySourceFields(t *testing.T) {
	t.Parallel()
	assertParserAheadNodeArraySource(t, "fields", "-1|7|true|1|1|1\ntrue\n", "a cast the runtime can't check")
}

func TestParserAheadNodeArraySourcePresence(t *testing.T) {
	t.Parallel()
	assertParserAheadNodeArraySource(t, "presence", "false|true\ntrue|true\ntrue|ready\n", "an unproven relation")
}

func assertParserAheadNodeArraySource(t *testing.T, name, stdout, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/node-array", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onSliceNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
		t.Fatalf("source Node %+v", truth)
	}
	program, err := loweredParser(t, path)
	if name == "fields" && err == nil {
		actual, binary := nativelyUncached(t, program)
		for _, got := range []run{actual, onSliceJavaScript(t, program)} {
			if d := disagreement(truth, got); d != "" {
				t.Fatal(d)
			}
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
		t.Log("all four NodeArray metadata fields match source Node; delivered placeholder machinery")
		return
	}
	if err == nil || !strings.Contains(err.Error(), reason) {
		t.Fatalf("pinned optional-only array boundary changed: %v", err)
	}
	t.Log("optional-only array relation remains refused; this boundary test does not certify acceptance")
}

// Display source positions relative to the repository, so diagnostics keep no machine path.
func parserDisplayedPath(t *testing.T, path string) string {
	t.Helper()
	if !filepath.IsAbs(path) {
		return path
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(relative)
}

func loweredParser(t *testing.T, path string) (*ir.Program, error) {
	t.Helper()
	program, err := load.LoadDisplayed([]string{path}, func(name string) string { return parserDisplayedPath(t, name) })
	if err != nil {
		t.Fatal(err)
	}
	return lower.Lower(context.Background(), program)
}

func TestParserPlaceholderProductsUseDisplayedPaths(t *testing.T) {
	t.Parallel()
	assertParserProductsUseDisplayedPaths(t, "internal/oracle/testdata/placeholder_nonnull_saved_leak.a")
}

func TestParserFactoryProductsUseDisplayedPaths(t *testing.T) {
	t.Parallel()
	assertParserProductsUseDisplayedPaths(t, "stage3/parser-ahead/rulings/factory-use.a")
}

func TestParserSpeculationProductsUseDisplayedPaths(t *testing.T) {
	t.Parallel()
	assertParserProductsUseDisplayedPaths(t, "stage3/parser-next/speculation/misfit.a")
}

func assertParserProductsUseDisplayedPaths(t *testing.T, fixture string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	var previous *ir.Program
	for range 2 {
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, source, 0600); err != nil {
			t.Fatal(err)
		}
		checked, err := load.LoadDisplayed([]string{path}, filepath.Base)
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), checked)
		if err != nil {
			t.Fatal(err)
		}
		if previous != nil {
			if native.C(previous) != native.C(program) || javascript.JavaScript(previous) != javascript.JavaScript(program) ||
				!reflect.DeepEqual(previous.PlaceholderSources, program.PlaceholderSources) ||
				!reflect.DeepEqual(previous.PlaceholderViews, program.PlaceholderViews) ||
				!reflect.DeepEqual(previous.PlaceholderChecks, program.PlaceholderChecks) {
				t.Fatal("parser product changed with its source directory")
			}
		}
		previous = program
	}
}

// Dedicated slice checks observe Node directly without preparing the result cache's three runtimes.
func onSliceNode(t *testing.T, path string) run {
	t.Helper()
	return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
}

func onSliceJavaScript(t *testing.T, program *ir.Program) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0644); err != nil {
		t.Fatal(err)
	}
	return onSliceNode(t, path)
}
