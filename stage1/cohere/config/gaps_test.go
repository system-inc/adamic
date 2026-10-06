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
