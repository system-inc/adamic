package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A fresh container written into a union slot is fresh at each member it can be (#rkybez3): tsc's
// commandLineParser writes type: new Map(Object.entries({ ... })) of an enum's values into
// CommandLineOptionBase.type, "string" | ... | Map<string, number | string>, and Stage 0 refused it as a
// Map<string, WatchFileKind> seen as Map<string, number | string>, since the fresh exemption looked for
// containers only in a slot that isn't a union. Nothing else holds the new Map to read it as its own
// type, so the controls lower and agree with Node on release, sanitized native and the JavaScript
// backend, in .a and .ts, each reading the map back through instanceof Map and writing through the wider
// member: the Object.entries form (c1), a new Map of literal entries (c2) and of the enum's members (c3),
// an array literal (c4) and a copy (c5). Still refused, each in invariant-mutable's words: a named Map seen
// through the union (r1), a named array inside the fresh Map, still walked (r2), and a fresh Map one
// member takes and another refuses, since every member it's assignable to must pass (r3). Each program's
// whole stop is pinned in its .stop file, or "lowers". Mutant: the union branch undone in freshOrWidened
// turns c1, c2, c3 and c5 red.
func TestFreshContainerUnionSlot(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/fresh_container_union_slot"
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
	if programs != 16 {
		t.Fatalf("found %d programs, want 16", programs)
	}
}
