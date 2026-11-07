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

func TestRepresentationClockSourceMutant(test *testing.T) {
	test.Parallel()
	test.Run("clock-source-ignore-payload", func(test *testing.T) {
		path, absolutePathError := filepath.Abs(filepath.Join(repository, representationClockSourceFixture))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		program, loweredError := lowered(test, path)
		if loweredError != nil {
			test.Fatal(loweredError)
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
			test.Fatalf("want one describe specialization, changed %d", changed)
		}
		want := onNode(test, path)
		if want.exitCode != 0 || string(want.stdout) != "7:source\n" || len(want.stderr) != 0 {
			test.Fatalf("unexpected Node observation: %+v", want)
		}
		got, sanitized := natively(test, program)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			test.Fatalf("mutant must finish cleanly: %+v", got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			test.Fatalf("want stdout kill, got %q", difference)
		}
		if report := leaks(test, program, sanitized); report != "" {
			test.Fatalf("mutant leaked: %s", report)
		}
		if difference := disagreement(want, onJavaScriptBackend(test, program)); difference != "stdout differs" {
			test.Fatalf("JavaScript mutant: %q", difference)
		}
		test.Log("clock-source-ignore-payload killed by Node stdout; mutant exits zero without leaks")
	})
}
