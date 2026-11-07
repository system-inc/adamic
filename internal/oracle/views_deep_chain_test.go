package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestViewsDeepChain(test *testing.T) {
	test.Parallel()
	loomViewsDAgrees(test, "internal/oracle/testdata/views_deep_chain.ts")
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/views_deep_chain.ts"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	got, binary := nativelyUncached(test, program)
	if got.exitCode != 0 || string(got.stdout) != "true\n" {
		test.Fatalf("deep chain: %#v", got)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatalf("chain release: %s", report)
	}
}

// Restore D's recursive matcher in a private runtime. The source and the
// receiver tag stay unchanged, so Node observes the same checked-view cast.
func TestViewsDeepChainRecursiveWalkMutant(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/views_deep_chain.ts"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	if want.exitCode != 0 || string(want.stdout) != "true\n" {
		test.Fatalf("Node: %#v", want)
	}
	source := native.C(program)
	if !strings.Contains(source, "adamic_view_untagged_plain_select(") {
		test.Fatal("checked-view matcher was not emitted")
	}
	if diff := disagreement(want, onWASI(test, source)); diff != "" {
		test.Fatalf("control: %s", diff)
	}
	directory := test.TempDir()
	// Reads: internal/native/runtime/*
	files, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
	if err != nil {
		test.Fatal(err)
	}
	recursive, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/views_recursive_walk.c.txt"))
	if err != nil {
		test.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", file.Name()))
		if err != nil {
			test.Fatal(err)
		}
		if file.Name() == "view_unions_untagged.c" {
			text := string(data)
			start, end := strings.Index(
				text,
				"// Each frame owns",
			), strings.Index(text, "size_t adamic_view_untagged_plain_select")
			if start < 0 || end <= start {
				test.Fatal("matcher boundaries moved")
			}
			data = []byte(text[:start] + string(recursive) + text[end:])
		}
		if err := os.WriteFile(filepath.Join(directory, file.Name()), data, 0644); err != nil {
			test.Fatal(err)
		}
	}
	options := native.Options{Target: "wasm32-wasi"}
	library, err := native.RuntimeLibraryForSource(directory, source, options)
	if err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(test.TempDir(), "mutant.wasm")
	if err := native.BuildWithRuntime(source, binary, options, library); err != nil {
		test.Fatal(err)
	}
	got := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/wasi.mjs"),
		binary,
	)
	if got.exitCode != 70 || string(got.stderr) != "adamic: panic: RangeError: Maximum call stack size exceeded\n" {
		test.Fatalf("recursive walk failed for an unexpected reason: %#v", got)
	}
	if diff := disagreement(want, got); diff == "" {
		test.Fatal("recursive walk mutant matched Node")
	}
	test.Logf(
		"Node exit %d stdout %q; recursive walk exit %d stdout %q stderr %q",
		want.exitCode,
		want.stdout,
		got.exitCode,
		got.stdout,
		got.stderr,
	)
}
