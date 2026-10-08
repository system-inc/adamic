package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"01-binary-complete", "02-variable-complete", "03-numeric-conditional"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/parser_factory_" + name + ".a", true, false})
	}
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/parser_factory_deferred.a", true, false}, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/parser_factory_deferred_read.a", true, true})
}

func TestParserFactoryDeferredRead(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/parser_factory_deferred_read.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("before\nundefined\n")}, observed); difference != "" {
		t.Fatal(difference)
	}
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: read before assignment: field 'bindDiagnostics' in node.bindDiagnostics\n"), exitCode: 70}
	baseline, _ := nativelyUncached(t, program)
	for _, got := range []run{baseline, onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %#v", difference, got)
		}
	}
	changed := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if field, ok := node.(ir.Property); ok && field.Name == "bindDiagnostics" && field.Readiness != "" {
			field.Readiness = ""
			changed++
			return field
		}
		return node
	})
	if changed == 0 {
		t.Fatal("mutant changed nothing")
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 0 || disagreement(want, mutant) == "" {
		t.Fatalf("missing read check mutant escaped or crashed: %#v", mutant)
	}
	t.Logf("read check mutant caught: stdout %q exit %d", mutant.stdout, mutant.exitCode)
}
