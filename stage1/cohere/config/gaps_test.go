package config

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONLibraryGap(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("gaps/json.ts")
	if err != nil {
		t.Fatal(err)
	}
	got := execute(t, nil, "node", path)
	if got.exitCode != 0 || string(got.stdout) != "parsed\n" {
		t.Fatalf("Node: %+v", got)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if err == nil {
		t.Fatal("JSON library gap closed: remove the recorded gap")
	}
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "JSON.parse: its result's type can't be proven from the text" {
		t.Fatalf("gap changed: %v", err)
	}
	t.Log(err)
}

// The settings/tsconfig oracle probes live in cohere's package to call its real loaders.
func TestSettingsBehaviorCensus(t *testing.T) {
	t.Parallel()
	input := filepath.Join(t.TempDir(), "cases")
	if err := os.WriteFile(input, []byte("settings-census\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := goCohere(t, input)
	for _, want := range []string{
		"comments: parsing lint config ROOT/comments.json: invalid character '/' looking for beginning of object key string",
		"trailing: parsing lint config ROOT/trailing.json: invalid character '}' looking for beginning of object key string",
		"tsconfig-comments: files ROOT/a.ts",
		"missing: reading lint config ROOT/missing.json: open ROOT/missing.json: no such file or directory",
		"extends itself:", "no tsconfig at ROOT/missing-tsconfig.json",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in census:\n%s", want, got)
		}
	}
	t.Logf("Go settings and tsconfig behavior:\n%s", got)
}

// Not a compiler refusal: the typed port explicitly declines a source-contributing directory link.
// Go's production parser follows links and canonicalizes visits. This counterexample must close when
// realpath is added, rather than letting a generic failure be counted as the expected gap.
func TestDirectoryLinkGap(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "z-target"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "z-target/a.ts"), []byte("export const a=1;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(`{"include":["**/*.ts"]}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("z-target", filepath.Join(root, "a-alias")); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "cases")
	if err := os.WriteFile(input, []byte("tsconfig\t"+filepath.Join(root, "tsconfig.json")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	want := goCohere(t, input)
	if !strings.Contains(want, "/a-alias/a.ts\"") || strings.Contains(want, "/z-target/a.ts\"") {
		t.Fatalf("Go symlink semantics changed: %s", want)
	}
	source, err := filepath.Abs("gaps/directory_link.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, source)
	expected := "stage1 tsconfig directory symlink needs realpath: " + filepath.Join(root, "a-alias") + "\n"
	for _, side := range []struct {
		name   string
		result run
	}{{"native", nativelyRun(t, program, filepath.Join(root, "tsconfig.json"))}, {"Node", onNode(t, source, filepath.Join(root, "tsconfig.json"))}, {"backend", onJavaScriptBackend(t, program, filepath.Join(root, "tsconfig.json"))}} {
		if side.result.exitCode != 0 || len(side.result.stderr) > 0 || string(side.result.stdout) != expected {
			t.Fatalf("%s: directory link gap closed or changed: exit %d stdout %q stderr %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	t.Logf("Go follows and deduplicates the directory alias; port explicitly declines: %s", expected)
}
