package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A for...in over an object literal with a computed key named the key with Node.Text, which panics on a computed
// name: "compiler bug: Unhandled case in Node.Text: *ast.ComputedPropertyName", exit 4 (#ttkhftg, found by Seats'
// review-i read 16). plainEnumerableObject now names a key as the literal's own lowering does (methodName, then
// numericFieldName), so a string-literal computed key is proven and enumerates as Node's does, and a key the
// literal can't name, numeric enum members inline among them, stops in for...in's own words. A literal typed as a
// record takes the record road, as it did on main. The programs are written to the test's own directory. Mutant
// (for_in_computed_key.mutants.md): the raw Node.Text call put back, which panics on the first two rows.
func TestForInComputedKey(t *testing.T) {
	t.Parallel()
	const stop = "for...in without a proven fixed plain-object origin"
	for _, probe := range []struct {
		name, source, node, stop string
	}{
		{"inline_enum.ts", "const enum P { Low, High }\nconst seen: string[] = [];\nfor (const k in { [P.Low]: 1, [P.High]: 2 }) {\n    seen.push(k);\n}\nconsole.log(seen.join(\",\"));\n", "0,1\n", stop},
		{"string_computed.ts", "const seen: string[] = [];\nfor (const k in { [\"a\"]: 1, b: 2 }) {\n    seen.push(k);\n}\nconsole.log(seen.join(\",\"));\n", "a,b\n", ""},
		// Node enumerates an integer-like key first, so a computed "1" after b reads 1,b; a wrong order here would be
		// silent (Library's should on #ttkhftg, J-xre8f1y).
		{"integer_like_computed.ts", "const seen: string[] = [];\nfor (const k in { b: 2, [\"1\"]: 1 }) {\n    seen.push(k);\n}\nconsole.log(seen.join(\",\"));\n", "1,b\n", ""},
		{"record_key.ts", "function key(n: number): string {\n    return \"k\" + String(n);\n}\nconst seen: string[] = [];\nfor (const k in { [key(1)]: 1 }) {\n    seen.push(k);\n}\nconsole.log(seen.join(\",\"));\n", "k1\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name)
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.node {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, err := lowered(t, path)
			if probe.stop != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.stop) {
					t.Fatalf("got %v; want a NotYet saying %q", err, probe.stop)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sanitized, _ := natively(t, program)
			for index, got := range []run{sanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s", difference, []string{"sanitized native", "the JavaScript backend"}[index])
				}
			}
		})
	}
}
