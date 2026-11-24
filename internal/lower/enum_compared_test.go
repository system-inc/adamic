package lower

import (
	"os"
	"strings"
	"testing"
)

// The program's own comparison of a const local proves an open numeric enum's value is one of the members
// a slot takes, where it's written (enumComparedMember, Adamic's ruling of Nov 19 on #snypn8c, c3). Each
// admitted row is modeled on a tsc site that writes token() into a slot of a few members, and lowers; the
// oracle holds each to Node (internal/oracle/enum_compared_test.go). Mutant: the call removed from
// freshOrWidened, which refuses every row again (internal/oracle/testdata/enum_compared/mutants.md).
func TestEnumComparedMemberLowers(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"case_chain", "or_and", "and_chain", "conditional"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/enum_compared/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lowerSourceAs(t, string(source), ".a"); err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
		})
	}
}

// What the comparison doesn't prove is refused in today's words, at the read written into the slot: an
// exclusion by !== (excluded), a let local (let), a case reached by falling through from a clause with
// statements (fallthrough), a read in an arrow made in the then-branch (arrow), and a default over the
// open enum (default). Mutants, each red here (mutants.md): !== proving, let taken, the function boundary
// walked past, and the fallthrough check dropped.
func TestEnumComparedMemberStopsAreRefused(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"excluded":    "main.a:12:15: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot K.A | K.B; ",
		"let":         "main.a:13:15: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot K.A | K.B; ",
		"fallthrough": "main.a:19:16: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot K.A | K.C; ",
		"arrow":       "main.a:13:35: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot K.A | K.B; ",
		"default":     "main.a:15:16: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot K.B | K.C | K.D; ",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/enum_compared/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			want += "compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"
			if _, err := lowerSourceAs(t, string(source), ".a"); err == nil || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("want %q, got %v", want, err)
			}
		})
	}
}
