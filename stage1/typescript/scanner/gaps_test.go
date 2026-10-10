package scanner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: ASAN_OPTIONS process environment via t.Setenv.
func TestGapStandsWhereGapsMdSays(t *testing.T) {
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
	t.Setenv("ASAN_OPTIONS", "detect_leaks=1")
	actual := execute(t, "", binary).output
	emitted := filepath.Join(t.TempDir(), "push.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	backend := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted).output
	for _, side := range [][]byte{actual, backend} {
		if !bytes.Equal(side, result.output) {
			t.Fatalf("closed push gap: %q, Node %q", side, result.output)
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
