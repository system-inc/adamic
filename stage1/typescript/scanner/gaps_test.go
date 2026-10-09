package scanner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestGapStandsWhereGapsMdSays(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("gaps/1_push.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "ab\n" {
		t.Fatalf("Node: %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "push")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	observed := execute(t, "", binary)
	if string(observed.output) != string(result.output) {
		t.Fatalf("native push: %q, Node: %q", observed.output, result.output)
	}
	// Hold the restored two-entry supplementary mapping to Go, including byte
	// positions for a token following a supplementary identifier.
	input := filepath.Join(t.TempDir(), "supplementary.a")
	if err := os.WriteFile(input, []byte("const 𐐀 = 1; 𐐀;"), 0600); err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", goOracle(t), input, "scan")
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	source := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(root, "main.ts"), input, "scan")
	actual := execute(t, "", buildPort(t, ".", true), input, "scan")
	for _, side := range []execution{source, actual} {
		if diff := difference(side.output, want.output); diff != "" {
			t.Fatal(diff)
		}
	}
}

func TestBigintGapStandsWhereGapsMdSays(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("gaps/2_bigint.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1237940039285380274899124223\n" {
		t.Fatalf("Node: %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a value of type 1237940039285380274899124223n" {
		t.Fatalf("gap changed: %v; update GAPS.md and undo gap 2 workaround", err)
	}
}
