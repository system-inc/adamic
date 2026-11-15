package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// tsc saves a host interface's method before it patches it (program.ts:543, `const originalReadFile =
// host.readFile`), calls the saved value as `originalReadFile.call(host, fileName)` and sometimes writes it
// back (#fj0vqtf wave 2). Each program's .stop is "lowers", and then it agrees with Node on release,
// sanitized and the JavaScript backend and balances in the counted build, or the exact lowering stop with
// its path relative to the repository. The .a and .ts twins are the same text.
//
//   - v1 is tsc's shape: save, patch with an arrow calling the original through .call(host, ...), restore.
//   - v2 is Adamic's ruled pin 2 with a method-syntax member: .call with another literal of the same
//     layout. v2b is the same with a function expression, as the ruling words it.
//   - h3 is Compiler's reordered twin: the thisArg's literal puts prefix at another offset.
//   - h2 is Language's: a saved method that reads this, called with no receiver, where Node throws.
//   - r1 reads the member straight into another literal's field, r2 aliases the saved value and stores it
//     in a literal's field, and r3 passes it to a function that calls it: Node runs each with a receiver
//     the member's own literal isn't.
//   - p1 saves an arrow property, which takes no receiver, and calls it plainly.
//   - s1 reads the member off a class's instance, and s2 off an arrow's parameter, whose reach is unknown.
//
// Mutants (internal/lower/host-member-value.mutants.md).
func TestHostMemberValueAgreesOrStops(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/host_member_value"
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
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
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
			// Every admitted program is judged by the counted build too: allocations equal frees.
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			command, args := pinnedStack(binary)
			if report := leakcheck.Unbalanced(leakRun(execute(t, command, args...))); report != "" {
				t.Fatalf("counted build: %s", report)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 22 {
		t.Fatalf("found %d programs, want 22", programs)
	}
}
