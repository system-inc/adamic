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
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_read/trusted.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "0 4\n" {
		t.Fatalf("Node must print 0 4: %+v", want)
	}
	sanitized, _ := natively(t, program)
	for backend, got := range map[string]run{"native": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s; exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// The same switch handed a value its type lies about (never_read/lie.ts, through a computed key): the
// never read is reached, and it stops loudly at exit 70, naming itself at 13:14, natively and on the
// JavaScript backend alike, where Node goes on into fail and throws. The pin holds the exit, what ran
// before, and where the read is, not the trap's words.
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
	if node := onNode(t, path); node.exitCode == 0 || node.exitCode == 70 || string(node.stdout) != "0\n" {
		t.Fatalf("Node must print 0 and then throw from fail: %+v", node)
	}
	sanitized, _ := natively(t, program)
	for backend, got := range map[string]run{"native": sanitized, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "0\n" || !strings.HasPrefix(string(got.stderr), "adamic: panic: ") || !strings.Contains(string(got.stderr), "lie.ts:13:14") {
			t.Errorf("%s: want exit 70 after 0, a panic naming the read at lie.ts:13:14; exit %d, stdout %q, stderr %q", backend, got.exitCode, got.stdout, got.stderr)
		}
	}
}
