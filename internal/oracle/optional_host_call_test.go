package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A call through ?. to a Node host method stops unless the method's own host lowering guards an
// optional receiver, read from lower's optionalHostCallGuarded (#3625x53). n1 (Dirent), n2 (Stats)
// and n3 (Stats | Dirent | undefined) now guard their predicates through optionalIntrinsic's mark.
// Each runs on undefined and present receivers against Node on release, sanitized native and
// JavaScript; n3 presents both Stats and Dirent. These are the table's guarded-host pins.
// n4 calls process's cwd through ?. on a local NodeJS.Process | undefined: nodeHost lowers cwd
// to host_cwd without reading the receiver, but handing process to the local is a namespace object
// used as a value, which stops first. The cwd road there remains unreached; the rule closes it if
// that stop lifts. process?.cwd() on the process binding stops at the rule too, since a member's
// name can't tell its proven receiver from a local. n5 is entry?.isFile?.() on Dirent | undefined:
// optionalMethodCall leaves it to optionalCall's optional-function-call stop before this rule.
// n6 reaches the unguarded cwd road through a parameter handed only undefined, without handing
// the process namespace over as a value. It stops here; either removing the rule or admitting
// node:process.Process.cwd in the table makes it print present where Node prints undefined.
// p1, a method the program declares, called on undefined and present receivers, still lowers and
// agrees with Node. Each program's whole stop is pinned in its .stop file, or "lowers". The two
// mutants are recorded in internal/lower/optional-host-call.mutants.md and caught by n6's pair.
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
	if programs != 14 {
		t.Fatalf("found %d programs, want 14", programs)
	}
}
