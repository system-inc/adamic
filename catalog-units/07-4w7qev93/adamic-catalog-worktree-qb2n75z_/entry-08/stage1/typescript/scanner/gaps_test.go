package scanner

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

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
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "push with other than one value" {
		t.Fatalf("gap changed: %v; update GAPS.md and undo gap 1 workaround", err)
	}
}

func TestBigintGapStandsWhereGapsMdSays(t *testing.T) {
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
