package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

const representationClockSourceFixture = "internal/oracle/testdata/representation_clock_source.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{representationClockSourceFixture, true, false})
}

func TestRepresentationClockSourceMutant(t *testing.T) {
	t.Parallel()
	t.Run("clock-source-ignore-payload", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(repository, representationClockSourceFixture))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		constant := len(program.Strings)
		program.Strings = append(program.Strings, "7:mutant")
		changed := 0
		for index := range program.Functions {
			function := &program.Functions[index]
			if !strings.HasPrefix(function.Name, "describe") {
				continue
			}
			// Replace the whole specialization: retaining its old temporaries would leak.
			function.Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: constant}}}
			changed++
		}
		if changed != 1 {
			t.Fatalf("want one describe specialization, changed %d", changed)
		}
		want := onNode(t, path)
		if want.exitCode != 0 || string(want.stdout) != "7:source\n" || len(want.stderr) != 0 {
			t.Fatalf("unexpected Node observation: %+v", want)
		}
		got, sanitized := natively(t, program)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("mutant must finish cleanly: %+v", got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("want stdout kill, got %q", difference)
		}
		if report := leaks(t, program, sanitized); report != "" {
			t.Fatalf("mutant leaked: %s", report)
		}
		if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
			t.Fatalf("JavaScript mutant: %q", difference)
		}
		t.Log("clock-source-ignore-payload killed by Node stdout; mutant exits zero without leaks")
	})
}
