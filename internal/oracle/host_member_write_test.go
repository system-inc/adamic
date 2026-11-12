package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// tsc patches its host interfaces' method members on program object literals (program.ts:547,
// `host.readFile = fileName => {...}`), and lowering admits that write when the receiver provably holds such a
// literal (#fj0vqtf wave 1, lower's host_member_write.go): every allocation reaching it is a plain program
// object literal holding the member as a field of its own, made with no spread, and the value fits the
// member in every type the literal is held as. Each program's .stop is "lowers", and then it agrees with
// Node on release, sanitized and the JavaScript backend, or the exact lowering stop with its path relative
// to the repository. The .a and .ts twins are the same text. The deferral only admits: a write it can't
// admit stops exactly as main 1bb9663d stops it, each such .stop measured there, and only the holder rule
// (s1, s2) stops in words of its own.
//
//   - a1 is Compiler's devirtualization pin: a call through the literal's method (one reading this), the
//     method replaced by an arrow, and the call again, direct and through a parameter, prints Node's text.
//   - a2 writes a function expression.
//   - a3 and a4 are Compiler's ownership condition, as Memory measured it: the written closure captures the
//     receiver, so the literal and the closure hold each other (a3, made a thousand times), and a second
//     write releases the first closure (a4). Each is held by a graph region, and the counted build must
//     balance, not only LeakSanitizer.
//   - f1 freezes a plain literal bound to a const, which Object.freeze takes, and holds it as Host: the
//     write throws Node's TypeError (ir.Program's FrozenWriteMayThrow), caught, and the method stays.
//   - s1 is Language's h1 as she wrote it: the literal held as Narrow, written through Wide, where Node
//     then throws reading a string that's undefined. s2 is the same with object results.
//   - s3 holds a class's instance, whose method is no own field; s4 writes through a parameter of an
//     exported function nothing calls, and s5 through an arrow's, whose reach isn't known; s7 makes the
//     literal with a spread. Each stops at main's refusal of a method read as a value.
//   - s6 implements the member with a get and set pair, whose setter Node runs on the write, and s6b with
//     a getter alone returning a function, where Node throws on the write. Each stops at main's refusal
//     too: s6's call through the getter stops first ("a method sharing its name with a getter"), and a
//     stop before the writes are judged is main's at the write. hostMemberReach's accessor arm is defense
//     behind both, which no row here reaches.
//   - s8 writes a literal it would admit after a refusal on an earlier line (==). Main reports that
//     refusal first, and so does lowering: refuse's walk stops at it before it reaches the write, so no
//     stop is deferred, and a deferred write's stop never replaces it (hostMemberPending).
//
// Mutants (internal/lower/host-member-write.mutants.md): (a) the reach rule dropped, and s3, s4 and s5
// lower; (b) the holder rule judging the write site's type alone, and s1 and s2 lower.
func TestHostMemberWriteAgreesOrStops(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/host_member_write"
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
			// Every admitted write is judged by the counted build too: allocations equal frees.
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
	if programs != 28 {
		t.Fatalf("found %d programs, want 28", programs)
	}
}
