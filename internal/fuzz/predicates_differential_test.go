//go:build adamic_predicates

// The predicate differential lane (review #fxspptb), an opt-in run outside the gate while main still gets these
// shapes wrong: go test -tags adamic_predicates ./internal/fuzz -run TestPredicateDifferential, with
// ADAMIC_PREDICATE_CHECKOUT naming the checkout (ADAMIC_PREDICATE_REPORT=1 records findings instead of failing). Behind
// a build tag, it is never compiled into the gate's run, so it never skips there.

package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Run explicitly against any checkout. Report mode records known open findings;
// strict mode is the regression gate after the compiler fixes land. Preparation
// has its own subtest so each timed unit remains below Kirk's 30-second limit.
func TestPredicateDifferential(t *testing.T) {
	root := os.Getenv("ADAMIC_PREDICATE_CHECKOUT")
	if root == "" {
		t.Fatal("set ADAMIC_PREDICATE_CHECKOUT to the checkout whose compiler this lane runs")
	}
	directory := t.TempDir()
	var checkout *Checkout
	t.Run("prepare", func(t *testing.T) {
		var err error
		checkout, err = Prepare(root, filepath.Join(directory, "compiler"))
		if err != nil {
			t.Fatal(err)
		}
	})
	if checkout == nil {
		t.Fatal("compiler preparation failed")
	}
	run := func(t *testing.T, source string) {
		start := time.Now()
		programDirectory := t.TempDir()
		path := filepath.Join(programDirectory, "program.a")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		node := execute(programDirectory, nil, 5*time.Second, "node", "--disable-warning=ExperimentalWarning", filepath.Join(checkout.Root, "oracle", "node.mjs"), path)
		if node.ExitCode != 0 || node.TimedOut {
			t.Fatalf("Node oracle failed: %+v", node)
		}
		result := checkout.TryFile(path, programDirectory)
		t.Logf("seconds=%.3f verdict=%s key=%s detail=%s Node=%q native=%q", time.Since(start).Seconds(), result.Verdict, result.Key, result.Detail, node.Stdout, result.Native.Stdout)
		if time.Since(start) >= 30*time.Second {
			t.Fatal("exceeded Kirk's 30-second unit budget")
		}
		switch result.Verdict {
		case Agreed, Checked, NotYet:
		case Invalid:
			if !strings.HasPrefix(result.Key, "refused:") {
				t.Fatalf("checker rejected generated source: %s\n%s", result.Detail, source)
			}
		case Finding:
			if os.Getenv("ADAMIC_PREDICATE_REPORT") != "1" {
				t.Fatalf("%s\n%s\n%s", result.Key, result.Detail, source)
			}
			t.Logf("finding program:\n%s", source)
		default:
			t.Fatalf("generator/corpus failure: %+v\n%s", result, source)
		}
	}
	// Generator-only runs never read corpus files. Seed is the random generator seed.
	t.Run("generator", func(t *testing.T) {
		for seed := uint64(1); seed <= 36; seed++ {
			t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
				t.Run("helper", func(t *testing.T) { run(t, predicateGenerator(seed).helperPredicateSource()) })
				t.Run("array", func(t *testing.T) { run(t, predicateGenerator(seed).arrayPredicateSource()) })
			})
		}
	})
	t.Run("corpus", func(t *testing.T) {
		paths, err := filepath.Glob("testdata/predicates/*.a")
		if err != nil || len(paths) != 10 {
			t.Fatalf("want ten seed programs: %v %v", paths, err)
		}
		for _, path := range paths {
			t.Run(filepath.Base(path), func(t *testing.T) {
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				run(t, string(source))
			})
		}
	})
}
