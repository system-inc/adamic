package lint

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

// Warm the real caches, edit exactly one source byte at the same path, and
// require changed findings on every runtime. Restoring that byte restores the
// original output. No cache is cleared or disabled between these observations.
func TestCheckerCacheSourceByte(t *testing.T) {
	directory := mutant(t, "", "")
	source := filepath.Join(t.TempDir(), "source.ts")
	if err := os.WriteFile(source, []byte("debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{source + "\tno-debugger"})
	want := execute(t, "", goOracle(t), "--manifest", path).output
	message := filepath.Join(directory, "rules/no-debugger/messages.ts")
	original, err := os.ReadFile(message)
	if err != nil {
		t.Fatal(err)
	}
	changed := append([]byte(nil), original...)
	at := bytes.Index(changed, []byte("A debugger statement"))
	if at < 0 {
		t.Fatal("source-byte anchor changed")
	}
	changed[at] = 'B'
	run := func() [][]byte {
		binary := buildPort(t, directory, true)
		return [][]byte{execute(t, "", binary, "--manifest", path).output, node(t, directory, path, false).output, emittedNode(t, directory, path, false).output}
	}
	before := run()
	for i, got := range before {
		if diff := difference(got, want); diff != "" {
			t.Fatalf("baseline runtime %d: %s", i, diff)
		}
	}
	baseline := checkerCompile(t, directory)
	if baseline != checkerCompile(t, directory) {
		t.Fatal("unchanged source missed warm compilation cache")
	}
	if err := os.WriteFile(message, changed, 0644); err != nil {
		t.Fatal(err)
	}
	after := run()
	expected := bytes.ReplaceAll(want, []byte("A debugger statement"), []byte("B debugger statement"))
	if bytes.Equal(expected, want) {
		t.Fatal("wire anchor changed")
	}
	for i, got := range after {
		if bytes.Equal(got, want) {
			t.Fatalf("stale cache hid source-byte mutant on runtime %d", i)
		}
		if diff := difference(got, expected); diff != "" {
			t.Fatalf("rebuilt runtime %d: %s", i, diff)
		}
		t.Logf("runtime %d: source-byte mutant caught by findings comparison; before=%x after=%x", i, sha256.Sum256(before[i]), sha256.Sum256(got))
	}
	rebuilt := checkerCompile(t, directory)
	if baseline.c == rebuilt.c || baseline.javascript == rebuilt.javascript {
		t.Fatal("source byte did not change generated output")
	}
	t.Logf("generated C before=%x after=%x; emitted JavaScript before=%x after=%x", sha256.Sum256([]byte(baseline.c)), sha256.Sum256([]byte(rebuilt.c)), sha256.Sum256([]byte(baseline.javascript)), sha256.Sum256([]byte(rebuilt.javascript)))
	if err := os.WriteFile(message, original, 0644); err != nil {
		t.Fatal(err)
	}
	for i, got := range run() {
		if diff := difference(got, want); diff != "" {
			t.Fatalf("restored runtime %d: %s", i, diff)
		}
	}
	if baseline != checkerCompile(t, directory) {
		t.Fatal("restored source did not reuse original content key")
	}
}
