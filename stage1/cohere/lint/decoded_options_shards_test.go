package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testDecodedOptionsAndMutantShards = 6

// This is a pinned set of assertions over one decoded-option fixture. Each
// assertion is a case, including both independent mutant-must-fail checks.
func decodedOptionsCases() []string {
	return []string{"Node", "emitted JavaScript", "native", "Go count", "mutant Node", "mutant emitted JavaScript"}
}

func decodedOptionsAssignments(cases []string) [][]int {
	shards := make([][]int, testDecodedOptionsAndMutantShards)
	for i := range cases {
		shards[i%len(shards)] = append(shards[i%len(shards)], i)
	}
	return shards
}

func runDecodedOptionsShards(t *testing.T) {
	t.Helper()
	t.Parallel()
	started := time.Now()
	cases := decodedOptionsCases()
	shards := decodedOptionsAssignments(cases)
	if len(shards) != testDecodedOptionsAndMutantShards {
		t.Fatal("decoded-option shard enumeration disagrees with constant")
	}
	seen := make([]int, len(cases))
	for _, shard := range shards {
		for _, index := range shard {
			seen[index]++
		}
	}
	for i, count := range seen {
		if count != 1 {
			t.Fatalf("case %s assigned %d times", cases[i], count)
		}
	}
	// Plant one disagreement in the live assignment and require exactly one
	// shard to observe it through the same byte comparison as the real checks.
	planted := 1
	catches := 0
	caughtBy := -1
	for shard, indices := range shards {
		caught := false
		for _, index := range indices {
			got, want := []byte("identical"), []byte("identical")
			if index == planted {
				got = []byte("planted disagreement")
			}
			if !bytes.Equal(got, want) {
				caught = true
			}
		}
		if caught {
			catches++
			caughtBy = shard
		}
	}
	if catches != 1 {
		t.Fatalf("planted disagreement caught by %d shards", catches)
	}
	t.Logf("union: %d cases, each exactly once; planted case %s caught by shard-%03d", len(cases), cases[planted], caughtBy)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(t.TempDir(), "catch.ts")
	if err := os.WriteFile(fixture, []byte("try { work(); } catch(e) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{fixture + "\tno-empty\t\t\tfalse\t{\"AllowEmptyCatch\":true}"})
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	module := emittedJavaScript(t, directory)
	changed := mutant(t, "'allowemptycatch'", "'ignored-allowemptycatch'", "main.ts")
	changedModule := emittedJavaScript(t, changed)
	want := execute(t, "", oracle, "--manifest", path).output
	for shard, indices := range shards {
		t.Run(fmt.Sprintf("shard-%03d", shard), func(t *testing.T) {
			t.Parallel()
			for _, index := range indices {
				var got []byte
				switch index {
				case 0:
					got = node(t, directory, path, false).output
				case 1:
					got = runJavaScript(t, module, path, false).output
				case 2:
					got = execute(t, "", binary, "--manifest", path).output
				case 3:
					count := execute(t, "", oracle, "--manifest", path, "--count")
					if string(count.output) != "0\n" {
						t.Fatal("JSON catch option did not override the legacy default")
					}
					continue
				case 4:
					got = node(t, changed, path, false).output
				case 5:
					got = runJavaScript(t, changedModule, path, false).output
				default:
					t.Fatalf("unhandled decoded-option case %d", index)
				}
				if index >= 4 {
					if bytes.Equal(got, want) {
						t.Fatalf("ignored decoded-option mutant survived on %s", cases[index])
					}
					t.Logf("ignored decoded-option mutant caught on %s: %s", cases[index], difference(got, want))
				} else if diff := difference(got, want); diff != "" {
					t.Fatalf("%s: %s", cases[index], diff)
				}
			}
		})
	}
	t.Logf("TestDecodedOptionsAndMutant (setup): %s", time.Since(started))
}
