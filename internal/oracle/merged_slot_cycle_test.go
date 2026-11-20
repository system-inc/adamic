package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A merged member's slot is reserved Absent, and a class whose layout holds an Absent slot is built down
// emit_objects.go's constructionFields, whose object is never adopted into a graph region. So a merged slot
// that could point onward could close a cycle nothing frees (Language's c01 on #86cmqrp, counted: one
// allocation, no frees, and LeakSanitizer clean). A merged member that can hold a reference gets no slot and
// stops at compile time (mergedLeaf); a string member keeps its slot (TestMergedMemberSlot). Its controls: the
// same cycle through a field the class declares is in a graph region and frees whole, counted.
func TestMergedSlotCycleStops(t *testing.T) {
	t.Parallel()
	const stop = "accessing other from the interface merge of class M, whose slot could hold a reference to another object, and a merged slot can't hold one yet, since a cycle through it would never be freed; declare other on class M itself"
	for _, c := range []struct{ name, source string }{
		// c01: the cycle through the merged slot.
		{"cycle", "class M { a = 1 }\ninterface M { other?: M }\nconst m = new M();\nm.other = m;\nconsole.log(String(m.other === m));\n"},
		// c00: the same write to a second instance. The rule is the slot's, not the program's, so it stops too.
		{"second instance", "class M { a = 1 }\ninterface M { other?: M }\nconst m = new M();\nconst n = new M();\nm.other = n;\nconsole.log(String(m.other === n));\n"},
		// A read alone of the merged reference member stops the same way.
		{"read", "class M { a = 1 }\ninterface M { other?: M }\nconst m = new M();\nconsole.log(String(m.other === undefined));\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, stop) {
				t.Fatalf("got %v, want NotYet %q", err, stop)
			}
		})
	}
}

// c02: the declared-field twin of the cycle. It agrees with Node and frees whole on the counted build.
func TestMergedSlotCycleDeclaredFieldFrees(t *testing.T) {
	t.Parallel()
	const source = "class M { a = 1; other: M | undefined = undefined }\nconst m = new M();\nm.other = m;\nconsole.log(String(m.other === m));\n"
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	got, _ := natively(t, program)
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
	}
	counted := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), counted, native_runtime.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	ran := execute(t, counted)
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: ran.stdout, Stderr: ran.stderr, ExitCode: ran.exitCode}); report != "" {
		t.Fatal(report)
	}
}
