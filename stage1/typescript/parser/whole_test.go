package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wholeNode(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest, "--whole"}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}
func TestWholeCompilerAgrees(t *testing.T) {
	t.Parallel()
	manifest, files := compilerManifest(t)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole")
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	got := wholeNode(t, directory, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node: %s", diff)
	}
	got = execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native: %s", diff)
	}
	t.Logf("%d whole compiler files, %d identical whole-tree bytes", files, len(want.output))
}
func wholeManifest(t *testing.T, cases []string) string {
	t.Helper()
	directory := t.TempDir()
	var manifest strings.Builder
	for i, source := range cases {
		path := filepath.Join(directory, fmt.Sprintf("whole-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
