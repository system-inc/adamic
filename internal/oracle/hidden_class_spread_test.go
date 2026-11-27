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
// generic_spread_copy_reads.go). A spread that isn't generic is judged only where the graph can't prove its
// source holds plain literals alone, so literal_with_class (a literal with its own weigh through the same
// copy, in a program where a Scale could be passed and isn't) lowers, as control_literal (no class in the
// program) and control_instance (a real Scale seen as Weighs, no spread) do, each agreeing with Node.
// Mutants: noteGenericSpread keys type parameters alone again, and interface_call, interface_truthy and
// interface_setter_write lower; plainSource answers true, and they lower too.
func TestHiddenClassSpreadMember(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, stop string }{
		{"interface_call.ts", "threw\n", "interface_call.ts:18:21: " + hiddenClassSpreadStop("a read", "weigh", "Scale", "method")},
		{"interface_truthy.ts", "true\n", hiddenClassSpreadStop("a read", "weigh", "Scale", "method")},
		{"interface_setter_write.ts", "{\"label\":\"t\",\"weight\":5}\n", hiddenClassSpreadStop("a write", "weight", "Sized", "getter")},
		{"generic_setter_write.ts", "{\"label\":\"t\",\"weight\":5}\n", "stage 0 can't lower a write of weight from { ...source }, which may be a generic spread's copy of a Sized instance"},
		{"literal_with_class.ts", "x 5\n", ""},
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
			if probe.stop != "" {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.name) || !strings.Contains(err.Error(), probe.stop) {
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

// hiddenClassSpreadStop is the words a use of name from the probes' copy, { ...source }, of a value whose
// type doesn't name its class, stops with.
func hiddenClassSpreadStop(use, name, class, kind string) string {
	return "stage 0 can't lower " + use + " of " + name + " from { ...source }, which may be a spread's copy, through a type that doesn't name its class, of a " + class + " instance, whose " + name + " is a " + kind + " on its prototype"
}
