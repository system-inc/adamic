package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
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
//   - v3 makes one literal twice through a factory and calls a's method through .call with b: one
//     allocation site, so it lowers. v4's thisArg may be either of two literals, the read's and another.
//     v5 is h4 through .call: a's method, reading this.x as a number, handed b, whose x is a string.
//   - h2 is Language's: a saved method that reads this, called with no receiver, where Node throws.
//   - h4 is Compiler's: a method that reads this.x, restored to an object of another literal whose x sits
//     at the same index but holds a string. h4b saves, patches and restores a method on one literal.
//   - r1 reads the member straight into another literal's field, r2 aliases the saved value and stores it
//     in a literal's field, and r3 passes it to a function that calls it: Node runs each with a receiver
//     the member's own literal isn't.
//   - p1 saves an arrow property, which takes no receiver, and calls it plainly.
//   - s1 reads the member off a class's instance, and s2 off an arrow's parameter, whose reach is unknown.
//   - s3 is s2's reach with an arrow property called plainly, so no use stops first and the reach rule's
//     stop is its own.
//
// Measured on head 1 (#fj0vqtf wave 2): p1 lowers and agrees. h2 stops at its plain call, r1 and r2 at the
// field they store the value in and r3 at the argument it passes. h3, s1, s2, v1, v2 and v2b call through
// .call and stop during lowering at main's refusal of a method read as a value, until head 2 lowers .call.
// s3 stops at the same refusal by the reach rule. h4 stops at its restore, a method moved between objects
// of different literals, and h4b lowers and agrees.
//
// Head 2 lowers .call(t, ...args) as bind's road does and admits it under the restore's one-site rule:
// t's reach is known and it and the read are one allocation site. Measured on head 2 (J-msx1c62): v1 and
// v3 lower and agree, v2 and h3 stop at their .call (an object of another literal), and v2b keeps head 1's
// stop, since a function expression reading this is refused at the read before .call is reached. s1 and s2
// now lower their .call and stop at the reach rule, in main's words as before. v4 and v5 hand .call an
// object of another literal and stop there too.
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
			if err := native.Build(native.C(program), binary, native_runtime.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			command, args := pinnedStack(binary)
			if report := leakcheck.Unbalanced(leakRun(execute(t, command, args...))); report != "" {
				t.Fatalf("counted build: %s", report)
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 34 {
		t.Fatalf("found %d programs, want 34", programs)
	}
}
