package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A getter or a method in an object literal is its property's value, seen as the slot's property, and is
// judged against it exactly as { p: r } is (#wjqtypf). Before, nothing judged either, since neither is an
// expression: `{ get p(): R { return r; } }` returned as `{ p: W }` let `make().p.t = 1` write r.t, which R
// says is readonly, and Node and native both printed 1. p01 to p05 are Language's getter probes (a plain
// return slot, `{ p: W } | undefined`, a const, an inferred return, an argument), and p06 a method
// `{ get(): R }` returned as `{ get(): W }`; each is refused. t01 to t06 are their honest twins, W's own
// field returned, and each lowers and agrees with Node on release, sanitized native and the JavaScript
// backend. s01, a setter alone under a `{ p: W }` slot, is not yet: a member of a kind refuseWidening
// doesn't judge fails closed wherever a slot names its property. n01, an arrow returning R seen as one
// returning W, and n02, a class instance whose method returns R seen through `{ get(): W }`, are the
// neighbours, refused already. Each program's whole stop is pinned in its .stop file, or "lowers".
// Mutants: MUTANTS.
func TestLiteralMemberSlots(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/literal_member_slot"
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
	if programs != 30 {
		t.Fatalf("found %d programs, want 30", programs)
	}
}
