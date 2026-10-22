package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestViewsDeepChain(t *testing.T) {
	t.Parallel()
	loomViewsDAgrees(t, "internal/oracle/testdata/views_deep_chain.a")
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/views_deep_chain.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, binary := nativelyUncached(t, program)
	if got.exitCode != 0 || string(got.stdout) != "true\n" {
		t.Fatalf("deep chain: %#v", got)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatalf("chain release: %s", report)
	}
}

// Restore D's recursive matcher in a private runtime. The source and the
// receiver tag stay unchanged, so Node observes the same checked-view cast.
func TestViewsDeepChainRecursiveWalkMutant(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/views_deep_chain.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "true\n" {
		t.Fatalf("Node: %#v", want)
	}
	source := native.C(program)
	if !strings.Contains(source, "adamic_view_untagged_plain_select(") {
		t.Fatal("checked-view matcher was not emitted")
	}
	if diff := disagreement(want, onWASI(t, source)); diff != "" {
		t.Fatalf("control: %s", diff)
	}
	directory := t.TempDir()
	files, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
	if err != nil {
		t.Fatal(err)
	}
	recursive, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/views_recursive_walk.c.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", file.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if file.Name() == "view_unions_untagged.c" {
			text := string(data)
			start, end := strings.Index(text, "// Each frame owns"), strings.Index(text, "size_t adamic_view_untagged_plain_select")
			if start < 0 || end <= start {
				t.Fatal("matcher boundaries moved")
			}
			data = []byte(text[:start] + string(recursive) + text[end:])
		}
		if err := os.WriteFile(filepath.Join(directory, file.Name()), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	options := native.Options{Target: "wasm32-wasi"}
	library, err := native.RuntimeLibraryForSource(directory, source, options)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "mutant.wasm")
	if err := native.BuildWithRuntime(source, binary, options, library); err != nil {
		t.Fatal(err)
	}
	got := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
	if got.exitCode != 70 || string(got.stderr) != "adamic: panic: RangeError: Maximum call stack size exceeded\n" {
		t.Fatalf("recursive walk failed for an unexpected reason: %#v", got)
	}
	if diff := disagreement(want, got); diff == "" {
		t.Fatal("recursive walk mutant matched Node")
	}
	t.Logf("Node exit %d stdout %q; recursive walk exit %d stdout %q stderr %q", want.exitCode, want.stdout, got.exitCode, got.stdout, got.stderr)
}
