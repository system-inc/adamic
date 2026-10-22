package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/json_shorthand.a", true, false})
}

func TestJSONStringifyShorthand(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/json_shorthand.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	got, binary := natively(t, program)
	for name, run := range map[string]run{"sanitized": got, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, run); difference != "" {
			t.Errorf("%s: %s; Node %q, got %q; stderr %q", name, difference, truth.stdout, run.stdout, run.stderr)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}

func TestJSONStringifyShorthandWrongBindingMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/json_shorthand.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	outer := -1
	for index, local := range program.Locals {
		if local.Name == "status" && local.Function < 0 {
			outer = index
			break
		}
	}
	if outer < 0 {
		t.Fatal("outer binding missing")
	}
	changed := false
	for _, function := range program.Functions {
		if function.Name != "report" {
			continue
		}
		returned := function.Body[0].(ir.Return).Value.(ir.Coalesce).Value.(ir.JSONStringify)
		literal := returned.Value.(ir.ObjectLiteral)
		literal.Fields[0].Value = ir.Read{Local: outer, Of: ir.String}
		changed = true
	}
	if !changed {
		t.Fatal("shorthand mutation site missing")
	}
	truth := onNode(t, path)
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant caught by %q", difference)
	}
	t.Logf("Node %q; wrong binding %q", truth.stdout, got.stdout)
}
