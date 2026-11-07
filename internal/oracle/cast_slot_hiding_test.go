package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A cast's own type goes into its slot as a named value's does (#8v8112n): f(wide as Node), wide holding
// name: 5, into a Positioned that declares name?: string, hid name and segfaulted natively where Node
// printed 5, since the hiding rule judged only the cast's operand against its type. Every road into a
// slot now refuses it in the hiding rule's words, in .a and .ts, where s2, its twin through a named
// const, already was: an argument (s1, s4), an initializer (s3), a return (s5), an assignment (s6), a
// class field (s7), a literal's property (s8), a parenthesized cast (s9), a cast of a cast (s11) and a
// cast of a literal that names name (s14). On the parent each of these lowered and exited 139 natively.
// A conditional arm holding a cast (s10), satisfies (s12), ! (s13, refused as ! in .a) and an angle
// assertion (s15) had no gap and keep their stops. The controls lower and agree with Node on release,
// sanitized native and the JavaScript backend: x as const, judged as its operand (c1), and a cast to the
// slot's own type (c2). Each program's whole stop is pinned in its .stop file, or "lowers". Mutant: the
// cast's own slot left unjudged turns s1 back into the native segfault, and every row the fix refuses
// red.
func TestCastSlotHiding(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/cast_slot_hiding"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		test.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				test.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				test.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(test, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					test.Fatalf("lowered, want %q", stop)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					test.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				test.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(test, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"release", released(test, program)},
				{"sanitized", sanitized},
				{"JavaScript", onJavaScriptBackend(test, program)},
			} {
				if difference := disagreement(truth, backend.got); difference != "" {
					test.Fatalf(
						"%s: %s: Node %q, %s %q",
						backend.name,
						difference,
						truth.stdout,
						backend.name,
						backend.got.stdout,
					)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 34 {
		test.Fatalf("found %d programs, want 34", programs)
	}
}
