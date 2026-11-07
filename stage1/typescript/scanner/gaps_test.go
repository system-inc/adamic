package scanner

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestGapStandsWhereGapsMdSays(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("gaps/1_push.ts")
	if err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	result := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "ab\n" {
		test.Fatalf("Node: %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(test.TempDir(), "push")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	actual := execute(test, "", binary)
	emitted := filepath.Join(test.TempDir(), "push.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		test.Fatal(err)
	}
	backend := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted)
	for _, side := range []execution{actual, backend} {
		if string(side.output) != string(result.output) {
			test.Fatalf("closed push output %q, Node %q", side.output, result.output)
		}
	}
}

func TestBigintGapStandsWhereGapsMdSays(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("gaps/2_bigint.ts")
	if err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	result := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1237940039285380274899124223\n" {
		test.Fatalf("Node: %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a value of type 1237940039285380274899124223n" {
		test.Fatalf("gap changed: %v; update GAPS.md and undo gap 2 workaround", err)
	}
}
