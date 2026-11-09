package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A call through ?. to a Node host method stops unless the method's own host lowering guards an
// optional receiver, read from lower's optionalHostCallGuarded (#3625x53). n1 calls a Dirent's isFile
// through ?. on an entry that is undefined at run time: nodeFSDirectoryCall lowers the predicate
// without reading the optional flag, so before the rule it lowered and ran isFile on undefined, a null
// access natively and a TypeError on the JavaScript backend, where Node prints undefined. n2 is a
// Stats' isFile, which nodeFSFile already refused through ?. and still stops, at the rule now. n3 is
// the Stats | Dirent | undefined shape, a union property whose roots are both members, so both are
// named. n4 calls process's cwd through ?. on a local NodeJS.Process | undefined: nodeHost lowers cwd
// to host_cwd without reading the receiver, but handing process to the local is a namespace object used
// as a value, which stops first at the parent and here alike (Codex 01a1beee), so n4 pins that stop and
// the cwd road is unreached today; the rule closes it if that stop lifts. processPath's own ?. admission,
// process?.cwd() on the process binding, stops at the rule, since a member's name can't
// tell its proven receiver from a local. p1, a method the program declares, called through ?. on a
// receiver undefined and then present, still lowers and agrees with Node on release, sanitized native
// and the JavaScript backend. The table is empty, so no row holds a guarded host method to Node. Each
// program's whole stop is pinned in its .stop file, or "lowers". Mutants
// (internal/lower/optional-host-call.mutants.md): the rule removed, and n1 and n3 lower; an unguarded
// member put in the table, and n1 lowers.
func TestOptionalHostCalls(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/optional_host_call"
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
	if programs != 10 {
		t.Fatalf("found %d programs, want 10", programs)
	}
}
