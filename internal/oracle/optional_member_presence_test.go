package oracle

import (
	"path/filepath"
	"testing"
)

// A presence test of an optional member a program's interface declares (#p2n17c4 shape 1,
// lower/optional_member_presence.go), read off its slot as it runs: each program in
// testdata/optional_member_presence agrees with Node natively, sanitized and on the JavaScript backend,
// and leaks nothing, as .a and as .ts. method and field are the two forms, m?(): number and
// m?: () => number, at typeof, !, ?:, the left of && and a condition guarding the call, which keeps its
// receiver (a literal's method reads this.n), on an arrow, a literal's method shorthand and a literal
// without m, and through ?. on a receiver that's missing; written is the field form's slot written after
// its literal is built, a function in and undefined back, which only a read that runs can follow (a write
// to an optional method stays refused, so it has no method form); getter_subclass is the field form read
// through finishAccessors' dispatcher, on a getter's owner, a subclass that inherits it and two literals.
// Node's own output is pinned, so a
// fixture that stopped exercising its rows shows here before it shows as an agreement. These aren't in
// the oracle's fixtures list, so no counted row moves with them. Mutants:
// lower/optional_member_presence.mutants.md.
func TestOptionalMemberPresenceAgreesWithNode(t *testing.T) {
	t.Parallel()
	const tested = "function false has and\nfunction false has and\nundefined true none neither\n3\nnone\nnone\nhas\nnone\n"
	const written = "undefined none\nfunction has\nundefined none\nfunction has\nundefined none\n"
	for _, program := range []struct{ name, want string }{
		{"method.a", tested},
		{"method.ts", tested},
		{"field.a", tested},
		{"field.ts", tested},
		{"written.a", written},
		{"written.ts", written},
		{"getter_subclass.a", "has\nhas\nnone\nhas\n"},
		{"getter_subclass.ts", "has\nhas\nnone\nhas\n"},
	} {
		t.Run(program.name, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/optional_member_presence/" + program.name
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); string(node.stdout) != program.want || node.exitCode != 0 {
				t.Fatalf("Node: exit %d, stdout %q, want exit 0 and %q", node.exitCode, node.stdout, program.want)
			}
			// generatorAgrees holds any program to Node on every backend and to leaking nothing.
			generatorAgrees(t, relative)
		})
	}
}
