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

func TestJSONStringifyShorthand(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("testdata/json_shorthand.a")
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	got, binary := natively(test, program)
	for name, run := range map[string]run{
		"sanitized":  got,
		"release":    released(test, program),
		"JavaScript": onJavaScriptBackend(test, program),
	} {
		if difference := disagreement(truth, run); difference != "" {
			test.Errorf("%s: %s; Node %q, got %q; stderr %q", name, difference, truth.stdout, run.stdout, run.stderr)
		}
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
}

func TestJSONStringifyShorthandWrongBindingMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("testdata/json_shorthand.a")
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	outer := -1
	for index, local := range program.Locals {
		if local.Name == "status" && local.Function < 0 {
			outer = index
			break
		}
	}
	if outer < 0 {
		test.Fatal("outer binding missing")
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
		test.Fatal("shorthand mutation site missing")
	}
	truth := onNode(test, path)
	got, binary := natively(test, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		test.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "stdout differs" {
		test.Fatalf("JavaScript mutant caught by %q", difference)
	}
	test.Logf("Node %q; wrong binding %q", truth.stdout, got.stdout)
}
