package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// An admission-only compiler check must not rerun a changed program using old
// native evidence. Compare every accepted C byte with the compiler used by the
// complete Node-held survey; changed refusals go through the unchanged attempt.
func TestRegExpProtocolCompilerEquivalence(t *testing.T) {
	corpus, previous, current, results, output := os.Getenv("ADAMIC_REGEXP_EQ_CORPUS"), os.Getenv("ADAMIC_REGEXP_EQ_BEFORE"), os.Getenv("ADAMIC_REGEXP_EQ_AFTER"), os.Getenv("ADAMIC_REGEXP_EQ_RESULTS"), os.Getenv("ADAMIC_REGEXP_EQ_OUTPUT")
	if corpus == "" || previous == "" || current == "" || results == "" || output == "" {
		t.Skip("explicit corpus, compiler pair, result log and scratch output required")
	}
	data, err := os.ReadFile(results)
	if err != nil {
		t.Fatal(err)
	}
	type checked struct {
		Path          string
		Outcome       outcomeKind
		Before, After string
		IdenticalC    bool
		SHA256        string
		Update        *result `json:",omitempty"`
	}
	rows := []checked{}
	var lock sync.Mutex
	t.Run("files", func(t *testing.T) {
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}
			var prior result
			if err := json.Unmarshal([]byte(line), &prior); err != nil {
				t.Fatal(err)
			}
			if prior.Kind == outcomeSkipped {
				continue
			}
			t.Run(prior.Path, func(t *testing.T) {
				t.Parallel()
				source, err := os.ReadFile(filepath.Join(corpus, "test", prior.Path))
				if err != nil {
					t.Fatal(err)
				}
				test := classify(prior.Path, string(source), true)
				if test.Skip != "" {
					t.Fatalf("classification changed: %s", test.Skip)
				}
				work := t.TempDir()
				file := filepath.Join(work, "program.ts")
				if err := os.WriteFile(file, []byte(test.Program), 0644); err != nil {
					t.Fatal(err)
				}
				old := runCommandWithLimit(2*time.Minute, nil, 16<<20, previous, "c", file)
				next := runCommandWithLimit(2*time.Minute, nil, 16<<20, current, "c", file)
				oldKind, _ := compileClass(old.Stderr, old.Exit, old.TimedOut)
				newKind, _ := compileClass(next.Stderr, next.Exit, next.TimedOut)
				row := checked{Path: prior.Path, Outcome: prior.Kind, Before: oldKind, After: newKind}
				if old.Exit == 0 && next.Exit == 0 {
					row.IdenticalC = old.Stdout == next.Stdout
					if !row.IdenticalC {
						t.Fatalf("accepted C changed for %s", prior.Path)
					}
					sum := sha256.Sum256([]byte(next.Stdout))
					row.SHA256 = hex.EncodeToString(sum[:])
				} else if oldKind == "refused" && newKind == "refused" {
					// Neither version emitted a runnable program.
				} else if old.Exit == 0 && newKind == "refused" && prior.Kind == outcomeRefused {
					// The unchanged runner determines the new refusal, without bypassing its
					// constructor guard or awarding a pass without untouched Node evidence.
					e := &engine{adamic: current, work: work}
					actual := e.attempt(test)
					if actual.Kind != outcomeRefused {
						t.Fatalf("changed compiler refusal: %+v", actual)
					}
					row.Update = &actual
				} else {
					t.Fatalf("compiler admission changed unexpectedly: %s: %d/%s -> %d/%s", prior.Path, old.Exit, oldKind, next.Exit, newKind)
				}
				if prior.Kind == outcomePass && !row.IdenticalC {
					t.Fatalf("a surveyed pass lost its identical C evidence: %s", prior.Path)
				}
				lock.Lock()
				rows = append(rows, row)
				lock.Unlock()
			})
		}
	})
	if t.Failed() {
		return
	}
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, append(encoded, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("checked %d non-skipped corpus files against the Node-held complete survey", len(rows))
}
