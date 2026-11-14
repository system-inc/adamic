package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// A never read trusted as the checker's proof (#p6vawef, never_read.go) lowers as a trap, and a valid run
// never reaches it: never_read/trusted.ts runs every member of the switch, and the sanitized build, the
// release build and the JavaScript backend each print what Node does.
func TestNeverReadTrustedAgreesWithNode(t *testing.T) {
	t.Parallel()
	neverReadAgreesWithNode(t, "trusted.ts", "0 4\n")
}

// Two of the premises the own-function trust leans on (never_read.go, Language on #p6vawef), held to Node:
// a class serving the union's member through a literal-typed getter answers only its literal, so the
// never read past both checks is trapped and never reached (never_read/getter.ts); and an open numeric
// enum parameter a switch exhausts, run with each member, agrees too, its default held by enum_never.go's
// trap (never_read/enum_in_range.ts).
func TestNeverReadLiteralGetterAgreesWithNode(t *testing.T) {
	t.Parallel()
	neverReadAgreesWithNode(t, "getter.ts", "circle 2 square\n")
}

func TestNeverReadOpenEnumInRangeAgreesWithNode(t *testing.T) {
	t.Parallel()
	neverReadAgreesWithNode(t, "enum_in_range.ts", "a b\n")
}

// neverReadAgreesWithNode holds testdata/never_read/name, which Node runs to stdout, to Node on the
// sanitized build, the release build and the JavaScript backend.
func neverReadAgreesWithNode(t *testing.T, name, stdout string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_read", name))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != stdout {
		t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", stdout, want.exitCode, want.stdout, want.stderr)
	}
	sanitized, _ := natively(t, program)
	for backend, got := range map[string]run{"native": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s; exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// The same switch handed a shape whose numeric enum tag lies (never_read/lie.ts): the never read is
// reached, and it stops loudly at exit 70, naming itself at 15:14, natively and on the JavaScript
// backend alike, before fail runs, where Node goes on into fail and throws. The pin holds the exit, what
// ran before, where the read is and that fail didn't run, not the trap's words.
func TestNeverReadTrustedTrapsOnALie(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_read/lie.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if node := onNode(t, path); node.exitCode == 0 || string(node.stdout) != "0\n" || !strings.Contains(string(node.stderr), "unexpected shape") {
		t.Fatalf("Node must print 0 and then throw from fail: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	sanitized, _ := natively(t, program)
	for backend, got := range map[string]run{"native": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "0\n" || !strings.HasPrefix(string(got.stderr), "adamic: panic: ") || !strings.Contains(string(got.stderr), "lie.ts:15:14") || strings.Contains(string(got.stderr), "unexpected shape") {
			t.Errorf("%s: want exit 70 after 0, a panic naming the read at lie.ts:15:14, before fail runs; exit %d, stdout %q, stderr %q", backend, got.exitCode, got.stdout, got.stderr)
		}
	}
}
