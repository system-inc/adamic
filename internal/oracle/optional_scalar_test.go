package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// A ?. read whose result the checker calls a number is held as the maybe it is (#0d56xpd, Adamic's rule (d),
// Nov 26). Every program reads N.parent, declared present, through ?. on a real parent and on root's undefined!
// placeholder parent, the way tsc's accessKind and dumpTypes read a SourceFile's. Each use that observes the
// value agrees with Node on both, natively sanitized and in release, leak-free, and on the JavaScript backend:
// a switch over const parent = node.parent (u1_switch_plain), === and !== (u2), ?? 0 (u3), || 0 (u4), & as a
// truth test (u5, undefined converted as ToNumber and ToInt32 convert it, so it's 0), a field typed number |
// undefined (u6), a chain of two (u7), and accessKind's own shape, const { parent } = node (u1_switch), whose
// binding is a placeholder slot as the plain const is (#s59xxn3), where #jjf9ppe checked it at the binding and
// stopped it on the placeholder before the ?. was reached. A use that
// needs the number agrees with Node on the real parent, and on the placeholder stops at its check in Adamic's
// words, where Node carries undefined on: a non-null assertion into a number (s1), a return, an argument, a
// field store and an element of a number[]. A plain read through the placeholder still stops in the
// placeholder's words (s2). Mutants (testdata/optional_scalar/mutants.md): the read made plain, the check dropped, the ?.
// not counted as a presence test, every read through the placeholder counted as one, the bitwise conversion
// dropped, and a switch over a number pair the checker calls a number refused.
func TestOptionalScalarRead(t *testing.T) {
	t.Parallel()
	const absent = "a ?. read whose receiver the checker calls present was absent, through a non-null or definite-assignment assertion: node.parent?.kind"
	for _, probe := range []struct{ name, node, panic string }{
		{"u1_switch.ts", "one\nother\n", ""},
		{"u1_switch_plain.ts", "one\nother\n", ""},
		{"u2_strict_equal.ts", "true false\nfalse true\n", ""},
		{"u3_coalesce.ts", "1\n0\n", ""},
		{"u4_or.ts", "1\n0\n", ""},
		{"u5_bitwise.ts", "odd\neven or none\n", ""},
		{"u6_field.ts", "1\nundefined\n", ""},
		{"u7_chain.ts", "1\nundefined\nundefined\n", ""},
		{"s1_assertion.ts", "1\nundefined\n", absent},
		{"s1_return.ts", "1\nundefined\n", absent},
		{"s1_argument.ts", "1\nundefined\n", absent},
		{"s1_field.ts", "1\nundefined\n", absent},
		{"s1_element.ts", "1\nundefined\n", absent},
		{"s2_plain.ts", "1\n", "placeholder 'parent' is unset at property read via node.parent"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/optional_scalar", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if string(node.stdout) != probe.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q, want stdout %q", node.exitCode, node.stdout, node.stderr, probe.node)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			if probe.panic == "" {
				for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(node, got); difference != "" {
						t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
					}
				}
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
			if strings.HasPrefix(probe.name, "s2") && (node.exitCode == 0 || !strings.Contains(string(node.stderr), "TypeError")) {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q, want a TypeError", node.exitCode, node.stdout, node.stderr)
			}
			// The real parent's line, then the check's stop where Node printed undefined or threw.
			printed := probe.node[:strings.Index(probe.node, "\n")+1]
			for index, got := range []run{sanitized, releasedUncached(t, program)} {
				if got.exitCode != 70 || string(got.stdout) != printed || string(got.stderr) != "adamic: panic: "+probe.panic+"\n" {
					t.Fatalf("the check on %s: exit %d, stdout %q, stderr %q", []string{"sanitized native", "release native"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if difference := disagreement(onJavaScriptBackend(t, program), sanitized); difference != "" {
				t.Fatalf("the JavaScript backend against native: %s", difference)
			}
		})
	}
}
