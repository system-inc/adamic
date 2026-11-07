package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMigratedNonNullFixturesRemainCovered(test *testing.T) {
	test.Parallel()
	data, err := os.ReadFile(filepath.Join(repository, "docs/verification/non-null-fixture-migration.json"))
	if err != nil {
		test.Fatal(err)
	}
	var report struct {
		Moves []struct {
			From, To, SHA256 string
			Counted          bool `json:"had_count_row"`
		}
	}
	if err := json.Unmarshal(data, &report); err != nil {
		test.Fatal(err)
	}
	if len(report.Moves) != 50 {
		test.Fatalf("want 50 preserved programs, got %d", len(report.Moves))
	}
	counts, err := recordedCounts()
	if err != nil {
		test.Fatal(err)
	}
	predicateKeys, err := predicateCountKeys()
	if err != nil {
		test.Fatal(err)
	}
	predicateCounts, err := readCountRows(countsDirectory, "predicates", predicateKeys)
	if err != nil {
		test.Fatal(err)
	}
	registered := map[string]bool{}
	for _, fixture := range fixtures {
		registered[fixture.path] = fixture.lowers
	}
	for _, move := range report.Moves {
		test.Run(move.To, func(test *testing.T) {
			source, err := os.ReadFile(filepath.Join(repository, move.To))
			if err != nil {
				test.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(source)); got != move.SHA256 {
				test.Fatalf("program changed while migrating %s: %s", move.From, got)
			}
			if move.Counted && counts[move.To] == "" && predicateCounts[move.To] == "" {
				test.Fatal("original measured fixture lost its count row")
			}
			if strings.HasPrefix(move.To, "internal/oracle/testdata/") && filepath.Base(
				move.To,
			) != "non_null_weak_freed.ts" || strings.HasPrefix(move.To, "stage3/fixtures/nested-functions/") {
				if !registered[move.To] {
					test.Fatal("original runnable fixture lost its registration")
				}
			}
		})
	}
}

func TestNonNullLiteralUnionInitializers(test *testing.T) {
	test.Parallel()
	for _, literal := range []string{"undefined", "null"} {
		test.Run(literal, func(test *testing.T) {
			path := filepath.Join(test.TempDir(), "main.ts")
			source := "let value: string | number = " + literal + "!; value = 1; console.log(value.toString());\n"
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			want := onSliceNode(test, path)
			native, binary := nativelyUncached(test, program)
			for _, got := range []run{native, onSliceJavaScript(test, program)} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatal(difference)
				}
			}
			if report := leaksUncached(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
}

// These unchanged sources used readiness markers under the historical ruling.
// As .ts programs they stop at the assertion itself, before any later read.
func assertMigratedNonNullCheck(test *testing.T, path, expression, stdout string, mutant bool) {
	test.Helper()
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	if program.NonNullChecks.Checked == 0 {
		test.Fatal("migrated fixture lost its inserted non-null check")
	}
	node := onSliceNode(test, path)
	if node.exitCode == 70 {
		test.Fatal("source Node unexpectedly fired an Adamic check")
	}
	native, _ := nativelyUncached(test, program)
	if native.exitCode != 70 ||
		string(native.stdout) != stdout ||
		!strings.HasPrefix(string(native.stderr), "adamic: panic: non-null assertion failed at "+path+":") ||
		!strings.HasSuffix(string(native.stderr), ": "+expression+" is null or undefined\n") {
		test.Fatalf("want eager checked %s with stdout %q, got %#v", expression, stdout, native)
	}
	if difference := disagreement(native, onSliceJavaScript(test, program)); difference != "" {
		test.Fatal(difference)
	}
	if difference := disagreement(native, releasedUncached(test, program)); difference != "" {
		test.Fatal(difference)
	}
	if !mutant {
		return
	}
	// Give each removed check a harmless fallback in the same representation.
	// A missing panic must be caught by output, never a C type error or sanitizer.
	empty := len(program.Strings)
	program.Strings = append(program.Strings, "")
	changes := 0
	drop := func(node any) any {
		check, ok := node.(ir.Coalesce)
		if !ok || check.Panic == nil {
			return node
		}
		check.Panic = nil
		switch check.Of {
		case ir.Number:
			check.Fallback = ir.NumberConstant{}
		case ir.Boolean:
			check.Fallback = ir.BooleanConstant{}
		case ir.String:
			check.Fallback = ir.StringConstant{Index: empty}
		case ir.Union:
			check.Fallback = ir.Box{Value: ir.NumberConstant{}}
		default:
			check.Fallback = ir.Undefined{Of: check.Of}
		}
		changes++
		return check
	}
	program.Main = mutateReadiness(program.Main, drop)
	for index := range program.Functions {
		program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, drop)
	}
	if changes == 0 {
		test.Fatal("drop-check mutant changed nothing")
	}
	got := releasedUncached(test, program)
	if string(got.stdout) == string(native.stdout) || got.exitCode != 0 && got.exitCode != 70 {
		test.Fatalf("drop-check mutant was not caught by harmless output: %#v", got)
	}
	test.Logf("drop-check mutant caught for %s: stdout %q", filepath.Base(path), got.stdout)
}
