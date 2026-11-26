package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A spread needn't be generic to hide a class (#r8mazrh): copy(source: Weighs) { return { ...source }; }
// over new Scale() types its copy Weighs, weigh and all, where the copy holds own data only. Main printed
// 27 natively for interface_call where Node throws (native reuse kept the source object and its methods),
// read interface_truthy's !value.weigh false where Node reads true, and both setter writes, generic and not,
// stopped natively on a compiler bug where Node defines an own field. Each now stops at compile time, naming
// the member, the spread and the class (checkGenericSpreadCopyReads, internal/lower/
// generic_spread_copy_reads.go). literal_with_class is the floor's price, named: a literal with its own
// weigh through the same copy, in a program where Scale may be what copy is given, stops too, since the
// stop keys on the source's type. #fgqj0tm's real fix lifts it. control_literal (the same copy, no class in
// the program) and control_instance (a real Scale seen as Weighs, no spread) lower and agree with Node.
// Mutant: noteGenericSpread keys type parameters alone again, and interface_call, interface_truthy and
// interface_setter_write lower.
func TestHiddenClassSpreadMember(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, stop string }{
		{"interface_call.ts", "threw\n", "MEASURE"},
		{"interface_truthy.ts", "true\n", "MEASURE"},
		{"interface_setter_write.ts", "{\"label\":\"t\",\"weight\":5}\n", "MEASURE"},
		{"generic_setter_write.ts", "{\"label\":\"t\",\"weight\":5}\n", "MEASURE"},
		{"literal_with_class.ts", "x 5\n", "MEASURE"},
		{"control_literal.ts", "x 5\n", ""},
		{"control_instance.ts", "t 27\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_class_spread", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.node)}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			t.Logf("LOWERED %s: %v", probe.name, err)
			if probe.stop != "" {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.name) || (probe.stop != "MEASURE" && !strings.Contains(err.Error(), probe.stop)) {
					t.Fatalf("want a not yet with %q, got %v", probe.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
