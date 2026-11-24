package oracle

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Ruling 2 (Adamic on #bbssdcy, Nov 24): a field every declaring member declares never takes the plain
// member's storage, which is no slot of its own, and every read of it is a :76 read (rulings.md:76), a
// loud stop. Each program in testdata/never_field is held here to Node on the sanitized build, the
// release build and the JavaScript backend, the stops beside what Node prints. m01.ts and m03.ts are
// Language's programs from #hatdcsp, and the three readonly-view programs the earlier lane's with
// Adamic's two readings (#bbssdcy, Nov 24 18:03Z), each byte for byte. internal/lower/never_field_test.go
// pins the lowering: the conflict, the write and the invariance refusal, and that the stop loads no slot.

// neverFieldPath is testdata/never_field/name's absolute path.
func neverFieldPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_field", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// The programs ruling 2 admits agree with Node: site.ts, tsc's Expression & ArrowFunction in small, read
// by its other field; m01.ts, an Arrow read through its Base's name?: string, as at the pin; the earlier
// lane's live writer, which writes the real object's own name through the writable alias it held from
// before the cast and never reads the Arrow's, as at the pin; and three shapes past tsc's, an intersection
// of three members, one held as an array's element and read in a loop, and one a function returns.
func TestNeverFieldAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout string }{
		{"site.ts", "kind 1\nkind 2\n"},
		{"m01.ts", "undefined\n"},
		{"readonly-view-live-writer.ts", "true 1 before\ntrue 1 written\n"},
		{"three_members.ts", "kind 3\nkind 4\n"},
		{"array_loop.ts", "7\n"},
		{"returned.ts", "kind 5\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := neverFieldPath(t, probe.name)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 || string(want.stdout) != probe.stdout {
				t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", probe.stdout, want.exitCode, want.stdout, want.stderr)
			}
			sanitized, _ := natively(t, program)
			for backend, got := range map[string]run{"native": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s; exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// A read of arrow.name, which Arrow declares never, stops loudly at exit 70 on every leg, naming the read
// and where it is, with what ran before it printed as Node prints it: direct_read.ts, whose real object
// has no name slot (Node reads undefined and goes on), and Adamic's two readings of the earlier lane's
// program, one before the write through base and one after it (Node reads before and written). After the
// write the field isn't the narrowing road's, whose write between is a not yet
// (TestNeverReadAcrossAWriteIsNotYet): it is the run-time stop. The two readings' stderr differ only in
// their location, and neither holds the slot's value. The stdout before the stop is Node's up to the read,
// byte for byte. Mutant: tmp-audit/wrong-read-unchecked.patch reads the slot, and direct_read.ts and
// readonly-view-read-after.ts print what Node prints, exit 0.
func TestNeverFieldReadStops(t *testing.T) {
	t.Parallel()
	stops := map[string]string{}
	for _, probe := range []struct {
		name, node, stdout string
		line               int
	}{
		{"direct_read.ts", "kind 1\nundefined\nafter\n", "kind 1\n", 17},
		{"readonly-view-read-before.ts", "true 1 before\nbefore\ntrue 1 written\n", "true 1 before\n", 18},
		{"readonly-view-read-after.ts", "true 1 before\ntrue 1 written\nwritten\n", "true 1 before\ntrue 1 written\n", 20},
	} {
		path := neverFieldPath(t, probe.name)
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != probe.node {
			t.Fatalf("%s: Node must print %q: exit %d, stdout %q, stderr %q", probe.name, probe.node, node.exitCode, node.stdout, node.stderr)
		}
		if !strings.HasPrefix(string(node.stdout), probe.stdout) {
			t.Fatalf("%s: the stdout before the stop must be Node's up to the read", probe.name)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		location := path + ":" + strconv.Itoa(probe.line) + ":20"
		want := run{stdout: []byte(probe.stdout), stderr: []byte("adamic: panic: unreachable read of arrow.name at " + location + ", a field declared never\n"), exitCode: 70}
		sanitized, _ := natively(t, program)
		release := released(t, program)
		for backend, got := range map[string]run{"native": sanitized, "release": release, "javascript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(want, got); difference != "" {
				t.Errorf("%s on %s: %s; exit %d, stdout %q, stderr %q", probe.name, backend, difference, got.exitCode, got.stdout, got.stderr)
			}
		}
		// The location names the program's file, whose name holds before for one reading, so what the slot
		// holds is looked for in the rest.
		stops[probe.name] = strings.Replace(string(release.stderr), location, "<location>", 1)
	}
	before, after := stops["readonly-view-read-before.ts"], stops["readonly-view-read-after.ts"]
	if before != after {
		t.Errorf("the two readings' stops differ past their location: %q and %q", before, after)
	}
	for name, stop := range stops {
		if strings.Contains(stop, "before") || strings.Contains(stop, "written") {
			t.Errorf("%s: the stop names what the slot holds: %q", name, stop)
		}
	}
}

// Language's m03 (#hatdcsp), a Named | Arrow union built by an as unknown as pair, keeps its stop at the
// cast with the pin's words, unchanged by ruling 2.
func TestNeverFieldUnionCastKeepsItsStop(t *testing.T) {
	t.Parallel()
	path := neverFieldPath(t, "m03.ts")
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.Where != path+":15:45" || !strings.HasPrefix(refused.What, "a cast the runtime can't check: an `as unknown as` pair from ") {
		t.Fatalf("want the as unknown as pair refused at m03.ts:15:45, got %v", err)
	}
}
