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
	path, absolutePathError := filepath.Abs("gaps/1_push.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "ab\n" {
		test.Fatalf("Node: %q", result.output)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := filepath.Join(test.TempDir(), "push")
	if buildError := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := execute(test, "", binary)
	emitted := filepath.Join(test.TempDir(), "push.mjs")
	if writeError := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); writeError != nil {
		test.Fatal(writeError)
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
	path, absolutePathError := filepath.Abs("gaps/2_bigint.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	result := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1237940039285380274899124223\n" {
		test.Fatalf("Node: %q", result.output)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(lowerError, &notYet) || notYet.What != "a value of type 1237940039285380274899124223n" {
		test.Fatalf("gap changed: %v; update GAPS.md and undo gap 2 workaround", lowerError)
	}
}
