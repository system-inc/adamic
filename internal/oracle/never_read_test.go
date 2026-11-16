package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
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

// The same switch handed a shape whose numeric enum tag lies (never_read/lie.ts): since #egmxf20, the lie
// never reaches the switch. `{ kind }` with kind a whole open Kind goes into a Shape only by tsc
// decomposing the union, and that's refused at the value written into the tag slot, at 20:30, before
// anything runs. Every other road a probe tried for a lying tag into the union stops too (a cast to the
// union or to a member, a generic wrapper, a spread, a conditional, a return, a parameter, an array's
// element: #egmxf20's comments), so the never read's trap is defense in depth with no lowerable lie in
// front of it. trusted.ts holds it never firing on valid input. A whole open enum switched on directly
// keeps its own trap, held by enums_open_test.go.
func TestNeverReadLieIsRefusedAtTheSlot(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_read/lie.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "enum member slot") || !strings.Contains(err.Error(), "lie.ts:20:30:") {
		t.Fatalf("want an enum member slot refused at lie.ts:20:30, got %v", err)
	}
}
