package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A const enum declared inside a function or a block lowers as one at module scope does (#xpjcdpf): it
// makes no runtime object, and every member read folds to the checker's constant by the member's symbol.
// p1 is debug.ts formatControlFlowGraph's shape (879, 893), a string enum and a flags enum whose members
// are built from members, read in a nested function's switch, an arrow and `|=` on a local typed by the
// enum. p2 shadows a module-level const enum of the same name inside two functions, each with its own
// values. p3 declares one in an if block. Each lowers and agrees with Node on release, sanitized native and
// the JavaScript backend. p4, two declarations of one const enum in a function body, which TypeScript
// merges, stops at the merged-declaration check as it does at module scope, and n1, a runtime enum inside
// a function, still stops, its object being one that would have to be made each time its block runs. It
// stops at the runtime-enum readiness check, which runs before the block-scope one. Each program's whole
// stop is pinned in its .stop file, or "lowers". Mutant: the const test at enums.go's block-scope check
// dropped, and p1 to p3 stop with "an enum inside a function or block".
func TestLocalConstEnums(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/local_const_enum"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 12 {
		t.Fatalf("found %d programs, want 12", programs)
	}
}
