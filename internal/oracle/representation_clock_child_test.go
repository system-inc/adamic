package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"internal/oracle/testdata/representation_clock_child.a", true, false,
	})
}

func TestRepresentationClockChildMutant(t *testing.T) {
	t.Run("clock-child-ignore-argument", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/representation_clock_child.a"))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		want := onNode(t, path)
		if want.exitCode != 0 || string(want.stdout) != "1:leaf\n" || len(want.stderr) != 0 {
			t.Fatalf("unexpected Node result: %+v", want)
		}
		index := len(program.Strings)
		program.Strings = append(program.Strings, "mutant")
		changed := 0
		for i := range program.Functions {
			function := &program.Functions[i]
			if function.Name != "render" && !strings.HasPrefix(function.Name, "render_") {
				continue
			}
			function.Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: index}}}
			changed++
		}
		if changed != 1 {
			t.Fatalf("changed %d render specializations, want 1", changed)
		}
		got, sanitized := natively(t, program)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("mutant must finish cleanly: %+v", got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("want stdout disagreement, got %q", difference)
		}
		if report := leaks(t, program, sanitized); report != "" {
			t.Fatalf("mutant leaked: %s", report)
		}
		t.Logf("Node stdout %q; clean sanitized mutant stdout %q", want.stdout, got.stdout)
	})
}
