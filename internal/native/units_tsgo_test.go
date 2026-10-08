package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Not parallel: this test explicitly compares the object cache with its bypass on the same input.
func TestSplitTSGoAgrees(t *testing.T) {
	archive := os.Getenv("ADAMIC_CLANG_TSGO_ARCHIVE")
	if archive == "" {
		// census: required-input ADAMIC_CLANG_TSGO_ARCHIVE: bridge/tsgo built c-archive for the pinned cohere TypeScript checker, using go build -buildmode=c-archive -o /path/tsgo.a ./bridge/tsgo/archive; see docs/gate-inputs.md.
		t.Skip("set ADAMIC_CLANG_TSGO_ARCHIVE to a built checker archive")
	}
	program, err := load.Load([]string{"../../stage1/cohere/typeaware/main.ts"})
	if err != nil {
		t.Fatal(err)
	}
	program.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source, err := TSGoC(lowered)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	input := filepath.Join(directory, "sample.a")
	config := filepath.Join(directory, "tsconfig.json")
	manifest := filepath.Join(directory, "manifest")
	for name, contents := range map[string]string{
		input:    "const value = 1; const negative = -value;\n",
		config:   `{"compilerOptions":{"strict":true,"target":"esnext"},"files":["sample.a"]}`,
		manifest: input + "\n",
	} {
		if err := os.WriteFile(name, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	whole := filepath.Join(directory, "whole")
	options := Options{Sanitize: true, Jobs: 5}
	if err := BuildTSGo(source, whole, archive, options); err != nil {
		t.Fatal(err)
	}
	want, err := exec.Command(whole, config, manifest).CombinedOutput()
	if err != nil {
		t.Fatalf("whole checker: %v\n%s", err, want)
	}
	for _, uncached := range []string{"0", "1"} {
		t.Setenv("ADAMIC_GATE_UNCACHED", uncached)
		split := filepath.Join(directory, "split")
		if err := BuildSplitTSGo(source, split, archive, options); err != nil {
			t.Fatal(err)
		}
		got, err := exec.Command(split, config, manifest).CombinedOutput()
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("checker uncached=%s: %v\nwhole=%s\nsplit=%s", uncached, err, want, got)
		}
		t.Logf("checker uncached=%s: %d identical bytes: %s", uncached, len(got), got)
	}
}
